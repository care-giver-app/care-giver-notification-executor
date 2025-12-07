package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/care-giver-app/care-giver-notification-executor/internal/event"
	"github.com/stretchr/testify/assert"
)

type MockEventDB struct{}

func (m *MockEventDB) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if av, found := params.Item["event_id"]; found {
		if id, ok := av.(*types.AttributeValueMemberS); ok {
			switch id.Value {
			case "Event#123":
				return nil, nil
			case "Error":
				return nil, errors.New("An error occured during Put Item")
			}
		}
	}
	return nil, errors.New("unsupported mock")
}

func (m *MockEventDB) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return nil, errors.New("unsupported mock")
}

func (m *MockEventDB) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return nil, errors.New("unsupported mock")
}

func (m *MockEventDB) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if av, found := params.ExpressionAttributeValues[":rid"]; found {
		if id, ok := av.(*types.AttributeValueMemberS); ok {
			switch id.Value {
			case "Receiver#123":
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"type": &types.AttributeValueMemberS{Value: "Shower"},
						},
						{
							"type": &types.AttributeValueMemberS{Value: "Medication"},
						},
					},
				}, nil
			case "BadData":
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"type": &types.AttributeValueMemberBOOL{Value: false},
						},
					},
				}, nil
			case "Error":
				return nil, errors.New("error querying db")
			}
		}
	}
	return nil, errors.New("unsupported mock")
}

func (m *MockEventDB) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	if av, found := params.Key["event_id"]; found {
		if id, ok := av.(*types.AttributeValueMemberS); ok {
			switch id.Value {
			case "Event#123":
				return nil, nil
			case "Error":
				return nil, errors.New("error deleting item")
			}
		}
	}
	return nil, errors.New("unsupported mock")
}

func TestAddEvent(t *testing.T) {
	appCfg := appconfig.NewAppConfig()
	testEventRepo := NewEventRespository(context.Background(), appCfg, &MockEventDB{})

	tests := map[string]struct {
		entry       *event.Entry
		expectError bool
	}{
		"Happy Path - Event Added": {
			entry: &event.Entry{
				EventID: "Event#123",
			},
		},
		"Sad Path - Put Item Error": {
			entry: &event.Entry{
				EventID: "Error",
			},
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := testEventRepo.AddEvent(tc.entry)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEvents(t *testing.T) {
	appCfg := appconfig.NewAppConfig()
	testEventRepo := NewEventRespository(context.Background(), appCfg, &MockEventDB{})

	tests := map[string]struct {
		rid           string
		expectedValue []event.Entry
		expectError   bool
	}{
		"Happy Path - Got Events": {
			rid: "Receiver#123",
			expectedValue: []event.Entry{
				{
					Type: "Shower",
				},
				{
					Type: "Medication",
				},
			},
		},
		"Sad Path - Query Error": {
			rid:         "Error",
			expectError: true,
		},
		"Sad Path - Bad Data Error": {
			rid:         "BadData",
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			events, err := testEventRepo.GetEvents(tc.rid)
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, events)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValue, events)
			}
		})
	}
}

func TestDeleteEvent(t *testing.T) {
	appCfg := appconfig.NewAppConfig()
	testEventRepo := NewEventRespository(context.Background(), appCfg, &MockEventDB{})

	tests := map[string]struct {
		rid         string
		eid         string
		expectError bool
	}{
		"Happy Path - Event Deleted": {
			rid: "Receiver#123",
			eid: "Event#123",
		},
		"Sad Path - Delete Error": {
			rid:         "Receiver#123",
			eid:         "Error",
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := testEventRepo.DeleteEvent(tc.rid, tc.eid)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
