package users

import (
	"server/internal/models"
	"server/internal/services"
)

func GetUserById(id float64) (models.User, error) {
	db := services.PostgresDB()
	var user models.User
	res := db.First(&user, "id = ?", id)
	if res.Error != nil {
		return user, res.Error
	}
	return user, nil
}

func GetUserByProviderId(provider string, id string) (models.User, error) {
	db := services.PostgresDB()
	var user models.User
	res := db.Where("provider = ?", provider).First(&user, "provider_id = ?", id)
	if res.Error != nil {
		return user, res.Error
	}
	return user, nil
}

func GetUserByEmail(email string) (models.User, error) {
	db := services.PostgresDB()
	var user models.User
	res := db.Where("provider = ?", "").First(&user, "email = ?", email)
	if res.Error != nil {
		return user, res.Error
	}
	return user, nil
}
