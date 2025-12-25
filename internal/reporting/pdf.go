package reporting

import (
	"fmt"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"
	"stoneweigh/internal/models"
)

// dateToIndonesian formats time to "02 Januari 2006"
func dateToIndonesian(t time.Time) string {
	months := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return fmt.Sprintf("%02d %s %d %02d:%02d", t.Day(), months[t.Month()], t.Year(), t.Hour(), t.Minute())
}

// GenerateInvoice creates a PDF invoice for a weighing transaction (Indonesian & Modern)
func GenerateInvoice(record models.WeighingRecord) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// --- Header Section ---
	// Company Logo Placeholder (Text for now)
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(19, 109, 236) // Primary Blue
	pdf.Cell(0, 10, "STONEWEIGH INDONESIA")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, "Jalan Tambang Raya No. 123, Jakarta Selatan")
	pdf.Ln(5)
	pdf.Cell(0, 5, "Telp: (021) 555-0123 | Email: info@stoneweigh.id")
	pdf.Ln(15)

	// Divider Line
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, 45, 200, 45)
	pdf.Ln(10)

	// --- Title & Ticket Info ---
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, "SURAT JALAN / BUKTI TIMBANG")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "No. Tiket")
	pdf.SetFont("Courier", "B", 12)
	pdf.Cell(60, 8, ": "+record.TicketNumber)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(30, 8, "Tanggal")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(60, 8, ": "+dateToIndonesian(record.WeighedAt))
	pdf.Ln(12)

	// --- Details Grid ---
	// Background for Details
	pdf.SetFillColor(245, 247, 250)
	pdf.Rect(10, 80, 190, 60, "F")

	pdf.SetY(85)
	pdf.SetX(15)

	// Column 1
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 8, "Nomor Polisi")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 8, ": "+record.PlateNumber)

	// Column 2
	pdf.SetX(110)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 8, "Supir")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 8, ": "+record.DriverName)
	pdf.Ln(10)

	pdf.SetX(15)
	// Column 1
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 8, "Jenis Muatan")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 8, ": "+record.Product)

	// Column 2
	pdf.SetX(110)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 8, "Operator")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 8, ": "+record.ManagerName)
	pdf.Ln(10)

	// --- Weight Data Box ---
	pdf.SetY(120)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "DATA PENIMBANGAN (Kg)")
	pdf.Ln(10)

	pdf.SetFont("Courier", "", 12)

	// Header Row
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(60, 10, "Berat Kotor (Gross)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 10, "Berat Kosong (Tare)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 10, "Berat Bersih (Netto)", "1", 1, "C", true, 0, "")

	// Value Row
	pdf.SetFont("Courier", "B", 14)
	pdf.CellFormat(60, 12, fmt.Sprintf("%.0f", record.GrossWeight), "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 12, fmt.Sprintf("%.0f", record.TareWeight), "1", 0, "C", false, 0, "")
	pdf.SetTextColor(0, 150, 0) // Green for Net
	pdf.CellFormat(60, 12, fmt.Sprintf("%.0f", record.NetWeight), "1", 1, "C", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	// --- Signatures ---
	pdf.Ln(30)

	ySig := pdf.GetY()

	// Driver Sig
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 5, "Diserahkan Oleh (Supir),")

	// Manager Sig
	pdf.SetX(115)
	pdf.Cell(95, 5, "Diterima Oleh (Pengelola),")
	pdf.Ln(25)

	// Names
	pdf.SetY(ySig + 25)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(95, 5, "( "+record.DriverName+" )")

	pdf.SetX(115)
	pdf.Cell(95, 5, "( "+record.ManagerName+" )")

	// --- Footer ---
	pdf.SetY(260)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(150, 150, 150)
	pdf.Cell(0, 5, "Dokumen ini dicetak secara komputerisasi dan sah tanpa cap basah.")
	pdf.Ln(5)
	pdf.Cell(0, 5, "StoneWeigh System v1.0 - "+time.Now().Format("2006"))

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
