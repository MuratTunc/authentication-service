package mailer

import (
	"authentication-service/src/config"
	"fmt"
	"net/smtp"
)

func SendPasswordResetMail(toEmail, resetCode string, cfg *config.Config) error {
	from := cfg.SMTPEmail
	password := cfg.SMTPPassword
	host := cfg.SMTPHost
	port := cfg.SMTPPort

	auth := smtp.PlainAuth("", from, password, host)

	subject := "Subject: Password Reset Request\n"
	body := fmt.Sprintf("Your password reset code is: %s\nThis code is valid for 15 minutes.", resetCode)
	msg := []byte(subject + "\n" + body)

	addr := host + ":" + port

	err := smtp.SendMail(addr, auth, from, []string{toEmail}, msg)
	if err != nil {
		return err
	}
	return nil
}

func SendAuthenticationCode(toEmail, authentication_code string, cfg *config.Config) error {
	from := cfg.SMTPEmail
	password := cfg.SMTPPassword
	host := cfg.SMTPHost
	port := cfg.SMTPPort

	auth := smtp.PlainAuth("", from, password, host)

	subject := "Subject: authentication_code Request\n"
	body := fmt.Sprintf("Your authentication code is: %s\nThis code is valid for 15 minutes.", authentication_code)
	msg := []byte(subject + "\n" + body)

	addr := host + ":" + port

	err := smtp.SendMail(addr, auth, from, []string{toEmail}, msg)
	if err != nil {
		return err
	}
	return nil
}
