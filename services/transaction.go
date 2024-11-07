package services

import (
	"bajetapp/model"
	"context"
	"time"

	"cloud.google.com/go/civil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TransactionService struct {
	db *mongo.Database
}

func NewTransactionService(db *mongo.Database) *TransactionService {
	return &TransactionService{
		db: db,
	}
}

func (ts *TransactionService) GetTransactions(ctx context.Context, userID string, dateStart civil.Date, dateEnd civil.Date) ([]model.Transaction, error) {
	collection := ts.db.Collection("transactions")
	filter := bson.M{
		"user_id": userID,
		"transaction_at": bson.M{
			"$gte": dateStart,
			"$lte": dateEnd,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "transaction_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []model.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (ts *TransactionService) CreateTransaction(ctx context.Context, transaction model.Transaction) error {
	transaction.ID = primitive.NewObjectID()
	transaction.CreatedAt = time.Now()
	collection := ts.db.Collection("transactions")
	_, err := collection.InsertOne(ctx, transaction)
	return err
}

func (ts *TransactionService) DeleteTransaction(ctx context.Context, transactionID string) error {
	collection := ts.db.Collection("transactions")
	filter := bson.M{"_id": transactionID}
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

func (ts *TransactionService) GetTransactionStats(ctx context.Context, userID string, dateStart civil.Date, dateEnd civil.Date) (model.TransactionStats, error) {
	collection := ts.db.Collection("transactions")
	pipeline := mongo.Pipeline{{{Key: "$match", Value: bson.D{
		{Key: "user_id", Value: userID},
		{Key: "transaction_at", Value: bson.D{
			{Key: "$gte", Value: dateStart},
			{Key: "$lte", Value: dateEnd},
		}},
	}}}, {{Key: "$group", Value: bson.D{
		{Key: "_id", Value: nil},
		{Key: "total_expense", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{bson.D{{Key: "$lt", Value: bson.A{"$amount", 0}}}, "$amount", 0}}}}}},
		{Key: "total_income", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{bson.D{{Key: "$gt", Value: bson.A{"$amount", 0}}}, "$amount", 0}}}}}},
		{Key: "date_start", Value: bson.D{{Key: "$first", Value: dateStart}}},
		{Key: "date_end", Value: bson.D{{Key: "$last", Value: dateEnd}}},
		{Key: "transaction_count", Value: bson.D{{Key: "$sum", Value: 1}}},
	}}}}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return model.TransactionStats{}, err
	}
	defer cursor.Close(ctx)

	var stats []model.TransactionStats
	if err = cursor.All(ctx, &stats); err != nil {
		return model.TransactionStats{}, err
	}

	if len(stats) > 0 {
		return stats[0], nil
	}

	return model.TransactionStats{}, nil
}

func (ts *TransactionService) GetAllCategories(userID string) ([]string, error) {
	collection := ts.db.Collection("transactions")
	distinctCategories, err := collection.Distinct(context.Background(), "category", bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	var categories []string
	for _, emoji := range distinctCategories {
		if e, ok := emoji.(string); ok {
			categories = append(categories, e)
		}
	}

	return categories, nil
}
