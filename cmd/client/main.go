package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	viewport viewport.Model
	textarea textarea.Model
	messages []string
	conn     net.Conn
}

type msgReceived string

func initialModel(conn net.Conn) model {
	ta := textarea.New()
	ta.Placeholder = "Type message..."
	ta.Focus()
	ta.SetHeight(1)
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(80, 20)
	
	return model{
		viewport: vp,
		textarea: ta,
		messages: []string{},
		conn:     conn,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, waitForMessage(m.conn))
}

func waitForMessage(conn net.Conn) tea.Cmd {
	return func() tea.Msg {
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			return msgReceived(scanner.Text())
		}
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			text := strings.TrimSpace(m.textarea.Value())
			if text != "" {
				m.conn.Write([]byte(text + "\n"))
				m.textarea.Reset()
			}
			return m, nil
		}

	case msgReceived:
		m.messages = append(m.messages, string(msg))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, waitForMessage(m.conn)

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		m.textarea.SetWidth(msg.Width)
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	return fmt.Sprintf(
		"%s\n\n%s\n\nCtrl+C to quit",
		style.Render(m.viewport.View()),
		m.textarea.View(),
	)
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Cannot connect:", err)
	}
	defer conn.Close()

	p := tea.NewProgram(initialModel(conn), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}