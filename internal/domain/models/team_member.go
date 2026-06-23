package models

import "time"

type TeamMember struct {
	TeamID   int64
	UserID   int64
	Role     Role
	JoinedAt time.Time
}
