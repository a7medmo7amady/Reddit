package service

import (
	"context"
	"errors"
	"log"
	"regexp"
	"sync"
	"time"

	"chat-service/internal/dto"
	"chat-service/internal/events"
	"chat-service/internal/models"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserClient interface {
	UserExists(ctx context.Context, userID string) (bool, error)
	IsBlocked(ctx context.Context, senderID, receiverID string) (bool, error)
}

type RealtimeDispatcher interface {
	PublishToUsers(ctx context.Context, userIDs []string, payload any)
}

type ChatService struct {
	conversations *mongo.Collection
	participants  *mongo.Collection
	messages      *mongo.Collection
	reports       *mongo.Collection
	users         UserClient
	producer      events.Producer
	rt            RealtimeDispatcher

	mongoCB *gobreaker.CircuitBreaker
	userCB  *gobreaker.CircuitBreaker

	bufferMu  sync.Mutex
	buffer    []models.Message
	maxBuffer int
}

func NewChatService(db *mongo.Database, users UserClient, producer events.Producer, rt RealtimeDispatcher) *ChatService {
	mongoCB := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "mongo",
		MaxRequests: 1,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})
	userCB := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "user-service",
		MaxRequests: 1,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})

	return &ChatService{
		conversations: db.Collection("conversations"),
		participants:  db.Collection("conversation_participants"),
		messages:      db.Collection("messages"),
		reports:       db.Collection("message_reports"),
		users:         users,
		producer:      producer,
		rt:            rt,
		mongoCB:       mongoCB,
		userCB:        userCB,
		maxBuffer:     1000,
	}
}

func (s *ChatService) Start(ctx context.Context) {
	// Flush buffered messages when Mongo recovers.
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.flushBuffer(ctx)
			}
		}
	}()
}

