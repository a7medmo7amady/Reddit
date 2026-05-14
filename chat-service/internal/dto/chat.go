package dto

type CreateConversationRequest struct {
	// One participant creates/returns a DM; multiple participants create a group chat.
	ParticipantIDs []string `json:"participantIds" binding:"required,min=1"`
	Name           string   `json:"name" binding:"omitempty,max=80"`
}

type RenameConversationRequest struct {
	Name string `json:"name" binding:"required,max=80"`
}

type AddGroupParticipantRequest struct {
	ParticipantID string `json:"participantId" binding:"required"`
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
