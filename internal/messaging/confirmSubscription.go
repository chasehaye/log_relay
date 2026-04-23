package messaging

import (
    "fmt"
    "log_relay/internal/services"
)

// SendConfirmationEmail sends the double opt-in link to a new subscriber
func SendConfirmationEmail(toEmail string, confirmLink string) error {
	htmlBody := fmt.Sprintf(`
		<h1>Confirm Your Subscription</h1>
		<p>Thanks for signing up! Please click the button below to confirm your subscription and start receiving updates:</p>
		<div style="margin: 20px 0;">
			<a href="%s" style="background: #28a745; color: white; padding: 12px 20px; text-decoration: none; border-radius: 5px; font-weight: bold;">Confirm Subscription</a>
		</div>
		<p>If you didn't request this, you can safely ignore this email.</p>
		<p>Alternatively, copy and paste this link into your browser:</p>
		<p>%s</p>
	`, confirmLink, confirmLink)

	return services.SendMail(toEmail, "Please confirm your subscription", htmlBody)
}