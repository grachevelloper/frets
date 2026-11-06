package auth

import "time"

type Profile struct {
	ID        string
	Nickname  string
	Email     string
	CreatedAt time.Time
}
