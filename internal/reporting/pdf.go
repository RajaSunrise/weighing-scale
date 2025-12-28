package reporting

import (
	"fmt"
	"log"
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
	log.Printf("Generating PDF for ticket %s", record.TicketNumber)
	log.Printf("  - PlateNumber: '%s'", record.PlateNumber)
	log.Printf("  - DriverName: '%s'", record.DriverName)
	log.Printf("  - CompanyName: '%s'", record.CompanyName)
	log.Printf("  - Product: '%s'", record.Product)
	log.Printf("  - GrossWeight: %.2f", record.GrossWeight)
	log.Printf("  - TareWeight: %.2f", record.TareWeight)
	log.Printf("  - NetWeight: %.2f", record.NetWeight)

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
	pdf.Cell(0, 10, "Timbang Batu Lombok")

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(10, 18)
	pdf.Cell(0, 5, "Solusi Penimbangan Digital Terintegrasi")

	// Reset Position
	pdf.SetY(40)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, "Jalan Tambang Raya No. 123, Lombok Barat")
	pdf.Ln(5)
	pdf.Cell(0, 5, "Telp: 087805815285 | Email: indraaryadi@gmail.com")
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
		pdf.SetX(xOffset)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(35, 6, label)
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(55, 6, ": "+value)
	}

	// Get values with defaults
	plate := record.PlateNumber
	if plate == "" || len(plate) == 0 {
		plate = "-"
	}
	driver := record.DriverName
	if driver == "" || len(driver) == 0 {
		driver = "-"
	}
	company := record.CompanyName
	if company == "" {
		company = "-"
	}
	manager := record.ManagerName
	if manager == "" {
		manager = "-"
	}
	product := record.Product
	if product == "" {
		product = "-"
	}

	// Left Column
	printRow("Nomor Polisi", plate, 10)
	printRow("Supir", driver, 110)
	pdf.Ln(8)

	printRow("Perusahaan", company, 10)
	printRow("Operator", manager, 110)
	pdf.Ln(8)

	printRow("Jenis Muatan", product, 10)
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

	grossStr := fmt.Sprintf("%.0f", record.GrossWeight)
	if record.GrossWeight == 0 {
		grossStr = "0"
	}
	tareStr := fmt.Sprintf("%.0f", record.TareWeight)
	if record.TareWeight == 0 {
		tareStr = "0"
	}
	netStr := fmt.Sprintf("%.0f", record.NetWeight)
	if record.NetWeight == 0 {
		netStr = "0"
	}

	// Gross
	pdf.CellFormat(63, 15, grossStr, "1", 0, "C", false, 0, "")
	// Tare
	pdf.CellFormat(63, 15, tareStr, "1", 0, "C", false, 0, "")
	// Net (Green Text)
	pdf.SetTextColor(0, 150, 0)
	pdf.CellFormat(64, 15, netStr, "1", 1, "C", false, 0, "")

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
	pdf.Cell(80, 5, "Diserahkan Ke (Supir),")

	// Manager Sig
	pdf.SetX(120)
	pdf.Cell(80, 5, "Diterima Oleh (Pengelola),")

	// Names
	pdf.SetY(ySig + 40)
	pdf.SetFont("Arial", "B", 10)

	driverSig := record.DriverName
	if driverSig == "" {
		driverSig = "-"
	}
	pdf.SetX(20)
	pdf.Cell(80, 5, "( "+driverSig+" )")
	pdf.Line(20, ySig+45, 80, ySig+45)

	pdf.SetX(120)
	pdf.Cell(80, 5, "( "+record.ManagerName+" )")
	pdf.Line(120, ySig+45, 180, ySig+45)

	// --- Footer ---
	pdf.SetY(265)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.Cell(0, 4, "Dokumen ini dicetak secara komputerisasi dan sah tanpa cap basah.")
	pdf.Ln(4)
	pdf.Cell(0, 4, fmt.Sprintf("Dicetak pada: %s", dateToIndonesian(time.Now())))

	// Ensure directory exists
	if _, err := os.Stat("web/static/reports"); os.IsNotExist(err) {
		os.MkdirAll("web/static/reports", 0755)
	}

	filename := fmt.Sprintf("web/static/reports/inv_%s.pdf", record.TicketNumber)
	pdf.SetJavascript("this.print(true);")
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
