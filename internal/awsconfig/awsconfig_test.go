package awsconfig

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/care-giver-app/care-giver-notification-executor/internal/appconfig"
	"github.com/stretchr/testify/assert"
)

var (
	localCreds = aws.Credentials{
		AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
		Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
	}
)

func TestGetAWSConfig(t *testing.T) {
	tests := map[string]struct {
		env            string
		expectedRegion string
	}{
		"Happy Path - Local Env": {
			env:            appconfig.LocalEnv,
			expectedRegion: USEastTwoRegion,
		},
		"Happy Path - Non Local Env": {
			env:            "dev",
			expectedRegion: USEastTwoRegion,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg, err := GetAWSConfig(context.Background(), tc.env)
			creds, _ := cfg.Credentials.Retrieve(context.Background())

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedRegion, cfg.Region)
			if tc.env == appconfig.LocalEnv {
				assert.Equal(t, localCreds, creds)
			} else {
				assert.NotEqual(t, localCreds, creds)
			}
		})
	}
}
