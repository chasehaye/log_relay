package messaging

import "fmt"

func SendResetEmail(toEmail string, resetLink string) error {
    htmlBody := fmt.Sprintf(`
        <h1>Password Reset</h1>
        <p>You requested a password reset. Click the button below to continue:</p>
        <div style="margin: 20px 0;">
            <a href="%s" style="background: #007bff; color: white; padding: 12px 20px; text-decoration: none; border-radius: 5px;">Reset Password</a>
        </div>
        <p>If the button doesn't work, copy and paste this link:</p>
        <p>%s</p>
    `, resetLink, resetLink)

    return SendMailAWS(toEmail, "Reset your password", htmlBody)
}