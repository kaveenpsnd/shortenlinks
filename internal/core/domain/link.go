package domain

import "time"

type Link struct {
	ID          int64      `json:"id"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Clicks      int64      `json:"clicks"`
	UserID      *string    `json:"user_id,omitempty"`
}
