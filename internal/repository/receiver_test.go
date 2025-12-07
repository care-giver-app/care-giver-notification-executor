package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/care-giver-app/care-giver-notification-executor/internal/receiver"
	"github.com/stretchr/testify/assert"
)

type MockReceiverDB struct{}

func (m *MockReceiverDB) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if av, found := params.Item["receiver_id"]; found {
		if id, ok := av.(*types.AttributeValueMemberS); ok {
			switch id.Value {
			case "Receiver#123":
				return nil, nil
			case "Error":
				return nil, errors.New("An error occured during Put Item")
			}
		}
	}
	return nil, errors.New("unsupported mock")
}

func (m *MockReceiverDB) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if av, found := params.Key["receiver_id"]; found {
		if id, ok := av.(*types.AttributeValueMemberS); ok {
			switch id.Value {
			case "Receiver#123":
				return &dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{
						"receiver_id": &types.AttributeValueMemberS{Value: id.Value},
						"first_name":  &types.AttributeValueMemberS{Value: "testFirstName"},
						"last_name":   &types.AttributeValueMemberS{Value: "testLastName"},
					},
				}, nil
			case "Get Item Error":
				return nil, errors.New("An error occured during Get Item")
			case "Unmarshal Error":
				return &dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{
						"receiver_id": &types.AttributeValueMemberS{Value: id.Value},
						"first_name":  &types.AttributeValueMemberS{Value: "testFirstName"},
						"last_name":   &types.AttributeValueMemberBOOL{Value: false},
					},
				}, nil
			}
		}
	}
	return nil, errors.New("unsupported mock")
}

func (m *MockReceiverDB) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	if av, found := params.Key["receiver_id"]; found {
		if id, ok := av.(*types.AttributeValueMemberS); ok {
			switch id.Value {
			case "Receiver#123":
				return &dynamodb.UpdateItemOutput{}, nil
			case "Update Item Error":
				return nil, errors.New("An error occured during Update Item")
			}
		}
	}
	return nil, nil
}

func (m *MockReceiverDB) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return nil, errors.New("unsupported mock")
}

func (m *MockReceiverDB) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return nil, errors.New("unsupported mock")
}

func TestCreateReceiver(t *testing.T) {
	appCfg := appconfig.NewAppConfig()
	testReceiverRepo := NewReceiverRespository(context.Background(), appCfg, &MockReceiverDB{})

	tests := map[string]struct {
		receiver    receiver.Receiver
		expectError bool
	}{
		"Happy Path - Receiver Created": {
			receiver: receiver.Receiver{
				ReceiverID: "Receiver#123",
				FirstName:  "testName",
				LastName:   "testLastName",
			},
		},
		"Sad Path - Error Putting Item": {
			receiver: receiver.Receiver{
				ReceiverID: "Error",
			},
			expectError: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := testReceiverRepo.CreateReceiver(tc.receiver)

			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetReceiver(t *testing.T) {
	appCfg := appconfig.NewAppConfig()
	testReceiverRepo := NewReceiverRespository(context.Background(), appCfg, &MockReceiverDB{})

	tests := map[string]struct {
		receiverID       string
		expectedReceiver receiver.Receiver
		expectError      bool
	}{
		"Happy Path - Got Receiver": {
			receiverID: "Receiver#123",
			expectedReceiver: receiver.Receiver{
				ReceiverID: "Receiver#123",
				FirstName:  "testFirstName",
				LastName:   "testLastName",
			},
		},
		"Sad Path - Error Getting Item": {
			receiverID:  "Get Item Error",
			expectError: true,
		},
		"Sad Path - Error Unmarshalling Item": {
			receiverID:  "Unmarshal Error",
			expectError: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			receiver, err := testReceiverRepo.GetReceiver(tc.receiverID)

			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedReceiver, receiver)
			}
		})
	}
}
