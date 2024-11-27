package utils

import (
	"gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "FMS@test.com")
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/plain", body)

	dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", "ed3092461eda54302535bcee6a6aeed1")

	return dialer.DialAndSend(mailer)
}
