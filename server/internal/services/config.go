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

	OAUTH struct {
		Google struct {
			Redirect string `mapstructure:"REDIRECT"`
		} `mapstructure:"GOOGLE"`
	} `mapstructure:"OAUTH"`
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
