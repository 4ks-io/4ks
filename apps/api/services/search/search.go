// Package search is the search service
package search

import (
	"4ks/apps/api/dtos"
	"4ks/libs/go/models"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
)

// Service is the interface for the search service
type Service interface {
	CreateSearchRecipeCollection() error
	RemoveSearchRecipeDocument(string) error
	SearchRecipesByAuthor(string, string, int) ([]*dtos.CreateSearchRecipe, error)
	UpsertSearchRecipeDocument(*models.Recipe) error
}

type searchService struct {
	client *typesense.Client
}

// ServiceConfig holds connection parameters for the Typesense search backend.
type ServiceConfig struct {
	URL string
	Key string
}

// New creates a new search service
func New(client *typesense.Client) Service {
	return &searchService{
		client,
	}
}

func (s searchService) UpsertSearchRecipeDocument(r *models.Recipe) error {
	ing := []string{}
	for _, v := range r.CurrentRevision.Ingredients {
		ing = append(ing, v.Name)
	}

	// ins := []string{}
	// for _, v := range r.CurrentRevision.Instructions {
	// 	ins = append(ins, v.Text)
	// }

	var banner string
	for _, b := range r.CurrentRevision.Banner {
		if b.Alias == "md" {
			banner = b.URL
		}
	}

	document := dtos.CreateSearchRecipe{
		ID:          r.ID,
		Author:      r.Author.Username,
		Name:        r.CurrentRevision.Name,
		Ingredients: ing,
		ImageURL:    banner,
	}

	_, err := s.client.Collection("recipes").Documents().Upsert(document)
	if err != nil {
		return err
	}

	return nil
}

func (s searchService) RemoveSearchRecipeDocument(id string) error {
	_, err := s.client.Collection("recipes").Document(id).Delete()
	if err != nil {
		return err
	}

	return nil
}

func (s searchService) SearchRecipesByAuthor(query string, author string, perPage int) ([]*dtos.CreateSearchRecipe, error) {
	if query == "" {
		query = "*"
	}
	if perPage <= 0 {
		perPage = 20
	}

	filterBy := fmt.Sprintf("author:=%s", author)
	params := &api.SearchCollectionParams{
		Q:        query,
		QueryBy:  "name,ingredients",
		FilterBy: &filterBy,
		PerPage:  &perPage,
	}

	result, err := s.client.Collection("recipes").Documents().Search(params)
	if err != nil {
		return nil, err
	}

	if result.Hits == nil {
		return []*dtos.CreateSearchRecipe{}, nil
	}

	out := make([]*dtos.CreateSearchRecipe, 0, len(*result.Hits))
	for _, hit := range *result.Hits {
		if hit.Document == nil {
			continue
		}

		body, err := json.Marshal(hit.Document)
		if err != nil {
			return nil, err
		}

		var recipe dtos.CreateSearchRecipe
		if err := json.Unmarshal(body, &recipe); err != nil {
			return nil, err
		}
		out = append(out, &recipe)
	}

	return out, nil
}

// note: schema must overlap with additionalSearchParameters in search-context.tsx
func (s searchService) CreateSearchRecipeCollection() error {
	False := false
	True := true
	schema := &api.CollectionSchema{
		Name: "recipes",
		Fields: []api.Field{
			{
				Name: "author",
				Type: "string",
			},
			{
				Name: "name",
				Type: "string",
			},
			{
				Name: "ingredients",
				Type: "string[]",
			},
			{
				Name:     "imageUrl",
				Type:     "string",
				Index:    &False,
				Optional: &True,
			},
		},
	}

	_, err := s.client.Collections().Create(schema)
	log.Error().Err(err).Msg("failed to create search collection")
	if err != nil {
		return err
	}

	return nil
}
