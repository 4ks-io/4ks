// Package fetcher provides a set of Cloud Functions samples.
package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

var (
	debug       bool
	apiSecret   string
	apiEndpoint string
)

func getMandatoryEnvVar(n string) string {
	v, ok := os.LookupEnv(n)
	if !ok || v == "" {
		log.Fatal().Err(errors.New("missing env var")).Caller().Msgf("env var %s required", n)
	}
	return v
}

func init() {
	debug = os.Getenv("DEBUG") == "false"

	apiSecret = getMandatoryEnvVar("API_FETCHER_PSK")
	apiEndpoint = getMandatoryEnvVar("API_ENDPOINT_URL")

	n := getMandatoryEnvVar("PUBSUB_PROJECT_ID")
	t := getMandatoryEnvVar("PUBSUB_TOPIC_ID")
	functions.CloudEvent(fmt.Sprintf("projects/%s/topics/%s", n, t), fetcher)
}

// MessagePublishedData contains the full Pub/Sub message
// See the documentation for more details:
// https://cloud.google.com/eventarc/docs/cloudevents#pubsub
type MessagePublishedData struct {
	Message PubSubMessage
}

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// FetcherRequest is a struct to hold the fetcher request data
type FetcherRequest struct {
	URL         string    `json:"url"`
	UserID      string    `json:"userId"`
	UserEventID uuid.UUID `json:"userEventId"`
}

// fetcher consumes a CloudEvent message and extracts the Pub/Sub message.
func fetcher(ctx context.Context, e event.Event) error {
	// event type validation
	if e.Type() != "google.cloud.pubsub.topic.v1.messagePublished" {
		log.Error().Err(errors.New("unexpected cloud even type")).Caller().
			Str("type", e.Type()).Msg("unexpected cloud even type")
	}

	// unmarshal event message
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		log.Error().Err(err).Caller().Msg("failed to unmarshal event data")
		return fmt.Errorf("event.DataAs: %w", err)
	}

	// unmarshal data
	var f FetcherRequest
	if err := json.Unmarshal(msg.Message.Data, &f); err != nil {
		log.Error().Err(err).Caller().Msg("failed to unmarshal msg data")
		return fmt.Errorf("event.DataAs: %w", err)
	}

	validateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	validatedURL, err := validateFetchURL(validateCtx, f.URL)
	if err != nil {
		return err
	}

	// scrape recipe
	recipe, err := visit(ctx, debug, validatedURL.Normalized)
	if err != nil {
		log.Error().Err(err).Msg("failed to visit")
	}

	// format reponse data
	dto := FetcherCreateRecipe{
		UserID:      f.UserID,
		UserEventID: f.UserEventID,
		Recipe:      createRecipeDtoFromRecipe(recipe),
	}

	// PrintStruct(dto)
	// tr@ck: validate dto and post errors to api

	// marshall data to json
	data, err := json.Marshal(dto)
	if err != nil {
		log.Error().Err(err).Caller().Msg("failed to marshal recipe")
		return err
	}

	nonce, err := newNonce()
	if err != nil {
		log.Error().Err(err).Caller().Msg("failed to create auth nonce")
		return err
	}

	// api callback
	client := http.Client{}
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal().Err(err).Caller().Msg("failed to create api callback request")
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	headers := buildSignatureHeaders([]byte(apiSecret), req.Method, req.URL.Host, req.URL.RequestURI(), data, time.Now(), nonce)
	applySignatureHeaders(req, headers)

	// perform request
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Caller().Msg("failed to perform api callback")
		return err
	}
	defer resp.Body.Close()

	// read response
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Caller().Msg("failed read api callback response")
		return err
	}

	return nil
}

// PrintStruct prints a struct
func PrintStruct(t interface{}) {
	j, _ := json.MarshalIndent(t, "", "  ")
	fmt.Println(string(j))
}

// tr@ck: import from a dtos package?

// FetcherCreateRecipe godoc
type FetcherCreateRecipe struct {
	Recipe      CreateRecipe `json:"recipe"`
	UserID      string       `json:"userId"`
	UserEventID uuid.UUID    `json:"userEventId"`
}

// CreateRecipe godoc
type CreateRecipe struct {
	Name         string               `json:"name"`
	Link         string               `json:"link"`
	Author       UserSummary          `json:"-"` // Author is auto-populated using the request context
	Instructions []Instruction        `json:"instructions"`
	Ingredients  []Ingredient         `json:"ingredients"`
	Banner       []RecipeMediaVariant `json:"banner"`
}

type RecipeMediaVariant struct {
	MaxWidth int    `firestore:"maxWidth" json:"maxWidth"`
	URL      string `firestore:"url" json:"url"`
	Filename string `firestore:"filename" json:"filename"`
	Alias    string `firestore:"alias" json:"alias"`
}

type Instruction struct {
	ID   int    `firestore:"id" json:"id"`
	Type string `firestore:"type" json:"type"`
	Name string `firestore:"name" json:"name"`
	Text string `firestore:"text" json:"text"`
}

type Ingredient struct {
	ID       int    `firestore:"id" json:"id"`
	Type     string `firestore:"type" json:"type"`
	Name     string `firestore:"name" json:"name"`
	Quantity string `firestore:"quantity" json:"quantity"`
}

type UserSummary struct {
	ID          string `firestore:"id,omitempty" json:"id"`
	Username    string `firestore:"username,omitempty" json:"username"`
	DisplayName string `firestore:"displayName,omitempty" json:"displayName"`
}

func createRecipeDtoFromRecipe(r Recipe) CreateRecipe {
	instructions := []Instruction{}
	for _, v := range r.Instructions {
		instructions = append(instructions, Instruction{
			Text: v,
		})
	}

	ingredients := []Ingredient{}
	for _, v := range r.Ingredients {
		ingredients = append(ingredients, Ingredient{
			Name: v,
		})
	}

	// recipe response
	return CreateRecipe{
		Name:         r.Title,
		Link:         r.SourceURL,
		Instructions: instructions,
		Ingredients:  ingredients,
	}
}
