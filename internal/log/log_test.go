package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetLogger(t *testing.T) {
	logger, err := GetLogger(InfoLevel)
	assert.Nil(t, err)
	assert.Equal(t, logger.Level(), zap.InfoLevel)
}

func TestGetLoggerWithEnv(t *testing.T) {
	logger, err := GetLoggerWithEnv(InfoLevel, "test")
	assert.Nil(t, err)
	assert.Equal(t, logger.Level(), zap.InfoLevel)
}
