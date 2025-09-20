package controllers

import (
	"context"
	"mime/multipart"
	"net/http"
	"server/internal/models"
	"server/internal/services"
	"server/internal/services/users"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
)

func GetUsers(c echo.Context) error {
	users, err := users.GetUsers()
	if err != nil {
		response := echo.Map{
			"message": err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	response := echo.Map{
		"data": users,
	}

	return c.JSON(http.StatusOK, response)
}

func GetMe(c echo.Context) error {
	return c.JSON(http.StatusOK, c.Get("model"))
}

type UploadPictureReq struct {
	Picture *multipart.FileHeader `form:"picture" validate:"required"`
}

type UploadPictureRes struct {
	Message string `json:"message" example:"success"`
}

func UploadPicture(c echo.Context) error {
	var form UploadPictureReq

	err := c.Bind(&form)
	if err != nil {
		return err
	}

	err = services.ValidateStruct(form)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	file, err := form.Picture.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user := c.Get("model").(models.User)

	key := user.FirstName + ".png"

	input := &s3.PutObjectInput{
		Bucket: aws.String(services.AWSBucketName),
		Key:    aws.String(key),
		Body:   file,
	}

	scv := services.AWS_Client()

	_, err = scv.PutObject(context.TODO(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	user.Avatar = "https://" + services.AWSBucketName + ".s3." + services.AWSBucketRegion + ".amazonaws.com/" + key

	err = users.UpdateUser(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, UploadPictureRes{
		Message: "success",
	})
}
