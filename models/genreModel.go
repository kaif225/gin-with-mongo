package models

type Genre struct {
	GenreID   int    `json:"genre_id"  bson:"genre_id" validate:"required"`
	GenreName string `json:"genre_name"  bson:"genre_name" validate:"required,min=2,max=100"`
}
