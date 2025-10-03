package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"moviestreaming/database"
	"moviestreaming/models"
	"moviestreaming/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/tmc/langchaingo/llms/openai"
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

	token, refreshToken, err := utils.GenerateAllToken(userLogin.Email, userExist.FirstName, userExist.LastName, userExist.Role, userExist.UserID)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token not created properly"})
		return
	}

	err = utils.UpdateAllToken(userExist.UserID, token, refreshToken)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in updating token"})
		return
	}
	// c.JSON(http.StatusOK, models.UserResponse{
	// 	UserID:          userExist.UserID,
	// 	FirstName:       userExist.FirstName,
	// 	LastName:        userExist.LastName,
	// 	Email:           userLogin.Email,
	// 	Role:            userExist.Role,
	// 	Token:           token,
	// 	RefreshToken:    refreshToken,
	// 	FavouriteGenres: userExist.FavouriteGenres,
	// })

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "Bearer",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	response := struct {
		Status string `json:"status"`
		Token  string `json:"token"`
	}{
		Status: "Login Successfull",
		Token:  token,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func Logout(c *gin.Context) {

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
}

func AdminReview(c *gin.Context) {
	imbdID := c.Param("imbd_id")

	if imbdID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	var req struct {
		AdminReview string `json:"admin_review"`
	}

	var resp struct {
		RankingName string `json:"ranking_name"`
		AdminReview string `json:"admin_review"`
	}

	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

}

func GetReviewRanking(admin_review string) (string, int, error) {
	rankings, err := GetRankings()

	if err != nil {
		return "", 0, err
	}
	sentimentDelimited := ""

	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited = sentimentDelimited + ranking.RankingName + ","
		}
	}
	sentimentDelimited = strings.Trim(sentimentDelimited, ",")
	openAiKey := os.Getenv("OPENAI_API_KEY")
	if openAiKey == "" {
		return "", 0, errors.New("could not read api key")
	}

	llm, err := openai.New(openai.WithToken(openAiKey))
	if err != nil {
		return "", 0, err
	}

	base_prompt_template := os.Getenv("BASE_PROMPT_TEMPLATE")

	base_prompt := strings.Replace(base_prompt_template, "{rankings}", sentimentDelimited, 1)
	response, err := llm.Call(context.Background(), base_prompt+admin_review)
	if err != nil {
		return "", 0, err
	}
	rankVal := 0

	for _, ranking := range rankings {
	}
	return "", 0, nil
}

func GetRankings() ([]models.Ranking, error) {
	var rankings []models.Ranking
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	dbName := os.Getenv("DATABASE_NAME")
	cursor, err := database.Client.Database(dbName).Collection("rankings").Find(ctx, bson.M{})
	if err != nil {
		log.Println("Find error:", err)
		return nil, err
	}

	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &rankings); err != nil {
		log.Println("Cursor decode error:", err)
		return nil, err
	}

	return rankings, nil
}
