package sqs

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
)

func CreateClient(cfg *appconfig.AppConfig) *sqs.Client {
	// TODO: Add local endpoint support if needed
	cfg.Logger.Info("creating sqs client")
	return sqs.NewFromConfig(cfg.AWSConfig)
}
