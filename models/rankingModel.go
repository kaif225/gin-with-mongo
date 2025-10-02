package models

type Ranking struct {
	RankingValue int    `json:"ranking_value" bson:"ranking_value" validate:"required"`
	RankingName  string `json:"ranking_name" bson:"ranking_name" validate:"required"`
}
