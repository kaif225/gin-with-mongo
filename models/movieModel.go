package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Movie struct {
	ID          bson.ObjectID `json:"_id,omitempty"  bson:"_id,,omitempty"`
	ImdbID      string        `json:"imdb_id"  bson:"imdb_id" validate:"required"`
	Title       string        `json:"title"  bson:"title" validate:"required,min=2,max=400"`
	PosterPath  string        `json:"poster_path"  bson:"poster_path" validate:"required,url"`
	YouTubeID   string        `json:"youtube_id"  bson:"youtube_id" validate:"required"`
	Genre       []Genre       `json:"genre"  bson:"genre" validate:"required,dive"` // here dive means that the corresponding key will also be validated.
	AdminReview string        `json:"admin_review"  bson:"admin_review" validate:"required"`
	Ranking     Ranking       `json:"ranking"  bson:"ranking" validate:"required"`
}
