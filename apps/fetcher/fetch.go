package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gocolly/colly/v2"
)

var (
	errNotRecipe               = errors.New("not a recipe")
	errUnsupportedContentType  = errors.New("unsupported content type")
	errUnexpectedRedirectCount = errors.New("too many redirects")
)

const (
	// ApplicationJSONLDScriptTag is the selector for json-ld script tags
	ApplicationJSONLDScriptTag = `script[type="application/ld+json"]`

	fetchDeadline       = 20 * time.Second
	maxRedirects        = 5
	maxResponseBodySize = 2 * 1024 * 1024
)

var supportedContentTypes = []string{
	"text/html",
	"application/xhtml+xml",
}

func collyOnRequest() func(r *colly.Request) {
	return func(r *colly.Request) {
		log.Info().Str("target", r.URL.String()).Msg("visiting recipe url")
	}
}

func visit(ctx context.Context, debug bool, target string) (Recipe, error) {
	if target == "" {
		err := errors.New("target is required")
		log.Error().Err(err).Caller().Msg("missing target host")
		return Recipe{}, err
	}

	deadlineCtx, cancel := context.WithTimeout(ctx, fetchDeadline)
	defer cancel()

	validatedURL, err := validateFetchURL(deadlineCtx, target)
	if err != nil {
		log.Error().Err(err).Caller().Msg("failed to validate target host")
		return Recipe{}, err
	}

	c, err := initCollector(deadlineCtx, validatedURL, debug)
	if err != nil {
		log.Error().Err(err).Caller().Msg("failed to initialize collector")
		return Recipe{}, err
	}
	c.MaxBodySize = maxResponseBodySize
	var recipe Recipe

	c.OnRequest(collyOnRequest())
	c.OnResponseHeaders(func(r *colly.Response) {
		if !isSupportedContentType(r.Headers.Get("Content-Type")) {
			r.Request.Abort()
			err = fmt.Errorf("%w: %s", errUnsupportedContentType, r.Headers.Get("Content-Type"))
		}
	})

	c.OnHTML(ApplicationJSONLDScriptTag, func(e *colly.HTMLElement) {
		if recipe.Title != "" {
			return
		}
		recipe, err = getRecipeFromJSONLD(e, validatedURL.Normalized)
	})

	if visitErr := c.Visit(validatedURL.Normalized); visitErr != nil {
		log.Error().Err(visitErr).Caller().Msg("failed to visit")
		return recipe, visitErr
	}
	if err != nil {
		return recipe, err
	}

	return recipe, nil
}

func isSupportedContentType(value string) bool {
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	for _, allowed := range supportedContentTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func getRecipeFromJSONLD(e *colly.HTMLElement, u string) (Recipe, error) {
	// json scrubbing
	jsonData := strings.TrimSpace(e.Text)
	jsonData = strings.ReplaceAll(jsonData, "\n", "")

	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return Recipe{}, err
	}

	node := searchJSON(data)
	if node == nil {
		return Recipe{}, errNotRecipe
	}

	ingredients, err := getIngredients(node["recipeIngredient"])
	if err != nil {
		log.Error().Err(err).Caller().Msg("failed to read recipe ingredients")

	}

	var instructions []string
	if node["recipeInstructions"] != nil {
		instructions, err = getInstructions(node["recipeInstructions"])
		if err != nil {
			log.Error().Err(err).Caller().Msg("failed to read recipe instructions")
			// tr@ck: log this url/html
		}
	}

	return Recipe{
		SourceURL:    u,
		Title:        node["name"].(string),
		Ingredients:  ingredients,
		Instructions: instructions,
	}, nil
}

