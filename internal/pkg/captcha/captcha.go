package captcha

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
)

// Store is the captcha storage
// Use DefaultMemStore to keep things simple
var Store = base64Captcha.DefaultMemStore

// GenerateCaptcha generates a new captcha
func GenerateCaptcha() (string, string, error) {
	// Configure driver (height, width, length, maxSkew, dotCount)
	// Larger width to accommodate digits
	driver := base64Captcha.NewDriverDigit(80, 240, 6, 0.7, 80)

	// If you want string (letters + numbers), use this:
	// driver := base64Captcha.NewDriverString(80, 240, 0, 0, 6, "1234567890abcdefghijklmnopqrstuvwxyz", &color.RGBA{0, 0, 0, 0}, nil, []string{"wqy-microhei.ttc"})

	c := base64Captcha.NewCaptcha(driver, Store)
	id, b64s, _, err := c.Generate()
	return id, b64s, err
}

// GenerateCustomCaptcha allows more customization if needed
func GenerateCustomCaptcha() (string, string, error) {
    // Width: 240, Height: 80, Length: 5, MaxSkew: 0.7, DotCount: 80
	// ShowHollowLine -> false or true
	driver := base64Captcha.NewDriverMath(80, 240, 0, base64Captcha.OptionShowHollowLine, &color.RGBA{0, 0, 0, 0}, nil, []string{"wqy-microhei.ttc"})
	c := base64Captcha.NewCaptcha(driver, Store)
    id, b64s, _, err := c.Generate()
	return id, b64s, err
}

// VerifyCaptcha verifies the captcha solution
func VerifyCaptcha(id string, answer string) bool {
	return Store.Verify(id, answer, true)
}
