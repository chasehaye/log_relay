package templates

import (
    "fmt"
    "log_relay/internal/services"
)

func SendResetEmailChangeEmail(toEmail string, resetLink string) error {
    htmlBody := fmt.Sprintf(`
        <h1 style="text-align: center;">Confirm Email Change</h1>
        <p style="text-align: center;">
            You requested to change your email. Click below to confirm:
        </p>
        <div style="margin: 20px 0; text-align: center;">
            <a href="%s"
            style="background: black;
                    color: #8f8f8f;
                    padding: 12px 20px;
                    text-decoration: none;
                    display: inline-block;">
                Confirm Email Change
            </a>
        </div>
        <p style="word-break: break-all; text-align: center;">%s</p>
    `, resetLink, resetLink)

    return services.SendMail(toEmail, "Confirm your email change", htmlBody)
}