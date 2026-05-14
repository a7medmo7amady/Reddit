package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"chat-service/internal/events"
	"chat-service/internal/http/middleware"
	"chat-service/internal/models"
	"chat-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// stubUserClient satisfies service.UserClient for integration tests
// without requiring the real user-service to be running.
// ---------------------------------------------------------------------------

type stubUserClient struct{}

func (s *stubUserClient) UserExists(_ context.Context, _ string) (bool, error) { return true, nil }
func (s *stubUserClient) IsBlocked(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

// ---------------------------------------------------------------------------
// Package-level test infrastructure
// ---------------------------------------------------------------------------

var (
	testRouter       *gin.Engine
	testDB           *mongo.Database
	testService      *service.ChatService
	testKafkaBrokers string
)

func TestMain(m *testing.M) {
	mongoURI := os.Getenv("TEST_MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Fprintf(os.Stderr, "SKIP integration tests – cannot connect to MongoDB: %v\n", err)
		os.Exit(0)
	}
	if err := client.Ping(ctx, nil); err != nil {
		fmt.Fprintf(os.Stderr, "SKIP integration tests – cannot ping MongoDB: %v\n", err)
		os.Exit(0)
	}

	testDB = client.Database("chat_service_integration_test")

	// Kafka – use real broker if available, else fall back to LogProducer.
	testKafkaBrokers = os.Getenv("TEST_KAFKA_BROKERS")
	var producer events.Producer
	if testKafkaBrokers != "" {
		kp := events.NewKafkaProducer(strings.Split(testKafkaBrokers, ","))
		defer kp.Close()
		producer = kp
		log.Printf("[test] kafka producer → %s", testKafkaBrokers)
	} else {
		producer = events.NewLogProducer()
		log.Println("[test] kafka not configured, using log producer")
	}

	testService = service.NewChatService(testDB, &stubUserClient{}, producer, nil)

	h := NewChatHandler(testService)

	gin.SetMode(gin.TestMode)
	testRouter = gin.New()
	api := testRouter.Group("/api")
	api.Use(middleware.Auth())

	api.GET("/inbox", h.GetInbox)
	api.POST("/conversations", h.CreateConversation)
	api.POST("/messages", h.SendMessage)
	api.DELETE("/messages/:messageId", h.DeleteMessage)
	api.POST("/messages/:messageId/report", h.ReportMessage)
	api.GET("/communities/:communityId/room", h.GetOrCreateCommunityRoom)
	api.GET("/conversations/:conversationId/messages", h.GetMessages)
	api.POST("/conversations/:conversationId/read", h.MarkRead)
	api.GET("/chat/:room/messages", h.GetRoomMessagesSince)

	code := m.Run()

	// Cleanup
	_ = testDB.Drop(context.Background())
	_ = client.Disconnect(context.Background())
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func doRequest(method, path string, body any, userID string, headers ...string) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-User-Id", userID)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d, want %d; body = %s", w.Code, want, w.Body.String())
	}
}

// createConversation is a test helper that creates a DM and returns the ID.
func createConversation(t *testing.T, creator string, participant string) string {
	t.Helper()
	w := doRequest(http.MethodPost, "/api/conversations",
		map[string]any{"participantIds": []string{participant}}, creator)
	if w.Code != http.StatusCreated {
		t.Fatalf("createConversation failed: %d %s", w.Code, w.Body.String())
	}
	var conv models.Conversation
	json.Unmarshal(w.Body.Bytes(), &conv)
	return conv.ID.Hex()
}

// ---------------------------------------------------------------------------
// Tests: Auth middleware (no infra needed)
// ---------------------------------------------------------------------------

func TestMissingAuth(t *testing.T) {
	w := doRequest(http.MethodGet, "/api/inbox", nil, "")
	assertStatus(t, w, http.StatusUnauthorized)
}

// ---------------------------------------------------------------------------
// Tests: CreateConversation (real MongoDB)
// ---------------------------------------------------------------------------

func TestCreateConversation_Success(t *testing.T) {
	w := doRequest(http.MethodPost, "/api/conversations",
		map[string]any{"participantIds": []string{"user2"}}, "user1")
	assertStatus(t, w, http.StatusCreated)

	var conv models.Conversation
	if err := json.Unmarshal(w.Body.Bytes(), &conv); err != nil {
		t.Fatal(err)
	}
	if conv.Type != "direct" {
		t.Errorf("type = %q, want %q", conv.Type, "direct")
	}
}

