package models

import (
	"github.com/lib/pq"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model `json:"-"`
	Tokens     pq.StringArray `gorm:"type:text[]" json:"-"`
	FirstName  string
	LastName   string
	Email      string `gorm:"uniqueIndex" json:"-"`
	Password   string `json:"-"`
}
