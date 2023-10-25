package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Pokemon struct {
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
}

type PokemonInfo struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Weight float64  `json:"weight"`
	Height float64  `json:"height"`
	Types  []string `json:"types"`
	Images []string `json:"images"`
	Moves  []string `json:"moves"`
}

var pokemon Pokemon

func main() {
	var port = envPortOr("3000")

	rand.Seed(time.Now().UnixNano())

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, Gin!")
	})

	db, err := sql.Open("mysql", "root:n8hgdm64f@fokpvdps0$o05j_s@jpchf@tcp(roundhouse.proxy.rlwy.net:34843)/railway")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.POST("/catch-pokemon", func(c *gin.Context) {
		var requestInfo PokemonInfo
		if err := c.ShouldBindJSON(&requestInfo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate a random number between 0 and 1
		randomNumber := rand.Float64()

		// Determine the catch result based on the 50% probability
		var catchResult string
		if randomNumber < 0.5 {
			catchResult = "Success! You caught the Pokemon!"
			// Convert slices to JSON strings
			typesJSON, _ := json.Marshal(requestInfo.Types)
			imagesJSON, _ := json.Marshal(requestInfo.Images)
			movesJSON, _ := json.Marshal(requestInfo.Moves)
			// Insert the Pokemon data into the database
			_, err := db.Exec("INSERT INTO pokemongo (id, name, weight, height, types, images, moves) VALUES (?, ?, ?, ?, ?, ?, ?)",
				requestInfo.ID, requestInfo.Name, requestInfo.Weight, requestInfo.Height, string(typesJSON), string(imagesJSON), string(movesJSON))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			catchResult = "Oops! The Pokemon got away."
		}

		// Return the catch result and the received Pokemon info as JSON
		c.JSON(http.StatusOK, gin.H{
			"message": catchResult,
			"info":    requestInfo,
		})
	})

	r.POST("/change-nickname", func(c *gin.Context) {
		var requestPokemon Pokemon
		if err := c.ShouldBindJSON(&requestPokemon); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pokemongo WHERE id = ?)", requestPokemon.ID).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pokemon with the specified ID not found"})
			return
		}
		_, err = db.Exec("UPDATE pokemongo SET name = ? WHERE id = ?", requestPokemon.Nickname, requestPokemon.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Nickname of Pokemon changed successfully"})
	})

	r.Run(port)
}

func envPortOr(port string) string {
	// If `PORT` variable in environment exists, return it
	if envPort := os.Getenv("PORT"); envPort != "" {
		return ":" + envPort
	}
	// Otherwise, return the value of `port` variable from function argument
	return ":" + port
}
