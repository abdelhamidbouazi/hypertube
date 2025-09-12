package models

import (
	"github.com/lib/pq"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model `json:"-"`
	Tokens     pq.StringArray `gorm:"type:text[]" json:"-"`
	FirstName  string         `json:"firstname" example:"Alan"`
	LastName   string         `json:"lastname" example:"Turing"`
	Email      string         `json:"email" example:"example@email.com"`
	Password   string         `json:"-" example:"aK62p1HYiC1f"`
	Avatar     string         `json:"avatar" example:"data:image/png;base64,iVBORw0KGgoA"`
	Provider   string         `json:"-"`
	ProviderId string         `json:"-"`
}
