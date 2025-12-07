package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/care-giver-app/care-giver-notification-executor/internal/event"
	"github.com/care-giver-app/care-giver-notification-executor/internal/log"
	"go.uber.org/zap"
)

type EventRepositoryProvider interface {
	AddEvent(e *event.Entry) error
	GetEvents(rid string) ([]event.Entry, error)
	DeleteEvent(rid, eid string) error
}

type EventRepository struct {
	Ctx       context.Context
	Client    DynamodbClientProvider
	TableName string
	logger    *zap.Logger
}

func NewEventRespository(ctx context.Context, cfg *appconfig.AppConfig, client DynamodbClientProvider) *EventRepository {
	return &EventRepository{
		Ctx:       ctx,
		Client:    client,
		TableName: cfg.EventTableName,
		logger:    cfg.Logger.With(zap.String(log.TableNameLogKey, cfg.EventTableName)),
	}
}

func (er *EventRepository) AddEvent(e *event.Entry) error {
	er.logger.Info("adding receiver event to db")

	er.logger.Info("marshalling receiver event struct")
	av, err := attributevalue.MarshalMap(e)
	if err != nil {
		return err
	}

	er.logger.Info("inserting item into db", zap.Any("item", av))
	_, err = er.Client.PutItem(er.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String(er.TableName),
		Item:      av,
	})
	if err != nil {
		return err
	}
	er.logger.Info("successfully inserted item")

	return nil
}

func (er *EventRepository) GetEvents(rid string) ([]event.Entry, error) {
	er.logger.Info("retrieving receiver events from db", zap.String(log.ReceiverIDLogKey, string(rid)))

	keyCondition := "receiver_id = :rid"
	expressionAttributeValues := map[string]types.AttributeValue{
		":rid": &types.AttributeValueMemberS{Value: string(rid)},
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(er.TableName),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: expressionAttributeValues,
	}

	result, err := er.Client.Query(er.Ctx, queryInput)
	if err != nil {
		return nil, err
	}

	var eventsList []event.Entry
	err = attributevalue.UnmarshalListOfMaps(result.Items, &eventsList)
	if err != nil {
		er.logger.Error("error unmarshalling events list", zap.Error(err))
		return nil, err
	}

	return eventsList, nil
}

func (er *EventRepository) DeleteEvent(rid, eid string) error {
	er.logger.Info("deleting receiver event from db", zap.String(log.EventIDLogKey, eid))

	_, err := er.Client.DeleteItem(er.Ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(er.TableName),
		Key: map[string]types.AttributeValue{
			"receiver_id": &types.AttributeValueMemberS{Value: rid},
			"event_id":    &types.AttributeValueMemberS{Value: eid},
		},
	})

	if err != nil {
		return err
	}

	er.logger.Info("successfully deleted event")
	return nil
}
