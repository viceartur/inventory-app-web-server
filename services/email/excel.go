package email

import (
	"bytes"
	"log"

	"github.com/xuri/excelize/v2"
)

// GenerateFormattedExcel creates an Excel file with headers and optional rows.
func GenerateFormattedExcel() (*bytes.Buffer, error) {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Column width and row height
	f.SetColWidth(sheet, "A", "C", 30)
	f.SetRowHeight(sheet, 1, 28)

	// Header values
	headers := []string{"Name", "Email", "Status"}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		f.SetCellValue(sheet, cell, h)
	}

	// Style for headers
	styleID, err := CreateHeaderStyle(f, sheet)
	if err != nil {
		log.Fatalf("Failed to create style: %v", err)
	}

	// Set style
	f.SetCellStyle(sheet, "A1", "C1", styleID)

	// Add rows
	f.SetCellValue(sheet, "A2", "Alice")
	f.SetCellValue(sheet, "B2", "alice@example.com")
	f.SetCellValue(sheet, "C2", "Active")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func CreateHeaderStyle(f *excelize.File, sheet string) (int, error) {
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

	styleID, err := f.NewStyle(style)
	if err != nil {
		return 0, err
	}

	return styleID, nil
}
