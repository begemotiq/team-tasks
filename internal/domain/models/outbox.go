package models

import (
	"encoding/json"
	"time"
)

type OutboxEventType string

const (
	OutboxEventTypeTeamInviteEmail OutboxEventType = "team_invite_email"
)

type OutboxEventStatus string

const (
	OutboxEventStatusPending    OutboxEventStatus = "pending"
	OutboxEventStatusProcessing OutboxEventStatus = "processing"
	OutboxEventStatusProcessed  OutboxEventStatus = "processed"
	OutboxEventStatusDeadLetter OutboxEventStatus = "dead_letter"
)

type OutboxEvent struct {
	ID          int64
	Type        OutboxEventType
	Payload     []byte
	Status      OutboxEventStatus
	Attempts    int
	ClaimToken  string
	AvailableAt time.Time
	ProcessedAt *time.Time
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TeamInviteEmailPayload struct {
	Email    string `json:"email"`
	TeamName string `json:"team_name"`
}

func NewTeamInviteEmailEvent(email, teamName string) (*OutboxEvent, error) {
	payload, err := json.Marshal(TeamInviteEmailPayload{
		Email:    email,
		TeamName: teamName,
	})
	if err != nil {
		return nil, err
	}
	return &OutboxEvent{
		Type:    OutboxEventTypeTeamInviteEmail,
		Payload: payload,
		Status:  OutboxEventStatusPending,
	}, nil
}

func (e OutboxEvent) TeamInviteEmailPayload() (TeamInviteEmailPayload, error) {
	var payload TeamInviteEmailPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return TeamInviteEmailPayload{}, err
	}
	return payload, nil
}
