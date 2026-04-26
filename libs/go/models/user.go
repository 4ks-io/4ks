package models

import (
	"time"

	"github.com/google/uuid"
)

// User godoc
type User struct {
	ID            string      `firestore:"id" json:"id"`
	Events        []UserEvent `firestore:"events,omitempty" json:"events"`
	Username      string      `firestore:"username,omitempty" json:"username"`
	UsernameLower string      `firestore:"usernameLower,omitempty" json:"usernameLower"`
	DisplayName   string      `firestore:"displayName,omitempty" json:"displayName"`
	EmailAddress  string      `firestore:"emailAddress,omitempty" json:"emailAddress"`
	CreatedDate   time.Time   `firestore:"createdDate,omitempty" json:"createdDate"`
	UpdatedDate   time.Time   `firestore:"updatedDate,omitempty" json:"updatedDate"`
}

// UserSummary is a compact user reference nested inside Recipe and RecipeRevision documents.
type UserSummary struct {
	ID          string `firestore:"id,omitempty" json:"id"`
	Username    string `firestore:"username,omitempty" json:"username"`
	DisplayName string `firestore:"displayName,omitempty" json:"displayName"`
}

// Username is the validation response for a username availability check.
type Username struct {
	Valid bool   `json:"valid" binding:"required"`
	Msg   string `json:"msg" binding:"required"`
}

// UserEvent godoc
type UserEvent struct {
	ID          uuid.UUID       `firestore:"id" json:"id"`
	Type        UserEventType   `firestore:"type" json:"type"`
	Status      UserEventStatus `firestore:"status" json:"status"`
	CreatedDate time.Time       `firestore:"createdDate,omitempty" json:"createdDate"`
	UpdatedDate time.Time       `firestore:"updatedDate,omitempty" json:"updatedDate"`
	// data is to be unmarshalled based on UserEventStatus ONLY
	Data  interface{}    `firestore:"data,omitempty" json:"data"`
	Error UserEventError `firestore:"error,omitempty" json:"error"`
}

// UserEventError carries error details when a UserEvent has failed status.
type UserEventError struct {
	Message string `firestore:"message,omitempty" json:"message"`
	Code    int    `firestore:"code,omitempty" json:"code"`
}

// UserEventStatus represents the processing state of a user event.
type UserEventStatus int

// UserEventStatus constants.
const (
	UserEventCreated      UserEventStatus = 0
	UserEventProcessing   UserEventStatus = 1
	UserEventReady        UserEventStatus = 2
	UserEventAcknowledged UserEventStatus = 3
	UserEventExpired      UserEventStatus = 9
	UserEventErrorState   UserEventStatus = 60
)

// UserEventType identifies the kind of action that triggered a user event.
type UserEventType int

// UserEventType constants.
const (
	UserEventTypeNewUser      UserEventType = 0
	UserEventTypeCreateRecipe UserEventType = 2
	UserEventTypeForkRecipe   UserEventType = 3
	UserEventTypeFetchRecipe  UserEventType = 9
)
