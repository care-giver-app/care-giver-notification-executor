package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"slices"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/care-giver-app/care-giver-golang-common/pkg/awsconfig"
	"github.com/care-giver-app/care-giver-golang-common/pkg/dynamo"
	"github.com/care-giver-app/care-giver-golang-common/pkg/receiver"
	"github.com/care-giver-app/care-giver-golang-common/pkg/relationship"
	"github.com/care-giver-app/care-giver-golang-common/pkg/repository"
	"github.com/care-giver-app/care-giver-golang-common/pkg/user"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
)

const (
	functionName = "care-giver-notification-executor"
)

var (
	dynamoClient *dynamodb.Client
	appCfg       *appconfig.AppConfig
	receiverRepo *repository.ReceiverRepository
	userRepo     *repository.UserRepository
	sesClient    *ses.Client
)

//go:embed templates/*.html
var templateFS embed.FS

type Notification struct {
	NotificationType string   `json:"notification_type"`
	Channel          []string `json:"channel"`
	ExecutionData    any      `json:"execution_data"`
}

type ReminderNotification struct {
	Relationship relationship.Relationship `json:"relationship"`
}

type FeedbackNotification struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

type EmailService struct {
	sesClient     *ses.Client
	appCfg        *appconfig.AppConfig
	emailTemplate *template.Template
}

func NewEmailService(sesClient *ses.Client, appCfg *appconfig.AppConfig, templateName string) (*EmailService, error) {
	var templateFile string
	switch templateName {
	case "reminder":
		templateFile = "email_reminder_notification.html"
	case "feedback":
		templateFile = "email_feedback.html"
	default:
		return nil, fmt.Errorf("unsupported template name: %s", templateName)
	}

	tmpl, err := template.ParseFS(templateFS, fmt.Sprintf("templates/%s", templateFile))
	if err != nil {
		return nil, err
	}

	return &EmailService{
		sesClient:     sesClient,
		appCfg:        appCfg,
		emailTemplate: tmpl,
	}, nil
}

type ReminderEmailTemplateData struct {
	UserName         string
	ReceiverName     string
	NotificationType string
}

type FeedbackEmailTemplateData struct {
	FeedbackMessage string
}

