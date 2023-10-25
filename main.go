package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.GET("/gin", func(c *gin.Context) {
		c.String(200, "Hello, Gin!")
	})
	r.Run("0.0.0.0:3000")
}
