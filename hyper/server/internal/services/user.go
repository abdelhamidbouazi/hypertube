package services

import (
	"context"
	"mime/multipart"
	"server/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers() ([]*models.User, error) {
	var users []*models.User
	db := PostgresDB()

	res := db.Find(&users)

	for _, v := range users {
		v.Tokens = nil
	}

	if res.Error != nil {
		return nil, res.Error
	}

	return users, nil
}

type CreateUserType struct {
	FirstName string `validate:"required,min=4"`
	LastName  string `validate:"required,min=4"`
	Email     string `validate:"required,email"`
	Password  string
}

func CreateUser(newUser CreateUserType) error {
	db := PostgresDB()
	var user models.User
	user.Email = newUser.Email
	user.FirstName = newUser.FirstName
	user.LastName = newUser.LastName
	hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 8)
	if err != nil {
		return err
	}
	user.Password = string(hashed)
	res := db.Create(&user)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func GetUserByEmail(email string) (models.User, error) {
	db := PostgresDB()
	var user models.User
	res := db.First(&user, "email = ?", email)
	if res.Error != nil {
		return user, res.Error
	}
	return user, nil
}

func UpdateUser(user models.User) error {
	db := PostgresDB()
	res := db.Save(&user)
	return res.Error
}

func UploadPicture(pic *multipart.FileHeader, username string) (string, error) {
	var url string

	file, err := pic.Open()
	if err != nil {
		return url, err
	}

	key := username + ".png"

	input := &s3.PutObjectInput{
		Bucket: aws.String(AWSBucketName),
		Key:    aws.String(key),
		Body:   file,
	}

	scv := AWS_Client()

	_, err = scv.PutObject(context.TODO(), input)
	if err != nil {
		return url, err
	}

	return "https://" + AWSBucketName + ".s3." + AWSBucketRegion + ".amazonaws.com/" + key, nil
}
