package kitchenpasssvc

import (
	"encoding/json"
	"fmt"
	"strings"
)

const apiBaseURL = "https://api.4ks.io"

type SkillDocument struct {
	Version           string              `json:"version"`
	Name              string              `json:"name"`
	Description       string              `json:"description"`
	APIBaseURL        string              `json:"apiBaseUrl"`
	Authentication    SkillAuthentication `json:"authentication"`
	AllowedActions    []string            `json:"allowedActions"`
	DisallowedActions []string            `json:"disallowedActions"`
	DecisionRules     []string            `json:"decisionRules"`
	ErrorGuidance     []SkillErrorRule    `json:"errorGuidance"`
	Examples          SkillExamples       `json:"examples"`
}

type SkillAuthentication struct {
	Type          string `json:"type"`
	Header        string `json:"header"`
	BearerToken   string `json:"bearerToken"`
	HeaderExample string `json:"headerExample"`
}

type SkillErrorRule struct {
	StatusCode      int      `json:"statusCode"`
	Classification  string   `json:"classification"`
	Retry           string   `json:"retry"`
	RecommendedWait string   `json:"recommendedWait,omitempty"`
	When            string   `json:"when"`
	ExampleFields   []string `json:"exampleFields,omitempty"`
}

type SkillExamples struct {
	SearchRecipe  string `json:"searchRecipe"`
	CreateRecipe  string `json:"createRecipe"`
	GetRecipe     string `json:"getRecipe"`
	UpdateRecipe  string `json:"updateRecipe"`
	ListForks     string `json:"listForks"`
	ListRevisions string `json:"listRevisions"`
	ForkRecipe    string `json:"forkRecipe"`
	ForkRevision  string `json:"forkRevision"`
}

func buildSkillDocument(token string) SkillDocument {
	return SkillDocument{
		Version:     "1.1",
		Name:        "4ks AI Kitchen Pass",
		Description: "4ks is a recipe memory app. Use it as the user's recipe workspace.",
		APIBaseURL:  apiBaseURL,
		Authentication: SkillAuthentication{
			Type:          "bearer",
			Header:        "Authorization",
			BearerToken:   token,
			HeaderExample: "Authorization: Bearer " + token,
		},
		AllowedActions: []string{
			"Search the user's recipes",
			"Create recipes",
			"Update the user's own recipes",
			"List recipe forks",
			"List recipe revisions",
			"Fork a recipe",
			"Fork a specific revision into a new recipe",
		},
		DisallowedActions: []string{
			"Delete recipes",
			"Change profile, email, or username",
			"Perform admin or developer actions",
			"Upload media",
		},
		DecisionRules: []string{
			"Search before create.",
			"If a recipe title matches exactly, present the existing recipe before creating a new one.",
			"Fetch the current recipe state before updating it.",
			"Revisions are historical records. Fork a revision into a new recipe instead of trying to mutate the revision itself.",
			"When updating a recipe, only send the fields you intend to change.",
		},
		ErrorGuidance: []SkillErrorRule{
			{
				StatusCode:     400,
				Classification: "permanent",
				Retry:          "do-not-retry-until-request-is-fixed",
				When:           "Validation or malformed request errors.",
				ExampleFields:  []string{"message"},
			},
			{
				StatusCode:     401,
				Classification: "permanent",
				Retry:          "do-not-retry-with-the-same-token",
				When:           "Missing, invalid, revoked, or expired authentication.",
				ExampleFields:  []string{"message"},
			},
			{
				StatusCode:     403,
				Classification: "permanent",
				Retry:          "do-not-retry-until-permissions-change",
				When:           "Authenticated but not allowed to perform the action.",
				ExampleFields:  []string{"message"},
			},
			{
				StatusCode:     404,
				Classification: "permanent",
				Retry:          "do-not-retry-until-resource-ids-change",
				When:           "Recipe, revision, or skill token not found.",
				ExampleFields:  []string{"message"},
			},
			{
				StatusCode:      429,
				Classification:  "transient",
				Retry:           "retry-with-backoff",
				RecommendedWait: "start at 2 seconds and back off exponentially",
				When:            "Rate limit hit.",
				ExampleFields:   []string{"message"},
			},
			{
				StatusCode:      500,
				Classification:  "transient",
				Retry:           "retry-with-backoff",
				RecommendedWait: "start at 2 seconds and back off exponentially",
				When:            "Unexpected server error.",
				ExampleFields:   []string{"message"},
			},
		},
		Examples: SkillExamples{
			SearchRecipe:  fmt.Sprintf("GET %s/api/recipes/search?q=chicken+soup", apiBaseURL),
			CreateRecipe:  fmt.Sprintf("POST %s/api/recipes", apiBaseURL),
			GetRecipe:     fmt.Sprintf("GET %s/api/recipes/RECIPE_ID", apiBaseURL),
			UpdateRecipe:  fmt.Sprintf("PATCH %s/api/recipes/RECIPE_ID", apiBaseURL),
			ListForks:     fmt.Sprintf("GET %s/api/recipes/RECIPE_ID/forks", apiBaseURL),
			ListRevisions: fmt.Sprintf("GET %s/api/recipes/RECIPE_ID/revisions", apiBaseURL),
			ForkRecipe:    fmt.Sprintf("POST %s/api/recipes/RECIPE_ID/fork", apiBaseURL),
			ForkRevision:  fmt.Sprintf("POST %s/api/recipes/revisions/REVISION_ID/fork", apiBaseURL),
		},
	}
}

