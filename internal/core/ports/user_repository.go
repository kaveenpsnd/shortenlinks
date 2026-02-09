package ports

import (
	"context"

	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
)

type UserRepository interface {
	// Upsert: Create user if not exists, otherwise update (e.g. email)
	Upsert(ctx context.Context, user domain.User) error
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetAll(ctx context.Context) ([]domain.User, error)
	Delete(ctx context.Context, id string) error
}
