package receiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReceiver(t *testing.T) {
	testFirstName := "Demo"
	testLastName := "Dan"
	expectedReceiver := &Receiver{
		FirstName: testFirstName,
		LastName:  testLastName,
	}

	receiver := NewReceiver(testFirstName, testLastName)
	expectedReceiver.ReceiverID = receiver.ReceiverID

	assert.Equal(t, expectedReceiver, receiver)
}
