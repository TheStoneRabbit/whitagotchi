package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/whitagotchi/whitagotchi/client/internal/api"
	"github.com/whitagotchi/whitagotchi/client/internal/config"
	"github.com/whitagotchi/whitagotchi/client/internal/render"
	"github.com/whitagotchi/whitagotchi/shared"
)

// Run launches the interactive TUI.
func Run(client *api.Client, cfg *config.Config) error {
	if cfg.Token == "" {
		return fmt.Errorf("not registered. run: whitagotchi register <username>")
	}
	m := newModel(client, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

const (
	animInterval    = 600 * time.Millisecond
	statusInterval  = 10 * time.Second
	maxChatMessages = 200
	inputWidth      = 78 // visible width of the input row, in runes
)

type mode int

const (
	modeHome mode = iota
	modePickPeer
	modeChat
)

type model struct {
	client *api.Client
	cfg    *config.Config

	creature *shared.Creature
	frame    int
	message  string
	err      error

	mode  mode
	input string // text being typed in pickpeer / chat

	peer     string
	peerInfo *shared.PeerInfoResponse
	conn     *api.ChatConn
	cancel   context.CancelFunc
	chatLog  []chatLine
}

type chatLine struct {
	from    string
	species shared.Species
	text    string
	self    bool
}

type tickAnim time.Time
type tickStatus time.Time
type statusMsg struct {
	c   *shared.Creature
	err error
}
type actionMsg struct {
	label string
	c     *shared.Creature
	err   error
}
type chatConnectedMsg struct {
	conn   *api.ChatConn
	cancel context.CancelFunc
	err    error
}
type chatRecvMsg struct {
	msg *shared.ChatOutbound
	err error
}
type peerInfoMsg struct {
	info *shared.PeerInfoResponse
	err  error
}

func newModel(client *api.Client, cfg *config.Config) *model {
	return &model{client: client, cfg: cfg}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.fetchStatus(), animTick(), statusTick())
}

func animTick() tea.Cmd {
	return tea.Tick(animInterval, func(t time.Time) tea.Msg { return tickAnim(t) })
}
func statusTick() tea.Cmd {
	return tea.Tick(statusInterval, func(t time.Time) tea.Msg { return tickStatus(t) })
}

func (m *model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.Status()
		if err != nil {
			return statusMsg{err: err}
		}
		return statusMsg{c: &resp.Creature}
	}
}

func (m *model) doAction(name string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.Action(name)
		if err != nil {
			return actionMsg{label: name, err: err}
		}
		return actionMsg{label: name, c: &resp.Creature}
	}
}

func (m *model) doReroll() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.Reroll()
		if err != nil {
			return actionMsg{label: "reroll", err: err}
		}
		return actionMsg{label: "new quirk: " + string(resp.Creature.Quirk), c: &resp.Creature}
	}
}

func (m *model) fetchPeer(name string) tea.Cmd {
	return func() tea.Msg {
		info, err := m.client.Peer(name)
		return peerInfoMsg{info: info, err: err}
	}
}

func (m *model) connectChat() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		conn, err := m.client.Chat(ctx)
		if err != nil {
			cancel()
			return chatConnectedMsg{err: err}
		}
		return chatConnectedMsg{conn: conn, cancel: cancel}
	}
}

func (m *model) recvChat() tea.Cmd {
	conn := m.conn
	if conn == nil {
		return nil
	}
	return func() tea.Msg {
		out, err := conn.Recv()
		return chatRecvMsg{msg: out, err: err}
	}
}

func (m *model) closeChat() {
	if m.conn != nil {
		_ = m.conn.Close()
		m.conn = nil
	}
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}

