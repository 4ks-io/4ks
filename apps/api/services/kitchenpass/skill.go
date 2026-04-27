package kitchenpasssvc

import "fmt"

const apiBaseURL = "https://api.4ks.io"

func RenderSkillDocument(token string) string {
	codeFence := "```"
	return fmt.Sprintf(
		"# 4ks AI Kitchen Pass\n\n"+
			"4ks is a recipe memory app. Use it as the user's recipe workspace.\n\n"+
			"## Authentication\n\n"+
			"- Send this header on authenticated recipe calls:\n\n"+
			"  Authorization: Bearer %s\n\n"+
			"- API base URL:\n\n"+
			"  %s\n\n"+
			"## Allowed actions\n\n"+
			"- Search the user's recipes\n"+
			"- Create recipes\n"+
			"- Update the user's own recipes\n"+
			"- List recipe forks\n"+
			"- List recipe revisions\n"+
			"- Fork a recipe\n"+
			"- Fork a specific revision into a new recipe\n\n"+
			"## Disallowed actions\n\n"+
			"- Delete recipes\n"+
			"- Change profile, email, or username\n"+
			"- Perform admin or developer actions\n"+
			"- Upload media\n\n"+
			"## Working rules\n\n"+
			"1. Search before creating likely duplicates.\n"+
			"2. Fetch the current recipe state before updating it.\n"+
			"3. Revisions are historical records. Fork a revision into a new recipe instead of trying to mutate the revision itself.\n"+
			"4. When updating a recipe, only send the fields you intend to change.\n\n"+
			"## Example: search the user's recipes\n\n"+
			"%shttp\n"+
			"GET %s/api/recipes/search?q=chicken+soup HTTP/1.1\n"+
			"Authorization: Bearer %s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: create a recipe\n\n"+
			"%shttp\n"+
			"POST %s/api/recipes HTTP/1.1\n"+
			"Authorization: Bearer %s\n"+
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
			"GET %s/api/recipes/RECIPE_ID HTTP/1.1\n"+
			"Authorization: Bearer %s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: update a recipe\n\n"+
			"%shttp\n"+
			"PATCH %s/api/recipes/RECIPE_ID HTTP/1.1\n"+
			"Authorization: Bearer %s\n"+
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
			"GET %s/api/recipes/RECIPE_ID/forks HTTP/1.1\n"+
			"Authorization: Bearer %s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: list revisions\n\n"+
			"%shttp\n"+
			"GET %s/api/recipes/RECIPE_ID/revisions HTTP/1.1\n"+
			"Authorization: Bearer %s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: fork a recipe\n\n"+
			"%shttp\n"+
			"POST %s/api/recipes/RECIPE_ID/fork HTTP/1.1\n"+
			"Authorization: Bearer %s\n"+
			"Accept: application/json\n"+
			"%s\n\n"+
			"## Example: fork a revision\n\n"+
			"%shttp\n"+
			"POST %s/api/recipes/revisions/REVISION_ID/fork HTTP/1.1\n"+
			"Authorization: Bearer %s\n"+
			"Accept: application/json\n"+
			"%s\n",
		token,
		apiBaseURL,
		codeFence, apiBaseURL, token, codeFence,
		codeFence, apiBaseURL, token, codeFence,
		codeFence, apiBaseURL, token, codeFence,
		codeFence, apiBaseURL, token, codeFence,
		codeFence, apiBaseURL, token, codeFence,
		codeFence, apiBaseURL, token, codeFence,
		codeFence, apiBaseURL, token, codeFence,
		codeFence, apiBaseURL, token, codeFence,
	)
}
