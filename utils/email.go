package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends a styled HTML email using Gmail SMTP configurations.
func SendEmail(to string, subject string, bodyHTML string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpHost == "" || smtpPort == "" || smtpEmail == "" || smtpPassword == "" {
		return fmt.Errorf("SMTP credentials not configured. Please set SMTP_PASSWORD in your backend .env file")
	}

	// Message headers and body
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("myLaundry <%s>", smtpEmail)
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + bodyHTML

	// Setup SMTP auth
	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	// Trigger SendMail using net/smtp package (handles STARTTLS on port 587 automatically)
	err := smtp.SendMail(addr, auth, smtpEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send SMTP email: %w", err)
	}

	return nil
}