func (m *model) appendChat(line chatLine) {
	m.chatLog = append(m.chatLog, line)
	if len(m.chatLog) > maxChatMessages {
		m.chatLog = m.chatLog[len(m.chatLog)-maxChatMessages:]
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tickAnim:
		m.frame++
		return m, animTick()

	case tickStatus:
		return m, tea.Batch(m.fetchStatus(), statusTick())

	case statusMsg:
		if msg.err == nil {
			m.creature = msg.c
		} else if m.creature == nil {
			m.err = msg.err
		}
		return m, nil

	case actionMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("%s failed: %v", msg.label, msg.err)
		} else {
			m.message = msg.label + "'d!"
			m.creature = msg.c
		}
		return m, nil

	case chatConnectedMsg:
		if msg.err != nil {
			m.message = "chat connect failed: " + msg.err.Error()
			m.mode = modeHome
			return m, nil
		}
		m.conn = msg.conn
		m.cancel = msg.cancel
		m.mode = modeChat
		m.appendChat(chatLine{from: "system", text: "connected. talking to " + m.peer})
		return m, tea.Batch(m.recvChat(), m.fetchPeer(m.peer))

	case peerInfoMsg:
		if msg.err != nil {
			m.appendChat(chatLine{from: "system", text: "couldn't load peer info: " + msg.err.Error()})
			return m, nil
		}
		m.peerInfo = msg.info
		return m, nil

	case chatRecvMsg:
		if msg.err != nil {
			m.appendChat(chatLine{from: "system", text: "disconnected: " + msg.err.Error()})
			m.closeChat()
			return m, nil
		}
		if msg.msg != nil {
			m.appendChat(chatLine{
				from:    msg.msg.From,
				species: msg.msg.FromSpecies,
				text:    msg.msg.Paraphrased,
			})
		}
		return m, m.recvChat()
	}
	return m, nil
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeHome:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "f":
			m.message = "feeding..."
			return m, m.doAction("feed")
		case "p":
			m.message = "playing..."
			return m, m.doAction("play")
		case "b":
			m.message = "bathing..."
			return m, m.doAction("bathe")
		case "r":
			m.message = "refreshing..."
			return m, m.fetchStatus()
		case "R":
			m.message = "rerolling quirk..."
			return m, m.doReroll()
		case "c":
			m.mode = modePickPeer
			m.input = ""
			m.message = ""
			return m, nil
		}

	case modePickPeer:
		switch msg.String() {
		case "esc":
			m.mode = modeHome
			m.input = ""
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			peer := strings.TrimSpace(m.input)
			if peer == "" {
				return m, nil
			}
			m.peer = peer
			m.input = ""
			m.message = "connecting to " + peer + "..."
			return m, m.connectChat()
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil
		default:
			if isPrintable(msg) {
				m.input += msg.String()
			}
			return m, nil
		}

	case modeChat:
		switch msg.String() {
		case "esc":
			m.closeChat()
			m.peerInfo = nil
			m.mode = modeHome
			m.input = ""
			return m, nil
		case "ctrl+c":
			m.closeChat()
			return m, tea.Quit
		case "enter":
			text := strings.TrimSpace(m.input)
			m.input = ""
			if text == "" || m.conn == nil {
				return m, nil
			}
			if err := m.conn.Send(m.peer, text); err != nil {
				m.appendChat(chatLine{from: "system", text: "send failed: " + err.Error()})
				return m, nil
			}
			m.appendChat(chatLine{from: m.cfg.Username, text: text, self: true})
			return m, nil
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil
		default:
			if isPrintable(msg) {
				m.input += msg.String()
			}
			return m, nil
		}
	}
	return m, nil
}

func isPrintable(msg tea.KeyMsg) bool {
	if msg.Type != tea.KeyRunes && msg.Type != tea.KeySpace {
		return false
	}
	return true
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213"))
	helpStyle  = lipgloss.NewStyle().Faint(true)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	msgStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	frameStyle = lipgloss.NewStyle().Padding(0, 2)
	selfStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	peerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	sysStyle   = lipgloss.NewStyle().Faint(true).Italic(true)
	inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	chatBox    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	petMini    = lipgloss.NewStyle().Padding(0, 1).Faint(true)
)

