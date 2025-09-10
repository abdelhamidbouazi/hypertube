package main

import (
	_ "server/docs"
	"server/internal"
	"server/internal/services"
)

// Swagger
//
//	@title						Hypertube API
//	@version					0.1.0
//	@description				A comprehensive API for searching movies, streaming them instantly, and managing users profiles.
//	@host						http://localhost:8080
//	@BasePath					/
//	@schemes					http
//	@securityDefinitions.apiKey	JWT
//	@in							header
//	@name						Authorization
//	@description				JWT security accessToken. Please add it in the format "Bearer {AccessToken}" to authorize your requests.
func main() {
	internal.Init("hypertube.yml")
	services.Logger.Info("Run all configurations!")
	internal.StartServer()
}
