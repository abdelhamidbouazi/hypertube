package users

import (
	"server/internal/models"
	"server/internal/services"
)

func UpdateUser(user models.User) error {
	db := services.PostgresDB()
	res := db.Save(&user)
	return res.Error
}
