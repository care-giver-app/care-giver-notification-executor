package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/care-giver-app/care-giver-notification-executor/internal/relationship"
	"github.com/stretchr/testify/assert"
)

type MockRelationshipDB struct {
	pages       []*dynamodb.QueryOutput
	queryError  error
	currentPage int
}

func (m *MockRelationshipDB) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}

	if params.IndexName != nil && *params.IndexName == "email_notifications" {
		if av, found := params.ExpressionAttributeValues[":email_notifications_gsi_pk"]; found {
			if emailNotif, ok := av.(*types.AttributeValueMemberN); ok && emailNotif.Value == "1" {
				if m.currentPage < len(m.pages) {
					result := m.pages[m.currentPage]
					m.currentPage++
					return result, nil
				}
			}
		}
	}

	return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
}

func TestGetRelationshipsByEmailNotifications(t *testing.T) {
	tests := map[string]struct {
		mockSetup             func() *MockRelationshipDB
		expectedRelationships []relationship.Relationship
		expectedCount         int
		expectError           bool
	}{
		"Happy Path - Single Page": {
			mockSetup: func() *MockRelationshipDB {
				return &MockRelationshipDB{
					pages: []*dynamodb.QueryOutput{
						{
							Items: []map[string]types.AttributeValue{
								{
									"user_id":                    &types.AttributeValueMemberS{Value: "User#123"},
									"receiver_id":                &types.AttributeValueMemberS{Value: "Receiver#123"},
									"primary_care_giver":         &types.AttributeValueMemberBOOL{Value: true},
									"email_notifications":        &types.AttributeValueMemberBOOL{Value: true},
									"email_notifications_gsi_pk": &types.AttributeValueMemberN{Value: "1"},
								},
								{
									"user_id":                    &types.AttributeValueMemberS{Value: "User#456"},
									"receiver_id":                &types.AttributeValueMemberS{Value: "Receiver#456"},
									"primary_care_giver":         &types.AttributeValueMemberBOOL{Value: false},
									"email_notifications":        &types.AttributeValueMemberBOOL{Value: true},
									"email_notifications_gsi_pk": &types.AttributeValueMemberN{Value: "1"},
								},
							},
							Count: 2,
						},
					},
				}
			},
			expectedRelationships: []relationship.Relationship{
				{
					UserID:             "User#123",
					ReceiverID:         "Receiver#123",
					PrimaryCareGiver:   true,
					EmailNotifications: true,
				},
				{
					UserID:             "User#456",
					ReceiverID:         "Receiver#456",
					PrimaryCareGiver:   false,
					EmailNotifications: true,
				},
			},
			expectedCount: 2,
		},
		"Happy Path - Multiple Pages": {
			mockSetup: func() *MockRelationshipDB {
				return &MockRelationshipDB{
					pages: []*dynamodb.QueryOutput{
						{
							Items: []map[string]types.AttributeValue{
								{
									"user_id":                    &types.AttributeValueMemberS{Value: "User#1"},
									"receiver_id":                &types.AttributeValueMemberS{Value: "Receiver#1"},
									"primary_care_giver":         &types.AttributeValueMemberBOOL{Value: true},
									"email_notifications":        &types.AttributeValueMemberBOOL{Value: true},
									"email_notifications_gsi_pk": &types.AttributeValueMemberN{Value: "1"},
								},
							},
							LastEvaluatedKey: map[string]types.AttributeValue{
								"user_id": &types.AttributeValueMemberS{Value: "User#1"},
							},
						},
						{
							Items: []map[string]types.AttributeValue{
								{
									"user_id":                    &types.AttributeValueMemberS{Value: "User#2"},
									"receiver_id":                &types.AttributeValueMemberS{Value: "Receiver#2"},
									"primary_care_giver":         &types.AttributeValueMemberBOOL{Value: false},
									"email_notifications":        &types.AttributeValueMemberBOOL{Value: true},
									"email_notifications_gsi_pk": &types.AttributeValueMemberN{Value: "1"},
								},
							},
						},
					},
				}
			},
			expectedRelationships: []relationship.Relationship{
				{
					UserID:             "User#1",
					ReceiverID:         "Receiver#1",
					PrimaryCareGiver:   true,
					EmailNotifications: true,
				},
				{
					UserID:             "User#2",
					ReceiverID:         "Receiver#2",
					PrimaryCareGiver:   false,
					EmailNotifications: true,
				},
			},
			expectedCount: 2,
		},
		"Happy Path - No Results": {
			mockSetup: func() *MockRelationshipDB {
				return &MockRelationshipDB{
					pages: []*dynamodb.QueryOutput{
						{
							Items: []map[string]types.AttributeValue{},
							Count: 0,
						},
					},
				}
			},
			expectedRelationships: nil,
			expectedCount:         0,
		},
		"Error Path - Query Error": {
			mockSetup: func() *MockRelationshipDB {
				return &MockRelationshipDB{
					queryError: errors.New("DynamoDB query failed"),
				}
			},
			expectError: true,
		},
		"Error Path - Unmarshal Error": {
			mockSetup: func() *MockRelationshipDB {
				return &MockRelationshipDB{
					pages: []*dynamodb.QueryOutput{
						{
							Items: []map[string]types.AttributeValue{
								{
									"user_id": &types.AttributeValueMemberBOOL{Value: false},
								},
							},
						},
					},
				}
			},
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockDB := tc.mockSetup()
			appCfg := appconfig.NewAppConfig()

			testRepo := NewRelationshipRepository(context.Background(), appCfg, mockDB)

			relationships, err := testRepo.GetRelationshipsByEmailNotifications()

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, relationships)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRelationships, relationships)
				assert.Len(t, relationships, tc.expectedCount)
			}
		})
	}
}
