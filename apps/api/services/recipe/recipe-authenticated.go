package recipesvc

import (
	"context"
	"time"

	firestore "cloud.google.com/go/firestore"
	mp "github.com/geraldo-labs/merge-struct"

	"4ks/apps/api/dtos"
	"4ks/apps/api/middleware"
	models "4ks/libs/go/models"
)

func (s recipeService) DeleteRecipe(ctx context.Context, id string, sub string) error {
	result, _ := s.recipeCollection.Doc(id).Get(ctx)
	if !result.Exists() {
		return ErrRecipeNotFound
	}

	recipe := new(models.Recipe)
	if err := result.DataTo(recipe); err != nil {
		return err
	}

	// is author
	c, _ := middleware.EnforceContributor(sub, recipe.Contributors)
	// is admin
	a, _ := middleware.Enforce(sub, "/recipes/*", "delete")
	if !c && !a {
		return ErrUnauthorized
	}

	_, err := s.recipeCollection.Doc(id).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s recipeService) CreateRecipe(ctx context.Context, recipe *dtos.CreateRecipe) (*models.Recipe, error) {
	newRecipeDoc := s.recipeCollection.NewDoc()
	newRevisionDoc := s.recipeRevisionsCollection.NewDoc()

	recipeCreatedDate := time.Now().UTC()

	recipeRevision := &models.RecipeRevision{
		ID:           newRevisionDoc.ID,
		Name:         recipe.Name,
		Link:         recipe.Link,
		RecipeID:     newRecipeDoc.ID,
		Author:       recipe.Author,
		Instructions: recipe.Instructions,
		Ingredients:  recipe.Ingredients,
		Banner:       recipe.Banner,
		CreatedDate:  recipeCreatedDate,
		UpdatedDate:  recipeCreatedDate,
	}
	newRecipe := &models.Recipe{
		ID:           newRecipeDoc.ID,
		Root:         newRecipeDoc.ID,
		Author:       recipe.Author,
		Contributors: []models.UserSummary{recipe.Author},
		Metadata: models.RecipeMetadata{
			Stars: 0,
			Forks: 0,
		},
		CurrentRevision: *recipeRevision,
		CreatedDate:     recipeCreatedDate,
		UpdatedDate:     recipeCreatedDate,
	}

	// tr@ck
	_, err := s.fire.Batch().Create(newRevisionDoc, recipeRevision).Create(newRecipeDoc, newRecipe).Commit(ctx)

	if err != nil {
		return nil, ErrUnableToCreateRecipe
	}

	return newRecipe, nil
}

func (s recipeService) UpdateRecipeByID(ctx context.Context, recipeID string, recipeUpdate *dtos.UpdateRecipe) (*models.Recipe, error) {
	recipeDoc, err := s.recipeCollection.Doc(recipeID).Get(ctx)

	if err != nil {
		return nil, ErrRecipeNotFound
	}

	recipe := new(models.Recipe)
	recipeDoc.DataTo(recipe)

	// e, err := middleware.EnforceAuthor(&recipeUpdate.Author.ID, &recipe.Author)
	e, err := middleware.EnforceContributor(recipeUpdate.Author.ID, recipe.Contributors)
	if err != nil {
		return nil, ErrUnableToUpdateRecipe
	} else if !e {
		return nil, ErrUnauthorized
	}

	recipeUpdatedDate := time.Now().UTC()
	newRevisionDocRef := s.recipeRevisionsCollection.NewDoc()
	newRevision := &models.RecipeRevision{}

	// Copy the existing revision data into the new revision
	// Set the id of the new revision
	mp.Struct(newRevision, recipe.CurrentRevision)
	newRevision.ID = newRevisionDocRef.ID

	// Apply the new revision updates
	mp.Struct(newRevision, recipeUpdate)

	newRevision.CreatedDate = recipeUpdatedDate
	newRevision.UpdatedDate = recipeUpdatedDate

	recipe.CurrentRevision = *newRevision
	recipe.UpdatedDate = recipeUpdatedDate

	_, err = s.fire.Batch().Create(newRevisionDocRef, newRevision).Set(recipeDoc.Ref, recipe).Commit(ctx)

	if err != nil {
		return nil, ErrUnableToUpdateRecipe
	}

	return recipe, nil
}

func (s recipeService) ForkRecipeByID(ctx context.Context, recipeID string, forkAuthor models.UserSummary) (*models.Recipe, error) {
	recipeDoc, err := s.recipeCollection.Doc(recipeID).Get(ctx)

	if err != nil {
		return nil, ErrRecipeNotFound
	}

	recipe := new(models.Recipe)
	recipeDoc.DataTo(recipe)

	newRecipeDocRef := s.recipeCollection.NewDoc()
	newRevisionDocRef := s.recipeRevisionsCollection.NewDoc()

	recipe.Branch = recipe.ID
	recipe.ID = newRecipeDocRef.ID

	recipe.Author = forkAuthor
	recipe.Contributors = []models.UserSummary{forkAuthor}
	recipe.CurrentRevision.Author = forkAuthor
	recipe.CurrentRevision.ID = newRevisionDocRef.ID
	recipe.CurrentRevision.RecipeID = newRecipeDocRef.ID
	recipe.Metadata.Forks = 0
	recipe.Metadata.Stars = 0

	_, err = s.fire.Batch().Create(newRevisionDocRef, recipe.CurrentRevision).Create(newRecipeDocRef, recipe).Update(recipeDoc.Ref, []firestore.Update{
		{Path: "metadata.forks", Value: firestore.Increment(1)},
	}).Commit(ctx)

	if err != nil {
		return nil, ErrUnableToForkRecipe
	}

	return recipe, nil
}

func (s recipeService) StarRecipeByID(ctx context.Context, recipeID string, author models.UserSummary) (bool, error) {
	recipeStarDocs, err := s.recipeStarsCollection.Where("user.id", "==", author.ID).Where("recipe.id", "==", recipeID).Documents(ctx).GetAll()

	if err != nil {
		return false, err
	}

	if len(recipeStarDocs) > 0 {
		return false, ErrRecipeAlreadyStarred
	}

	recipeDoc, err := s.recipeCollection.Doc(recipeID).Get(ctx)

	if err != nil {
		return false, ErrRecipeNotFound
	}

	recipe := new(models.Recipe)
	recipeDoc.DataTo(recipe)

	starredDate := time.Now().UTC()
	recipeStarDoc := models.RecipeStar{
		User: author,
		Recipe: models.RecipeSummary{
			ID:   recipe.ID,
			Name: recipe.CurrentRevision.Name,
		},
		CreatedDate: starredDate,
		UpdatedDate: starredDate,
	}

	_, err = s.fire.Batch().Create(s.recipeStarsCollection.NewDoc(), recipeStarDoc).Update(recipeDoc.Ref, []firestore.Update{
		{Path: "metadata.stars", Value: firestore.Increment(1)},
	}).Commit(ctx)

	if err != nil {
		return false, err
	}

	return true, nil
}
