package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/care-giver-app/care-giver-notification-executor/internal/dynamo"
	"github.com/care-giver-app/care-giver-notification-executor/internal/log"
	"github.com/care-giver-app/care-giver-notification-executor/internal/relationship"
	"go.uber.org/zap"
)

type RelationshipRepositoryProvider interface {
	GetRelationshipsByEmailNotifications() ([]relationship.Relationship, error)
}

type RelationshipRepository struct {
	Ctx       context.Context
	Client    dynamo.DynamodbClientProvider
	TableName string
	logger    *zap.Logger
}

func NewRelationshipRepository(ctx context.Context, cfg *appconfig.AppConfig, client dynamo.DynamodbClientProvider) *RelationshipRepository {
	return &RelationshipRepository{
		Ctx:       ctx,
		Client:    client,
		TableName: cfg.RelationshipTableName,
		logger:    cfg.Logger.With(zap.String(log.TableNameLogKey, cfg.RelationshipTableName)),
	}
}

func (r *RelationshipRepository) GetRelationshipsByEmailNotifications() ([]relationship.Relationship, error) {
	r.logger.Info("getting relationships with email notifications enabled")

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		IndexName:              aws.String("email_notifications"),
		KeyConditionExpression: aws.String("email_notifications_gsi_pk = :email_notifications_gsi_pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email_notifications_gsi_pk": &types.AttributeValueMemberN{Value: "1"},
		},
	}

	var relationships []relationship.Relationship

	paginator := dynamodb.NewQueryPaginator(r.Client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(r.Ctx)
		if err != nil {
			r.logger.Error("failed to query relationships by email notification",
				zap.Error(err))
			return nil, err
		}

		var pageRelationships []relationship.Relationship
		err = attributevalue.UnmarshalListOfMaps(page.Items, &pageRelationships)
		if err != nil {
			r.logger.Error("failed to unmarshal relationships",
				zap.Error(err))
			return nil, err
		}

		relationships = append(relationships, pageRelationships...)
	}

	r.logger.Info("successfully retrieved relationships with email notifications",
		zap.Int("count", len(relationships)))

	return relationships, nil
}
