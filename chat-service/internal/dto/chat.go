package dto

type CreateConversationRequest struct {
	// DM is 1:1 (creator + 1 participant).
	ParticipantIDs []string `json:"participantIds" binding:"required,min=1,max=1"`
}

type SendMessageRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
	// Plain text only, max 2,000 chars.
	Content string `json:"content" binding:"required,max=2000"`
	Type    string `json:"type" binding:"omitempty,oneof=text"`
}

type TypingEvent struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversationId"`
}
