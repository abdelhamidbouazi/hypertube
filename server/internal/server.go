package internal

import (
	"fmt"
	"net/http"
	"server/internal/controllers"
	"server/internal/middlewares"
	"server/internal/routes"
	"server/internal/services"
	"strconv"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var Server *echo.Echo

func Init(config string) {
	services.LoadConfig(config)
	services.LoadTemplateConfig()
	services.LoadMailDialer()
	services.LoadValidator()
	services.LoadDatabase()
	LoadServer()

	fmt.Println("mail=", services.Conf.SMTP.Gmail.Mail, " password=", services.Conf.SMTP.Gmail.Password)

	// seeds.AddUsersSeeds()
}

func LoadServer() {
	services.Logger.Debug("Loading Server")

	Server = echo.New()
	Server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowOrigins: services.Conf.CORS.Origins,
	}))
	Server.Use(middleware.Logger())
	config := echojwt.Config{
		SigningKey: []byte(services.Conf.JWT.SigningKey),
	}

	middlewares.SetupJWT(config)

	Server.POST("/forgot-password", controllers.ForgotPassword)
	Server.POST("/reset-password", controllers.ResetPassword)
	routes.AddAuthRouter(Server.Group("/auth"))
	routes.AddUserRouter(Server.Group("/users"))
}

func StartServer() {
	services.Logger.Info("Starting Server")
	port := strconv.Itoa(services.Conf.HTTP.PORT)
	if err := Server.Start(":" + port); err != http.ErrServerClosed {
		Server.Logger.Fatal(err)
	}
}
