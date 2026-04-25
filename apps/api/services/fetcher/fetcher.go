// Package fetchervc is the user service
package fetchervc

import (
	recipeService "4ks/apps/api/services/recipe"
	searchService "4ks/apps/api/services/search"
	staticService "4ks/apps/api/services/static"
	userService "4ks/apps/api/services/user"
	"4ks/libs/go/models"
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"

	"4ks/apps/api/utils"
)

// FetcherOpts is the options for the fetcher service
type FetcherOpts struct {
	ProjectID string
	TopicID   string
}

// FetcherPubSubConnection is the connection to the fetcher topic
type FetcherPubSubConnection struct {
	ProjectID string
	TopicID   string
	Topic     *pubsub.Topic
}

// Service is the interface for the user service
type Service interface {
	Send(context.Context, *models.FetcherRequest) (string, error)
	// Start() error
}

type fetcherService struct {
	sysFlags      *utils.SystemFlags
	sender        *FetcherPubSubConnection
	context       context.Context
	recipeService recipeService.Service
	userService   userService.Service
	searchService searchService.Service
	staticService staticService.Service
	// ready         bool
}

// New creates a new user service
func New(ctx context.Context, sysFlags *utils.SystemFlags, client *pubsub.Client, reso FetcherOpts, userSvc userService.Service, recipeSvc recipeService.Service, searchSvc searchService.Service, staticSvc staticService.Service) Service {
	// connect to sender topic
	t := client.Topic(reso.TopicID)

	// check if topic exists
	ok, err := t.Exists(ctx)
	if err != nil {
		log.Fatal().Err(err).Str("project", reso.ProjectID).Str("topic", reso.TopicID).Msg("failed to check fetcher topic")
	}
	if !ok {
		log.Fatal().Str("project", reso.ProjectID).Str("topic", reso.TopicID).Msg("fetcher topic not found — run init-pubsub.sh")
	}

	// create sender
	sender := &FetcherPubSubConnection{
		ProjectID: reso.ProjectID,
		TopicID:   reso.TopicID,
		Topic:     t,
	}

	return &fetcherService{
		sysFlags:      sysFlags,
		staticService: staticSvc,
		searchService: searchSvc,
		recipeService: recipeSvc,
		userService:   userSvc,
		sender:        sender,
		context:       ctx,
	}
}

func (s *fetcherService) Send(ctx context.Context, data *models.FetcherRequest) (string, error) {
	var id string
	t := s.sender.Topic

	d, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode message")
		return id, err
	}

	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(d),
		Attributes: map[string]string{
			"URL":         data.URL,
			"UserID":      data.UserID,
			"UserEventID": (data.UserEventID).String(),
		},
	})

	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err = result.Get(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to publish message")
	}

	return id, nil
}
