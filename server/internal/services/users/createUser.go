package users

import (
	"server/internal/models"
	"server/internal/services"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
)

type CreateUserType struct {
	FirstName  string `validate:"required,min=4"`
	LastName   string `validate:"required,min=4"`
	Email      string `validate:"required,email"`
	Password   string
	Provider   string `validate:"required"`
	ProviderId string
	Avatar     string
	Username   string `validate:"required"`
}

func CreateUser(newUser CreateUserType) (models.User, error) {
	db := services.PostgresDB()
	var user models.User
	user.Email = newUser.Email
	user.FirstName = newUser.FirstName
	user.LastName = newUser.LastName
	user.Avatar = newUser.Avatar
	user.Username = newUser.Username
	hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 8)
	if err != nil {
		return user, err
	}
	user.Password = string(hashed)
	user.Provider = newUser.Provider
	user.ProviderId = newUser.ProviderId
	res := db.Clauses(clause.Returning{}).Create(&user)
	if res.Error != nil {
		return user, res.Error
	}
	db.First(&user, user.ID)
	return user, err
}
