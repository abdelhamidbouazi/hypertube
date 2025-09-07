package resetpassword

import (
	"server/internal/models"
	"server/internal/services"
)

type NewResetPassword struct {
	Token string
	Email string
}

func CreateNew(record NewResetPassword) error {
	db := services.PostgresDB()

	r := models.ResetPassword{
		Token: record.Token,
		Email: record.Email,
	}

	res := db.Create(&r)
	return res.Error
}

func DeleteByEmail(email string) error {
	db := services.PostgresDB()

	res := db.Delete(&models.ResetPassword{}, "email = ?", email)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func CountTokensByEmail(token string, email string) (int64, error) {
	db := services.PostgresDB()

	toFind := models.ResetPassword{
		Email: email,
		Token: token,
	}

	res := db.Find(&toFind)

	return res.RowsAffected, res.Error
}