func removeEmptyStrings(s []string) []string {
	re := regexp.MustCompile(`^\s*$`)
	filteredLines := []string{}
	for _, line := range s {
		if line != "" || !re.MatchString(line) {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}

// getInstructions returns a slice of strings from a slice of interface{}
func getInstructions(in interface{}) ([]string, error) {
	var o []string

	// marshal
	jsonData, err := json.Marshal(in)
	if err != nil {
		// tr@ck: log this url/html
		return o, err
	}

	// unmarsal -- confirm is array
	var data []interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		// error unmarshaling -- expected array
		// tr@ck: log this url/html
		return o, err
	}

	// handle known instruction shapes

	// [{
	//   "@type": "HowToStep"
	//   "text": "foo",
	// }]
	// -- OR --
	// [{
	// 	"@type": "HowToStep",
	// 	"name": "foo",
	// 	"text": "foo",
	// },
	var howtostep []HowToStep
	if err = json.Unmarshal(jsonData, &howtostep); err == nil { // inverted check
		if howtostep[0].Type == "HowToStep" {
			for _, v := range howtostep {
				o = append(o, html.UnescapeString(v.Text))
			}
			return removeEmptyStrings(o), nil
		}
	}

	// [{
	//   "@type": "HowToSection",
	//   "name": "bar",
	//   "itemListElement": [
	//     {
	//       "@type": "HowToStep",
	//       "text": "foo."
	//     },
	var howtosection []HowToSection
	if err = json.Unmarshal(jsonData, &howtosection); err == nil { // inverted check
		if howtosection[0].Type == "HowToSection" {
			for _, v := range howtosection[0].ItemList {
				o = append(o, html.UnescapeString(v.Text))
			}
		}
		return removeEmptyStrings(o), nil
	}

	// [ "1. foo 2. bar" ] || [ "foo", "bar" ]
	var simplestring []string
	if err = json.Unmarshal(jsonData, &simplestring); err == nil { // inverted check
		// handle very stupid case where instructions are possibly concatenated into a single string
		if len(simplestring) == 1 && len(simplestring[0]) >= 2 && simplestring[0][:2] == "1." {
			re := regexp.MustCompile(`\d+\. `)

			// Split the text into lines
			o = re.Split(simplestring[0], -1)

			// Trim the leading and trailing spaces from each line
			for i, line := range o {
				s := html.UnescapeString(line)
				o[i] = strings.TrimSpace(s)
			}

			return removeEmptyStrings(o), nil
		}

		return simplestring, nil
	}

	return o, nil
}

// getIngredients returns a slice of strings from a slice of interface{}
func getIngredients(in interface{}) ([]string, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return []string{}, err
	}

	var o []string
	err = json.Unmarshal(data, &o)
	if err != nil {
		return []string{}, err
	}

	for i, v := range o {
		o[i] = html.UnescapeString(v)
	}

	return removeEmptyStrings(o), nil
}

// Recipe is a struct to hold the scraped recipe data
type Recipe struct {
	SourceURL    string
	Title        string
	Ingredients  []string
	Instructions []string
}

// HowToSection is a struct to hold the scraped recipe data
type HowToSection struct {
	Type     string      `json:"@type"`
	ItemList []HowToStep `json:"itemListElement"`
}

// HowToStep is a struct to hold the scraped recipe data
type HowToStep struct {
	Type string `json:"@type"`
	Text string `json:"text"`
}

type searchStack struct {
	value  interface{}
	parent map[string]interface{}
}

// searchJSON iteratively searches a map[string]interface{} with a @type key with a value of "Recipe"
func searchJSON(data interface{}) map[string]interface{} {
	stack := []searchStack{{value: data}}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		switch t := current.value.(type) {
		case map[string]interface{}:
			for key, value := range t {
				if key == "@type" {
					switch tt := value.(type) {
					case string:
						if tt == "Recipe" {
							return t
						}
					case []interface{}:
						for _, item := range tt {
							if item == "Recipe" {
								return t
							}
						}
					default:
						// fmt.Printf("Unsupported type: %T\n", data)
					}
				}
				if key == "@type" && value == "Recipe" {
					return t
				}

				stack = append(stack, searchStack{value: value, parent: t})
			}
		case []interface{}:
			for _, item := range t {
				stack = append(stack, searchStack{value: item, parent: nil})
			}
		default:
			// fmt.Printf("Unsupported type: %T\n", data)
		}
	}

	return nil
}
