package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/internal/controllers/auth"
	"server/internal/services"
	oauthService "server/internal/services/oauth2"
	"server/internal/services/users"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func Google(c echo.Context) error {
	config := oauthService.Providers()["google"]
	if config == nil {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "Google method not implemented",
		})
	}

	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusFound, url)
}

type OAuth2GoogleRedirect struct {
	Code string `validate:"required" query:"code"`
}

type GoogleUserResult struct {
	Id          string
	Email       string
	Name        string
	Given_name  string
	Family_name string
	Picture     string
}

func GoogleCallback(c echo.Context) error {
	config := oauthService.Providers()["google"]
	if config == nil {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "Google method not implemented",
		})
	}

	var body OAuth2GoogleRedirect

	err := c.Bind(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}

	token, err := config.Exchange(context.Background(), body.Code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "Unable to retrive token: " + err.Error(),
		})
	}

	err = services.ValidateStruct(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		services.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	token.SetAuthHeader(req)
	client := http.Client{
		Timeout: time.Second * 30,
	}
	res, err := client.Do(req)
	if err != nil {
		services.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if res.StatusCode != http.StatusOK {
		services.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var resBody bytes.Buffer
	_, err = io.Copy(&resBody, res.Body)
	if err != nil {
		return err
	}

	var GoogleUserRes map[string]interface{}
	if err := json.Unmarshal(resBody.Bytes(), &GoogleUserRes); err != nil {
		return err
	}

	providerId := GoogleUserRes["id"].(string)
	user, err := users.GetUserByProviderId("google", providerId)
	if err == gorm.ErrRecordNotFound {
		newUser := users.CreateUserType{
			Provider:   "google",
			ProviderId: providerId,
			Email:      GoogleUserRes["email"].(string),
			FirstName:  GoogleUserRes["name"].(string),
			Avatar:     GoogleUserRes["picture"].(string),
		}
		createdUser, err := users.CreateUser(newUser)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": err.Error(),
			})
		}
		user = createdUser
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	response, err := auth.RevokeToken(user, "")
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "AccessToken"
	cookie.Value = response["AccessToken"].(string)
	cookie.Path = "/"
	cookie.HttpOnly = false
	cookie.Secure = false
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	cookie = new(http.Cookie)
	cookie.Name = "AccessTokenExpiresIn"
	cookie.Value = fmt.Sprintf("%d", response["ExpiresIn"].(int64))
	cookie.Path = "/"
	cookie.HttpOnly = false
	cookie.Secure = false
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	cookie = new(http.Cookie)
	cookie.Name = "RefreshToken"
	cookie.Value = response["RefreshToken"].(string)
	cookie.Path = "/"
	cookie.HttpOnly = false
	cookie.Secure = false
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	return c.Redirect(http.StatusFound, services.Conf.UI.Address+"/"+services.Conf.UI.OauthCallbackRoute)
}
