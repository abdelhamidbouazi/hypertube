package oauth2

import (
	"fmt"
	"server/internal/services"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var providers = make(map[string]*oauth2.Config)

func LoadFortyTwoConfig() {
	uid := viper.GetString("OAUTH_FORTYTWO_UID")
	secret := viper.GetString("OAUTH_FORTYTWO_SECRET")
	redirect := services.Conf.OAUTH.FortyTwo.Redirect
	if uid != "" && secret != "" && redirect != "" {
		fmt.Println("uid=", uid, "secret=", secret, "redirect=", redirect)
		providers["42"] = &oauth2.Config{
			ClientID:     uid,
			ClientSecret: secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://api.intra.42.fr/oauth/authorize",
				TokenURL: "https://api.intra.42.fr/oauth/token",
			},
			RedirectURL: redirect,
			Scopes:      []string{"public"},
		}
	}
}

func LoadGoogleConfig() {
	uid := viper.GetString("OAUTH_GOOGLE_UID")
	secret := viper.GetString("OAUTH_GOOGLE_SECRET")
	redirect := services.Conf.OAUTH.Google.Redirect
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

func LoadConfig() {
	LoadGoogleConfig()
	LoadFortyTwoConfig()
}

func Providers() map[string]*oauth2.Config {
	return providers
}
