package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable not set")
	}

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS albums (
		id INT AUTO_INCREMENT PRIMARY KEY,
		artist VARCHAR(255) NOT NULL,
		title VARCHAR(255) NOT NULL,
		year INT NOT NULL,
		image_url VARCHAR(512) NOT NULL
	) ENGINE=InnoDB;
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/albums", func(c *gin.Context) {
		var album struct {
			Artist   string `json:"artist" binding:"required"`
			Title    string `json:"title" binding:"required"`
			Year     int    `json:"year" binding:"required"`
			ImageURL string `json:"image_url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&album); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request data"})
			return
		}

		result, err := db.Exec("INSERT INTO albums (artist, title, year, image_url) VALUES (?, ?, ?, ?)",
			album.Artist, album.Title, album.Year, album.ImageURL)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		id, _ := result.LastInsertId()
		c.JSON(200, gin.H{"id": id})
	})

	r.GET("/albums/:id", func(c *gin.Context) {
		id := c.Param("id")

		var album struct {
			ID       int    `json:"id"`
			Artist   string `json:"artist"`
			Title    string `json:"title"`
			Year     int    `json:"year"`
			ImageURL string `json:"image_url"`
		}

		err := db.QueryRow("SELECT id, artist, title, year, image_url FROM albums WHERE id = ?", id).
			Scan(&album.ID, &album.Artist, &album.Title, &album.Year, &album.ImageURL)

		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Album not found"})
			return
		} else if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, album)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s ...", port)
	r.Run(":" + port)
}
