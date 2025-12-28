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

	// --- Colors ---
	// Primary Blue: #196DEC -> 25, 109, 236
	// Light Blue: #F0F7FF -> 240, 247, 255
	// Dark Grey: #333333 -> 51, 51, 51

	// --- Header Section ---
	// Background Banner
	pdf.SetFillColor(25, 109, 236)
	pdf.Rect(0, 0, 210, 30, "F")

	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(255, 255, 255) // White
	pdf.SetXY(10, 8)
	pdf.Cell(0, 10, "STONEWEIGH INDONESIA")

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(10, 18)
	pdf.Cell(0, 5, "Solusi Penimbangan Digital Terintegrasi")

	// Reset Position
	pdf.SetY(40)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, "Jalan Tambang Raya No. 123, Jakarta Selatan")
	pdf.Ln(5)
	pdf.Cell(0, 5, "Telp: (021) 555-0123 | Email: info@stoneweigh.id")
	pdf.Ln(10)

	// Divider Line
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.5)
	pdf.Line(10, 60, 200, 60)
	pdf.Ln(10)

	// --- Title & Ticket Info ---
	pdf.SetY(65)
	pdf.SetTextColor(25, 109, 236)
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, "SURAT JALAN / BUKTI TIMBANG")
	pdf.Ln(12)

	// Ticket Details Box
	pdf.SetFillColor(240, 247, 255)
	pdf.SetDrawColor(25, 109, 236)
	pdf.SetLineWidth(0.2)
	pdf.RoundedRect(10, 80, 190, 25, 2, "1234", "FD")

	pdf.SetY(85)
	pdf.SetTextColor(51, 51, 51)

	// Row 1 inside box
	pdf.SetX(15)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(25, 6, "No. Tiket")
	pdf.SetFont("Courier", "B", 12)
	pdf.Cell(60, 6, ": "+record.TicketNumber)

	pdf.SetX(110)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(25, 6, "Tanggal")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(60, 6, ": "+dateToIndonesian(record.WeighedAt))
	pdf.Ln(10)

	// --- Main Details ---
	pdf.SetY(115)

	// Helper to print label: value
	printRow := func(label, value string, xOffset float64) {
		if value == "" {
			value = "-"
		}
		pdf.SetX(xOffset)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(35, 6, label)
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(55, 6, ": "+value)
	}

	// Left Column
	printRow("Nomor Polisi", record.PlateNumber, 10)
	printRow("Supir", record.DriverName, 110)
	pdf.Ln(8)

	printRow("Perusahaan", record.CompanyName, 10)
	printRow("Operator", record.ManagerName, 110)
	pdf.Ln(8)

	printRow("Jenis Muatan", record.Product, 10)
	printRow("Status", record.Status, 110)
	pdf.Ln(15)

	// --- Weight Table ---
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(25, 109, 236)
	pdf.Cell(0, 10, "DATA PENIMBANGAN (Kg)")
	pdf.Ln(10)

	// Table Header
	pdf.SetFillColor(25, 109, 236)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetLineWidth(0.3)
	pdf.SetDrawColor(200, 200, 200)

	pdf.CellFormat(63, 10, "Berat Kotor (Gross)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(63, 10, "Berat Kosong (Tare)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(64, 10, "Berat Bersih (Netto)", "1", 1, "C", true, 0, "")

	// Table Content
	pdf.SetFillColor(255, 255, 255)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetFont("Courier", "B", 14)

	// Gross
	pdf.CellFormat(63, 15, fmt.Sprintf("%.0f", record.GrossWeight), "1", 0, "C", false, 0, "")
	// Tare
	pdf.CellFormat(63, 15, fmt.Sprintf("%.0f", record.TareWeight), "1", 0, "C", false, 0, "")
	// Net (Green Text)
	pdf.SetTextColor(0, 150, 0)
	pdf.CellFormat(64, 15, fmt.Sprintf("%.0f", record.NetWeight), "1", 1, "C", false, 0, "")

	// --- Signatures ---
	pdf.SetTextColor(51, 51, 51)
	pdf.Ln(25)

	ySig := pdf.GetY()

	// Box for signatures
	pdf.SetDrawColor(230, 230, 230)
	pdf.Rect(10, ySig, 190, 50, "D")

	// Driver Sig
	pdf.SetY(ySig + 5)
	pdf.SetX(20)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(80, 5, "Diserahkan Oleh (Supir),")

	// Manager Sig
	pdf.SetX(120)
	pdf.Cell(80, 5, "Diterima Oleh (Pengelola),")

	// Names
	pdf.SetY(ySig + 40)
	pdf.SetFont("Arial", "B", 10)

	dName := record.DriverName
	if dName == "" {
		dName = "-"
	}
	pdf.SetX(20)
	pdf.Cell(80, 5, "( "+dName+" )")
	pdf.Line(20, ySig+45, 80, ySig+45)

	mName := record.ManagerName
	if mName == "" {
		mName = "-"
	}
	pdf.SetX(120)
	pdf.Cell(80, 5, "( "+mName+" )")
	pdf.Line(120, ySig+45, 180, ySig+45)

	// --- Footer ---
	pdf.SetY(265)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.Cell(0, 4, "Dokumen ini dicetak secara komputerisasi dan sah tanpa cap basah.")
	pdf.Ln(4)
	pdf.Cell(0, 4, fmt.Sprintf("Dicetak pada: %s", dateToIndonesian(time.Now())))
	pdf.Ln(4)
	pdf.Cell(0, 4, "StoneWeigh System v1.0")

	// Ensure directory exists
	if _, err := os.Stat("web/static/reports"); os.IsNotExist(err) {
		os.MkdirAll("web/static/reports", 0755)
	}

	// Auto-print
	pdf.SetJavascript("this.print(true);")

	filename := fmt.Sprintf("web/static/reports/inv_%s.pdf", record.TicketNumber)
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
