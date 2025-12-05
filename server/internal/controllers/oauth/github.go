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
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// Github OAuth2 godoc
//
//	@Summary		Register using Github API OAuth2
//	@Description	Register new user using Github API OAuth2
//	@Tags			oauth2
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	auth.RevokeTokenRes
//	@Router			/oauth2/github [post]
func Github(c echo.Context) error {
	config := oauthService.Providers()["github"]
	if config == nil {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "Github API method not implemented",
		})
	}
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusFound, url)
}

type OAuth2GithubRedirect struct {
	Code string `validate:"required" query:"code"`
}

func GithubCallback(c echo.Context) error {
	config := oauthService.Providers()["github"]
	if config == nil {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"message": "Github API method not implemented",
		})
	}

	var body OAuth2GithubRedirect

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

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
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

	type Source struct {
		Id        int    `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"name"`
		LastName  string `json:"last_name"`
		Avatar    string `json:"avatar_url"`
		Login     string `json:"login"`
	}

	var APIUserRes Source
	if err := json.Unmarshal(resBody.Bytes(), &APIUserRes); err != nil {
		return err
	}

	provideId := strconv.Itoa(APIUserRes.Id)
	user, err := users.GetUserByProviderId("github", provideId)
	if err == gorm.ErrRecordNotFound {
		newUser := users.CreateUserType{
			Provider:   "github",
			ProviderId: provideId,
			Email:      APIUserRes.Email,
			FirstName:  APIUserRes.FirstName,
			LastName:   APIUserRes.LastName,
			Avatar:     APIUserRes.Avatar,
			Username:   APIUserRes.Login,
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
	cookie.Value = response.AccessToken
	cookie.Path = "/"
	cookie.HttpOnly = false
	cookie.Secure = false
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	cookie = new(http.Cookie)
	cookie.Name = "AccessTokenExpiresIn"
	cookie.Value = fmt.Sprintf("%d", response.ExpiresIn)
	cookie.Path = "/"
	cookie.HttpOnly = false
	cookie.Secure = false
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	cookie = new(http.Cookie)
	cookie.Name = "RefreshToken"
	cookie.Value = response.RefreshToken
	cookie.Path = "/"
	cookie.HttpOnly = false
	cookie.Secure = false
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	return c.Redirect(http.StatusFound, services.Conf.UI.Address+"/"+services.Conf.UI.OauthCallbackRoute)
}
