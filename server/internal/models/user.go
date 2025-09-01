package models

import (
	"github.com/lib/pq"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model `json:"-"`
	Tokens     pq.StringArray `gorm:"type:text[]" json:"-"`
	Name       string
	Email      string `gorm:"uniqueIndex"`
	Password   string `json:"-"`
}
