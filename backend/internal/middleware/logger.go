package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Printf("%s %s %d %s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
