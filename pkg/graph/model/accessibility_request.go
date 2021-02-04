package model

import (
	"time"

	"github.com/google/uuid"
)

// AccessibilityRequest models a 508 request
type AccessibilityRequest struct {
	ID        uuid.UUID
	Name      string
	CreatedAt *time.Time `db:"created_at" gqlgen:"submittedAt"`
	UpdatedAt *time.Time `db:"updated_at"`
}
