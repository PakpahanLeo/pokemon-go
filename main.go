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

type Moves struct {
	Move MoveName `json:"move"`
}
type MoveName struct {
	Name string `json:"name"`
}

type Type struct {
	Type TypeName `json:"type"`
}
type TypeName struct {
	Name string `json:"name"`
}

type PokemonInfo struct {
	Height int     `json:"height"`
	ID     int     `json:"id"`
	Images string  `json:"images"`
	Moves  []Moves `json:"moves"`
	Name   string  `json:"name"`
	Types  []Type  `json:"types"`
	Weight int     `json:"weight"`
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
			// imagesJSON, _ := json.Marshal(requestInfo.Images)
			movesJSON, _ := json.Marshal(requestInfo.Moves)
			// Insert the Pokemon data into the database
			_, err := db.Exec("INSERT INTO pokemongo (id, name, weight, height, types, images, moves) VALUES (?, ?, ?, ?, ?, ?, ?)",
				requestInfo.ID, requestInfo.Name, requestInfo.Weight, requestInfo.Height, typesJSON, requestInfo.Images, movesJSON)
			if err != nil {
				// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				// c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
					"info":    make(map[string]interface{}),
				})
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

	r.GET("/get-pokemon-catch", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, height, weight, images, moves, types FROM pokemongo ORDER BY id ASC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		pokemonMap := make(map[int]*PokemonInfo)

		for rows.Next() {
			var id int
			var name, moveData, typeData string
			var height, weight int
			var images string

			if err := rows.Scan(&id, &name, &height, &weight, &images, &moveData, &typeData); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			pokemon, exists := pokemonMap[id]
			if !exists {
				pokemon = &PokemonInfo{
					ID:     id,
					Name:   name,
					Height: height,
					Weight: weight,
					Images: images,
				}
				pokemonMap[id] = pokemon
			}

			// Unmarshal JSON data from the "moves" column
			var movesData []struct {
				Move struct {
					Name string `json:"name"`
				} `json:"move"`
			}

			if err := json.Unmarshal([]byte(moveData), &movesData); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for _, m := range movesData {
				pokemon.Moves = append(pokemon.Moves, Moves{Move: MoveName{Name: m.Move.Name}})
			}

			// Unmarshal JSON data from the "types" column
			var typesData []struct {
				Type struct {
					Name string `json:"name"`
				} `json:"type"`
			}

			if err := json.Unmarshal([]byte(typeData), &typesData); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for _, m := range typesData {
				pokemon.Types = append(pokemon.Types, Type{Type: TypeName{Name: m.Type.Name}})
			}
		}

		var pokemonList []PokemonInfo
		for _, pokemon := range pokemonMap {
			pokemonList = append(pokemonList, *pokemon)
		}

		c.JSON(http.StatusOK, pokemonList)
	})

	//KASIR
	type User struct {
		ID       int    // or uint, depending on your database schema
		Username string `json:"username"`
		Password string `json:"password"`
	}

	r.POST("/login", func(c *gin.Context) {
		var requestUser User
		if err := c.ShouldBindJSON(&requestUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Query the database to check if the user exists
		var foundUser User
		err := db.QueryRow("SELECT username, password FROM users WHERE username = ?", requestUser.Username).Scan(&foundUser.Username, &foundUser.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Check if the password matches (use a secure password hashing library in production)
		if requestUser.Password != foundUser.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password disini"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Login successful", "userID": foundUser.ID})
	})

	type Category struct {
		Name string `json:"name"`
	}

	r.POST("/addCategory", func(c *gin.Context) {
		var category Category
		if err := c.ShouldBindJSON(&category); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Insert the new category into the database
		_, err := db.Exec("INSERT INTO category (name) VALUES (?)", category.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert category into the database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Category added successfully"})
	})

	type Product struct {
		Name        string `json:"name"`
		Category    string `json:"category"`
		Price       int    `json:"price"`
		Description string `json:"description"`
		Discount    int    `json:"discount"`
		Image       string `json:"image"`
		Status      string `json:"status"`
	}

	r.POST("/addProduct", func(c *gin.Context) {
		// Parse form data from the request
		err := c.Request.ParseMultipartForm(10 << 20) // 10 MB limit for form data
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get form data values
		productName := c.Request.FormValue("name")
		category := c.Request.FormValue("category")
		price := c.Request.FormValue("price")
		description := c.Request.FormValue("description")
		discount := c.Request.FormValue("discount")
		status := c.Request.FormValue("status")

		// Validate form data
		if productName == "" || category == "" || price == "" || description == "" || discount == "" || status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
			return
		}

		// Handle image upload from form field named "image"
		// file, err := c.FormFile("image")
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
		// 	return
		// }

		_, err = db.Exec("INSERT INTO product (name, category, price, description, discount, image, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
			productName, category, price, description, discount, "", status)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert product into the database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product added successfully"})
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
