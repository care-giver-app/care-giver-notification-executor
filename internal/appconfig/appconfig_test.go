package appconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadEnvVars(t *testing.T) {
	t.Setenv("ENV", "TEST")
	t.Setenv("USER_TABLE_NAME", "user-table-test")

	ac := &AppConfig{}
	ac.ReadEnvVars()

	assert.Equal(t, "TEST", ac.Env)
	assert.Equal(t, "user-table-test", ac.UserTableName)
	assert.Equal(t, "receiver-table-local", ac.ReceiverTableName)
	assert.Equal(t, "event-table-local", ac.EventTableName)
}
