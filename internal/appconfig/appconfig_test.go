package appconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	ac = NewAppConfig()
)

func TestReadEnvVars(t *testing.T) {
	os.Setenv("ENV", "TEST")
	os.Setenv("USER_TABLE_NAME", "user-table-test")
	ac.ReadEnvVars()

	assert.Equal(t, "TEST", ac.Env)
	assert.Equal(t, "user-table-test", ac.UserTableName)
	assert.Equal(t, "receiver-table-local", ac.ReceiverTableName)
	assert.Equal(t, "event-table-local", ac.EventTableName)
	assert.Equal(t, "relationship-table-local", ac.RelationshipTableName)
}
