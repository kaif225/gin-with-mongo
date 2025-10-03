package main

import (
	"log"
	"moviestreaming/database"
	"moviestreaming/routes"

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
	routes.Protected(router)
	routes.Unprotected(router)

	err = router.Run(":8007")

	if err != nil {
		log.Println(err)
		return
	}
}
