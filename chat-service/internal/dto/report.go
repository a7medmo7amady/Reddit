package dto

type ReportMessageRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=500"`
}
