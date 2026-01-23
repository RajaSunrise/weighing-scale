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

// GenerateInvoice creates a PDF invoice for a weighing transaction
func GenerateInvoice(record models.WeighingRecord) (string, error) {
	log.Printf("Generating PDF for ticket %s", record.TicketNumber)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Colors
	primaryR, primaryG, primaryB := 30, 58, 95             // Navy
	accentR, accentG, accentB := 184, 134, 11              // Gold
	textR, textG, textB := 44, 62, 80                      // Dark text
	borderR, borderG, borderB := 200, 200, 200             // Lighter border
	mediumGreyR, mediumGreyG, mediumGreyB := 100, 100, 100 // Medium grey

	// --- Header Section ---
	// White background for header
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(0, 0, 210, 30, "F")

	// Add logo
	pdf.Image("web/static/images/logo-invoice.png", 15, 5, 30, 20, false, "", 0, "")

	// --- Title Section ---
	pdf.SetY(36)
	pdf.SetTextColor(primaryR, primaryG, primaryB)
	pdf.SetFont("Arial", "B", 15)
	pdf.CellFormat(0, 8, "SURAT JALAN / BUKTI TIMBANG", "", 1, "C", false, 0, "")

	// Decorative underline
	pdf.SetDrawColor(accentR, accentG, accentB)
	pdf.SetLineWidth(0.5)
	lineY := pdf.GetY() - 1
	pdf.Line(70, lineY, 140, lineY)
	pdf.Ln(4)

	plate := record.PlateNumber
	if plate == "" {
		plate = "-"
	}
	driver := record.DriverName
	if driver == "" {
		driver = "-"
	}
	company := record.CompanyName
	if company == "" {
		company = "-"
	}
	product := record.Product
	if product == "" {
		product = "-"
	}
	manager := record.ManagerName
	if manager == "" {
		manager = "-"
	}

	// Koordinat awal tabel
	startX := 10.0
	pdf.SetX(startX)

	// Konfigurasi ukuran sel
	colLabelW := 30.0
	colValW := 65.0
	rowH := 7.5

	// Fungsi helper untuk menggambar satu baris grid (2 kolom data)
	drawRow := func(lbl1, val1, lbl2, val2 string, isBoldVal1, isBoldVal2 bool, val2ColorOverride []int) {
		// Set Border Color
		pdf.SetDrawColor(borderR, borderG, borderB)
		pdf.SetLineWidth(0.2)
		pdf.SetFillColor(255, 255, 255)

		// --- Kolom Kiri ---
		// Label
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(mediumGreyR, mediumGreyG, mediumGreyB)
		pdf.CellFormat(colLabelW, rowH, " "+lbl1, "LTB", 0, "L", false, 0, "")

		// Value
		if isBoldVal1 {
			pdf.SetFont("Arial", "B", 9)
			pdf.SetTextColor(0, 0, 0)
		} else {
			pdf.SetFont("Arial", "", 9)
			pdf.SetTextColor(textR, textG, textB)
		}
		pdf.CellFormat(colValW, rowH, " "+val1, "RTB", 0, "L", false, 0, "")

		// --- Kolom Kanan ---
		// Label
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(mediumGreyR, mediumGreyG, mediumGreyB)
		pdf.CellFormat(colLabelW, rowH, " "+lbl2, "LTB", 0, "L", false, 0, "")

		// Value
		if val2ColorOverride != nil {
			// Custom color (misal untuk Status)
			pdf.SetTextColor(val2ColorOverride[0], val2ColorOverride[1], val2ColorOverride[2])
			pdf.SetFont("Arial", "B", 9)
		} else if isBoldVal2 {
			pdf.SetFont("Arial", "B", 9)
			pdf.SetTextColor(0, 0, 0)
		} else {
			pdf.SetFont("Arial", "", 9)
			pdf.SetTextColor(textR, textG, textB)
		}
		pdf.CellFormat(colValW, rowH, " "+val2, "RTB", 1, "L", false, 0, "") // '1' untuk ganti baris
	}

	// Baris 1: No. Tiket | Perusahaan
	drawRow("No. Tiket", record.TicketNumber, "Perusahaan", company, true, false, nil)

	// Baris 2: Tanggal | Produk
	drawRow("Tanggal", dateToIndonesian(record.WeighedAt), "Produk", product, false, false, nil)

	// Baris 3: No. Polisi | Status
	// Tentukan warna status (Tanpa background, hanya warna teks)
	statusColor := []int{234, 179, 8} // Yellow/Orange default
	if record.Status == "SELESAI" {
		statusColor = []int{34, 197, 94} // Green
	}
	drawRow("No. Polisi", plate, "Status", record.Status, true, true, statusColor)

	// Baris 4: Supir | Operator
	drawRow("Supir", driver, "Operator", manager, false, false, nil)

	// --- Weight Table (Enhanced) ---
	// Posisikan tabel berat sedikit di bawah tabel detail
	pdf.Ln(5)

	// Table header
	pdf.SetFillColor(primaryR, primaryG, primaryB)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetDrawColor(primaryR, primaryG, primaryB)
	pdf.SetLineWidth(0.1)

	headerHeight := 8.0
	// Menggunakan startX yang sama agar lurus dengan grid atas
	pdf.SetX(startX)

	// Hitung lebar total agar sama dengan grid atas (30+65+30+65 = 190)
	totalW := (colLabelW + colValW) * 2
	colW := totalW / 3.0 // Bagi 3 kolom rata

	pdf.CellFormat(colW, headerHeight, "BERAT KOTOR", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW, headerHeight, "BERAT KOSONG", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW, headerHeight, "BERAT BERSIH", "1", 1, "C", true, 0, "")

	// Sub-header
	pdf.SetX(startX)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(mediumGreyR, mediumGreyG, mediumGreyB)
	pdf.CellFormat(colW, 4, "(Gross Weight)", "LR", 0, "C", true, 0, "")
	pdf.CellFormat(colW, 4, "(Tare Weight)", "LR", 0, "C", true, 0, "")
	pdf.CellFormat(colW, 4, "(Net Weight)", "LR", 1, "C", true, 0, "")

	// Isi Tabel Berat
	pdf.SetX(startX)
	pdf.SetDrawColor(borderR, borderG, borderB)

	grossStr := fmt.Sprintf("%.0f kg", record.GrossWeight)
	tareStr := fmt.Sprintf("%.0f kg", record.TareWeight)
	netStr := fmt.Sprintf("%.0f kg", record.NetWeight)

	pdf.SetFont("Courier", "B", 14)

	// Gross - Hitam
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(colW, 10, grossStr, "1", 0, "C", false, 0, "")

	// Tare - Hitam
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(colW, 10, tareStr, "1", 0, "C", false, 0, "")

	// Net - Hijau (Tanpa background box, hanya teks)
	pdf.SetTextColor(21, 128, 61)
	pdf.CellFormat(colW, 10, netStr, "1", 1, "C", false, 0, "")

	// --- Signatures Section ---
	pdf.SetTextColor(textR, textG, textB)
	pdf.Ln(5)

	ySig := pdf.GetY()

	// Supir Box
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.SetLineWidth(0.5)
	pdf.Rect(15, ySig, 80, 20, "D") // Kotak tanda tangan diperkecil sedikit tingginya agar muat

	pdf.SetXY(15, ySig+2)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(mediumGreyR, mediumGreyG, mediumGreyB)
	pdf.CellFormat(80, 4, "Diserahkan oleh (Supir)", "", 1, "C", false, 0, "")

	pdf.SetXY(15, ySig+15)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(80, 4, driver, "", 0, "C", false, 0, "")

	// Manager Box
	pdf.Rect(115, ySig, 80, 20, "D")
	pdf.SetXY(115, ySig+2)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(mediumGreyR, mediumGreyG, mediumGreyB)
	pdf.CellFormat(80, 4, "Diterima oleh (Pengelola)", "", 1, "C", false, 0, "")

	pdf.SetXY(115, ySig+15)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(80, 4, manager, "", 0, "C", false, 0, "")

	// --- Footer ---
	// Posisikan footer tepat di atas garis potong
	pdf.SetY(138)
	pdf.SetFont("Arial", "I", 7)
	pdf.SetTextColor(mediumGreyR, mediumGreyG, mediumGreyB)
	pdf.CellFormat(0, 3, fmt.Sprintf("Dicetak: %s | Dokumen: %s", dateToIndonesian(time.Now()), record.TicketNumber), "", 1, "C", false, 0, "")

	// --- Decorative Cut Line (Fixed at 148mm) ---
	pdf.SetY(148)
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.3)
	pdf.SetDashPattern([]float64{5, 3}, 0)
	pdf.Line(0, 148, 210, 148)

	// Gunting Text
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.SetXY(85, 146.5)
	pdf.SetFillColor(255, 255, 255) // Background putih untuk teks gunting agar tidak tertimpa garis
	pdf.CellFormat(40, 3, " POTONG DI SINI ", "", 0, "C", true, 0, "")

	pdf.SetDashPattern([]float64{}, 0)

	// Ensure directory exists
	if _, err := os.Stat("web/static/reports"); os.IsNotExist(err) {
		os.MkdirAll("web/static/reports", 0755)
	}

	filename := fmt.Sprintf("web/static/reports/inv_%s.pdf", record.TicketNumber)

	// Optional: Auto print
	// pdf.SetJavascript("this.print(true);")

	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
