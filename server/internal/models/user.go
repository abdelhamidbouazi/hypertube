package models

import (
	"github.com/lib/pq"

	"gorm.io/gorm"
)

// UserPublicInfo represents user information without sensitive data like email
type UserPublicInfo struct {
	Username          string `json:"username" example:"fturing"`
	FirstName         string `json:"firstname" example:"Alan"`
	LastName          string `json:"lastname" example:"Turing"`
	Avatar            string `json:"avatar" example:"data:image/png;base64,iVBORw0KGgoA"`
	PreferredLanguage string `json:"preferred_language" example:"en"`
}

type User struct {
	gorm.Model     `json:"-"`
	Tokens         pq.StringArray `gorm:"type:text[]" json:"-"`
	UserPublicInfo `json:",inline"`
	Email          string `json:"email" example:"example@email.com"`
	Password       string `json:"-" example:"aK62p1HYiC1f"`
	Provider       string `json:"-"`
	ProviderId     string `json:"-"`
}
