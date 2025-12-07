package receiver

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	DBPrefix = "Receiver"
	ParamID  = "receiverId"
)

type Receiver struct {
	ReceiverID string `json:"receiverId" dynamodbav:"receiver_id"`
	FirstName  string `json:"firstName" dynamodbav:"first_name"`
	LastName   string `json:"lastName" dynamodbav:"last_name"`
}

func NewReceiver(firstName string, lastName string) *Receiver {
	return &Receiver{
		ReceiverID: fmt.Sprintf("%s#%s", DBPrefix, uuid.New()),
		FirstName:  firstName,
		LastName:   lastName,
	}
}
