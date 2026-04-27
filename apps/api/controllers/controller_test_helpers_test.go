package controllers

import (
	"4ks/apps/api/dtos"
	fetcherService "4ks/apps/api/services/fetcher"
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	recipeService "4ks/apps/api/services/recipe"
	searchService "4ks/apps/api/services/search"
	staticService "4ks/apps/api/services/static"
	userService "4ks/apps/api/services/user"
	"4ks/apps/api/utils"
	models "4ks/libs/go/models"
	"context"
	"sync"

	"github.com/google/uuid"
)

type stubUserService struct {
	getAllUsersFn                  func(context.Context) ([]*models.User, error)
	getUserByIDFn                  func(context.Context, string) (*models.User, error)
	getUserByUsernameFn            func(context.Context, string) (*models.User, error)
	getUserByEmailFn               func(context.Context, string) (*models.User, error)
	createUserFn                   func(context.Context, string, string, *dtos.CreateUser) (*models.User, error)
	updateUserByIDFn               func(context.Context, string, *dtos.UpdateUser) (*models.User, error)
	deleteUserFn                   func(context.Context, string) error
	createUserEventByUserIDFn      func(context.Context, string, *dtos.CreateUserEvent) (*models.UserEvent, error)
	updateUserEventByUserIDEventFn func(context.Context, string, *dtos.UpdateUserEvent) (*models.UserEvent, error)
	removeUserEventFn              func(context.Context, string, uuid.UUID) error
	testNameFn                     func(context.Context, string) error
	testValidNameFn                func(string) bool
	testReservedWordFn             func(string) bool
	testAvailableNameFn            func(context.Context, string) (bool, error)
}

type stubKitchenPassService struct {
	getStatusFn      func(context.Context, string) (*dtos.KitchenPassResponse, error)
	createOrRotateFn func(context.Context, string) (*dtos.KitchenPassResponse, error)
	revokeFn         func(context.Context, string) error
	validateTokenFn  func(context.Context, string) (*models.PersonalAccessToken, error)
	recordUsageFn    func(context.Context, string, string) error
}

func (s stubUserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	if s.getAllUsersFn != nil {
		return s.getAllUsersFn(ctx)
	}
	return nil, nil
}

func (s stubUserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if s.getUserByIDFn != nil {
		return s.getUserByIDFn(ctx, id)
	}
	return nil, nil
}

