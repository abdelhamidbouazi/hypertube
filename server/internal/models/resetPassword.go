package models

import "gorm.io/gorm"

type ResetPassword struct {
	gorm.Model
	Token string
	Email string
}
