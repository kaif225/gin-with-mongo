package routes

import (
	"moviestreaming/controller"
	mw "moviestreaming/middlewares"

	"github.com/gin-gonic/gin"
)

func Protected(router *gin.Engine) {

	protected := router.Group("/")
	protected.Use(mw.JWT()) // middleware only for these routes

	protected.GET("/movies/:imdb_id", controller.GetMovie)
	protected.POST("/addMovies", controller.AddMovies)
}
