package routes

import (
	"moviestreaming/controller"

	"github.com/gin-gonic/gin"
)

func Unprotected(router *gin.Engine) {
	router.GET("/movies", controller.GetMovies)
	router.POST("/register", controller.RegisterUser)
	router.POST("/login", controller.LoginUser)
	router.POST("/logout", controller.Logout)
}
