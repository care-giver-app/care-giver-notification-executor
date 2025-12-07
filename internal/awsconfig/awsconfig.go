package awsconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
)

var (
	USEastTwoRegion = "us-east-2"
)

func GetAWSConfig(ctx context.Context, env string) (aws.Config, error) {
	if env == appconfig.LocalEnv {
		return getLocalAWSConfig(ctx)
	}
	return getAWSConfig(ctx)
}

func getAWSConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(USEastTwoRegion),
	)

	return cfg, err
}

func getLocalAWSConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(USEastTwoRegion),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
	)

	return cfg, err
}
