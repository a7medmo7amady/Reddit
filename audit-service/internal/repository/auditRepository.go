package repository

import "fmt"

type AuditRepository struct{}

func NewAuditRepository() *AuditRepository {
	return &AuditRepository{}
}

func (r *AuditRepository) Save(log string) {
	fmt.Println("Saving log:", log)
}