func (s stubUserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if s.getUserByUsernameFn != nil {
		return s.getUserByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (s stubUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if s.getUserByEmailFn != nil {
		return s.getUserByEmailFn(ctx, email)
	}
	return nil, nil
}

func (s stubUserService) CreateUser(ctx context.Context, userID string, userEmail string, payload *dtos.CreateUser) (*models.User, error) {
	if s.createUserFn != nil {
		return s.createUserFn(ctx, userID, userEmail, payload)
	}
	return nil, nil
}

func (s stubUserService) UpdateUserByID(ctx context.Context, userID string, payload *dtos.UpdateUser) (*models.User, error) {
	if s.updateUserByIDFn != nil {
		return s.updateUserByIDFn(ctx, userID, payload)
	}
	return nil, nil
}

func (s stubUserService) DeleteUser(ctx context.Context, userID string) error {
	if s.deleteUserFn != nil {
		return s.deleteUserFn(ctx, userID)
	}
	return nil
}

func (s stubUserService) CreateUserEventByUserID(ctx context.Context, id string, payload *dtos.CreateUserEvent) (*models.UserEvent, error) {
	if s.createUserEventByUserIDFn != nil {
		return s.createUserEventByUserIDFn(ctx, id, payload)
	}
	return nil, nil
}

func (s stubUserService) UpdateUserEventByUserIDEventID(ctx context.Context, id string, payload *dtos.UpdateUserEvent) (*models.UserEvent, error) {
	if s.updateUserEventByUserIDEventFn != nil {
		return s.updateUserEventByUserIDEventFn(ctx, id, payload)
	}
	return nil, nil
}

func (s stubUserService) RemoveUserEventByUserIDEventID(ctx context.Context, id string, eventID uuid.UUID) error {
	if s.removeUserEventFn != nil {
		return s.removeUserEventFn(ctx, id, eventID)
	}
	return nil
}

func (s stubUserService) TestName(ctx context.Context, name string) error {
	if s.testNameFn != nil {
		return s.testNameFn(ctx, name)
	}
	return nil
}

func (s stubUserService) TestValidName(name string) bool {
	if s.testValidNameFn != nil {
		return s.testValidNameFn(name)
	}
	return false
}

func (s stubUserService) TestReservedWord(name string) bool {
	if s.testReservedWordFn != nil {
		return s.testReservedWordFn(name)
	}
	return false
}

func (s stubUserService) TestAvailableName(ctx context.Context, name string) (bool, error) {
	if s.testAvailableNameFn != nil {
		return s.testAvailableNameFn(ctx, name)
	}
	return false, nil
}

func (s stubKitchenPassService) GetStatus(ctx context.Context, userID string) (*dtos.KitchenPassResponse, error) {
	if s.getStatusFn != nil {
		return s.getStatusFn(ctx, userID)
	}
	return nil, nil
}

func (s stubKitchenPassService) CreateOrRotate(ctx context.Context, userID string) (*dtos.KitchenPassResponse, error) {
	if s.createOrRotateFn != nil {
		return s.createOrRotateFn(ctx, userID)
	}
	return nil, nil
}

func (s stubKitchenPassService) Revoke(ctx context.Context, userID string) error {
	if s.revokeFn != nil {
		return s.revokeFn(ctx, userID)
	}
	return nil
}

func (s stubKitchenPassService) ValidateToken(ctx context.Context, token string) (*models.PersonalAccessToken, error) {
	if s.validateTokenFn != nil {
		return s.validateTokenFn(ctx, token)
	}
	return nil, kitchenpasssvc.ErrKitchenPassNotFound
}

func (s stubKitchenPassService) RecordUsage(ctx context.Context, tokenDigest string, action string) error {
	if s.recordUsageFn != nil {
		return s.recordUsageFn(ctx, tokenDigest, action)
	}
	return nil
}

type stubRecipeService struct {
	createRecipeFn               func(context.Context, *dtos.CreateRecipe) (*models.Recipe, error)
	createRecipeMediaFn          func(context.Context, *utils.MediaProps, string, string, *sync.WaitGroup) (*models.RecipeMedia, error)
	createRecipeMediaSignedURLFn func(context.Context, *utils.MediaProps, *sync.WaitGroup) (string, error)
	deleteRecipeFn               func(context.Context, string, string) error
	getAdminRecipeMediasFn       func(context.Context, string) ([]*models.RecipeMedia, error)
	getRecipesFn                 func(context.Context, int) ([]*models.Recipe, error)
	getRecipeByIDFn              func(context.Context, string) (*models.Recipe, error)
	getRecipesByUsernameFn       func(context.Context, string, int) ([]*models.Recipe, error)
	getRecipesByUserIDFn         func(context.Context, string, int) ([]*models.Recipe, error)
	getRecipeMediaFn             func(context.Context, string) ([]*models.RecipeMedia, error)
	getRecipeForksFn             func(context.Context, string) ([]*models.Recipe, error)
	getRecipeRevisionsFn         func(context.Context, string) ([]*models.RecipeRevision, error)
	getRecipeRevisionByIDFn      func(context.Context, string) (*models.RecipeRevision, error)
	forkRecipeByIDFn             func(context.Context, string, models.UserSummary) (*models.Recipe, error)
	forkRecipeByRevisionIDFn     func(context.Context, string, models.UserSummary) (*models.Recipe, error)
	starRecipeByIDFn             func(context.Context, string, models.UserSummary) (bool, error)
	updateRecipeByIDFn           func(context.Context, string, *dtos.UpdateRecipe) (*models.Recipe, error)
	createMockBannerFn           func(string, string) []models.RecipeMediaVariant
}

func (s stubRecipeService) CreateRecipe(ctx context.Context, payload *dtos.CreateRecipe) (*models.Recipe, error) {
	if s.createRecipeFn != nil {
		return s.createRecipeFn(ctx, payload)
	}
	return nil, nil
}

func (s stubRecipeService) CreateRecipeMedia(ctx context.Context, mp *utils.MediaProps, recipeID string, userID string, wg *sync.WaitGroup) (*models.RecipeMedia, error) {
	if s.createRecipeMediaFn != nil {
		return s.createRecipeMediaFn(ctx, mp, recipeID, userID, wg)
	}
	return nil, nil
}

func (s stubRecipeService) CreateRecipeMediaSignedURL(ctx context.Context, mp *utils.MediaProps, wg *sync.WaitGroup) (string, error) {
	if s.createRecipeMediaSignedURLFn != nil {
		return s.createRecipeMediaSignedURLFn(ctx, mp, wg)
	}
	return "", nil
}

func (s stubRecipeService) DeleteRecipe(ctx context.Context, recipeID string, userID string) error {
	if s.deleteRecipeFn != nil {
		return s.deleteRecipeFn(ctx, recipeID, userID)
	}
	return nil
}

func (s stubRecipeService) GetAdminRecipeMedias(ctx context.Context, recipeID string) ([]*models.RecipeMedia, error) {
	if s.getAdminRecipeMediasFn != nil {
		return s.getAdminRecipeMediasFn(ctx, recipeID)
	}
	return nil, nil
}

func (s stubRecipeService) GetRecipes(ctx context.Context, limit int) ([]*models.Recipe, error) {
	if s.getRecipesFn != nil {
		return s.getRecipesFn(ctx, limit)
	}
	return nil, nil
}

func (s stubRecipeService) GetRecipeByID(ctx context.Context, recipeID string) (*models.Recipe, error) {
	if s.getRecipeByIDFn != nil {
		return s.getRecipeByIDFn(ctx, recipeID)
	}
	return nil, nil
}

func (s stubRecipeService) GetRecipesByUsername(ctx context.Context, username string, limit int) ([]*models.Recipe, error) {
	if s.getRecipesByUsernameFn != nil {
		return s.getRecipesByUsernameFn(ctx, username, limit)
	}
	return nil, nil
}

func (s stubRecipeService) GetRecipesByUserID(ctx context.Context, userID string, limit int) ([]*models.Recipe, error) {
	if s.getRecipesByUserIDFn != nil {
		return s.getRecipesByUserIDFn(ctx, userID, limit)
	}
	return nil, nil
}

func (s stubRecipeService) GetRecipeMedia(ctx context.Context, recipeID string) ([]*models.RecipeMedia, error) {
	if s.getRecipeMediaFn != nil {
		return s.getRecipeMediaFn(ctx, recipeID)
	}
	return nil, nil
}

func (s stubRecipeService) GetRecipeForks(ctx context.Context, recipeID string) ([]*models.Recipe, error) {
	if s.getRecipeForksFn != nil {
		return s.getRecipeForksFn(ctx, recipeID)
	}
	return nil, nil
}

func (s stubRecipeService) GetRecipeRevisions(ctx context.Context, recipeID string) ([]*models.RecipeRevision, error) {
	if s.getRecipeRevisionsFn != nil {
		return s.getRecipeRevisionsFn(ctx, recipeID)
	}
	return nil, nil
}

func (s stubRecipeService) GetRecipeRevisionByID(ctx context.Context, revisionID string) (*models.RecipeRevision, error) {
	if s.getRecipeRevisionByIDFn != nil {
		return s.getRecipeRevisionByIDFn(ctx, revisionID)
	}
	return nil, nil
}

func (s stubRecipeService) ForkRecipeByID(ctx context.Context, recipeID string, author models.UserSummary) (*models.Recipe, error) {
	if s.forkRecipeByIDFn != nil {
		return s.forkRecipeByIDFn(ctx, recipeID, author)
	}
	return nil, nil
}

func (s stubRecipeService) ForkRecipeByRevisionID(ctx context.Context, revisionID string, author models.UserSummary) (*models.Recipe, error) {
	if s.forkRecipeByRevisionIDFn != nil {
		return s.forkRecipeByRevisionIDFn(ctx, revisionID, author)
	}
	return nil, nil
}

func (s stubRecipeService) StarRecipeByID(ctx context.Context, recipeID string, author models.UserSummary) (bool, error) {
	if s.starRecipeByIDFn != nil {
		return s.starRecipeByIDFn(ctx, recipeID, author)
	}
	return false, nil
}

func (s stubRecipeService) UpdateRecipeByID(ctx context.Context, recipeID string, payload *dtos.UpdateRecipe) (*models.Recipe, error) {
	if s.updateRecipeByIDFn != nil {
		return s.updateRecipeByIDFn(ctx, recipeID, payload)
	}
	return nil, nil
}

func (s stubRecipeService) CreateMockBanner(filename string, url string) []models.RecipeMediaVariant {
	if s.createMockBannerFn != nil {
		return s.createMockBannerFn(filename, url)
	}
	return nil
}

type stubSearchService struct {
	createSearchRecipeCollectionFn func() error
	removeSearchRecipeDocumentFn   func(string) error
	searchRecipesByAuthorFn        func(string, string, int) ([]*dtos.CreateSearchRecipe, error)
	upsertSearchRecipeDocumentFn   func(*models.Recipe) error
}

func (s stubSearchService) CreateSearchRecipeCollection() error {
	if s.createSearchRecipeCollectionFn != nil {
		return s.createSearchRecipeCollectionFn()
	}
	return nil
}

func (s stubSearchService) RemoveSearchRecipeDocument(id string) error {
	if s.removeSearchRecipeDocumentFn != nil {
		return s.removeSearchRecipeDocumentFn(id)
	}
	return nil
}

func (s stubSearchService) SearchRecipesByAuthor(query string, author string, perPage int) ([]*dtos.CreateSearchRecipe, error) {
	if s.searchRecipesByAuthorFn != nil {
		return s.searchRecipesByAuthorFn(query, author, perPage)
	}
	return nil, nil
}

func (s stubSearchService) UpsertSearchRecipeDocument(recipe *models.Recipe) error {
	if s.upsertSearchRecipeDocumentFn != nil {
		return s.upsertSearchRecipeDocumentFn(recipe)
	}
	return nil
}

type stubStaticService struct {
	getRandomFallbackImageURLFn func(string) string
	getRandomFallbackImageFn    func(context.Context) (string, error)
}

func (s stubStaticService) GetRandomFallbackImageURL(filename string) string {
	if s.getRandomFallbackImageURLFn != nil {
		return s.getRandomFallbackImageURLFn(filename)
	}
	return ""
}

func (s stubStaticService) GetRandomFallbackImage(ctx context.Context) (string, error) {
	if s.getRandomFallbackImageFn != nil {
		return s.getRandomFallbackImageFn(ctx)
	}
	return "", nil
}

type stubFetcherService struct {
	sendFn func(context.Context, *models.FetcherRequest) (string, error)
}

func (s stubFetcherService) Send(ctx context.Context, req *models.FetcherRequest) (string, error) {
	if s.sendFn != nil {
		return s.sendFn(ctx, req)
	}
	return "", nil
}

var (
	_ userService.Service    = stubUserService{}
	_ recipeService.Service  = stubRecipeService{}
	_ searchService.Service  = stubSearchService{}
	_ staticService.Service  = stubStaticService{}
	_ fetcherService.Service = stubFetcherService{}
)
