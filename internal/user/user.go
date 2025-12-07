package user

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	ParamID  = "userId"
	DBPrefix = "User"
)

type UserID string

type User struct {
	UserID    string `json:"userId" dynamodbav:"user_id"`
	Email     string `json:"email" dynamodbav:"email"`
	FirstName string `json:"firstName" dynamodbav:"first_name"`
	LastName  string `json:"lastName" dynamodbav:"last_name"`
}

func NewUser(email string, firstName string, lastName string) (*User, error) {
	return &User{
		UserID:    fmt.Sprintf("%s#%s", DBPrefix, uuid.New()),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}
