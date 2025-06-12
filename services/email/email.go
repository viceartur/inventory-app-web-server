package email

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gomail "gopkg.in/gomail.v2"
)

type EmailRequest struct {
	To      string   `json:"to"`
	Cc      []string `json:"cc,omitempty"`  // Carbon Copy
	Bcc     []string `json:"bcc,omitempty"` // Blind Carbon Copy
	Subject string   `json:"subject"`
	Body    string   `json:"body,omitempty"`
}

// SendEmail sends an email with an Excel attachment using gomail.
// Supports CC/BCC and dynamic Excel content.
func SendEmail(req EmailRequest) error {
	from := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	log.Println(from, password, smtpHost, smtpPort)

	// Generate Excel file
	excelBuf, err := GenerateFormattedExcel()
	if err != nil {
		return fmt.Errorf("excel generation failed: %w", err)
	}

	var body string
	if req.Body != "" {
		body = req.Body
	} else {
		body, err = loadEmailBody("email_body.html")
		if err != nil {
			return fmt.Errorf("email body load failed: %w", err)
		}
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", req.To)
	if len(req.Cc) > 0 {
		m.SetHeader("Cc", req.Cc...)
	}
	if len(req.Bcc) > 0 {
		m.SetHeader("Bcc", req.Bcc...)
	}
	m.SetHeader("Subject", req.Subject)
	m.SetHeader("Date", time.Now().Format(time.RFC1123Z))
	m.SetBody("text/html", body)
	m.Attach("report.xlsx",
		gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(excelBuf.Bytes())
			return err
		}),
		gomail.SetHeader(map[string][]string{
			"Content-Type":              {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
			"Content-Disposition":       {`attachment; filename="report.xlsx"`},
			"Content-Transfer-Encoding": {"base64"},
		}),
	)

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	d := gomail.NewDialer(smtpHost, port, from, password)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Gomail error: %v", err)
		return err
	}

	log.Println("Email with Excel attachment sent successfully")
	return nil
}

func loadEmailBody(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	body := string(content)
	body = strings.ReplaceAll(body, "{{Date}}", time.Now().Format("January 2, 2006"))

	return body, nil
}
