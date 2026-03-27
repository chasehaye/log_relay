package messaging
// fix me
import (
    "context"
    "encoding/base64"
    "fmt"
    "os"

    "google.golang.org/api/gmail/v1"
    "google.golang.org/api/option"
)

func SendGmail(to, subject, body string) error {

    ctx := context.Background()
    credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
    adminEmail := os.Getenv("GMAIL_USER")

    service, err := gmail.NewService(ctx, 
        option.WithCredentialsFile(credentialsFile),
    )
    if err != nil {
        return fmt.Errorf("unable to retrieve Gmail client: %v", err)
    }
	fmt.Print("after")
    messageStr := fmt.Sprintf(
        "From: Fude Software <%s>\r\n"+
            "To: %s\r\n"+
            "Subject: %s\r\n"+
            "MIME-Version: 1.0\r\n"+
            "Content-Type: text/html; charset=\"utf-8\"\r\n"+
            "\r\n%s",
        adminEmail, to, subject, body,
    )

    msg := &gmail.Message{
        Raw: base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(messageStr)),
    }

	
	fmt.Printf("Encoded Base64: %s\n", messageStr)
    fmt.Println("Attempting to send...\n")
    _, err = service.Users.Messages.Send("me", msg).Do()
    if err != nil {
        return fmt.Errorf("unable to send message: %v", err)
    }
	fmt.Print("sent")
    return nil
}