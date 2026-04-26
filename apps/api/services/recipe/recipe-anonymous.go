package recipesvc

import (
	"context"
	"sort"

	firestore "cloud.google.com/go/firestore"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"

	models "4ks/libs/go/models"
)

func (s recipeService) GetRecipeByID(ctx context.Context, id string) (*models.Recipe, error) {
	result, err := s.recipeCollection.Doc(id).Get(ctx)

	if err != nil {
		return nil, ErrRecipeNotFound
	}

	recipe := new(models.Recipe)
	if err = result.DataTo(recipe); err != nil {
		return nil, err
	}

	recipe.ID = result.Ref.ID
	return recipe, nil
}

func (s recipeService) GetRecipes(ctx context.Context, limit int) ([]*models.Recipe, error) {
	var all []*models.Recipe
	iter := s.recipeCollection.Limit(limit).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Handle error, possibly by returning the error
			// to the caller. Break the loop or return.
			return nil, err
		}
		var u models.Recipe
		if err := doc.DataTo(&u); err != nil {
			// Handle error, possibly by returning the error
			// to the caller. Continue the loop,
			// break the loop or return.
			return nil, err
		}
		all = append(all, &u)
	}

	return all, nil
}

// GetRecipesByUsername gets all recipes for a given username
func (s recipeService) GetRecipesByUsername(ctx context.Context, username string, limit int) ([]*models.Recipe, error) {
	log.Debug().Str("username", username).Msg("GetRecipesByUsername")
	var all []*models.Recipe
	iter := s.recipeCollection.Where("author.username", "==", &username).Limit(limit).Documents(ctx)

	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Handle error, possibly by returning the error
			// to the caller. Break the loop or return.
			return nil, err
		}
		var u models.Recipe
		if err := doc.DataTo(&u); err != nil {
			// Handle error, possibly by returning the error
			// to the caller. Continue the loop,
			// break the loop or return.
			return nil, err
		}
		all = append(all, &u)
	}

	return all, nil
}

// GetRecipesByUserID
func (s recipeService) GetRecipesByUserID(ctx context.Context, id string, limit int) ([]*models.Recipe, error) {
	log.Debug().Str("id", id).Msg("GetRecipesByUserID")

	var all []*models.Recipe
	iter := s.recipeCollection.Where("author.id", "==", &id).Limit(limit).Documents(ctx)

	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Handle error, possibly by returning the error
			// to the caller. Break the loop or return.
			return nil, err
		}
		var u models.Recipe
		if err := doc.DataTo(&u); err != nil {
			// Handle error, possibly by returning the error
			// to the caller. Continue the loop,
			// break the loop or return.
			return nil, err
		}
		all = append(all, &u)
	}

	return all, nil
}

func (s recipeService) GetRecipeRevisions(ctx context.Context, recipeID string) ([]*models.RecipeRevision, error) {
	recipeRevisionsDocs, err := s.recipeRevisionsCollection.Where("recipeId", "==", recipeID).OrderBy("createdDate", firestore.Desc).Documents(ctx).GetAll()

	if err != nil {
		return nil, err
	}

	numberOfRevisions := len(recipeRevisionsDocs)
	if numberOfRevisions == 0 {
		return nil, ErrRecipeNotFound
	}

	recipeRevisions := make([]*models.RecipeRevision, numberOfRevisions)
	for i, ds := range recipeRevisionsDocs {
		recipeRevision := new(models.RecipeRevision)
		ds.DataTo(recipeRevision)
		recipeRevisions[i] = recipeRevision
	}

	return recipeRevisions, nil
}

func (s recipeService) GetRecipeForks(ctx context.Context, recipeID string) ([]*models.Recipe, error) {
	recipeDoc, err := s.recipeCollection.Doc(recipeID).Get(ctx)

	if err != nil || !recipeDoc.Exists() {
		return nil, ErrRecipeNotFound
	}

	recipeForkDocs, err := s.recipeCollection.Where("branch", "==", recipeID).Documents(ctx).GetAll()

	if err != nil {
		return nil, err
	}

	recipeForks := make([]*models.Recipe, len(recipeForkDocs))
	for i, ds := range recipeForkDocs {
		recipeFork := new(models.Recipe)
		ds.DataTo(recipeFork)
		recipeForks[i] = recipeFork
	}

	sort.Slice(recipeForks, func(i, j int) bool {
		return recipeForks[i].UpdatedDate.After(recipeForks[j].UpdatedDate)
	})

	return recipeForks, nil
}

func (s recipeService) GetRecipeRevisionByID(ctx context.Context, revisionID string) (*models.RecipeRevision, error) {
	recipeRevisionDoc, err := s.recipeRevisionsCollection.Doc(revisionID).Get(ctx)

	if err != nil {
		return nil, ErrRecipeRevisionNotFound
	}

	recipeRevision := new(models.RecipeRevision)
	recipeRevisionDoc.DataTo(recipeRevision)

	return recipeRevision, nil
}
