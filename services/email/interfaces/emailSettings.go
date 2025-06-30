package interfaces

import (
	"bytes"
	"database/sql"
	"fmt"
	"inv_app/services/reports"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

type EmailSettings struct {
	To         []Email       `json:"to"`
	Cc         Email         `json:"cc,omitempty"`
	Bcc        []Email       `json:"bcc,omitempty"`
	Subject    string        `json:"subject"`
	Body       string        `json:"body,omitempty"`
	Attachment *bytes.Buffer `json:"attachment,omitempty"`
}

type EmailStatus struct {
	CustomerID
	Status
}

// NewReport initializes a new ReportSettings with a non-nil map.
func NewEmail() *EmailSettings {
	return &EmailSettings{}
}

func (*EmailSettings) LoadEmailBody(filter reports.SearchQuery) (string, error) {
	content, err := os.ReadFile("services/email/templates/email_body.html")
	if err != nil {
		return "", err
	}

	// Replace template variables
	body := string(content)
	body = strings.ReplaceAll(body, "{{start_date}}", filter.DateFrom)
	body = strings.ReplaceAll(body, "{{end_date}}", filter.DateTo)

	return body, nil
}

// GenerateFormattedExcel creates an Excel file with headers and rows.
func (*EmailSettings) GenerateFormattedExcel(report Report) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	sheet := "Customer Report"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	f.SetColWidth(sheet, "A", "J", 30)

	headers := []string{
		"Program Name",
		"Material Type",
		"Stock ID",
		"Qty Start",
		"Qty Received",
		"Qty Used",
		"Qty Spoiled",
		"Qty End",
		"6-Week Avg Usage",
		"Weeks Remaining",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	styleID, err := createHeaderStyle(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create style: %w", err)
	}
	f.SetCellStyle(sheet, "A1", "J1", styleID)

	for i, r := range report {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.ProgramName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.MaterialType)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.StockID)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.QtyStart)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.QtyReceived)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), r.QtyUsed)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), r.QtySpoiled)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), r.QtyEnd)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), r.WeekAvg)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), r.WeeksRemaining)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

// createHeaderStyle returns a style ID for Excel headers.
func createHeaderStyle(f *excelize.File) (int, error) {
	style := &excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	}
	return f.NewStyle(style)
}

func (r *EmailSettings) InitEmailStatus(customerId CustomerID) *EmailStatus {
	return &EmailStatus{CustomerID: customerId}
}

// AddSentStatus updates the customer's last report delivery status and timestamp in the database.
func (r *EmailSettings) AddEmailSentStatus(db *sql.DB, emailStatus *EmailStatus) error {
	_, err := db.Exec(`
		UPDATE
			customers
		SET
			last_report_delivery_status = $2,
			last_report_sent_at = NOW()
		WHERE
			customer_id = $1;
	`, emailStatus.CustomerID, emailStatus.Status)
	if err != nil {
		return err
	}

	return nil
}