func TestCreateConversation_BadBody(t *testing.T) {
	w := doRequest(http.MethodPost, "/api/conversations", map[string]any{}, "user1")
	assertStatus(t, w, http.StatusBadRequest)
}

// ---------------------------------------------------------------------------
// Tests: SendMessage (real MongoDB + Kafka)
// ---------------------------------------------------------------------------

func TestSendMessage_Success(t *testing.T) {
	convID := createConversation(t, "sender1", "receiver1")
	w := doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "hello integration"}, "sender1")
	assertStatus(t, w, http.StatusCreated)

	var msg models.Message
	json.Unmarshal(w.Body.Bytes(), &msg)
	if msg.Content != "hello integration" {
		t.Errorf("content = %q", msg.Content)
	}
}

func TestSendMessage_NotParticipant(t *testing.T) {
	convID := createConversation(t, "s2", "r2")
	w := doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "nope"}, "outsider")
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSendMessage_InvalidConversationID(t *testing.T) {
	w := doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": "not-hex", "content": "x"}, "user1")
	assertStatus(t, w, http.StatusBadRequest)
}

// ---------------------------------------------------------------------------
// Tests: GetMessages (real MongoDB)
// ---------------------------------------------------------------------------

func TestGetMessages_Success(t *testing.T) {
	convID := createConversation(t, "gm1", "gm2")
	doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "msg1"}, "gm1")
	doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "msg2"}, "gm1")

	w := doRequest(http.MethodGet, "/api/conversations/"+convID+"/messages", nil, "gm1")
	assertStatus(t, w, http.StatusOK)

	var msgs []models.Message
	json.Unmarshal(w.Body.Bytes(), &msgs)
	if len(msgs) < 2 {
		t.Errorf("messages count = %d, want >= 2", len(msgs))
	}
}

func TestGetMessages_NotParticipant(t *testing.T) {
	convID := createConversation(t, "gm3", "gm4")
	w := doRequest(http.MethodGet, "/api/conversations/"+convID+"/messages", nil, "nobody")
	assertStatus(t, w, http.StatusBadRequest)
}

// ---------------------------------------------------------------------------
// Tests: MarkRead (real MongoDB)
// ---------------------------------------------------------------------------

func TestMarkRead_Success(t *testing.T) {
	convID := createConversation(t, "mr1", "mr2")
	w := doRequest(http.MethodPost, "/api/conversations/"+convID+"/read", nil, "mr1")
	assertStatus(t, w, http.StatusNoContent)
}

func TestMarkRead_NotParticipant(t *testing.T) {
	convID := createConversation(t, "mr3", "mr4")
	w := doRequest(http.MethodPost, "/api/conversations/"+convID+"/read", nil, "unknown")
	assertStatus(t, w, http.StatusBadRequest)
}

// ---------------------------------------------------------------------------
// Tests: GetInbox (real MongoDB)
// ---------------------------------------------------------------------------

func TestGetInbox_Success(t *testing.T) {
	convID := createConversation(t, "inbox1", "inbox2")
	doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "hi inbox"}, "inbox1")

	w := doRequest(http.MethodGet, "/api/inbox", nil, "inbox1")
	assertStatus(t, w, http.StatusOK)

	var items []map[string]any
	json.Unmarshal(w.Body.Bytes(), &items)
	if len(items) == 0 {
		t.Error("inbox empty, want ≥ 1 item")
	}
}

// ---------------------------------------------------------------------------
// Tests: DeleteMessage (real MongoDB)
// ---------------------------------------------------------------------------

func TestDeleteMessage_BySender(t *testing.T) {
	convID := createConversation(t, "del1", "del2")
	w := doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "to delete"}, "del1")
	var msg models.Message
	json.Unmarshal(w.Body.Bytes(), &msg)

	w = doRequest(http.MethodDelete, "/api/messages/"+msg.ID.Hex(), nil, "del1")
	assertStatus(t, w, http.StatusNoContent)
}

func TestDeleteMessage_NotOwner(t *testing.T) {
	convID := createConversation(t, "del3", "del4")
	w := doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "owned by del3"}, "del3")
	var msg models.Message
	json.Unmarshal(w.Body.Bytes(), &msg)

	w = doRequest(http.MethodDelete, "/api/messages/"+msg.ID.Hex(), nil, "del4")
	assertStatus(t, w, http.StatusBadRequest)
}

// ---------------------------------------------------------------------------
// Tests: ReportMessage (real MongoDB + Kafka)
// ---------------------------------------------------------------------------

