package handlers

import (
	"fmt"
	"net"
	"strings"

	"chat-server/server/models"
)

var profilePics = map[string]string{
	"default":  "[@_@]",
	"cat":      "(=^･^=)",
	"dog":      "(ᵔᴥᵔ)",
	"cool":     "(⌐■_■)",
	"bear":     "ʕ•ᴥ•ʔ",
	"happy":    "(◕‿◕)",
	"star":     "☆彡",
	"fire":     "(🔥)",
	"alien":    "[👽]",
	"robot":    "[▪‿▪]",
	"love":     "(♥‿♥)",
	"wink":     "(^_~)",
	"dead":     "(x_x)",
	"shrug":    "¯\\_(ツ)_/¯",
	"music":    "♪(┌・。・)┌",
	"ninja":    "[忍]",
	"king":     "(♔‿♔)",
	"queen":    "(♕‿♕)",
	"devil":    "(ψ｀∇´)ψ",
	"angel":    "(◕ᴗ◕✿)",
	"sleep":    "(-.-)zzZ",
	"cry":      "(╥﹏╥)",
	"laugh":    "(≧▽≦)",
	"angry":    "(╬ಠ益ಠ)",
	"confused": "(・_・ヾ",
	"shocked":  "(°ロ°)",
	"peace":    "(✌ﾟ∀ﾟ)☞",
	"skull":    "[☠]",
	"heart":    "[❤]",
	"coffee":   "c[_]",
	"pizza":    "[🍕]",
	"ghost":    "(ー'`ー)",
	"fox":      "ᓚᘏᗢ",
	"owl":      "(◉Θ◉)",
	"penguin":  "(°<°)",
	"frog":     "( ･ั﹏･ั)",
	"bunny":    "(\\(•ᴗ•)/)",
	"snake":    "~>°)~~~",
	"dino":     "<コ:彡",
	"wizard":   "⊂(◉‿◉)つ",
	"pirate":   "(✪‿✪)ノ",
	"nerd":     "(⌐□_□)",
	"party":    "ヽ(^o^)ノ",
	"think":    "(¬‿¬)",
	"flex":     "ᕦ(ò_óˇ)ᕤ",
	"dance":    "┏(･o･)┛",
	"flip":     "(ノಠ益ಠ)ノ彡┻━┻",
}

func (h *CommandHandler) handleSetProfile(conn net.Conn, client *models.Client, cmd string) {
	content := strings.TrimSpace(strings.TrimPrefix(cmd, "/sp"))

	if content == "" || content == "default" {
		client.UserProfile = profilePics["default"]
		conn.Write([]byte(ColorGreen + "Profile picture reset to default.\n" + ColorReset))
		return
	}

	if content == "list" {
		showProfilePics(conn)
		return
	}

	pic, exists := profilePics[content]
	if !exists {
		conn.Write([]byte(ColorRed + "Profile picture not found. Use /sp list to see options.\n" + ColorReset))
		return
	}

	client.UserProfile = pic
	conn.Write([]byte(ColorGreen + fmt.Sprintf("Profile picture changed to: %s\n", pic) + ColorReset))
}

func showProfilePics(conn net.Conn) {
	msg := ColorCyan + "\n=== Available Profile Pictures ===\n" + ColorReset
	msg += ColorYellow + "Usage: /sp <name>\n\n" + ColorReset
	for name, pic := range profilePics {
		msg += fmt.Sprintf("  %s%-10s%s → %s\n", ColorWhite, name, ColorReset, pic)
	}
	msg += "\n"
	conn.Write([]byte(msg))
}
