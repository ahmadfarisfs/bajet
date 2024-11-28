package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID        primitive.ObjectID `json:"id" bson:"_id" validate:"-"`
	Icon      string             `json:"icon" bson:"icon" form:"icon" validate:"required"`
	Name      string             `json:"name" bson:"name" form:"name" validate:"required"`
	Type      string             `json:"type" bson:"type" form:"type" validate:"required,oneof=expense income"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at" validate:"-"`
}
