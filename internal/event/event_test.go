package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEntry(t *testing.T) {
	tests := map[string]struct {
		eventType     string
		timestamp     string
		data          []DataPoint
		note          string
		expectedEntry Entry
		expectErr     bool
	}{
		"Happy Path": {
			eventType: "Shower",
			expectedEntry: Entry{
				Type: "Shower",
			},
		},
		"Happy Path - With Timestamp": {
			eventType: "Medication",
			timestamp: "some time",
			expectedEntry: Entry{
				Type:      "Medication",
				Timestamp: "some time",
			},
		},
		"Happy Path - With Data": {
			eventType: "Weight",
			data: []DataPoint{
				{
					Name:  "Weight",
					Value: 120.3,
				},
			},
			expectedEntry: Entry{
				Type: "Weight",
				Data: []DataPoint{
					{
						Name:  "Weight",
						Value: 120.3,
					},
				},
			},
		},
		"Happy Path - With Note": {
			eventType: "Weight",
			note:      "some note",
			expectedEntry: Entry{
				Type: "Weight",
				Note: "some note",
			},
		},
		"Sad Path - Bad Event Type": {
			eventType: "BadEventType",
			expectErr: true,
		},
	}

	testRID := "Receiver#123"
	testUID := "User#123"

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.expectedEntry.ReceiverID = testRID
			tc.expectedEntry.UserID = testUID

			opts := []EntryOption{}
			if tc.timestamp != "" {
				opts = append(opts, WithTimestamp(tc.timestamp))
			}

			if len(tc.data) > 0 {
				opts = append(opts, WithData(tc.data))
				tc.expectedEntry.Data = tc.data
			}

			if tc.note != "" {
				opts = append(opts, WithNote(tc.note))
			}

			entry, err := NewEntry(testRID, testUID, tc.eventType, opts...)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, entry)
			} else {
				tc.expectedEntry.EventID = entry.EventID
				if tc.expectedEntry.Timestamp == "" {
					tc.expectedEntry.Timestamp = entry.Timestamp
				}
				assert.Equal(t, tc.expectedEntry, *entry)
			}
		})
	}

}
