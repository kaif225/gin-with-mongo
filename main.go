package main

import (
	"log"
	"moviestreaming/controller"
	"moviestreaming/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println(err)
		return
	}

	err = database.Connect()
	if err != nil {
		log.Println(err)
		return
	}
	router := gin.Default()
	router.GET("/movies", controller.GetMovies)
	router.GET("/movies/:imdb_id", controller.GetMovie)
	router.POST("/movies", controller.AddMovies)
	router.POST("/register", controller.RegisterUser)
	router.POST("/login", controller.LoginUser)
	err = router.Run(":8007")

	if err != nil {
		log.Println(err)
		return
	}
}
