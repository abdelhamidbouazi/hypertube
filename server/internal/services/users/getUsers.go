package users

import (
	"server/internal/models"
	"server/internal/services"
)

func GetUsers() ([]*models.User, error) {
	var users []*models.User
	db := services.PostgresDB()

	res := db.Find(&users)

	for _, v := range users {
		v.Tokens = nil
	}

	if res.Error != nil {
		return nil, res.Error
	}

	return users, nil
}