func (m *model) View() string {
	if m.err != nil && m.creature == nil {
		return errStyle.Render("error: "+m.err.Error()) + "\n\n" + helpStyle.Render("press q to quit")
	}
	if m.creature == nil {
		return helpStyle.Render("loading...")
	}
	switch m.mode {
	case modeChat:
		return m.viewChat()
	case modePickPeer:
		return m.viewPickPeer()
	default:
		return m.viewHome()
	}
}

func (m *model) viewHome() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("whitagotchi") + "\n\n")
	b.WriteString(frameStyle.Render(render.CreatureFrame(m.creature, m.frame)))
	b.WriteString("\n")
	if m.message != "" {
		b.WriteString(msgStyle.Render(m.message) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[f] feed  [p] play  [b] bathe  [c] chat  [R] reroll quirk  [r] refresh  [q] quit"))
	return b.String()
}

func (m *model) viewPickPeer() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("whitagotchi · new chat") + "\n\n")
	b.WriteString("who do you want to chat with?\n\n")
	b.WriteString(renderInput(m.input) + "\n\n")
	b.WriteString(helpStyle.Render("[enter] connect  [esc] back"))
	return b.String()
}

func (m *model) viewChat() string {
	header := titleStyle.Render("whitagotchi · chat with " + m.peer)

	// Our pet card
	youCard := petCard("you", m.cfg.Username, m.creature.Species, m.creature.Stage, m.creature.Quirk,
		render.Frame(m.creature.Species, m.creature.Stage, m.frame))

	// Peer pet card (loading state if not yet fetched)
	var peerCard string
	if m.peerInfo != nil {
		peerCard = petCard("them", m.peerInfo.Username, m.peerInfo.Species, m.peerInfo.Stage, m.peerInfo.Quirk,
			render.Frame(m.peerInfo.Species, m.peerInfo.Stage, m.frame))
	} else {
		peerCard = petMini.Render("(loading " + m.peer + "...)")
	}

	pets := lipgloss.JoinHorizontal(lipgloss.Top, youCard, "   ", peerCard)

	// Chat log
	var log strings.Builder
	for _, l := range m.chatLog {
		switch {
		case l.from == "system":
			log.WriteString(sysStyle.Render("· "+l.text) + "\n")
		case l.self:
			log.WriteString(selfStyle.Render(l.from+": ") + l.text + "\n")
		default:
			tag := l.from
			if l.species != "" {
				tag = fmt.Sprintf("%s the %s", l.from, l.species)
			}
			log.WriteString(peerStyle.Render(tag+": ") + l.text + "\n")
		}
	}

	body := chatBox.Width(80).Height(12).Render(log.String())
	input := renderInput(m.input)
	footer := helpStyle.Render("[enter] send  [esc] back to pet  [ctrl+c] quit")

	return header + "\n\n" + pets + "\n\n" + body + "\n" + input + "\n\n" + footer
}

// renderInput shows the input field with a horizontal scroll window. When the
// typed text is longer than inputWidth, the visible tail follows the caret and
// an ellipsis marks the truncated head.
func renderInput(s string) string {
	const prefix, caret = "> ", "_"
	runes := []rune(s)
	avail := inputWidth - len([]rune(prefix)) - len([]rune(caret))
	if avail < 1 {
		avail = 1
	}
	if len(runes) <= avail {
		return inputStyle.Render(prefix + string(runes) + caret)
	}
	// Reserve one rune for the leading ellipsis.
	visible := runes[len(runes)-(avail-1):]
	return inputStyle.Render(prefix + "…" + string(visible) + caret)
}

func petCard(role, name string, sp shared.Species, st shared.Stage, q shared.Quirk, art string) string {
	label := fmt.Sprintf("%s · %s the %s (%s)\nquirk: %s", role, name, sp, st, q)
	return petMini.Render(art + "\n" + label)
}
