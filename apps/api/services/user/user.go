// Package usersvc is the user service
package usersvc

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	firestore "cloud.google.com/go/firestore"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"

	"4ks/apps/api/dtos"
	"4ks/apps/api/utils"
	models "4ks/libs/go/models"
)

var (
	// ErrEmailInUse is returned when a user email is already in use
	ErrEmailInUse = errors.New("email already in use")
	// ErrUsernameInUse is returned when a username is already in use
	ErrUsernameInUse = errors.New("username already in use")
	// ErrInvalidUsername is returned when a username is invalid
	ErrInvalidUsername = errors.New("invalid username")
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrReservedWord is returned when a username is a reserved word
	ErrReservedWord = errors.New("reserved word")
	// ErrUserEventNotFound is returned when a user event is not found
	ErrUserEventNotFound = errors.New("user event not found")
)

// Service is the interface for the user service
type Service interface {
	GetAllUsers(context.Context) ([]*models.User, error)
	GetUserByID(context.Context, string) (*models.User, error)
	GetUserByUsername(context.Context, string) (*models.User, error)
	GetUserByEmail(context.Context, string) (*models.User, error)
	CreateUser(context.Context, string, string, *dtos.CreateUser) (*models.User, error)
	CreateUserFromMCP(ctx context.Context, email, username string) (*models.User, error)
	UpdateUserByID(context.Context, string, *dtos.UpdateUser) (*models.User, error)
	DeleteUser(context.Context, string) error
	CreateUserEventByUserID(context.Context, string, *dtos.CreateUserEvent) (*models.UserEvent, error)
	UpdateUserEventByUserIDEventID(context.Context, string, *dtos.UpdateUserEvent) (*models.UserEvent, error)
	RemoveUserEventByUserIDEventID(context.Context, string, uuid.UUID) error

	// test username
	TestName(context.Context, string) error
	TestValidName(string) bool
	TestReservedWord(string) bool
	TestAvailableName(context.Context, string) (bool, error)
}

type userService struct {
	userCollection *firestore.CollectionRef
	validator      *validator.Validate
	reservedWords  *[]string
	sysFlags       *utils.SystemFlags
}

// New creates a new user service
func New(sysFlags *utils.SystemFlags, store *firestore.Client, validator *validator.Validate, reservedWords *[]string) Service {
	return &userService{
		sysFlags:       sysFlags,
		validator:      validator,
		reservedWords:  reservedWords,
		userCollection: store.Collection("users"),
	}
}

func (s userService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var all []*models.User
	iter := s.userCollection.Documents(ctx)
	defer iter.Stop() // add this line to ensure resources cleaned up

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Handle error, possibly by returning the error
			// to the caller. Break the loop or return.
			return nil, err
		}
		var u models.User
		if err := doc.DataTo(&u); err != nil {
			// Handle error, possibly by returning the error
			// to the caller. Continue the loop,
			// break the loop or return.
			return nil, err
		}
		// fmt.Println(u.ID)
		all = append(all, &u)
	}

	return all, nil
}

func (s userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	result, err := s.userCollection.Doc(id).Get(ctx)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user := new(models.User)
	if err = result.DataTo(user); err != nil {
		return nil, err
	}

	user.ID = result.Ref.ID
	return user, nil
}

func (s userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	// result, err := userCollection.Where("usernameLower", "==", l).Documents(ctx).GetAll()
	result, err := s.userCollection.Where("usernameLower", "==", strings.ToLower(username)).Documents(ctx).GetAll()
	if err != nil || len(result) == 0 {
		return nil, ErrUserNotFound
	}
	// fmt.Print(result)

	userSnapshot := result[0]
	user := new(models.User)
	// fmt.Print(user.ID)

	err = userSnapshot.DataTo(user)
	if err != nil {
		return nil, err
	}

	user.ID = userSnapshot.Ref.ID
	return user, nil
}

