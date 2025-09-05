package models

import (
	"github.com/lib/pq"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model `json:"-"`
	Tokens     pq.StringArray `gorm:"type:text[]" json:"-"`
<<<<<<< Updated upstream
	FirstName  string
	LastName   string
	Email      string `gorm:"uniqueIndex" json:"-"`
	Password   string `json:"-"`
=======
	FirstName  string         `json:"firstname"`
	LastName   string         `json:"lastname"`
	Email      string         `json:"email"`
	Password   string         `json:"-"`
	Avatar     string         `json:"avatar"`
	Provider   string         `json:"-"`
	ProviderId string         `json:"-"`
>>>>>>> Stashed changes
}
