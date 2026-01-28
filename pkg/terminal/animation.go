package terminal

import "github.com/fatih/color"

// HueToRGB converts HSV to RGB for rainbow gradient.
// hue should be in the range 0.0 to 1.0
func HueToRGB(hue float64) (uint8, uint8, uint8) {
	// hue is 0.0 to 1.0
	h := hue * 6.0
	x := 1.0 - abs(mod(h, 2.0)-1.0)

	var r, g, b float64
	switch int(h) {
	case 0:
		r, g, b = 1.0, x, 0.0
	case 1:
		r, g, b = x, 1.0, 0.0
	case 2:
		r, g, b = 0.0, 1.0, x
	case 3:
		r, g, b = 0.0, x, 1.0
	case 4:
		r, g, b = x, 0.0, 1.0
	default:
		r, g, b = 1.0, 0.0, x
	}

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

// MakeRainbow creates a rainbow gradient across a string using 24-bit RGB colors.
// offset animates the gradient (0.0 to 1.0+, wraps around)
func MakeRainbow(text string, offset float64) string {
	if len(text) == 0 {
		return text
	}

	var result string
	textLen := len(text)

	for i, ch := range text {
		// Calculate which color to use based on position
		// Stretch the gradient by scaling the position (smaller = more stretched)
		position := float64(i) / float64(textLen-1) * 0.3 // 0.3 = stretch factor
		if textLen == 1 {
			position = 0
		}

		// Add animated offset and wrap around (subtract to move right)
		animatedPosition := position - offset
		// Wrap to keep in 0.0-1.0 range (handle negative wrapping)
		for animatedPosition < 0 {
			animatedPosition += 1.0
		}
		animatedPosition = animatedPosition - float64(int(animatedPosition))

		// Convert position to RGB color (full rainbow spectrum)
		r, g, b := HueToRGB(animatedPosition)

		// Use color.RGB for 24-bit true color
		c := color.RGB(int(r), int(g), int(b))
		result += c.Sprint(string(ch))
	}

	return result
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func mod(x, y float64) float64 {
	return x - y*float64(int(x/y))
}

// SpinnerChars returns the Braille spinner characters used for terminal animations
var SpinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
