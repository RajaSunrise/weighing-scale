package reporting

import (
	"os"
	"testing"
	"time"

	"stoneweigh/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGenerateInvoice(t *testing.T) {
	// Setup test data
	record := models.WeighingRecord{
		TicketNumber: "T-TEST-001",
		PlateNumber:  "B 1234 TEST",
		DriverName:   "Test Driver",
		CompanyName:  "Test Corp",
		Product:      "Test Product",
		GrossWeight:  20000,
		TareWeight:   5000,
		NetWeight:    15000,
		ManagerName:  "Test Manager",
		Status:       "COMPLETED",
		WeighedAt:    time.Now(),
	}

	// Ensure output directory exists (GenerateInvoice does this, but for test cleanup we might need to know)
	// GenerateInvoice writes to web/static/reports
	// We might want to temporarily redirect output or just clean up.
	// Since GenerateInvoice hardcodes path, we rely on it creating the folder.

	// Create temp output dir if strictly needed, but func creates it.

	path, err := GenerateInvoice(record)

	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.FileExists(t, path)

	// Cleanup
	defer os.Remove(path)
}

func TestDateToIndonesian(t *testing.T) {
	// 2023-01-01 12:00
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	formatted := dateToIndonesian(fixedTime)

	// "01 Januari 2023 12:00"
	// Note: time.Date with UTC might format hour differently if machine is local?
	// dateToIndonesian uses t.Hour(), so it uses whatever is in t.

	assert.Equal(t, "01 Januari 2023 12:00", formatted)
}