func (e *EmailService) SendFeedbackEmail(ctx context.Context, toEmail string, message string) error {
	logger := e.appCfg.Logger.Sugar()

	templateData := FeedbackEmailTemplateData{
		FeedbackMessage: message,
	}

	var emailBody bytes.Buffer
	if err := e.emailTemplate.Execute(&emailBody, templateData); err != nil {
		logger.Errorf("Failed to execute email template: %v", err)
		return fmt.Errorf("failed to execute email template: %w", err)
	}
	emailBodyStr := emailBody.String()

	subject := "You received feedback"

	input := &ses.SendEmailInput{
		Source: &e.appCfg.SenderEmailAddress,
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    &subject,
				Charset: &[]string{"UTF-8"}[0],
			},
			Body: &types.Body{
				Html: &types.Content{
					Data:    &emailBodyStr,
					Charset: &[]string{"UTF-8"}[0],
				},
			},
		},
	}

	result, err := e.sesClient.SendEmail(ctx, input)
	if err != nil {
		logger.Errorf("Failed to send email to %s: %v", toEmail, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Infof("Successfully sent email to %s. Message ID: %s", toEmail, *result.MessageId)
	return nil
}

func (e *EmailService) SendNotificationEmail(ctx context.Context, user user.User, receiver receiver.Receiver) error {
	logger := e.appCfg.Logger.Sugar()

	templateData := ReminderEmailTemplateData{
		UserName:     user.FirstName,
		ReceiverName: receiver.FirstName,
	}

	var emailBody bytes.Buffer
	if err := e.emailTemplate.Execute(&emailBody, templateData); err != nil {
		logger.Errorf("Failed to execute email template: %v", err)
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	subject := fmt.Sprintf("CareToSher Notification: %s - %s", templateData.NotificationType, receiver.FirstName)

	emailBodyStr := emailBody.String()

	input := &ses.SendEmailInput{
		Source: &e.appCfg.SenderEmailAddress,
		Destination: &types.Destination{
			ToAddresses: []string{user.Email},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    &subject,
				Charset: &[]string{"UTF-8"}[0],
			},
			Body: &types.Body{
				Html: &types.Content{
					Data:    &emailBodyStr,
					Charset: &[]string{"UTF-8"}[0],
				},
			},
		},
	}

	result, err := e.sesClient.SendEmail(ctx, input)
	if err != nil {
		logger.Errorf("Failed to send email to %s: %v", user.Email, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Infof("Successfully sent email to %s. Message ID: %s", user.Email, *result.MessageId)
	return nil
}

func init() {
	appCfg = appconfig.NewAppConfig()
	appCfg.Logger.Sugar().Infof("initializing %s", functionName)

	cfg, err := awsconfig.GetAWSConfig(context.TODO(), appCfg.Env)
	if err != nil {
		appCfg.Logger.Sugar().Fatalf("Unable to load SDK config: %v", err)
	}

	appCfg.AWSConfig = cfg

	dynamoClient = dynamo.CreateClient(appCfg.Env, appCfg.AWSConfig, appCfg.Logger)

	sesClient = ses.NewFromConfig(cfg)

	receiverRepo = repository.NewReceiverRespository(context.TODO(), appCfg.ReceiverTableName, dynamoClient, appCfg.Logger)
	userRepo = repository.NewUserRespository(context.TODO(), appCfg.UserTableName, dynamoClient, appCfg.Logger)

	appCfg.Logger.Info("initializing relationship repository")
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	logger := appCfg.Logger.Sugar()
	for _, message := range sqsEvent.Records {
		var notifMsg Notification
		err := json.Unmarshal([]byte(message.Body), &notifMsg)
		if err != nil {
			logger.Errorf("Failed to unmarshal message body: %v", err)
			continue
		}

		logger.Infof("Processing %s notification for the following channels: %v", notifMsg.NotificationType, notifMsg.Channel)

		switch notifMsg.NotificationType {
		case "reminder":
			if slices.Contains(notifMsg.Channel, "email") {
				if err := handleEmailReminder(ctx, notifMsg.ExecutionData); err != nil {
					logger.Errorf("Failed to handle reminder: %v", err)
					continue
				}
			}
		case "feedback":
			if err := handleEmailFeedback(ctx, notifMsg.ExecutionData); err != nil {
				logger.Errorf("Failed to handle feedback: %v", err)
				continue
			}
		default:
			logger.Warnf("Unknown notification type: %s", notifMsg.NotificationType)
		}
	}
	return nil
}

func handleEmailReminder(ctx context.Context, executionData any) error {
	logger := appCfg.Logger.Sugar()

	var reminderData ReminderNotification
	execDataBytes, err := json.Marshal(executionData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(execDataBytes, &reminderData); err != nil {
		return err

	}

	user, err := userRepo.GetUser(reminderData.Relationship.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	receiver, err := receiverRepo.GetReceiver(reminderData.Relationship.ReceiverID)
	if err != nil {
		return fmt.Errorf("failed to get receiver: %w", err)
	}

	if reminderData.Relationship.EmailNotifications {
		emailService, err := NewEmailService(sesClient, appCfg, "reminder")
		if err != nil {
			return err
		}

		err = emailService.SendNotificationEmail(ctx, user, receiver)
		if err != nil {
			return err
		}

		logger.Infof("Successfully sent email notification to %s for receiver %s", user.Email, receiver.FirstName)
	}

	return nil
}

func handleEmailFeedback(ctx context.Context, executionData any) error {
	logger := appCfg.Logger.Sugar()

	var feedbackData FeedbackNotification
	execDataBytes, err := json.Marshal(executionData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(execDataBytes, &feedbackData); err != nil {
		return err
	}

	emailService, err := NewEmailService(sesClient, appCfg, "feedback")
	if err != nil {
		return err
	}

	err = emailService.SendFeedbackEmail(ctx, feedbackData.Email, feedbackData.Message)
	if err != nil {
		return err
	}

	logger.Infof("Successfully sent feedback email")
	return nil
}

func main() {
	lambda.Start(handler)
}