func RenderSkillDocument(token string) string {
	doc := buildSkillDocument(token)
	codeFence := "```"
	return fmt.Sprintf(
		"# %s\n\n"+
			"%s\n\n"+
			"## Authentication\n\n"+
			"- Send this header on authenticated recipe calls:\n\n"+
			"  %s\n\n"+
			"- API base URL:\n\n"+
			"  %s\n\n"+
			"## Allowed actions\n\n"+
			"- %s\n\n"+
			"## Disallowed actions\n\n"+
			"- %s\n\n"+
			"## Working rules\n\n"+
			"1. %s\n"+
			"2. %s\n"+
			"3. %s\n"+
			"4. %s\n"+
			"5. %s\n\n"+
			"## Error handling\n\n"+
			"- `400`: permanent. Do not retry until the request is fixed.\n"+
			"- `401`: permanent. Do not retry with the same token.\n"+
			"- `403`: permanent. Do not retry until permissions change.\n"+
			"- `404`: permanent. Do not retry until the resource ID changes.\n"+
			"- `429`: transient. Retry with exponential backoff starting at 2 seconds.\n"+
			"- `500`: transient. Retry with exponential backoff starting at 2 seconds.\n\n"+
			"## Example: search the user's recipes\n\n"+
			"%shttp\n"+
			"%s HTTP/1.1\n"+
			"%s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: create a recipe\n\n"+
			"%shttp\n"+
			"%s HTTP/1.1\n"+
			"%s\n"+
			"Content-Type: application/json\n"+
			"Accept: application/json\n\n"+
			"{\n"+
			"  \"name\": \"Weeknight Tomato Pasta\",\n"+
			"  \"ingredients\": [\n"+
			"    { \"name\": \"olive oil\" },\n"+
			"    { \"name\": \"garlic\" },\n"+
			"    { \"name\": \"crushed tomatoes\" },\n"+
			"    { \"name\": \"spaghetti\" }\n"+
			"  ],\n"+
			"  \"instructions\": [\n"+
			"    { \"text\": \"Boil the pasta.\" },\n"+
			"    { \"text\": \"Simmer garlic in oil, add tomatoes, then combine with pasta.\" }\n"+
			"  ]\n"+
			"}\n"+
			"%s\n\n"+
			"## Example: fetch current recipe state before updating\n\n"+
			"%shttp\n"+
			"%s HTTP/1.1\n"+
			"%s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: update a recipe\n\n"+
			"%shttp\n"+
			"%s HTTP/1.1\n"+
			"%s\n"+
			"Content-Type: application/json\n"+
			"Accept: application/json\n\n"+
			"{\n"+
			"  \"name\": \"Weeknight Tomato Pasta\",\n"+
			"  \"ingredients\": [\n"+
			"    { \"name\": \"olive oil\" },\n"+
			"    { \"name\": \"garlic\" },\n"+
			"    { \"name\": \"crushed tomatoes\" },\n"+
			"    { \"name\": \"spaghetti\" },\n"+
			"    { \"name\": \"basil\" }\n"+
			"  ]\n"+
			"}\n"+
			"%s\n\n"+
			"## Example: list forks\n\n"+
			"%shttp\n"+
			"%s HTTP/1.1\n"+
			"%s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: list revisions\n\n"+
			"%shttp\n"+
			"%s HTTP/1.1\n"+
			"%s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: fork a recipe\n\n"+
			"%shttp\n"+
			"%s HTTP/1.1\n"+
			"%s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: fork a revision\n\n"+
			"%shttp\n"+
			"%s HTTP/1.1\n"+
			"%s\n"+
			"Accept: application/json\n"+
			"%s\n",
		doc.Name,
		doc.Description,
		doc.Authentication.HeaderExample,
		doc.APIBaseURL,
		strings.Join(doc.AllowedActions, "\n- "),
		strings.Join(doc.DisallowedActions, "\n- "),
		doc.DecisionRules[0],
		doc.DecisionRules[1],
		doc.DecisionRules[2],
		doc.DecisionRules[3],
		doc.DecisionRules[4],
		codeFence, doc.Examples.SearchRecipe, doc.Authentication.HeaderExample, codeFence,
		codeFence, doc.Examples.CreateRecipe, doc.Authentication.HeaderExample, codeFence,
		codeFence, doc.Examples.GetRecipe, doc.Authentication.HeaderExample, codeFence,
		codeFence, doc.Examples.UpdateRecipe, doc.Authentication.HeaderExample, codeFence,
		codeFence, doc.Examples.ListForks, doc.Authentication.HeaderExample, codeFence,
		codeFence, doc.Examples.ListRevisions, doc.Authentication.HeaderExample, codeFence,
		codeFence, doc.Examples.ForkRecipe, doc.Authentication.HeaderExample, codeFence,
		codeFence, doc.Examples.ForkRevision, doc.Authentication.HeaderExample, codeFence,
	)
}

