package utils

import (
	"context"
	"log"
	"moviestreaming/database"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/mgo.v2/bson"
)

type SignedDetails struct {
	Email                string
	FirstName            string
	LastName             string
	Role                 string
	UserID               string
	jwt.RegisteredClaims // it is a struct provided by jwt package
}

// var secretKey string = os.Getenv("SECRET_KEY")
// var refreshSecretKey string = os.Getenv("SECRET_REFRESH_KEY")

func GenerateAllToken(email, firstName, lastName, role, userId string) (string, string, error) {
	var secretKey string = os.Getenv("SECRET_KEY")
	var refreshSecretKey string = os.Getenv("SECRET_REFRESH_KEY")
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserID:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagciSteam",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))

	if err != nil {
		log.Println(err)
		return "", "", nil
	}

	// Code For refresh token

	refreshClaims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserID:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagciSteam",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(refreshSecretKey))

	if err != nil {
		log.Println(err)
		return "", "", nil
	}

	return signedToken, signedRefreshToken, nil
}

func UpdateAllToken(userId, token, refreshToken string) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	updateAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	updateData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			"update_at":     updateAt,
		},
	}

	dbName := os.Getenv("DATABASE_NAME")
	_, err = database.Client.Database(dbName).Collection("users").UpdateOne(ctx, bson.M{"user_id": userId}, updateData)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
