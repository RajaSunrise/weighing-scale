package report

import (
	"fmt"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"
	"stoneweigh/internal/models"
)

func GenerateTicketPDF(txn models.WeighingRecord) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "TICKET TIMBANGAN")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Ticket ID: %s", txn.TicketNumber))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Date: %s", txn.CreatedAt.Format(time.RFC1123)))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Plate Number: %s", txn.PlateNumber))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Driver: %s", txn.DriverName))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Net Weight: %.2f Kg", txn.NetWeight))

	// Ensure directory exists
	os.MkdirAll("web/static/reports", 0755)

	filename := fmt.Sprintf("web/static/reports/ticket_%s.pdf", txn.TicketNumber)
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
