package event

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	DBPrefix = "Event"
	ParamID  = "eventId"

	bowelMovementType = "Bowel Movement"
	medicationType    = "Medication"
	showerType        = "Shower"
	urinationType     = "Urination"
	weightType        = "Weight"
)

type Entry struct {
	EventID    string      `json:"eventId" dynamodbav:"event_id"`
	ReceiverID string      `json:"receiverId" dynamodbav:"receiver_id"`
	UserID     string      `json:"userId" dynamodbav:"user_id"`
	Timestamp  string      `json:"timestamp" dynamodbav:"timestamp"`
	Type       string      `json:"type" dynamodbav:"type"`
	Data       []DataPoint `json:"data,omitempty" dynamodbav:"data"`
	Note       string      `json:"note,omitempty" dynamodbav:"note"`
}

type DataPoint struct {
	Name  string `json:"name" dynamodbav:"name"`
	Value any    `json:"value" dynamodbav:"value"`
}

type EntryOption func(*Entry)

func WithTimestamp(timestamp string) EntryOption {
	return func(e *Entry) {
		e.Timestamp = timestamp
	}
}

func WithData(data []DataPoint) EntryOption {
	return func(e *Entry) {
		e.Data = data
	}
}

func WithNote(note string) EntryOption {
	return func(e *Entry) {
		e.Note = note
	}
}

func NewEntry(receiverID, userID, eventType string, opts ...EntryOption) (*Entry, error) {
	supportedEventTypes := map[string]bool{
		bowelMovementType: true,
		medicationType:    true,
		showerType:        true,
		urinationType:     true,
		weightType:        true,
	}

	if !supportedEventTypes[eventType] {
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}

	e := &Entry{
		EventID:    fmt.Sprintf("%s#%s", DBPrefix, uuid.New().String()),
		ReceiverID: receiverID,
		UserID:     userID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Type:       eventType,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}
