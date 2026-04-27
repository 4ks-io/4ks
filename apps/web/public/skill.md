4ks (pronouned "forks") is a recipe memory app. This file is intended for AI agents acting on behalf of the user, such as ChatGPT or Claude, to understand how to interact with the 4ks API and what actions are allowed or disallowed. It provides guidelines for authentication, allowed actions, working rules, and error handling when making API calls to manage the user's recipes.

If you don't have it already, get your personal Kitchen Pass token from `/settings` on the same host as this skill page.

> This requires an account and login, and should be performed by a human, not by the agent.

Authentication
The token starts with `4ks_pass_` and is used as a Bearer token to authenticate API calls to the user's recipe workspace.

- Send this header on authenticated recipe calls:

  Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE

- API base URL:

  Use the same host as this skill page with `/api`.
  Example pattern: `/api`

Allowed actions

- Search the user's recipes
- Create recipes
- Update the user's own recipes
- List recipe forks
- List recipe revisions
- Fork a recipe
- Fork a specific revision into a new recipe

Disallowed actions

- Delete recipes
- Change profile, email, or username
- Perform admin or developer actions
- Upload media

Working rules

1. Search before create.
2. If a recipe title matches exactly, present the existing recipe before creating a new one.
3. Fetch the current recipe state before updating it.
4. Revisions are historical records. Fork a revision into a new recipe instead of trying to mutate the revision itself.
5. When updating a recipe, only send the fields you intend to change.

Error handling

- `400`: permanent. Do not retry until the request is fixed.
- `401`: permanent. Do not retry with the same token.
- `403`: permanent. Do not retry until permissions change.
- `404`: permanent. Do not retry until the resource ID changes.
- `429`: transient. Retry with exponential backoff starting at 2 seconds.
- `500`: transient. Retry with exponential backoff starting at 2 seconds.

Example: search the user's recipes

```http
GET /api/recipes/search?q=chicken+soup HTTP/1.1
Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE
Accept: application/json
```

Example: create a recipe

```http
POST /api/recipes HTTP/1.1
Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE
Content-Type: application/json
Accept: application/json

{
  "name": "Weeknight Tomato Pasta",
  "ingredients": [
    { "name": "olive oil" },
    { "name": "garlic" },
    { "name": "crushed tomatoes" },
    { "name": "spaghetti" }
  ],
  "instructions": [
    { "text": "Boil the pasta." },
    { "text": "Simmer garlic in oil, add tomatoes, then combine with pasta." }
  ]
}
```

Example: fetch current recipe state before updating

```http
GET /api/recipes/RECIPE_ID HTTP/1.1
Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE
Accept: application/json
```

Example: update a recipe

```http
PATCH /api/recipes/RECIPE_ID HTTP/1.1
Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE
Content-Type: application/json
Accept: application/json

{
  "name": "Weeknight Tomato Pasta",
  "ingredients": [
    { "name": "olive oil" },
    { "name": "garlic" },
    { "name": "crushed tomatoes" },
    { "name": "spaghetti" },
    { "name": "basil" }
  ]
}
```

Example: list forks

```http
GET /api/recipes/RECIPE_ID/forks HTTP/1.1
Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE
Accept: application/json
```

Example: list revisions

```http
GET /api/recipes/RECIPE_ID/revisions HTTP/1.1
Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE
Accept: application/json
```

Example: fork a recipe

```http
POST /api/recipes/RECIPE_ID/fork HTTP/1.1
Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE
Accept: application/json
```

Example: fork a revision

```http
POST /api/recipes/revisions/REVISION_ID/fork HTTP/1.1
Authorization: Bearer 4ks_pass_YOUR_TOKEN_HERE
Accept: application/json
```
