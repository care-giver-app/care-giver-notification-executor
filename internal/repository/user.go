package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/care-giver-app/care-giver-notification-executor/internal/log"
	"github.com/care-giver-app/care-giver-notification-executor/internal/user"
	"go.uber.org/zap"
)

type DynamodbClientProvider interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

type UserRepositoryProvider interface {
	CreateUser(u user.User) error
	GetUser(uid string) (user.User, error)
	GetUserByEmail(email string) (user.User, error)
}

const (
	PrimaryReceiverList    string = "primary_care_receivers"
	AdditionalReceiverList string = "additional_care_receivers"
)

var (
	userID = "user_id"
)

type UserRepository struct {
	Ctx       context.Context
	Client    DynamodbClientProvider
	TableName string
	logger    *zap.Logger
}

func NewUserRespository(ctx context.Context, cfg *appconfig.AppConfig, client DynamodbClientProvider) *UserRepository {
	return &UserRepository{
		Ctx:       ctx,
		Client:    client,
		TableName: cfg.UserTableName,
		logger:    cfg.Logger.With(zap.String(log.TableNameLogKey, cfg.UserTableName)),
	}
}

func (ur *UserRepository) CreateUser(u user.User) error {
	ur.logger.Info("adding user to db", zap.Any(log.UserIDLogKey, u.UserID))

	ur.logger.Info("marshalling user struct")
	av, err := attributevalue.MarshalMap(u)
	if err != nil {
		return err
	}

	ur.logger.Info("inserting item into db", zap.Any("item", av))
	_, err = ur.Client.PutItem(ur.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String(ur.TableName),
		Item:      av,
	})
	if err != nil {
		return err
	}
	ur.logger.Info("successfully inserted item")

	return nil
}

func (ur *UserRepository) GetUser(uid string) (user.User, error) {
	ur.logger.Info("getting user from db", zap.Any(log.UserIDLogKey, uid))
	result, err := ur.Client.GetItem(ur.Ctx, &dynamodb.GetItemInput{
		TableName: &ur.TableName,
		Key: map[string]types.AttributeValue{
			userID: &types.AttributeValueMemberS{Value: string(uid)},
		},
	})
	if err != nil {
		return user.User{}, err
	}

	var u user.User
	err = attributevalue.UnmarshalMap(result.Item, &u)
	if err != nil {
		return user.User{}, err
	}

	return u, nil
}

func (ur *UserRepository) GetUserByEmail(email string) (user.User, error) {
	ur.logger.Info("getting user from db")

	keyCondition := "email = :email"
	expressionAttributeValues := map[string]types.AttributeValue{
		":email": &types.AttributeValueMemberS{Value: string(email)},
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(ur.TableName),
		IndexName:                 aws.String("email"),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: expressionAttributeValues,
	}

	result, err := ur.Client.Query(ur.Ctx, queryInput)
	if err != nil {
		return user.User{}, err
	}

	if len(result.Items) == 1 {
		var u user.User
		err = attributevalue.UnmarshalMap(result.Items[0], &u)
		if err != nil {
			return user.User{}, err
		}
		return u, nil
	}

	return user.User{}, fmt.Errorf("user with email %s not found", email)
}
