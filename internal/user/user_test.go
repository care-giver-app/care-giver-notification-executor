package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testFirstName = "Demo"
	testLastName  = "Daniel"
	testEmail     = "Demo.Daniel@email.com"
)

func TestNewUser(t *testing.T) {
	expectedUser := &User{
		FirstName: testFirstName,
		LastName:  testLastName,
		Email:     testEmail,
	}

	user, err := NewUser(testEmail, testFirstName, testLastName)
	assert.Nil(t, err)

	expectedUser.UserID = user.UserID
	assert.Equal(t, expectedUser, user)
}
