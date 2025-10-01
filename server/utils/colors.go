package utils

const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
	ColorPurple  = "\033[95m"
	ColorNeon    = "\033[96m"
	ColorGold    = "\033[93m"
	Bold         = "\033[1m"
	Dim          = "\033[2m"
	Blink        = "\033[5m"
	Reverse      = "\033[7m"
	Underline    = "\033[4m"
)

func BuildColor(text, color string) string {
	switch color {
	case "green":
		return ColorGreen + text + ColorReset
	case "red":
		return ColorRed + text + ColorReset
	case "yellow":
		return ColorYellow + text + ColorReset
	default:
		return text
	}
}
