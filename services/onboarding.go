package services

import "go.mongodb.org/mongo-driver/mongo"

type OnboardingService struct {
	db *mongo.Database
}

func NewOnboardingService(db *mongo.Database) *OnboardingService {
	return &OnboardingService{
		db: db,
	}
}

func (os *OnboardingService) PopulateDefault() error {
	// Implement logic to populate default data
	return nil
}
