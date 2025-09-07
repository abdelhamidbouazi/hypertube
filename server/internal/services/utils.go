package services

import (
	"html/template"
	"log"

	"github.com/spf13/viper"
	gomail "gopkg.in/mail.v2"
)

var (
	_reset_password_template *template.Template
	dialer                   *gomail.Dialer
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
