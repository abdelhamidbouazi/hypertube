package oauth2

import (
	"fmt"
	"server/internal/services"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var providers = make(map[string]*oauth2.Config)

func LoadConfig() {
	uid := viper.GetString("OAUTH_GOOGLE_UID")
	secret := viper.GetString("OAUTH_GOOGLE_SECRET")
	redirect := services.Conf.OAUTH.Google.Redirect
	fmt.Println("uid=", uid, "secret=", secret, "redirect=", redirect)
	if uid != "" && secret != "" && redirect != "" {
		providers["google"] = &oauth2.Config{
			ClientID:     uid,
			ClientSecret: secret,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirect,
			Scopes:       []string{"profile", "email"},
		}
	}
}

func Providers() map[string]*oauth2.Config {
	return providers
}
