package reporting

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"
	"stoneweigh/internal/models"
)

// dateToIndonesian formats time to "02 Januari 2006 15:04"
func dateToIndonesian(t time.Time) string {
	months := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return fmt.Sprintf("%02d %s %d %02d:%02d", t.Day(), months[t.Month()], t.Year(), t.Hour(), t.Minute())
}

// GenerateInvoice creates a PDF invoice for a weighing transaction (Indonesian & Modern)
// It is designed to fit in the top half of an A4 page (approx 148mm height).
func GenerateInvoice(record models.WeighingRecord) (string, error) {
	log.Printf("Generating PDF for ticket %s", record.TicketNumber)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// --- Colors ---
	// Primary Amber: #d97706 -> 217, 119, 6
	// Light Background: #f6f7f8 -> 246, 247, 248
	// Dark Grey: #333333 -> 51, 51, 51

	primaryR, primaryG, primaryB := 217, 119, 6
	bgR, bgG, bgB := 246, 247, 248
	textR, textG, textB := 51, 51, 51

	// --- Header Section (Top Banner) ---
	// Background Banner
	pdf.SetFillColor(primaryR, primaryG, primaryB)
	pdf.Rect(0, 0, 210, 25, "F")

	// Company Name
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(255, 255, 255) // White
	pdf.SetXY(10, 5)
	pdf.Cell(0, 10, "PT. KAYA RAYA BAROKAH")

	// Subtitle / Address
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(10, 14)
	pdf.Cell(0, 5, "Lombok Timur, Nusa Tenggara Barat")
	pdf.Ln(4)
	pdf.Cell(0, 5, "Solusi Penimbangan Digital Terintegrasi")

	// --- Title Section ---
	pdf.SetY(30)
	pdf.SetTextColor(primaryR, primaryG, primaryB)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "SURAT JALAN / BUKTI TIMBANG")
	pdf.Ln(8)

	// --- Transaction Details (Grid) ---
	// We use a light box for details
	detailsTopY := pdf.GetY()

	// Box Background
	pdf.SetFillColor(bgR, bgG, bgB)
	pdf.SetDrawColor(primaryR, primaryG, primaryB)
	pdf.SetLineWidth(0.1)
	// Rect(x, y, w, h, style)
	// Height approx 35mm
	pdf.RoundedRect(10, detailsTopY, 190, 32, 2, "FD", "1234")

	pdf.SetTextColor(textR, textG, textB)
	pdf.SetY(detailsTopY + 3)

	// Helper to print label: value pair
	printPair := func(x, y float64, label, value string, isBoldValue bool) {
		pdf.SetXY(x, y)
		pdf.SetFont("Arial", "", 8)
		pdf.Cell(25, 5, label)

		pdf.SetXY(x+25, y)
		if isBoldValue {
			pdf.SetFont("Arial", "B", 9)
		} else {
			pdf.SetFont("Arial", "", 9)
		}
		pdf.Cell(60, 5, ": "+value)
	}

	// Prepare Data
	plate := record.PlateNumber
	if plate == "" { plate = "-" }
	driver := record.DriverName
	if driver == "" { driver = "-" }
	company := record.CompanyName
	if company == "" { company = "-" }
	product := record.Product
	if product == "" { product = "-" }

	// Column 1 (Left)
	col1X := 15.0
	rowHeight := 6.0
	currentY := detailsTopY + 2

	printPair(col1X, currentY, "No. Tiket", record.TicketNumber, true)
	printPair(col1X, currentY+rowHeight, "Tanggal", dateToIndonesian(record.WeighedAt), false)
	printPair(col1X, currentY+rowHeight*2, "Nomor Polisi", plate, true)
	printPair(col1X, currentY+rowHeight*3, "Supir", driver, false)

	// Column 2 (Right)
	col2X := 110.0
	printPair(col2X, currentY, "Perusahaan", company, false)
	printPair(col2X, currentY+rowHeight, "Produk/Material", product, false)
	printPair(col2X, currentY+rowHeight*2, "Status", record.Status, true)
	// Operator/Manager
	manager := record.ManagerName
	if manager == "" { manager = "-" }
	printPair(col2X, currentY+rowHeight*3, "Operator", manager, false)

	// --- Weight Table ---
	pdf.SetY(detailsTopY + 36)

	// Table Header
	pdf.SetFillColor(primaryR, primaryG, primaryB)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(200, 200, 200)

	pdf.CellFormat(63, 7, "Berat Kotor (Gross)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(63, 7, "Berat Kosong (Tare)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(64, 7, "Berat Bersih (Netto)", "1", 1, "C", true, 0, "")

	// Table Content
	pdf.SetFillColor(255, 255, 255)
	pdf.SetTextColor(textR, textG, textB)
	pdf.SetFont("Courier", "B", 12)

	grossStr := fmt.Sprintf("%.0f", record.GrossWeight)
	tareStr := fmt.Sprintf("%.0f", record.TareWeight)
	netStr := fmt.Sprintf("%.0f", record.NetWeight)

	// Gross
	pdf.CellFormat(63, 10, grossStr, "1", 0, "C", false, 0, "")
	// Tare
	pdf.CellFormat(63, 10, tareStr, "1", 0, "C", false, 0, "")
	// Net (Green Text for Netto)
	pdf.SetTextColor(0, 128, 0)
	pdf.CellFormat(64, 10, netStr, "1", 1, "C", false, 0, "")

	// --- Signatures ---
	pdf.SetTextColor(textR, textG, textB)
	pdf.Ln(4) // Small gap

	ySig := pdf.GetY()

	// Driver Sig
	pdf.SetY(ySig)
	pdf.SetX(20)
	pdf.SetFont("Arial", "", 8)
	pdf.Cell(60, 4, "Diserahkan (Supir),")

	// Manager Sig
	pdf.SetX(130)
	pdf.Cell(60, 4, "Diterima (Pengelola),")

	// Space for signature
	pdf.Ln(15)

	// Names
	pdf.SetFont("Arial", "B", 9)
	pdf.SetX(20)
	pdf.Cell(60, 4, "( "+driver+" )")

	pdf.SetX(130)
	pdf.Cell(60, 4, "( "+manager+" )")

	// --- Footer & Disclaimer ---
	pdf.SetY(138)
	pdf.SetFont("Arial", "I", 7)
	pdf.SetTextColor(150, 150, 150)
	pdf.Cell(0, 4, fmt.Sprintf("Dicetak: %s | ID: %s", dateToIndonesian(time.Now()), record.TicketNumber))
	pdf.Ln(3)
	pdf.Cell(0, 4, "Dokumen ini sah tanpa cap basah.")

	// --- Cut Line (Dashed) ---
	pdf.SetY(148) // Approx half page
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.5)
	pdf.SetDashPattern([]float64{3, 3}, 0)
	pdf.Line(0, 148, 210, 148)
	pdf.SetDashPattern([]float64{}, 0) // Reset

	// Ensure directory exists
	if _, err := os.Stat("web/static/reports"); os.IsNotExist(err) {
		os.MkdirAll("web/static/reports", 0755)
	}

	filename := fmt.Sprintf("web/static/reports/inv_%s.pdf", record.TicketNumber)
	// Auto print JS
	pdf.SetJavascript("this.print(true);")

	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
