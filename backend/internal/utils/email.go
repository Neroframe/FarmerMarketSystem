package utils

import (
	"log"

	"gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "no-reply@mailtrap.io")
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/plain", body)

	dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", "ed3092461eda54302535bcee6a6aeed1")

	if err := dialer.DialAndSend(mailer); err != nil {
		log.Printf("Email send error: %v", err)
		return err
	}

	log.Println("Email sent successfully!")
	return nil
}
