package messaging

import "fmt"

func SendResetEmail(toEmail string, resetLink string) error {
    htmlBody := fmt.Sprintf(`
test
    `, resetLink, resetLink)

    return SendGmail(toEmail, "Reset your password", htmlBody)
}