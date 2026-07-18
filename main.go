package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/middlewares"
	"github.com/raihansyahrin/backend_laundry_app.git/routes"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize Gin router
	r := gin.Default()

	// Apply CORS middleware globally
	r.Use(middlewares.CORSMiddleware())

	// Connect to database
	config.ConnectDatabase()

	// Setup routes with middleware
	routes.SetupRoutes(r)

	// Start the server
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