func (s userService) GetUserByEmail(ctx context.Context, emailAddress string) (*models.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(emailAddress))
	log.Debug().Str("email", normalizedEmail).Msg("user service lookup by email")
	result, err := s.userCollection.Where("emailAddress", "==", normalizedEmail).Documents(ctx).GetAll()
	if err != nil || len(result) == 0 {
		log.Debug().Str("email", normalizedEmail).Msg("user service lookup by email missed")
		return nil, ErrUserNotFound
	}

	userSnapshot := result[0]
	user := new(models.User)

	err = userSnapshot.DataTo(user)
	if err != nil {
		return nil, err
	}

	user.ID = userSnapshot.Ref.ID
	log.Debug().
		Str("email", normalizedEmail).
		Str("resolvedUserID", user.ID).
		Str("resolvedEmail", user.EmailAddress).
		Msg("user service lookup by email resolved")
	return user, nil
}

func (s userService) CreateUserFromMCP(ctx context.Context, email, username string) (*models.User, error) {
	if err := s.TestName(ctx, username); err != nil {
		return nil, err
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	userID := uuid.NewString()
	newUser := &models.User{
		ID:               userID,
		Username:         username,
		UsernameLower:    strings.ToLower(username),
		EmailAddress:     normalizedEmail,
		OnboardingSource: "mcp",
		FirstLogin:       true,
		WelcomeEmailSent: false,
		CreatedDate:      time.Now().UTC(),
		UpdatedDate:      time.Now().UTC(),
	}
	log.Debug().
		Str("createdUserID", userID).
		Str("email", normalizedEmail).
		Str("username", username).
		Msg("user service creating MCP user")
	if _, err := s.userCollection.Doc(userID).Create(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}

func (s userService) CreateUser(ctx context.Context, userID string, userEmail string, user *dtos.CreateUser) (*models.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(userEmail))
	log.Debug().
		Str("claimUserID", userID).
		Str("email", normalizedEmail).
		Str("username", user.Username).
		Msg("user service creating web user")

	// Email is the identity key: if a user with this address already exists
	// (e.g., auto-created via MCP), return them instead of creating a duplicate.
	if existing, err := s.GetUserByEmail(ctx, normalizedEmail); err == nil {
		log.Debug().
			Str("claimUserID", userID).
			Str("email", normalizedEmail).
			Str("resolvedUserID", existing.ID).
			Str("resolvedEmail", existing.EmailAddress).
			Msg("user service web create resolved existing user by email")
		return existing, nil
	}

	if err := s.TestName(ctx, user.Username); err != nil {
		return nil, err
	}

	newUser := &models.User{
		ID:               userID,
		Username:         user.Username,
		UsernameLower:    strings.ToLower(user.Username),
		DisplayName:      user.DisplayName,
		EmailAddress:     normalizedEmail,
		OnboardingSource: "web",
		CreatedDate:      time.Now().UTC(),
		UpdatedDate:      time.Now().UTC(),
	}

	if _, err := s.userCollection.Doc(userID).Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s userService) UpdateUserByID(ctx context.Context, userID string, user *dtos.UpdateUser) (*models.User, error) {
	if err := s.TestName(ctx, user.Username); err != nil {
		return nil, err
	}

	if _, err := s.userCollection.Doc(userID).Update(ctx, []firestore.Update{
		{
			Path:  "username",
			Value: user.Username,
		},
		{
			Path:  "usernameLower",
			Value: strings.ToLower(user.Username),
		},
	}); err != nil {
		return nil, err
	}

	u, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// CreateUserEventByUserID creates a user event for the given user id
func (s userService) CreateUserEventByUserID(ctx context.Context, id string, e *dtos.CreateUserEvent) (*models.UserEvent, error) {
	result, err := s.userCollection.Doc(id).Get(ctx)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user := new(models.User)
	if err = result.DataTo(user); err != nil {
		return nil, err
	}

	if user.Events == nil {
		user.Events = make([]models.UserEvent, 0)
	}

	newUserEvent := models.UserEvent{
		ID:          uuid.New(),
		Type:        e.Type,
		Status:      e.Status,
		Data:        e.Data,
		CreatedDate: time.Now().UTC(),
		UpdatedDate: time.Now().UTC(),
	}

	user.Events = append(user.Events, newUserEvent)

	if _, err := s.userCollection.Doc(id).Update(ctx, []firestore.Update{
		{
			Path:  "events",
			Value: user.Events,
		},
	}); err != nil {
		return nil, err
	}

	return &newUserEvent, nil
}

// UpdateUserEventByUserIDEventID creates a user event for the given user id
func (s userService) UpdateUserEventByUserIDEventID(ctx context.Context, id string, e *dtos.UpdateUserEvent) (*models.UserEvent, error) {
	if zero, _ := uuid.Parse("00000000-0000-0000-0000-000000000000"); e.ID == uuid.Nil || e.ID == zero {
		return nil, ErrUserEventNotFound
	}

	result, err := s.userCollection.Doc(id).Get(ctx)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user := new(models.User)
	if err = result.DataTo(user); err != nil {
		return nil, err
	}

	if user.Events == nil {
		return nil, ErrUserEventNotFound
	}

	ie := -1
	for i, ue := range user.Events {
		if ue.ID == e.ID {
			ie = i
			user.Events[i].Status = e.Status
			user.Events[i].Error = e.Error
			user.Events[i].Data = e.Data
			user.Events[i].UpdatedDate = time.Now().UTC()
		}
	}
	if ie == -1 {
		return nil, ErrUserEventNotFound
	}

	// tr@ck: improve performance
	// https://stackoverflow.com/questions/46757614/how-to-update-an-array-of-objects-with-firestore/51794212#51794212

	if _, err := s.userCollection.Doc(id).Update(ctx, []firestore.Update{
		{
			Path:  "events",
			Value: user.Events,
		},
	}); err != nil {
		return nil, err
	}

	return &user.Events[ie], nil
}

// RemoveUserEventByUserIDEventID deletes the specified user event for a given user id
func (s userService) RemoveUserEventByUserIDEventID(ctx context.Context, id string, eventID uuid.UUID) error {
	result, err := s.userCollection.Doc(id).Get(ctx)
	if err != nil {
		return ErrUserNotFound
	}

	user := new(models.User)
	if err = result.DataTo(user); err != nil {
		return err
	}

	if user.Events == nil {
		return ErrUserEventNotFound
	}

	var ok bool
	var newEvents []models.UserEvent
	for _, ue := range user.Events {
		if ue.ID == eventID {
			ok = true
			continue
		}
		newEvents = append(newEvents, ue)
	}

	// return if not found
	if !ok {
		return ErrUserEventNotFound
	}

	if _, err := s.userCollection.Doc(id).Update(ctx, []firestore.Update{
		{
			Path:  "events",
			Value: newEvents,
		},
	}); err != nil {
		return err
	}

	return nil
}

// tr@ck: add with disableUser/enableUser as we never want to delete a recipe
func (s userService) DeleteUser(ctx context.Context, userID string) error {
	existingUserID, _ := s.userCollection.Doc(userID).Get(ctx)
	if !existingUserID.Exists() {
		return ErrUserNotFound
	}

	_, err := s.userCollection.Doc(userID).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

// TestReservedWord tests if the entityname is a reserved word
func (s userService) TestReservedWord(name string) bool {
	for _, word := range *s.reservedWords {
		if word == name {
			return true
		}
	}
	return false
}

// TestValidName tests if the entityname is valid
func (s userService) TestValidName(name string) bool {
	/*
		1. at least 8 characters
		2. no longer than 24 characters
		3. alphanumeric characters or hyphens
		4. no consecutive hyphens
		5. cannot begin or end with a hyphen
	*/

	// tr@ck: combine these 2 regex...
	if regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9-]{6,22}[a-zA-Z0-9]$").MatchString(name) {
		return !regexp.MustCompile("--").MatchString(name)
	}

	return false
}

// TestAvailableName tests if the name is available
func (s userService) TestAvailableName(ctx context.Context, username string) (bool, error) {
	// u, err := url.PathUnescape(*username)
	usersWithUsername, err := s.userCollection.Where("usernameLower", "==", strings.ToLower(username)).Documents(ctx).GetAll()
	if err != nil || len(usersWithUsername) > 0 {
		return true, ErrUsernameInUse
	}
	return false, nil
}

func (s userService) TestName(ctx context.Context, n string) error {
	// test validity
	if isValid := s.TestValidName(n); !isValid {
		return ErrInvalidUsername
	}

	// test reserved word
	if isReserved := s.TestReservedWord(n); isReserved {
		return ErrReservedWord
	}

	if _, err := s.TestAvailableName(ctx, n); err != nil {
		if err == ErrUsernameInUse {
			return ErrUsernameInUse
		} else if err != ErrUserNotFound {
			return err
		}
	}

	return nil
}
