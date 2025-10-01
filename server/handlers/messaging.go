package handlers

import (
	"fmt"
	"net"
	"strings"

	"chat-server/server/models"
)

func (h *CommandHandler) handlePrivateMessage(conn net.Conn, sender *models.Client, cmd string) {
	content := strings.TrimPrefix(cmd, "/msg ")
	parts := strings.SplitN(content, " ", 2)

	if len(parts) < 2 {
		conn.Write([]byte(ColorRed + "Usage: /msg <username> <message>\n" + ColorReset))
		return
	}

	targetName := parts[0]
	message := parts[1]

	target := h.ClientManager.GetClientByUsername(targetName)
	if target == nil {
		conn.Write([]byte(ColorRed + "User not found.\n" + ColorReset))
		return
	}

	targetMsg := fmt.Sprintf("%s[DM]%s %s%s%s %s—»%s You\n  %s╰─>%s %s\n",
		ColorMagenta, ColorReset, ColorCyan, sender.Username, ColorReset,
		ColorMagenta, ColorReset, ColorCyan, ColorReset, message)
	target.Conn.Write([]byte(targetMsg))

	senderMsg := fmt.Sprintf("%s[DM]%s You %s—»%s %s%s%s\n  %s╰─>%s %s\n",
		ColorMagenta, ColorReset, ColorMagenta, ColorReset,
		ColorCyan, targetName, ColorReset, ColorCyan, ColorReset, message)
	sender.Conn.Write([]byte(senderMsg))
}

func (h *CommandHandler) handleTagCommand(conn net.Conn, sender *models.Client, cmd string) {
	content := strings.TrimPrefix(cmd, "/tag ")
	parts := strings.SplitN(content, " ", 2)

	if len(parts) < 2 {
		conn.Write([]byte(ColorRed + "Usage: /tag <username> <message>\n" + ColorReset))
		return
	}

	targetName := parts[0]
	message := parts[1]

	target := h.ClientManager.GetClientByUsername(targetName)
	if target == nil {
		conn.Write([]byte(ColorRed + "User not found.\n" + ColorReset))
		return
	}

	fullMessage := fmt.Sprintf("@%s: %s", targetName, message)
	h.LobbyManager.StoreMessage(sender.CurrentLobby, sender.UserProfile, sender.Username, fullMessage)

	taggedMsg := fmt.Sprintf("%s%s %s%s @%s%s%s\n  %s╰─>%s %s\n",
		ColorYellow, sender.UserProfile, ColorCyan, sender.Username,
		ColorMagenta, targetName, ColorReset, ColorCyan, ColorReset, message)

	h.ClientManager.BroadcastToLobby(sender.CurrentLobby, taggedMsg)

	if target.Conn != nil && target.Username != sender.Username {
		notification := fmt.Sprintf("%s✦ %s tagged you%s\n",
			ColorMagenta, sender.Username, ColorReset)
		target.Conn.Write([]byte(notification))
	}
}
