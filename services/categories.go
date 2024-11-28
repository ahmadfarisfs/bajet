package services

import (
	"bajetapp/model"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryService struct {
	db *mongo.Database
}

func NewCategoryService(db *mongo.Database) *CategoryService {
	return &CategoryService{
		db: db,
	}
}

func (cs *CategoryService) GetCategories(ctx context.Context, userID string) ([]model.Category, error) {

	collection := cs.db.Collection("users")
	filter := bson.M{
		"_id": userID,
	}
	opts := options.FindOne().SetProjection(bson.M{
		"categories": 1,
	})

	var result struct {
		Categories []model.Category `bson:"categories"`
	}

	err := collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Categories, nil
}

func (cs *CategoryService) CreateCategory(ctx context.Context, userID string, category model.Category) error {

	// Check if the category name already exists for the user
	existingCategories, err := cs.GetCategories(ctx, userID)
	if err != nil {
		return err
	}

	for _, existingCategory := range existingCategories {
		if existingCategory.Name == category.Name {
			return model.ErrDuplicateCategoryName
		}
		if existingCategory.Icon == category.Icon {
			return model.ErrDuplicateCategoryIcon
		}
	}

	filter := bson.M{
		"_id": userID,
	}
	// Generate a new ObjectID for the category ID
	category.ID = primitive.NewObjectID()

	update := bson.M{
		"$push": bson.M{
			"categories": category,
		},
	}

	_, err = cs.db.Collection("users").UpdateOne(ctx, filter, update)
	return err
}
