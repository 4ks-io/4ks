package fetcher

import (
	"encoding/json"
	"testing"

	"github.com/gocolly/colly/v2"
)

func TestIsSupportedContentType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "html", input: "text/html", want: true},
		{name: "html with charset", input: "text/html; charset=utf-8", want: true},
		{name: "xhtml upper case", input: "APPLICATION/XHTML+XML", want: true},
		{name: "json", input: "application/json", want: false},
		{name: "empty", input: "", want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isSupportedContentType(tc.input); got != tc.want {
				t.Fatalf("isSupportedContentType(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestRemoveEmptyStrings(t *testing.T) {
	t.Parallel()

	got := removeEmptyStrings([]string{"salt", "", " ", "pepper"})
	if len(got) != 3 || got[0] != "salt" || got[1] != " " || got[2] != "pepper" {
		t.Fatalf("unexpected filtered strings: %#v", got)
	}
}

func TestGetInstructions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{
			name: "how to step list",
			input: []map[string]any{
				{"@type": "HowToStep", "text": "Mix &amp; stir"},
				{"@type": "HowToStep", "text": "Bake"},
			},
			want: []string{"Mix & stir", "Bake"},
		},
		{
			name: "how to section",
			input: []map[string]any{
				{
					"@type": "HowToSection",
					"itemListElement": []map[string]any{
						{"@type": "HowToStep", "text": "Prep"},
						{"@type": "HowToStep", "text": "Cook"},
					},
				},
			},
			want: []string{"Prep", "Cook"},
		},
		{
			name:  "numbered single string",
			input: []string{"1. Prep 2. Cook 3. Serve"},
			want:  []string{"Prep", "Cook", "Serve"},
		},
		{
			name:  "simple string list",
			input: []string{"Prep", "Cook"},
			want:  []string{"Prep", "Cook"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := getInstructions(tc.input)
			if err != nil {
				t.Fatalf("getInstructions returned error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("unexpected instruction count: got %v want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("instructions[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestGetInstructionsRejectsNonArray(t *testing.T) {
	t.Parallel()

	if _, err := getInstructions(map[string]any{"text": "Prep"}); err == nil {
		t.Fatal("expected non-array instructions to fail")
	}
}

func TestGetIngredients(t *testing.T) {
	t.Parallel()

	got, err := getIngredients([]string{"1 cup sugar", "", "Salt &amp; pepper"})
	if err != nil {
		t.Fatalf("getIngredients returned error: %v", err)
	}
	if len(got) != 2 || got[0] != "1 cup sugar" || got[1] != "Salt & pepper" {
		t.Fatalf("unexpected ingredients: %#v", got)
	}
}

func TestSearchJSON(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"@graph": []any{
			map[string]any{"@type": "WebPage", "name": "page"},
			map[string]any{
				"@type": []any{"Thing", "Recipe"},
				"name":  "Soup",
			},
		},
	}

	got := searchJSON(data)
	if got == nil || got["name"] != "Soup" {
		t.Fatalf("expected recipe node, got %#v", got)
	}
	if searchJSON(map[string]any{"@type": "WebPage"}) != nil {
		t.Fatal("did not expect non-recipe data to match")
	}
}

func TestGetRecipeFromJSONLD(t *testing.T) {
	t.Parallel()

	node := map[string]any{
		"@type": "Recipe",
		"name":  "Soup",
		"recipeIngredient": []string{
			"1 cup broth",
			"Salt &amp; pepper",
		},
		"recipeInstructions": []map[string]any{
			{"@type": "HowToStep", "text": "Mix"},
			{"@type": "HowToStep", "text": "Serve"},
		},
	}

	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	elem := &colly.HTMLElement{Text: string(raw)}
	recipe, err := getRecipeFromJSONLD(elem, "https://example.com/recipe")
	if err != nil {
		t.Fatalf("getRecipeFromJSONLD returned error: %v", err)
	}
	if recipe.Title != "Soup" || recipe.SourceURL != "https://example.com/recipe" {
		t.Fatalf("unexpected recipe metadata: %+v", recipe)
	}
	if len(recipe.Ingredients) != 2 || len(recipe.Instructions) != 2 {
		t.Fatalf("unexpected recipe body: %+v", recipe)
	}
}

func TestCreateRecipeDtoFromRecipe(t *testing.T) {
	t.Parallel()

	dto := createRecipeDtoFromRecipe(Recipe{
		SourceURL:    "https://example.com/recipe",
		Title:        "Soup",
		Ingredients:  []string{"broth", "salt"},
		Instructions: []string{"mix", "serve"},
	})

	if dto.Name != "Soup" || dto.Link != "https://example.com/recipe" {
		t.Fatalf("unexpected dto metadata: %+v", dto)
	}
	if len(dto.Ingredients) != 2 || dto.Ingredients[0].Name != "broth" {
		t.Fatalf("unexpected dto ingredients: %+v", dto.Ingredients)
	}
	if len(dto.Instructions) != 2 || dto.Instructions[1].Text != "serve" {
		t.Fatalf("unexpected dto instructions: %+v", dto.Instructions)
	}
}
