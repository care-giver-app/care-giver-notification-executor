package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/care-giver-app/care-giver-notification-executor/internal/awsconfig"
	"github.com/care-giver-app/care-giver-notification-executor/internal/dynamo"
	"github.com/care-giver-app/care-giver-notification-executor/internal/receiver"
	"github.com/care-giver-app/care-giver-notification-executor/internal/relationship"
	"github.com/care-giver-app/care-giver-notification-executor/internal/repository"
	"github.com/care-giver-app/care-giver-notification-executor/internal/user"
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

type NotificationMessage struct {
	Relationship     relationship.Relationship `json:"relationship"`
	NotificationType string                    `json:"notification_type"`
	Channel          []string                  `json:"channel"`
}

type EmailService struct {
	sesClient *ses.Client
	appCfg    *appconfig.AppConfig
}

func NewEmailService(sesClient *ses.Client, appCfg *appconfig.AppConfig) *EmailService {
	return &EmailService{
		sesClient: sesClient,
		appCfg:    appCfg,
	}
}

const emailTemplate = `
<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <title>CareGiver Notification</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f5f5f5;
        }

        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #ffffff;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
        }

        .header {
            background-color: #ffffff;
            color: #333;
            padding: 30px 20px;
            text-align: center;
            border-bottom: 3px solid #4a90e2;
        }

        .logo {
            max-width: 150px;
            height: auto;
            flex-shrink: 0;
        }

        .header-text {
            flex-grow: 1;
            text-align: center;
            margin-left: 0px;
        }

        .header h1 {
            margin: 0;
            color: #4a90e2;
            font-size: 28px;
            font-weight: bold;
        }

        .content {
            padding: 30px 20px;
            background-color: #ffffff;
        }

        .content h2 {
            color: #4a90e2;
            margin-top: 0;
        }

        .content a {
            color: #4a90e2;
            text-decoration: none;
        }

        .content a:hover {
            text-decoration: underline;
        }

        .footer {
            padding: 20px;
            text-align: center;
            font-size: 12px;
            color: #666;
            background-color: #f8f9fa;
            border-top: 1px solid #e9ecef;
        }

        .notification-type {
            background-color: #e8f4fd;
            padding: 15px;
            border-left: 4px solid #4a90e2;
            margin: 20px 0;
            border-radius: 0 4px 4px 0;
        }
    </style>
</head>

<body>
    <div class="container">
        <div class="header">
            <img src="https://www.caretosher.com/assets/caretosher-logo.png" alt="CareToSher Logo" class="logo">
            <div class="header-text">
                <h1>Daily Reminder</h1>
            </div>
        </div>
        <div class="content">
            <h2>Hello {{USER_NAME}},</h2>
            <p>Thank you for caring for {{RECEIVER_NAME}}! If you have additional events that need to be logged today,
                you
                can add them from your dashboard at <a href="https://caretosher.com/">caretosher.com</a>.
            </p>

            <div class="notification-type">
                <p><strong>Coming soon:</strong> Today's events will be shared with these emails.</p>
            </div>

            <p>We appreciate your dedication to providing excellent care!</p>
        </div>
        <div class="footer">
            <p>This is an automated notification from CareToSher.com</p>
            <p>© 2025 CareToSher. All rights reserved.</p>
        </div>
    </div>
</body>

</html>
`

func (e *EmailService) SendNotificationEmail(ctx context.Context, user user.User, receiver receiver.Receiver, notifMsg NotificationMessage) error {
	logger := e.appCfg.Logger.Sugar()

	// Replace template placeholders with actual values
	emailBody := strings.ReplaceAll(emailTemplate, "{{USER_NAME}}", user.FirstName)
	emailBody = strings.ReplaceAll(emailBody, "{{RECEIVER_NAME}}", receiver.FirstName)
	emailBody = strings.ReplaceAll(emailBody, "{{NOTIFICATION_TYPE}}", strings.ReplaceAll(notifMsg.NotificationType, "_", " "))

	// Create email subject
	subject := fmt.Sprintf("CareToSher Notification: %s - %s", strings.ReplaceAll(notifMsg.NotificationType, "_", " "), receiver.FirstName)

	// Prepare SES email input
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
					Data:    &emailBody,
					Charset: &[]string{"UTF-8"}[0],
				},
			},
		},
	}

	// Send the email
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

	dynamoClient = dynamo.CreateClient(appCfg)

	sesClient = ses.NewFromConfig(cfg)

	receiverRepo = repository.NewReceiverRespository(context.TODO(), appCfg, dynamoClient)
	userRepo = repository.NewUserRespository(context.TODO(), appCfg, dynamoClient)

	appCfg.Logger.Info("initializing relationship repository")
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	logger := appCfg.Logger.Sugar()
	for _, message := range sqsEvent.Records {
		logger.Infof("Processing message ID: %s", message.MessageId)
		logger.Infof("Message Body: %s", message.Body)
		var notifMsg NotificationMessage
		err := json.Unmarshal([]byte(message.Body), &notifMsg)
		if err != nil {
			logger.Errorf("Failed to unmarshal message body: %v", err)
			continue
		}

		logger.Infof("Processing notification for Receiver ID: %s & User ID: %s, Notification Type: %s, Channels: %v",
			notifMsg.Relationship.ReceiverID, notifMsg.Relationship.UserID, notifMsg.NotificationType, notifMsg.Channel)

		user, err := userRepo.GetUser(notifMsg.Relationship.UserID)
		if err != nil {
			logger.Errorf("Failed to get user: %v", err)
			continue
		}

		receiver, err := receiverRepo.GetReceiver(notifMsg.Relationship.ReceiverID)
		if err != nil {
			logger.Errorf("Failed to get receiver: %v", err)
			continue
		}

		if notifMsg.Relationship.EmailNotifications && contains(notifMsg.Channel, "email") {
			emailService := NewEmailService(sesClient, appCfg)
			err = emailService.SendNotificationEmail(
				ctx,
				user,
				receiver,
				notifMsg,
			)
			if err != nil {
				logger.Errorf("Failed to send email notification: %v", err)
				continue
			}

			logger.Infof("Successfully sent email notification to %s for receiver %s", user.Email, receiver.FirstName)
		} else {
			logger.Infof("Email notifications disabled or email not in channels for user %s", user.Email)
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func main() {
	lambda.Start(handler)
}