func TestReportMessage_Success(t *testing.T) {
	convID := createConversation(t, "rpt1", "rpt2")
	w := doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "bad stuff"}, "rpt1")
	var msg models.Message
	json.Unmarshal(w.Body.Bytes(), &msg)

	w = doRequest(http.MethodPost, "/api/messages/"+msg.ID.Hex()+"/report",
		map[string]any{"reason": "spam"}, "rpt2")
	assertStatus(t, w, http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Tests: Community room (real MongoDB)
// ---------------------------------------------------------------------------

func TestCommunityRoom_CreateAndGet(t *testing.T) {
	w := doRequest(http.MethodGet, "/api/communities/test-community/room", nil, "cu1")
	assertStatus(t, w, http.StatusOK)

	var conv models.Conversation
	json.Unmarshal(w.Body.Bytes(), &conv)
	if conv.Type != "community" {
		t.Errorf("type = %q, want community", conv.Type)
	}

	// Second call returns the same room.
	w2 := doRequest(http.MethodGet, "/api/communities/test-community/room", nil, "cu2")
	assertStatus(t, w2, http.StatusOK)
	var conv2 models.Conversation
	json.Unmarshal(w2.Body.Bytes(), &conv2)
	if conv.ID != conv2.ID {
		t.Errorf("expected same room ID, got %s vs %s", conv.ID.Hex(), conv2.ID.Hex())
	}
}

// ---------------------------------------------------------------------------
// Tests: GetRoomMessagesSince (real MongoDB)
// ---------------------------------------------------------------------------

func TestGetRoomMessagesSince(t *testing.T) {
	// Use a community room for this test.
	w := doRequest(http.MethodGet, "/api/communities/since-comm/room", nil, "since1")
	var conv models.Conversation
	json.Unmarshal(w.Body.Bytes(), &conv)

	doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": conv.ID.Hex(), "content": "old msg"}, "since1")

	mark := time.Now().UTC()
	time.Sleep(10 * time.Millisecond)

	doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": conv.ID.Hex(), "content": "new msg"}, "since1")

	w = doRequest(http.MethodGet,
		"/api/chat/"+conv.ID.Hex()+"/messages?since="+mark.Format(time.RFC3339Nano),
		nil, "since1")
	assertStatus(t, w, http.StatusOK)

	var msgs []models.Message
	json.Unmarshal(w.Body.Bytes(), &msgs)
	if len(msgs) != 1 {
		t.Errorf("messages = %d, want 1", len(msgs))
	}
}

// ---------------------------------------------------------------------------
// Tests: parseSince (unit – no infra needed)
// ---------------------------------------------------------------------------

func TestParseSince(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"RFC3339", "2026-01-01T00:00:00Z", false},
		{"RFC3339Nano", "2026-01-01T00:00:00.123456789Z", false},
		{"UnixSeconds", "1735689600", false},
		{"UnixMillis", "1735689600000", false},
		{"Invalid", "not-a-timestamp", true},
		{"Empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSince(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSince(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: Kafka end-to-end (skipped if TEST_KAFKA_BROKERS not set)
// ---------------------------------------------------------------------------

func TestSendMessage_KafkaEvent(t *testing.T) {
	if testKafkaBrokers == "" {
		t.Skip("TEST_KAFKA_BROKERS not set – skipping Kafka integration test")
	}

	topic := "message.sent"

	// Create a reader for the topic BEFORE publishing.
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(testKafkaBrokers, ","),
		Topic:    topic,
		GroupID:  "test-" + primitive.NewObjectID().Hex(),
		MaxWait:  3 * time.Second,
		MinBytes: 1,
		MaxBytes: 1e6,
	})
	defer reader.Close()

	// Send a message via the API.
	convID := createConversation(t, "kafka1", "kafka2")
	w := doRequest(http.MethodPost, "/api/messages",
		map[string]any{"conversationId": convID, "content": "kafka test"}, "kafka1")
	assertStatus(t, w, http.StatusCreated)

	// Read back from Kafka.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		t.Fatalf("kafka read: %v", err)
	}

	var event events.MessageSentEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if event.EventType != "message.sent" {
		t.Errorf("eventType = %q, want message.sent", event.EventType)
	}
	if event.ConversationID != convID {
		t.Errorf("conversationId = %q, want %q", event.ConversationID, convID)
	}
	if event.SenderID != "kafka1" {
		t.Errorf("senderId = %q, want kafka1", event.SenderID)
	}
}
