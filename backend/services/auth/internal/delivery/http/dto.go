package httpapi

import (
	"time"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type profileDTO struct {
	ID        string    `json:"id"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func profileResponse(p auth.Profile) profileDTO {
	return profileDTO{
		ID:        p.ID,
		Nickname:  p.Nickname,
		Email:     p.Email,
		CreatedAt: p.CreatedAt,
	}
}
