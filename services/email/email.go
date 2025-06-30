package email

import (
	"database/sql"
	"fmt"
	interfaces "inv_app/services/email/interfaces"
	"inv_app/services/reports"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
)

// HandleCustomerReportsEmail generates, formats, and sends customer usage reports via email.
// It fetches report data, groups by customer, retrieves email addresses, generates attachments,
// sends emails, and logs the status in the database.
func HandleCustomerReportsEmail(db *sql.DB, customerUsgFilter reports.SearchQuery) error {
	// Ensure the report period is specified.
	if customerUsgFilter.DateFrom == "" || customerUsgFilter.DateTo == "" {
		return fmt.Errorf("No period specified.")
	}

	// Initialize report settings.
	reportSettings := interfaces.NewReport()

	// Prepare the report query.
	report := &reports.Report{
		DB:                  db,
		CustomerUsageFilter: customerUsgFilter,
	}

	// Fetch the list of customer usage reports.
	customerReports, err := report.GetCustomerUsage()
	if err != nil {
		return fmt.Errorf("Error getting customer reports: %w", err)
	}

	// Group reports by CustomerID and collect unique customer IDs.
	customerIdSet := make(interfaces.CustomerSet)
	for _, rep := range customerReports {
		cid := interfaces.CustomerID(rep.CustomerID)
		reportSettings.AddReport(cid, rep)
		customerIdSet[cid] = struct{}{}
	}

	// Fetch customer and representative emails concurrently for each customer.
	var wg sync.WaitGroup
	for cid := range customerIdSet {
		_, ok := reportSettings.GetCustomerSettings(cid)
		if !ok {
			log.Printf("Customer settings not found for ID: %d", cid)
			continue
		}

		wg.Add(1)
		go func(id interfaces.CustomerID) {
			defer wg.Done()

			if err := reportSettings.SetCustomerEmails(db, id); err != nil {
				log.Println("Error setting customer emails:", err)
			}
			if err := reportSettings.SetRepresentativeEmail(db, id); err != nil {
				log.Println("Error setting rep email:", err)
			}
		}(cid)
	}
	wg.Wait()

	if reportSettings.IsCustomerSettingsEmpty() {
		return fmt.Errorf("No customer settings defined.")
	}

	// Send emails concurrently for each customer.
	var wgSend sync.WaitGroup
	for customerId, customerSettings := range reportSettings.CustomerSettings {
		if reportSettings.IsCustomerEmailsEmpty(customerId) {
			return fmt.Errorf(
				"No customer emails defined for the Customer ID: %d",
				customerId,
			)
		}

		if reportSettings.IsRepresentativeEmailEmpty(customerId) {
			return fmt.Errorf(
				"No representative email defined for the Customer ID: %d",
				customerId,
			)
		}

		emailSettings := interfaces.NewEmail()

		// Generate the Excel attachment for the report.
		emailSettings.Attachment, err = emailSettings.GenerateFormattedExcel(customerSettings.Report)
		if err != nil {
			return err
		}

		// Load the email body template.
		emailSettings.Body, err = emailSettings.LoadEmailBody(customerUsgFilter)
		if err != nil {
			return err
		}

		// Set email recipients.
		emailSettings.To = customerSettings.CustomerEmails
		if customerSettings.RepresentativeEmail != "" {
			emailSettings.Cc = interfaces.Email(customerSettings.RepresentativeEmail)
		}
		emailSettings.Subject = fmt.Sprintf(
			"Inventory Report | %s - %s",
			customerUsgFilter.DateFrom,
			customerUsgFilter.DateTo,
		)

		wgSend.Add(1)
		go func(emailSettings *interfaces.EmailSettings) {
			defer wgSend.Done()

			emailStatus := emailSettings.InitEmailStatus(customerId)

			// Send the email and update status.
			err := EmailCustomerReport(emailSettings)
			if err != nil {
				emailStatus.Status = interfaces.Status(fmt.Sprintf(
					"Error: %s", err,
				))
			} else {
				emailStatus.Status = interfaces.Status(fmt.Sprintf(
					"Succes: Report sent for [%s - %s].",
					customerUsgFilter.DateFrom,
					customerUsgFilter.DateTo,
				))
			}

			// Save status into the database.
			emailSettings.AddEmailSentStatus(db, emailStatus)

		}(emailSettings)
	}
	wgSend.Wait()

	return nil
}

// EmailCustomerReport sends an email with the given Excel attachment.
func EmailCustomerReport(email *interfaces.EmailSettings) error {
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")

	if user == "" || password == "" || smtpHost == "" || smtpPort == "" || from == "" {
		return fmt.Errorf("SMTP configuration is incomplete.")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)

	to := make([]string, len(email.To))
	for i, e := range email.To {
		to[i] = string(e)
	}
	if len(to) == 0 {
		return fmt.Errorf("No recipients specified.")
	}
	m.SetHeader("To", to...)

	if email.Cc != "" {
		m.SetHeader("Cc", string(email.Cc))
	}

	if len(email.Bcc) > 0 {
		bcc := make([]string, len(email.Bcc))
		for i, e := range email.Bcc {
			bcc[i] = string(e)
		}
		m.SetHeader("Bcc", bcc...)
	}

	m.SetHeader("Subject", email.Subject)
	m.SetHeader("Date", time.Now().Format(time.RFC1123Z))
	m.SetBody("text/html", email.Body)
	if email.Attachment != nil {
		m.Attach("customer_report.xlsx",
			gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(email.Attachment.Bytes())
				return err
			}),
			gomail.SetHeader(map[string][]string{
				"Content-Type":              {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
				"Content-Disposition":       {`attachment; filename="customer_report.xlsx"`},
				"Content-Transfer-Encoding": {"base64"},
			}),
		)
	}

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		return fmt.Errorf("Invalid SMTP_PORT: %w", err)
	}

	d := gomail.NewDialer(smtpHost, port, user, password)
	return d.DialAndSend(m)
}
