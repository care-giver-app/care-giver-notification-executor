package dynamo

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateClient(t *testing.T) {
	tests := map[string]struct {
		appCfg            appconfig.AppConfig
		expectedEndpoint  string
		expectNilEndpoint bool
	}{
		"Happy Path - Local Endpoint": {
			appCfg: appconfig.AppConfig{
				Logger:    zap.Must(zap.NewProduction()),
				AWSConfig: aws.Config{},
				Env:       appconfig.LocalEnv,
			},
			expectedEndpoint: localDockerEndpoint,
		},
		"Happy Path - Non Local Endpoint": {
			appCfg: appconfig.AppConfig{
				Logger:    zap.Must(zap.NewProduction()),
				AWSConfig: aws.Config{},
				Env:       "dev",
			},
			expectNilEndpoint: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateClient(&tc.appCfg)

			if tc.expectNilEndpoint {
				assert.Nil(t, client.Options().BaseEndpoint)
			} else {
				assert.Equal(t, tc.expectedEndpoint, *client.Options().BaseEndpoint)
			}
		})
	}
}