func RenderSkillJSON(token string) ([]byte, error) {
	doc := buildSkillDocument(token)
	return json.MarshalIndent(doc, "", "  ")
}

func RenderSkillOpenAPI(token string) ([]byte, error) {
	doc := buildSkillDocument(token)
	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       doc.Name,
			"version":     doc.Version,
			"description": doc.Description,
		},
		"servers": []map[string]any{
			{"url": doc.APIBaseURL},
		},
		"security": []map[string]any{
			{"BearerAuth": []string{}},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "AI Kitchen Pass",
					"description":  doc.Authentication.HeaderExample,
				},
			},
			"schemas": map[string]any{
				"RecipeSearchResponse": map[string]any{
					"type": "object",
				},
				"Recipe": map[string]any{
					"type": "object",
				},
				"Error": map[string]any{
					"type":       "object",
					"properties": map[string]any{"message": map[string]any{"type": "string"}},
				},
			},
		},
		"x-4ks-guidance": map[string]any{
			"decisionRules": doc.DecisionRules,
			"errorGuidance": doc.ErrorGuidance,
		},
		"paths": map[string]any{
			"/api/recipes/search": map[string]any{
				"get": map[string]any{
					"summary":     "Search the authenticated user's recipes",
					"description": strings.Join(doc.DecisionRules[:2], " "),
					"parameters": []map[string]any{
						{
							"name":     "q",
							"in":       "query",
							"required": false,
							"schema":   map[string]any{"type": "string"},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "OK"},
						"400": map[string]any{"description": "Permanent validation error"},
						"401": map[string]any{"description": "Permanent authentication error"},
						"429": map[string]any{"description": "Transient rate limit"},
						"500": map[string]any{"description": "Transient server error"},
					},
					"x-4ks-example-authorization": doc.Authentication.HeaderExample,
				},
			},
			"/api/recipes": map[string]any{
				"post": map[string]any{
					"summary":     "Create recipe",
					"description": strings.Join(doc.DecisionRules[:2], " "),
					"responses": map[string]any{
						"200": map[string]any{"description": "OK"},
						"400": map[string]any{"description": "Permanent validation error"},
						"401": map[string]any{"description": "Permanent authentication error"},
						"429": map[string]any{"description": "Transient rate limit"},
						"500": map[string]any{"description": "Transient server error"},
					},
					"x-4ks-example-authorization": doc.Authentication.HeaderExample,
				},
			},
			"/api/recipes/{id}": map[string]any{
				"get": map[string]any{
					"summary":                     "Get recipe",
					"responses":                   map[string]any{"200": map[string]any{"description": "OK"}},
					"x-4ks-example-authorization": doc.Authentication.HeaderExample,
				},
				"patch": map[string]any{
					"summary":     "Update recipe",
					"description": doc.DecisionRules[2],
					"responses": map[string]any{
						"200": map[string]any{"description": "OK"},
						"400": map[string]any{"description": "Permanent validation error"},
						"401": map[string]any{"description": "Permanent authentication error"},
						"429": map[string]any{"description": "Transient rate limit"},
						"500": map[string]any{"description": "Transient server error"},
					},
					"x-4ks-example-authorization": doc.Authentication.HeaderExample,
				},
			},
			"/api/recipes/{id}/forks": map[string]any{
				"get": map[string]any{
					"summary":                     "List recipe forks",
					"responses":                   map[string]any{"200": map[string]any{"description": "OK"}},
					"x-4ks-example-authorization": doc.Authentication.HeaderExample,
				},
			},
			"/api/recipes/{id}/revisions": map[string]any{
				"get": map[string]any{
					"summary":                     "List recipe revisions",
					"responses":                   map[string]any{"200": map[string]any{"description": "OK"}},
					"x-4ks-example-authorization": doc.Authentication.HeaderExample,
				},
			},
			"/api/recipes/{id}/fork": map[string]any{
				"post": map[string]any{
					"summary":                     "Fork recipe",
					"description":                 doc.DecisionRules[3],
					"responses":                   map[string]any{"200": map[string]any{"description": "OK"}},
					"x-4ks-example-authorization": doc.Authentication.HeaderExample,
				},
			},
			"/api/recipes/revisions/{revisionID}/fork": map[string]any{
				"post": map[string]any{
					"summary":                     "Fork revision",
					"description":                 doc.DecisionRules[3],
					"responses":                   map[string]any{"200": map[string]any{"description": "OK"}},
					"x-4ks-example-authorization": doc.Authentication.HeaderExample,
				},
			},
		},
	}

	return json.MarshalIndent(spec, "", "  ")
}