func (s *ChatService) CreateConversation(ctx context.Context, creatorID string, req dto.CreateConversationRequest) (*models.Conversation, error) {
	now := time.Now().UTC()

	allParticipants := append(req.ParticipantIDs, creatorID)

	for _, userID := range allParticipants {
		exists, err := s.userExists(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("participant does not exist")
		}
	}

	// Block check (fallback to allow if User Service is unavailable, per docs).
	for _, otherID := range req.ParticipantIDs {
		blocked := s.isBlockedDegradeToAllow(ctx, creatorID, otherID)
		if blocked {
			return nil, errors.New("cannot open dm (blocked)")
		}
	}

	conversation := models.Conversation{
		ID:        primitive.NewObjectID(),
		Type:      "direct",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.conversations.InsertOne(ctx, conversation)
	if err != nil {
		return nil, err
	}

	var participantDocs []any
	for _, userID := range allParticipants {
		participantDocs = append(participantDocs, models.ConversationParticipant{
			ID:             primitive.NewObjectID(),
			ConversationID: conversation.ID,
			UserID:         userID,
			JoinedAt:       now,
		})
	}

	_, err = s.participants.InsertMany(ctx, participantDocs)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (s *ChatService) GetOrCreateCommunityRoom(ctx context.Context, userID, communityID string) (*models.Conversation, error) {
	if communityID == "" {
		return nil, errors.New("missing community id")
	}

	now := time.Now().UTC()

	var conv models.Conversation
	err := s.conversations.FindOne(ctx, bson.M{
		"type":        "community",
		"communityId": communityID,
	}).Decode(&conv)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}

		conv = models.Conversation{
			ID:          primitive.NewObjectID(),
			Type:        "community",
			CommunityID: communityID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if _, err := s.conversations.InsertOne(ctx, conv); err != nil {
			return nil, err
		}
	}

	_, err = s.participants.UpdateOne(
		ctx,
		bson.M{"conversationId": conv.ID, "userId": userID},
		bson.M{
			"$setOnInsert": bson.M{
				"_id":            primitive.NewObjectID(),
				"conversationId": conv.ID,
				"userId":         userID,
				"joinedAt":       now,
				"lastReadAt":     now,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return nil, err
	}

	return &conv, nil
}

func (s *ChatService) SendMessage(ctx context.Context, senderID string, req dto.SendMessageRequest) (*models.Message, bool, error) {
	conversationID, err := primitive.ObjectIDFromHex(req.ConversationID)
	if err != nil {
		return nil, false, errors.New("invalid conversation id")
	}

	isParticipant, err := s.isParticipant(ctx, conversationID, senderID)
	if err != nil {
		return nil, false, err
	}
	if !isParticipant {
		return nil, false, errors.New("user is not a participant")
	}

	otherParticipants, err := s.getOtherParticipants(ctx, conversationID, senderID)
	if err != nil {
		return nil, false, err
	}

	for _, receiverID := range otherParticipants {
		blocked := s.isBlockedDegradeToAllow(ctx, senderID, receiverID)
		if blocked {
			return nil, false, errors.New("message blocked")
		}
	}

	msgType := "text"

	now := time.Now().UTC()

	message := models.Message{
		ID:             primitive.NewObjectID(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        req.Content,
		Type:           msgType,
		CreatedAt:      now,
	}

	queued, err := s.persistMessageWithFallback(ctx, message)
	if err != nil {
		return nil, false, err
	}

	_, _ = s.conversations.UpdateByID(ctx, conversationID, bson.M{"$set": bson.M{"updatedAt": now}})

	participantIDs, _ := s.getAllParticipants(ctx, conversationID)
	if s.rt != nil {
		s.rt.PublishToUsers(ctx, participantIDs, map[string]any{
			"type":           "chat.message",
			"conversationId": conversationID.Hex(),
			"message":        message,
			"queued":         queued,
		})
		// Mentions -> in-app notification.
		mentioned := extractMentions(req.Content)
		for _, m := range mentioned {
			s.rt.PublishToUsers(ctx, []string{m}, map[string]any{
				"type":           "notification.mention",
				"conversationId": conversationID.Hex(),
				"messageId":      message.ID.Hex(),
				"fromUserId":     senderID,
			})
		}
	}

	if !queued {
		_ = s.producer.Publish(ctx, "message.sent", events.MessageSentEvent{
			EventID:        primitive.NewObjectID().Hex(),
			EventType:      "message.sent",
			Version:        1,
			OccurredAt:     now,
			MessageID:      message.ID.Hex(),
			ConversationID: conversationID.Hex(),
			SenderID:       senderID,
		})
	}

	return &message, queued, nil
}

func (s *ChatService) persistMessage(ctx context.Context, message models.Message) error {
	// Mongo write retries (3 attempts exponential, 100ms base) + circuit breaker.
	_, execErr := s.mongoCB.Execute(func() (any, error) {
		bo := backoff.NewExponentialBackOff()
		bo.InitialInterval = 100 * time.Millisecond
		bo.MaxInterval = 800 * time.Millisecond
		bo.MaxElapsedTime = 0
		b := backoff.WithMaxRetries(bo, 3)
		op := func() error {
			writeCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()
			_, err := s.messages.InsertOne(writeCtx, message)
			return err
		}
		if err := backoff.Retry(op, backoff.WithContext(b, ctx)); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return execErr
}

func (s *ChatService) persistMessageWithFallback(ctx context.Context, message models.Message) (queued bool, err error) {
	if err := s.persistMessage(ctx, message); err == nil {
		return false, nil
	}

	// Mongo down (or breaker open): buffer in-memory, bounded.
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()
	if len(s.buffer) >= s.maxBuffer {
		return false, errors.New("message buffer full")
	}
	s.buffer = append(s.buffer, message)
	log.Printf("buffered message id=%s (mongo unavailable)", message.ID.Hex())
	return true, nil
}

func (s *ChatService) flushBuffer(ctx context.Context) {
	s.bufferMu.Lock()
	if len(s.buffer) == 0 {
		s.bufferMu.Unlock()
		return
	}
	// pop one FIFO to keep work bounded
	msg := s.buffer[0]
	s.buffer = s.buffer[1:]
	s.bufferMu.Unlock()

	if err := s.persistMessage(ctx, msg); err != nil {
		// push back to front if still failing
		s.bufferMu.Lock()
		s.buffer = append([]models.Message{msg}, s.buffer...)
		s.bufferMu.Unlock()
		return
	}

	// persisted after flush -> emit event
	_ = s.producer.Publish(ctx, "message.sent", events.MessageSentEvent{
		EventID:        primitive.NewObjectID().Hex(),
		EventType:      "message.sent",
		Version:        1,
		OccurredAt:     time.Now().UTC(),
		MessageID:      msg.ID.Hex(),
		ConversationID: msg.ConversationID.Hex(),
		SenderID:       msg.SenderID,
	})
}

func (s *ChatService) userExists(ctx context.Context, userID string) (bool, error) {
	op := func() (any, error) {
		callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return s.users.UserExists(callCtx, userID)
	}
	res, err := s.userCB.Execute(op)
	if err != nil {
		log.Printf("user-exists degraded (assume exists): %v", err)
		return true, nil
	}
	return res.(bool), nil
}

func (s *ChatService) isBlockedDegradeToAllow(ctx context.Context, senderID, receiverID string) bool {
	op := func() (any, error) {
		callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return s.users.IsBlocked(callCtx, senderID, receiverID)
	}
	res, err := s.userCB.Execute(op)
	if err != nil {
		log.Printf("block-check degraded (allow): %v", err)
		return false
	}
	blocked, _ := res.(bool)
	return blocked
}

var mentionRe = regexp.MustCompile(`(?m)(?:^|\s)@([A-Za-z0-9_\-]{1,64})`)

func extractMentions(content string) []string {
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	var out []string
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		id := m[1]
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *ChatService) getAllParticipants(ctx context.Context, conversationID primitive.ObjectID) ([]string, error) {
	cursor, err := s.participants.Find(ctx, bson.M{"conversationId": conversationID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var participants []models.ConversationParticipant
	if err := cursor.All(ctx, &participants); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(participants))
	for _, p := range participants {
		ids = append(ids, p.UserID)
	}
	return ids, nil
}

func (s *ChatService) GetConversationMessages(ctx context.Context, userID string, conversationIDHex string) ([]models.Message, error) {
	conversationID, err := primitive.ObjectIDFromHex(conversationIDHex)
	if err != nil {
		return nil, errors.New("invalid conversation id")
	}

	ok, err := s.isParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user is not a participant")
	}

	var conv models.Conversation
	_ = s.conversations.FindOne(ctx, bson.M{"_id": conversationID}).Decode(&conv)

	filter := bson.M{
		"conversationId": conversationID,
		"deletedAt":      bson.M{"$exists": false},
	}
	if conv.Type == "community" {
		filter["createdAt"] = bson.M{"$gte": time.Now().UTC().Add(-7 * 24 * time.Hour)}
	}

	cursor, err := s.messages.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *ChatService) GetRoomMessagesSince(ctx context.Context, userID, roomIDHex string, since *time.Time) ([]models.Message, error) {
	roomID, err := primitive.ObjectIDFromHex(roomIDHex)
	if err != nil {
		return nil, errors.New("invalid room id")
	}

	ok, err := s.isParticipant(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user is not a participant")
	}

	var conv models.Conversation
	if err := s.conversations.FindOne(ctx, bson.M{"_id": roomID}).Decode(&conv); err != nil {
		return nil, err
	}

	filter := bson.M{
		"conversationId": roomID,
		"deletedAt":      bson.M{"$exists": false},
	}

	if since != nil {
		filter["createdAt"] = bson.M{"$gt": since.UTC()}
	} else if conv.Type == "community" {
		filter["createdAt"] = bson.M{"$gte": time.Now().UTC().Add(-7 * 24 * time.Hour)}
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}})
	cursor, err := s.messages.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *ChatService) MarkRead(ctx context.Context, userID string, conversationIDHex string) error {
	conversationID, err := primitive.ObjectIDFromHex(conversationIDHex)
	if err != nil {
		return errors.New("invalid conversation id")
	}

	now := time.Now().UTC()

	res, err := s.participants.UpdateOne(ctx, bson.M{
		"conversationId": conversationID,
		"userId":         userID,
	}, bson.M{
		"$set": bson.M{"lastReadAt": now},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("participant not found")
	}

	return nil
}

func (s *ChatService) DeleteMessage(ctx context.Context, actorID, messageIDHex string, isModerator bool) error {
	messageID, err := primitive.ObjectIDFromHex(messageIDHex)
	if err != nil {
		return errors.New("invalid message id")
	}

	var msg models.Message
	if err := s.messages.FindOne(ctx, bson.M{"_id": messageID}).Decode(&msg); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("message not found")
		}
		return err
	}
	if msg.DeletedAt != nil {
		return nil
	}

	if msg.SenderID != actorID {
		if !isModerator {
			return errors.New("not allowed")
		}
		var conv models.Conversation
		if err := s.conversations.FindOne(ctx, bson.M{"_id": msg.ConversationID}).Decode(&conv); err != nil {
			return err
		}
		if conv.Type != "community" {
			return errors.New("not allowed")
		}
	}

	now := time.Now().UTC()
	deletedBy := actorID
	_, err = s.messages.UpdateByID(ctx, messageID, bson.M{
		"$set": bson.M{
			"deletedAt": now,
			"deletedBy": deletedBy,
		},
	})
	return err
}

func (s *ChatService) ReportMessage(ctx context.Context, reporterID, messageIDHex string, req dto.ReportMessageRequest) error {
	messageID, err := primitive.ObjectIDFromHex(messageIDHex)
	if err != nil {
		return errors.New("invalid message id")
	}

	var msg models.Message
	if err := s.messages.FindOne(ctx, bson.M{"_id": messageID}).Decode(&msg); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("message not found")
		}
		return err
	}

	ok, err := s.isParticipant(ctx, msg.ConversationID, reporterID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("user is not a participant")
	}

	now := time.Now().UTC()
	report := models.MessageReport{
		ID:         primitive.NewObjectID(),
		MessageID:  msg.ID,
		ReporterID: reporterID,
		Reason:     req.Reason,
		CreatedAt:  now,
	}

	if _, err := s.reports.InsertOne(ctx, report); err != nil {
		return err
	}

	_ = s.producer.Publish(ctx, "message.reported", map[string]any{
		"eventId":        primitive.NewObjectID().Hex(),
		"eventType":      "message.reported",
		"version":        1,
		"occurredAt":     now,
		"messageId":      msg.ID.Hex(),
		"reporterId":     reporterID,
		"conversationId": msg.ConversationID.Hex(),
	})

	return nil
}

func (s *ChatService) GetInbox(ctx context.Context, userID string) ([]dto.InboxItem, error) {
	// Aggregate on participants to find conversations, latest message and unread count.
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"userId": userID}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "conversations",
			"localField":   "conversationId",
			"foreignField": "_id",
			"as":           "conv",
		}}},
		{{Key: "$unwind", Value: "$conv"}},
		{{Key: "$lookup", Value: bson.M{
			"from": "messages",
			"let":  bson.M{"cid": "$conversationId"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{
					"deletedAt": bson.M{"$exists": false},
					"$expr":     bson.M{"$eq": bson.A{"$conversationId", "$$cid"}},
				}}},
				{{Key: "$sort", Value: bson.D{{Key: "createdAt", Value: -1}}}},
				{{Key: "$limit", Value: 1}},
				{{Key: "$project", Value: bson.M{
					"_id":       1,
					"senderId":  1,
					"content":   1,
					"createdAt": 1,
				}}},
			},
			"as": "lastMsg",
		}}},
		{{Key: "$addFields", Value: bson.M{
			"lastMsg": bson.M{"$arrayElemAt": bson.A{"$lastMsg", 0}},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "messages",
			"let":  bson.M{"cid": "$conversationId", "lastReadAt": "$lastReadAt"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{
					"deletedAt": bson.M{"$exists": false},
					"$expr": bson.M{"$and": bson.A{
						bson.M{"$eq": bson.A{"$conversationId", "$$cid"}},
						bson.M{"$ne": bson.A{"$senderId", userID}},
						bson.M{"$gt": bson.A{"$createdAt", bson.M{"$ifNull": bson.A{"$$lastReadAt", time.Time{}}}}},
					}},
				}}},
				{{Key: "$count", Value: "count"}},
			},
			"as": "unread",
		}}},
		{{Key: "$addFields", Value: bson.M{
			"unreadCount": bson.M{"$ifNull": bson.A{bson.M{"$arrayElemAt": bson.A{"$unread.count", 0}}, 0}},
			"updatedAt":   bson.M{"$ifNull": bson.A{"$lastMsg.createdAt", "$conv.updatedAt"}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "updatedAt", Value: -1}}}},
		{{Key: "$project", Value: bson.M{
			"conversationId": 1,
			"type":           "$conv.type",
			"communityId":    "$conv.communityId",
			"updatedAt":      1,
			"unreadCount":    1,
			"lastMsg":        1,
		}}},
	}

	cursor, err := s.participants.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var raw []struct {
		ConversationID primitive.ObjectID `bson:"conversationId"`
		Type           string             `bson:"type"`
		CommunityID    string             `bson:"communityId"`
		UpdatedAt      time.Time          `bson:"updatedAt"`
		UnreadCount    int                `bson:"unreadCount"`
		LastMsg        *struct {
			ID        primitive.ObjectID `bson:"_id"`
			SenderID  string             `bson:"senderId"`
			Content   string             `bson:"content"`
			CreatedAt time.Time          `bson:"createdAt"`
		} `bson:"lastMsg"`
	}

	if err := cursor.All(ctx, &raw); err != nil {
		return nil, err
	}

	items := make([]dto.InboxItem, 0, len(raw))
	for _, r := range raw {
		item := dto.InboxItem{
			ConversationID: r.ConversationID.Hex(),
			Type:           r.Type,
			CommunityID:    r.CommunityID,
			UnreadCount:    r.UnreadCount,
			UpdatedAt:      r.UpdatedAt,
		}
		if r.LastMsg != nil {
			item.LastMessage = &dto.InboxMessage{
				ID:        r.LastMsg.ID.Hex(),
				SenderID:  r.LastMsg.SenderID,
				Content:   r.LastMsg.Content,
				CreatedAt: r.LastMsg.CreatedAt,
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *ChatService) isParticipant(ctx context.Context, conversationID primitive.ObjectID, userID string) (bool, error) {
	count, err := s.participants.CountDocuments(ctx, bson.M{
		"conversationId": conversationID,
		"userId":         userID,
	})
	return count > 0, err
}

func (s *ChatService) getOtherParticipants(ctx context.Context, conversationID primitive.ObjectID, senderID string) ([]string, error) {
	cursor, err := s.participants.Find(ctx, bson.M{
		"conversationId": conversationID,
		"userId":         bson.M{"$ne": senderID},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var participants []models.ConversationParticipant
	if err := cursor.All(ctx, &participants); err != nil {
		return nil, err
	}

	var ids []string
	for _, p := range participants {
		ids = append(ids, p.UserID)
	}

	return ids, nil
}
