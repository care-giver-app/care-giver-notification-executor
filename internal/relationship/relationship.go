package relationship

type Relationship struct {
	UserID             string `json:"userId" dynamodbav:"user_id"`
	ReceiverID         string `json:"receiverId" dynamodbav:"receiver_id"`
	PrimaryCareGiver   bool   `json:"primaryCareGiver" dynamodbav:"primary_care_giver"`
	EmailNotifications bool   `json:"emailNotifications" dynamodbav:"email_notifications"`
}

func NewRelationship(uid, rid string, primaryCareGiver, emailNotifications bool) *Relationship {
	return &Relationship{
		UserID:             uid,
		ReceiverID:         rid,
		PrimaryCareGiver:   primaryCareGiver,
		EmailNotifications: emailNotifications,
	}
}

func IsACareGiver(uid string, rid string, relationships []Relationship) bool {
	for _, r := range relationships {
		if r.UserID == uid && r.ReceiverID == rid {
			return true
		}
	}

	return false
}

func IsAPrimaryCareGiver(uid string, rid string, relationships []Relationship) bool {
	for _, r := range relationships {
		if r.UserID == uid && r.ReceiverID == rid && r.PrimaryCareGiver {
			return true
		}
	}

	return false
}
