package messaging
// fix me
import (
    "fmt"
    "os"
    "net/smtp"
)

func SendMailAWS(to, subject, body string) error {

    user := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("AMAZON_HOST")
	from := os.Getenv("SENDER_ADDRESS")
	port := "587"

    addr := host + ":" + port

    messageStr := fmt.Sprintf(
		"From: Fude Software <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"utf-8\"\r\n"+
			"\r\n"+
			"%s",
		from, to, subject, body,
	)

    auth := smtp.PlainAuth("", user, password, host)

	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(messageStr))
	if err != nil {
		return fmt.Errorf("AWS SES send failed: %w", err)
	}

	return nil
}