// Package app contains process-level wiring types shared by api sub-servers.
package app

import (
	fetcherService "4ks/apps/api/services/fetcher"
	kitchenPassService "4ks/apps/api/services/kitchenpass"
	recipeService "4ks/apps/api/services/recipe"
	searchService "4ks/apps/api/services/search"
	staticService "4ks/apps/api/services/static"
	userService "4ks/apps/api/services/user"
)

// Services is the shared service bundle passed from main to api sub-servers.
type Services struct {
	User        userService.Service
	Recipe      recipeService.Service
	Search      searchService.Service
	Static      staticService.Service
	Fetcher     fetcherService.Service
	KitchenPass kitchenPassService.Service
}
