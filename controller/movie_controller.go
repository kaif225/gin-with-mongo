package controller

import (
	"context"
	"fmt"
	"log"
	"moviestreaming/database"
	"moviestreaming/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func GetMovies(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	var movies []models.Movie
	dbName := os.Getenv("DATABASE_NAME")
	cursor, err := database.Client.Database(dbName).Collection("movies").Find(ctx, bson.M{})
	//cursor, err := client.Database(dbName).Collection("movies").Find(ctx, bson.M{})
	if err != nil {
		log.Println("Find error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch movies"})
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &movies); err != nil {
		log.Println("Cursor decode error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode movies"})
		return
	}

	if movies == nil {
		movies = []models.Movie{}
	}

	c.JSON(http.StatusOK, movies)
}

func GetMovie(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	movieID := c.Param("imdb_id")

	fmt.Println("ID :", movieID)
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Movie ID is required"})
		return
	}
	var movie models.Movie
	dbName := os.Getenv("DATABASE_NAME")

	err := database.Client.Database(dbName).Collection("movies").FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Required Body"})
		return
	}

	c.JSON(http.StatusOK, movie)
}

// func AddMovies(c *gin.Context) {
// 	validate := validator.New()
// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)

// 	defer cancel()

// 	var movies []models.Movie
// 	c.BindJSON(&movies)
// 	dbName := os.Getenv("DATABASE_NAME")
// 	for _, movie := range movies {
// 		err := validate.Struct(movie)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"message": "Validation failed"})
// 			return
// 		}
// 		_, err = database.Client.Database(dbName).Collection("movies").InsertOne(ctx, movie)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Movies have not been added"})
// 			return
// 		}
// 		movies = append(movies, movie)
// 	}

// 	c.JSON(http.StatusOK, movies)

// }

// Using Insertmany in place of InserOne , above code is correct but Inertmany is a better choice.

func AddMovies(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	validate := validator.New()

	var movies []models.Movie
	err := c.BindJSON(&movies)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	for _, movie := range movies {
		err = validate.Struct(movie)

		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Validation Failed"})
			return
		}
	}
	dbName := os.Getenv("DATABASE_NAME")
	collection := database.Client.Database(dbName).Collection("movies")

	docs := make([]interface{}, len(movies))

	for i, movie := range movies {
		docs[i] = movie
	}
	result, err := collection.InsertMany(ctx, docs)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Adding movies"})
		return
	}
	c.JSON(http.StatusOK, result)
}
