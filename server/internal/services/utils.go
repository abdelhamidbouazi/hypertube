package services

import (
	"context"
	"html/template"
	"log"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
	gomail "gopkg.in/mail.v2"
)

var (
	_reset_password_template *template.Template
	dialer                   *gomail.Dialer
	aws_client               *s3.Client
	AWSBucketRegion          = "eu-north-1"
	AWSBucketName            = "hypertube-users-pictures"
)

func LoadTemplateConfig() {
	t, err := template.ParseFiles("./assets/reset_password.html")
	if err != nil {
		log.Fatal(err)
	}
	_reset_password_template = t
}

func ResetPasswordTemplate() *template.Template {
	return _reset_password_template
}

func LoadMailDialer() {
	Conf.SMTP.Gmail.Mail = viper.GetString("SMTP_GMAIL_MAIL")
	Conf.SMTP.Gmail.Password = viper.GetString("SMTP_GMAIL_PASSWORD")
	dialer = gomail.NewDialer(Conf.SMTP.Gmail.Host, Conf.SMTP.Gmail.Port, Conf.SMTP.Gmail.Mail, Conf.SMTP.Gmail.Password)
}

func MailDialer() *gomail.Dialer {
	return dialer
}

func LoadAWSBucket() {
	if viper.GetString("AWS_ACCESS_KEY_ID") == "" {
		log.Fatal("AWS_ACCESS_KEY_ID was not read")
	}

	if viper.GetString("AWS_SECRET_ACCESS_KEY") == "" {
		log.Fatal("AWS_SECRET_ACCESS_KEY was not read")
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion("eu-north-1"))
	if err != nil {
		log.Fatal(err)
	}

	aws_client = s3.NewFromConfig(cfg)
}

func AWS_Client() *s3.Client {
	return aws_client
}
