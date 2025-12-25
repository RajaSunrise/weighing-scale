package reporting

import (
	"fmt"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"
	"stoneweigh/internal/models"
)

// GenerateInvoice creates a PDF invoice for a weighing transaction
func GenerateInvoice(record models.WeighingRecord) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "STONEWEIGH WEIGHING SLIP")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Ticket No: %s", record.TicketNumber))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Date: %s", record.WeighedAt.Format("2006-01-02 15:04:05")))
	pdf.Ln(20)

	// Details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Vehicle Details")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(50, 10, "Plate Number:")
	pdf.Cell(50, 10, record.PlateNumber)
	pdf.Ln(8)
	pdf.Cell(50, 10, "Driver:")
	pdf.Cell(50, 10, record.DriverName)
	pdf.Ln(8)
	pdf.Cell(50, 10, "Product:")
	pdf.Cell(50, 10, record.Product)
	pdf.Ln(15)

	// Weight Data
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Weight Data (kg)")
	pdf.Ln(10)

	pdf.SetFont("Courier", "", 12)
	pdf.Cell(50, 10, "GROSS:")
	pdf.Cell(30, 10, fmt.Sprintf("%8.0f", record.GrossWeight))
	pdf.Ln(8)
	pdf.Cell(50, 10, "TARE:")
	pdf.Cell(30, 10, fmt.Sprintf("%8.0f", record.TareWeight))
	pdf.Ln(8)
	pdf.SetFont("Courier", "B", 14)
	pdf.Cell(50, 10, "NET:")
	pdf.Cell(30, 10, fmt.Sprintf("%8.0f", record.NetWeight))
	pdf.Ln(20)

	// Footer
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 10, "Thank you for your business.")

	// Ensure directory exists
	if _, err := os.Stat("web/static/reports"); os.IsNotExist(err) {
		os.MkdirAll("web/static/reports", 0755)
	}

	filename := fmt.Sprintf("web/static/reports/inv_%s.pdf", record.TicketNumber)
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// GenerateDailyReport creates a list of transactions for a specific day
func GenerateDailyReport(date time.Time, records []models.WeighingRecord) (string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("Daily Weighing Report - %s", date.Format("2006-01-02")))
	pdf.Ln(15)

	// Table Header
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 10, "Time")
	pdf.Cell(30, 10, "Ticket")
	pdf.Cell(30, 10, "Vehicle")
	pdf.Cell(40, 10, "Product")
	pdf.Cell(30, 10, "Gross")
	pdf.Cell(30, 10, "Tare")
	pdf.Cell(30, 10, "Net")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	var totalNet float64

	for _, r := range records {
		pdf.Cell(30, 8, r.WeighedAt.Format("15:04:05"))
		pdf.Cell(30, 8, r.TicketNumber)
		pdf.Cell(30, 8, r.PlateNumber)
		pdf.Cell(40, 8, r.Product)
		pdf.Cell(30, 8, fmt.Sprintf("%.0f", r.GrossWeight))
		pdf.Cell(30, 8, fmt.Sprintf("%.0f", r.TareWeight))
		pdf.Cell(30, 8, fmt.Sprintf("%.0f", r.NetWeight))
		pdf.Ln(8)
		totalNet += r.NetWeight
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(160, 10, "TOTAL NET WEIGHT:")
	pdf.Cell(30, 10, fmt.Sprintf("%.0f kg", totalNet))

	filename := fmt.Sprintf("web/static/reports/daily_%s.pdf", date.Format("20060102"))
	err := pdf.OutputFileAndClose(filename)
	return filename, err
}
