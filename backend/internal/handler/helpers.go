package handler

import "github.com/gin-gonic/gin"

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func respondMessage(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"message": message})
}

func userIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

func userRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("userRole")
	if !exists {
		return "", false
	}

	value, ok := role.(string)
	return value, ok
}

func usernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	value, ok := username.(string)
	return value, ok
}
