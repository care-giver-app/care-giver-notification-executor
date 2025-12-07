package relationship

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRelationship(t *testing.T) {
	r := NewRelationship("User#123", "Receiver#123", true, false)

	assert.Equal(t, "User#123", r.UserID)
	assert.Equal(t, "Receiver#123", r.ReceiverID)
	assert.True(t, r.PrimaryCareGiver)
	assert.False(t, r.EmailNotifications)
}

func TestIsACareGiver(t *testing.T) {
	tests := map[string]struct {
		uid      string
		rid      string
		expected bool
	}{
		"Is a caregiver": {
			uid:      "User#123",
			rid:      "Receiver#123",
			expected: true,
		},
		"Is a caregiver - not primary": {
			uid:      "User#456",
			rid:      "Receiver#123",
			expected: true,
		},
		"Is not a caregiver": {
			uid:      "User#456",
			rid:      "Receiver#456",
			expected: false,
		},
	}

	relationships := []Relationship{
		{
			UserID:             "User#123",
			ReceiverID:         "Receiver#123",
			PrimaryCareGiver:   true,
			EmailNotifications: false,
		},
		{
			UserID:             "User#456",
			ReceiverID:         "Receiver#123",
			PrimaryCareGiver:   false,
			EmailNotifications: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			isCareGiver := IsACareGiver(tc.uid, tc.rid, relationships)
			assert.Equal(t, tc.expected, isCareGiver)
		})
	}
}

func TestIsAPrimaryCareGiver(t *testing.T) {
	tests := map[string]struct {
		uid      string
		rid      string
		expected bool
	}{
		"Is a primary caregiver": {
			uid:      "User#123",
			rid:      "Receiver#123",
			expected: true,
		},
		"Is a caregiver - not primary": {
			uid:      "User#456",
			rid:      "Receiver#123",
			expected: false,
		},
		"Is not a caregiver": {
			uid:      "User#456",
			rid:      "Receiver#456",
			expected: false,
		},
	}

	relationships := []Relationship{
		{
			UserID:             "User#123",
			ReceiverID:         "Receiver#123",
			PrimaryCareGiver:   true,
			EmailNotifications: false,
		},
		{
			UserID:             "User#456",
			ReceiverID:         "Receiver#123",
			PrimaryCareGiver:   false,
			EmailNotifications: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			isCareGiver := IsAPrimaryCareGiver(tc.uid, tc.rid, relationships)
			assert.Equal(t, tc.expected, isCareGiver)
		})
	}
}
