package services

import (
	"log"
	"server/internal/utils"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var Conf AppConfig

type AppConfig struct {
	DB struct {
		NAME string `mapstructure:"NAME"`
		USER string `mapstructure:"USER"`
		PASS string `mapstructure:"PASS"`
		HOST string `mapstructure:"HOST"`
		PORT int    `mapstructure:"PORT"`
	} `mapstructure:"DB"`

	HTTP struct {
		PORT int
	} `mapstructure:"HTTP"`

	JWT struct {
		SigningKey            string `mapstructure:"SECRET_KEY"`
		AccessTkExpiresAtRaw  string `mapstructure:"ACCESS_TOKEN_EXPIRES_AT"`
		RefreshTkExpiresAtRaw string `mapstructure:"REFRESH_TOKEN_EXPIRES_AT"`
		AccessTkExpiresAt     time.Duration
		RefreshTkExpiresAt    time.Duration
	} `mapstructure:"JWT"`

	CORS struct {
		Origins []string `mapstructure:"ORIGINS"`
	} `mapstructure:"CORS"`

	SMTP struct {
		Gmail struct {
			Mail     string `mapstructure:"MAIL"`
			Password string `mapstructure:"PASSWORD"`
			Host     string `mapstructure:"HOST"`
			Port     int    `mapstructure:"PORT"`
		} `mapstructure:"GMAIL"`
	} `mapstructure:"SMTP"`

	UI struct {
		Address                 string `mapstructure:"ADDRESS"`
		ResetPasswordRoute      string `mapstructure:"RESET_PASSWORD_ROUTE"`
		ResetPasswordTokenQuery string `mapstructure:"RESET_PASSWORD_TOKEN_QUERY"`
		OauthCallbackRoute      string `mapstructure:"OAUTH_CALLBACK_ROUTE"`
	}

	OAUTH struct {
		Google struct {
			Redirect string `mapstructure:"REDIRECT"`
		} `mapstructure:"GOOGLE"`
		FortyTwo struct {
			Redirect string `mapstructure:"REDIRECT"`
		} `mapstructure:"FORTYTWO"`
		Github struct {
			Redirect string `mapstructure:"REDIRECT"`
		} `mapstructure:"GITHUB"`
	} `mapstructure:"OAUTH"`

	MOVIE_APIS struct {
		TMDB struct {
			APIKey string `mapstructure:"API_KEY"`
		} `mapstructure:"TMDB"`
		OMDB struct {
			APIKey string `mapstructure:"API_KEY"`
		} `mapstructure:"OMDB"`
	} `mapstructure:"MOVIE_APIS"`

	DOWNLOADS struct {
		Directory string `mapstructure:"DIRECTORY"`
	} `mapstructure:"DOWNLOADS"`

	STREAMING struct {
		DownloadDir              string  `mapstructure:"DOWNLOAD_DIR"`
		TMDBAPIKey               string  `mapstructure:"TMDB_API_KEY"`
		MinBytesForTranscoding   int64   `mapstructure:"MIN_BYTES_FOR_TRANSCODING"`
		MinPercentForTranscoding float64 `mapstructure:"MIN_PERCENT_FOR_TRANSCODING"`
	} `mapstructure:"STREAMING"`
}

func LoadConfig(config string) {
	Logger.Debug("Loading Config")
	viper.SetConfigFile(config)

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = viper.Unmarshal(&Conf)
	if err != nil {
		log.Fatal(err)
	}

	if Conf.STREAMING.MinBytesForTranscoding < 5*1024*1024 {
		log.Fatal("MinBytesForTranscoding should be greater or equal than 5MB")
	}

	if Conf.STREAMING.MinPercentForTranscoding < 0.5 {
		log.Fatal("MinPercentForTranscoding should be greater or equal than 0.5 percent")
	}

	setupExtra()
}

func setupExtra() {
	Conf.CORS.Origins = strings.Split(viper.GetString("CORS_ORIGINS"), ",")

	expAt, err := utils.ParseDuration(Conf.JWT.AccessTkExpiresAtRaw)
	if err != nil {
		log.Fatal(err)
	}

	Conf.JWT.AccessTkExpiresAt = expAt

	expAt, err = utils.ParseDuration(Conf.JWT.RefreshTkExpiresAtRaw)
	if err != nil {
		log.Fatal(err)
	}

	Conf.JWT.RefreshTkExpiresAt = expAt
}
