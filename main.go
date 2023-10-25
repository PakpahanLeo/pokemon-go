package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

// func main() {
// 	r := gin.Default()
// 	r.GET("/gin", func(c *gin.Context) {
// 		c.String(200, "Hello, Gin!")
// 	})
// 	r.Run("0.0.0.0:3000")
// }

func main() {
	var port = envPortOr("3000")

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, Gin!")
	})
	r.Run(port)
}

// Returns PORT from environment if found, defaults to
// value in `port` parameter otherwise. The returned port
// is prefixed with a `:`, e.g. `":3000"`.
func envPortOr(port string) string {
	// If `PORT` variable in environment exists, return it
	if envPort := os.Getenv("PORT"); envPort != "" {
		return ":" + envPort
	}
	// Otherwise, return the value of `port` variable from function argument
	return ":" + port
}
