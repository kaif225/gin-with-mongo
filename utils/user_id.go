package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func GetUserIdFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")

	if !exists {
		return "", errors.New("User id does not exists in the context")
	}
	id, ok := userID.(string)

	if !ok {
		return "", errors.New("Unable to retrive userId")
	}

	return id, nil
}
