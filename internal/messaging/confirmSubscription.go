package messaging

import (
    "fmt"
    "log_relay/internal/services"
)


func SendConfirmationEmail(toEmail string, confirmLink string) error {
	htmlBody := fmt.Sprintf(`
		<h1 style="text-align: center;">Confirm Your Subscription</h1>
        <p style="text-align: center;">
            Thanks for signing up! Please click the button below to confirm your subscription and start receiving updates:
        </p>
        <div style="margin: 20px 0; text-align: center;">
            <a href="%s"
               style="background: black;
                      color: #8f8f8f;
                      padding: 12px 20px;
                      text-decoration: none;
                      display: inline-block;">
                Confirm Subscription
            </a>
        </div>
        <p style="text-align: center;">
            If you didn't request this, you can safely ignore this email.
        </p>
        <p style="text-align: center;">
            Alternatively, copy and paste this link into your browser:
        </p>
        <p style="word-break: break-all; text-align: center;">%s</p>
	`, confirmLink, confirmLink)

	return services.SendMail(toEmail, "Please confirm your subscription", htmlBody)
}