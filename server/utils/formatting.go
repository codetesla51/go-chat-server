package utils

import (
	"fmt"
	"time"
)

// FormatTimeAgo returns a human-readable time string
func FormatTimeAgo(t time.Time) string {
	elapsed := time.Since(t)

	if elapsed < time.Second {
		return "just now"
	} else if elapsed < time.Minute {
		seconds := int(elapsed.Seconds())
		if seconds == 1 {
			return "1s ago"
		}
		return fmt.Sprintf("%ds ago", seconds)
	} else if elapsed < time.Hour {
		minutes := int(elapsed.Minutes())
		if minutes == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", minutes)
	} else if elapsed < 24*time.Hour {
		hours := int(elapsed.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	} else {
		days := int(elapsed.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}

// FormatMessage formats a chat message for display
func FormatMessage(senderProfile, username, text, colorYellow, colorWhite, colorCyan, colorReset string, timestamp time.Time) string {
	timeAgo := FormatTimeAgo(timestamp)
	return fmt.Sprintf("%s%s %s%s [%s%s%s]\n  %s╰─>%s %s\n",
		colorYellow,
		senderProfile,
		username,
		colorReset,
		colorWhite,
		timeAgo,
		colorReset,
		colorCyan,
		colorReset,
		text,
	)
}
