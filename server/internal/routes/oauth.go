package routes

import (
	"server/internal/controllers/oauth"

	"github.com/labstack/echo/v4"
)

func AddOAuthRouter(oauthRouter *echo.Group) {
	addGoogleOAuthRouter(oauthRouter.Group("/google"))
}

func addGoogleOAuthRouter(oauthRouter *echo.Group) {
	oauthRouter.POST("", oauth.Google)
	oauthRouter.GET("/callback", oauth.GoogleCallback)
}
