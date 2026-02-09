package domain

import "time"

type User struct {
	ID        string    `json:"id"` // Matches Firebase UID
	Email     string    `json:"email"`
	Role      string    `json:"role"` // 'admin' or 'user'
	CreatedAt time.Time `json:"created_at"`
}
