package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"moviestreaming/database"
	"moviestreaming/models"
	"moviestreaming/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/argon2"
)

//var userCollection *mongo.Collection = database.Client.Database("magic-stream-movies").Collection("users")

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		log.Println(err)
		return "", err
	}
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	Hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	HashBase64 := base64.StdEncoding.EncodeToString(Hash)
	encodedPass := fmt.Sprintf("%s.%s", saltBase64, HashBase64)
	return encodedPass, nil
}

func RegisterUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Input data"})
		return
	}
	validate := validator.New()

	err = validate.Struct(user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Failed", "details": err.Error()})
		return
	}
	HashPass, err := HashPassword(user.Password)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to Hash password."})
		return
	}
	//count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	dbName := os.Getenv("DATABASE_NAME")
	count, err := database.Client.Database(dbName).Collection("users").CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to check existing user"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exist"})
		return
	}
	user.UserID = bson.NewObjectID().Hex()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Password = HashPass
	result, err := database.Client.Database(dbName).Collection("users").InsertOne(ctx, user)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusOK, result)

}

func LoginUser(c *gin.Context) {
	var userLogin models.UserLogin
	if err := c.ShouldBindJSON(&userLogin); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Request Body"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(userLogin); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "detail": err.Error()})
		return
	}

	if userLogin.Email == "" || userLogin.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	dbName := os.Getenv("DATABASE_NAME")
	collection := database.Client.Database(dbName).Collection("users")

	userExist := &models.User{}
	// var userExist models.User
	err := collection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(userExist)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	err = utils.VerifyPassword(userLogin.Password, userExist.Password)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login Successful"})
}
