package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	ws "github.com/bubbletrack/server/internal/tui/websocket"
)

var debugKey = os.Getenv("DEBUG_KEY") == "true"

type Tab int

const (
	ChatTab Tab = iota
	GraphTab
)

type ViewState int

const (
	LoadingState ViewState = iota
	ReadyState
	ErrorState
	EmptyState
)

// keyMap defines keyboard shortcuts for the help panel.
type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Tab        key.Binding
	Enter      key.Binding
	NewSession key.Binding
	Quit       key.Binding
	Sort       key.Binding
	Refresh    key.Binding
	Help       key.Binding
}

var defaultKeys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch tab"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send"),
	),
	NewSession: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "new session"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "ctrl+q", "esc"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.NewSession, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab},
		{k.Enter, k.NewSession, k.Help, k.Quit},
	}
}

type model struct {
	graphRepo GraphRepository
	chatRepo  ChatRepository
	userID    string
	log       *slog.Logger
	wsClient  *ws.Client

	activeTab Tab
	graphData *GraphData
	viewState ViewState
	err       error

	selected  int
	sortMode  SortMode
	sortNames []string

	layout Layout

	width  int
	height int

	// Bubbles components
	chatTextInput textinput.Model
	chatViewport  viewport.Model
	spinner       spinner.Model
	helpModel     help.Model
	relTable      table.Model
	keys          keyMap

	chatMessages []ChatMessage
	chatStatus   string
	chatStatusOK bool
	chatLoading  bool
	showHelp     bool
}

type ChatMessage struct {
	ID        string
	Sender    string
	Content   string
	Timestamp string
	IsUser    bool
}

type GraphData struct {
	Nodes []Node
	Edges []Edge
	Stats GraphStats
}

type Node struct {
	ID           string
	Name         string
	Role         string
	Mood         string
	Energy       float64
	LastTopic    string
	InteractCount int
}

type Edge struct {
	Source       string
	Target       string
	Quality     string
	Strength    float64
	Protocol    string
	SourceW     float64
	TargetW     float64
	ReciprocityIndex float64
}

type GraphRepository interface {
	GetGraph(userID string) (*GraphData, error)
}

type ChatRepository interface {
	SaveMessage(userID, sender, content string, isUser bool) error
	LoadMessages(userID string, limit int) ([]ChatMessage, error)
}

type ChatService interface {
	SendMessage(sender, content, sessionID string) error
	GetMessages(sessionID string, limit int) ([]ChatMessage, error)
}

func newBubblesComponents() (textinput.Model, viewport.Model, spinner.Model, help.Model, table.Model) {
	ti := textinput.New()
	ti.Placeholder = "Type a message and press Enter..."
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(T().Title).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(T().MainFg)
	ti.CharLimit = 500
	ti.Width = 50

	vp := viewport.New(50, 10)
	vp.KeyMap.Up.SetEnabled(false)
	vp.KeyMap.Down.SetEnabled(false)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(T().Subtitle)

	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(T().Subtitle)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(T().MutedText)
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(T().Subtitle)
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(T().MutedText)

	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "Source", Width: 12},
			{Title: "Target", Width: 12},
			{Title: "Quality", Width: 10},
			{Title: "Strength", Width: 10},
			{Title: "Reciprocity", Width: 11},
		}),
		table.WithFocused(false),
		table.WithHeight(8),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Bold(true).Foreground(T().Title)
	s.Selected = s.Selected.Foreground(T().SelectedFg).Background(T().SelectedBg).Bold(true)
	tbl.SetStyles(s)

	return ti, vp, sp, h, tbl
}

func NewModel(graphRepo GraphRepository, chatRepo ChatRepository, userID string, log *slog.Logger) model {
	ti, vp, sp, h, tbl := newBubblesComponents()
	ti.Focus()

	return model{
		graphRepo:     graphRepo,
		chatRepo:      chatRepo,
		userID:        userID,
		log:           log,
		activeTab:     ChatTab,
		viewState:     LoadingState,
		sortMode:      SortByName,
		sortNames:     []string{"name", "quality", "strength"},
		chatTextInput: ti,
		chatViewport:  vp,
		spinner:       sp,
		helpModel:     h,
		relTable:      tbl,
		keys:          defaultKeys,
		chatMessages: []ChatMessage{
			{Sender: "BubbleTrack", Content: "Welcome to BubbleTrack! Ask me anything about your social graph.", Timestamp: "Now", IsUser: false},
		},
		chatStatus:   "Ready",
		chatStatusOK: true,
	}
}

func NewModelWithChatService(graphRepo GraphRepository, service ChatService, userID string, log *slog.Logger, wsClient *ws.Client) model {
	ti, vp, sp, h, tbl := newBubblesComponents()
	ti.Focus()

	return model{
		graphRepo: graphRepo,
		chatRepo: &chatServiceAdapter{
			service: service,
			userID:  userID,
		},
		userID:        userID,
		log:           log,
		wsClient:      wsClient,
		activeTab:     ChatTab,
		viewState:     LoadingState,
		sortMode:      SortByName,
		sortNames:     []string{"name", "quality", "strength"},
		chatTextInput: ti,
		chatViewport:  vp,
		spinner:       sp,
		helpModel:     h,
		relTable:      tbl,
		keys:          defaultKeys,
		chatMessages:  []ChatMessage{{Sender: "BubbleTrack", Content: "Welcome to BubbleTrack! Ask me anything about your social graph.", Timestamp: "Now", IsUser: false}},
		chatStatus:    "Ready",
		chatStatusOK:  true,
	}
}

type chatServiceAdapter struct {
	service ChatService
	userID  string
}

func (a *chatServiceAdapter) SaveMessage(userID, sender, content string, isUser bool) error {
	_ = userID
	_ = isUser
	return a.service.SendMessage(sender, content, "")
}

func (a *chatServiceAdapter) LoadMessages(userID string, limit int) ([]ChatMessage, error) {
	_ = userID
	return a.service.GetMessages("", limit)
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{spinner.Tick}
	if m.graphRepo != nil {
		cmds = append(cmds, loadGraph(m.graphRepo, m.userID))
	}
	if m.chatRepo != nil {
		cmds = append(cmds, loadChatMessages(m.chatRepo, m.userID))
	}
	if m.wsClient != nil {
		cmds = append(cmds, connectWebSocket(m.wsClient))
	}
	return tea.Batch(cmds...)
}

func loadGraph(repo GraphRepository, userID string) tea.Cmd {
	return func() tea.Msg {
		data, err := repo.GetGraph(userID)
		return graphLoadedMsg{data: data, err: err}
	}
}

func loadChatMessages(repo ChatRepository, userID string) tea.Cmd {
	return func() tea.Msg {
		messages, err := repo.LoadMessages(userID, 50)
		if err != nil {
			return chatMessagesLoadedMsg{err: err}
		}
		return chatMessagesLoadedMsg{messages: messages}
	}
}

func sendChatMessage(repo ChatRepository, userID, content string) tea.Cmd {
	return func() tea.Msg {
		err := repo.SaveMessage(userID, "You", content, true)
		if err != nil {
			return chatMessageSentMsg{err: err}
		}
		return chatMessageSentMsg{
			message: ChatMessage{
				ID:        fmt.Sprintf("pending-%d", time.Now().UnixNano()),
				Sender:    "You",
				Content:   content,
				IsUser:    true,
				Timestamp: time.Now().Format("15:04"),
			},
		}
	}
}

func connectWebSocket(wsClient *ws.Client) tea.Cmd {
	return func() tea.Msg {
		if err := wsClient.Connect(context.Background()); err != nil {
			return wsConnectionErrorMsg{err: err}
		}
		return wsConnectedMsg{}
	}
}

func receiveWebSocketMessage(wsClient *ws.Client) tea.Cmd {
	return func() tea.Msg {
		data, ok := <-wsClient.Receive()
		if !ok {
			return wsDisconnectedMsg{}
		}

		var wsMsg ws.ChatMessage
		if err := json.Unmarshal(data, &wsMsg); err != nil {
			return wsConnectionErrorMsg{err: err}
		}

		timeLabel := wsMsg.CreatedAt
		if ts, parseErr := time.Parse(time.RFC3339Nano, wsMsg.CreatedAt); parseErr == nil {
			timeLabel = ts.Local().Format("15:04")
		}

		chatMsg := ChatMessage{
			Sender:    wsMsg.Sender,
			Content:   wsMsg.Content,
			Timestamp: timeLabel,
			IsUser:    wsMsg.IsUser,
		}
		return wsMessageReceivedMsg{message: chatMsg}
	}
}

type graphLoadedMsg struct {
	data *GraphData
	err  error
}

type chatMessagesLoadedMsg struct {
	messages []ChatMessage
	err      error
}

type chatMessageSentMsg struct {
	message ChatMessage
	err     error
}

type wsConnectedMsg struct{}
type wsDisconnectedMsg struct{}
type wsMessageReceivedMsg struct {
	message ChatMessage
}
type wsConnectionErrorMsg struct {
	err error
}

type pollChatMessagesMsg struct{}

func pollChatMessages(repo ChatRepository, userID string, knownCount int) tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		messages, err := repo.LoadMessages(userID, 50)
		if err != nil {
			return pollChatMessagesMsg{}
		}
		if len(messages) > knownCount {
			return chatMessagesLoadedMsg{messages: messages}
		}
		return pollChatMessagesMsg{}
	})
}



func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		if m.activeTab != ChatTab {
			m.activeTab = ChatTab
		}
		return m, nil
	case graphLoadedMsg:
		m.viewState = ReadyState
		if msg.err != nil {
			m.viewState = ErrorState
			m.err = msg.err
		} else if msg.data == nil || (len(msg.data.Nodes) == 0 && len(msg.data.Edges) == 0) {
			m.viewState = EmptyState
		} else {
			m.graphData = msg.data
			m.updateRelTable()
		}
		return m, nil
	case chatMessagesLoadedMsg:
		if msg.err == nil {
			if len(msg.messages) > 0 {
				dedup := make(map[string]bool)
				var clean []ChatMessage
				for _, m := range msg.messages {
					key := m.Sender + "|" + m.Content + "|" + m.Timestamp
					if !dedup[key] {
						dedup[key] = true
						clean = append(clean, m)
					}
				}
				m.chatMessages = clean
			}
			m.updateViewportContent()
			m.chatStatus = fmt.Sprintf("Chat (%d)", len(m.chatMessages))
			m.chatStatusOK = true
		} else {
			m.chatStatus = "Load failed"
			m.chatStatusOK = false
		}
		return m, nil
	case chatMessageSentMsg:
		if msg.err == nil {
			key := msg.message.Sender + "|" + msg.message.Content + "|" + msg.message.Timestamp
			exists := false
			for _, m := range m.chatMessages {
				if m.Sender+"|"+m.Content+"|"+m.Timestamp == key {
					exists = true
					break
				}
			}
			if !exists {
				m.chatMessages = append(m.chatMessages, msg.message)
			}
			m.updateViewportContent()
			m.chatStatus = "Sent"
			m.chatStatusOK = true
			m.chatLoading = true
			if m.wsClient == nil || !m.wsClient.IsConnected() {
				return m, tea.Batch(pollChatMessages(m.chatRepo, m.userID, len(m.chatMessages)), m.spinner.Tick)
			}
			return m, m.spinner.Tick
		} else {
			m.chatStatus = "Send failed"
			m.chatStatusOK = false
		}
		return m, nil
	case wsConnectedMsg:
		m.chatStatus = "Connected (WS)"
		m.chatStatusOK = true
		go m.wsClient.Run(context.Background())
		return m, receiveWebSocketMessage(m.wsClient)
	case wsDisconnectedMsg:
		m.chatStatus = "Disconnected"
		m.chatStatusOK = false
		return m, nil
	case wsMessageReceivedMsg:
		m.chatMessages = append(m.chatMessages, msg.message)
		m.updateViewportContent()
		m.chatStatus = "Live"
		m.chatStatusOK = true
		if !msg.message.IsUser {
			m.chatLoading = false
		}
		return m, receiveWebSocketMessage(m.wsClient)
	case wsConnectionErrorMsg:
		m.chatStatus = "WS Error"
		m.chatStatusOK = false
		if m.wsClient.IsConnected() {
			m.wsClient.Close()
		}
		return m, nil
	case pollChatMessagesMsg:
		if m.chatLoading {
			return m, tea.Batch(pollChatMessages(m.chatRepo, m.userID, len(m.chatMessages)), m.spinner.Tick)
		}
		return m, pollChatMessages(m.chatRepo, m.userID, len(m.chatMessages))
	case spinner.TickMsg:
		if m.viewState == LoadingState || m.chatLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout = CalcLayout(m.width, m.height)
		m.chatTextInput.Width = min(m.width-10, 70)
		m.chatViewport.Width = min(m.width-4, 78)
		m.chatViewport.Height = m.height - 8
		if m.chatViewport.Height < 5 {
			m.chatViewport.Height = 5
		}
		m.updateViewportContent()
		m.relTable.SetWidth(min(m.width-4, 60))
		m.relTable.SetHeight(m.height - 8)
	}
	return m, nil
}

func (m *model) updateViewportContent() {
	if m.chatViewport.Width == 0 {
		return
	}
	var lines []string
	for _, msg := range m.chatMessages {
		lines = append(lines, msgLine(msg))
	}
	content := strings.Join(lines, "\n")
	wrapped := lipgloss.NewStyle().Width(m.chatViewport.Width).Render(content)
	m.chatViewport.SetContent(wrapped)
	m.chatViewport.GotoBottom()
}

func (m *model) updateRelTable() {
	if m.graphData == nil {
		return
	}
	sorted := SortEdges(m.graphData.Edges, m.sortMode)
	rows := make([]table.Row, len(sorted))
	for i, e := range sorted {
		rows[i] = table.Row{
			e.Source,
			e.Target,
			e.Quality,
			fmt.Sprintf("%.0f%%", e.Strength*100),
			fmt.Sprintf("%.2f", e.ReciprocityIndex),
		}
	}
	m.relTable.SetRows(rows)
	if m.selected >= len(rows) {
		m.selected = max(0, len(rows)-1)
	}
	m.relTable.SetCursor(m.selected)
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q", "esc":
		return m, tea.Quit
	case "ctrl+s":
		m.chatMessages = nil
		m.chatTextInput.Reset()
		m.updateViewportContent()
		m.chatStatus = "New session"
		m.chatStatusOK = true
		return m, nil
	case "tab":
		if m.activeTab == ChatTab {
			m.activeTab = GraphTab
		} else {
			m.activeTab = ChatTab
		}
		return m, nil
	case "shift+tab":
		if m.activeTab == GraphTab {
			m.activeTab = ChatTab
		} else {
			m.activeTab = GraphTab
		}
		return m, nil
	case "1":
		if m.activeTab != ChatTab {
			m.activeTab = ChatTab
		}
		return m, nil
	case "2":
		if m.activeTab == GraphTab {
			return m, nil
		}
		if m.activeTab != ChatTab {
			m.activeTab = GraphTab
		}
		return m, nil
	case "enter":
		if m.activeTab == ChatTab && m.chatTextInput.Value() != "" {
			content := m.chatTextInput.Value()
			m.chatTextInput.Reset()
			m.chatStatus = "Sending..."
			m.chatStatusOK = true
			if m.chatRepo != nil {
				return m, sendChatMessage(m.chatRepo, m.userID, content)
			} else {
				m.chatMessages = append(m.chatMessages, ChatMessage{
					Sender:    "You",
					Content:   content,
					IsUser:    true,
					Timestamp: time.Now().Format("15:04"),
				})
				m.updateViewportContent()
				m.chatStatus = "Sent (local)"
				m.chatStatusOK = true
			}
		}
	}

	if m.activeTab == GraphTab {
		return m.handleGraphKey(msg)
	}
	return m.handleChatKey(msg)
}

func (m model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chatTextInput, cmd = m.chatTextInput.Update(msg)
	return m, cmd
}

func (m model) handleGraphKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		m.viewState = LoadingState
		m.err = nil
		return m, tea.Batch(loadGraph(m.graphRepo, m.userID), m.spinner.Tick)
	case "s":
		m.sortMode = (m.sortMode + 1) % 3
		m.updateRelTable()
		return m, nil
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	}
	var cmd tea.Cmd
	m.relTable, cmd = m.relTable.Update(msg)
	m.selected = m.relTable.Cursor()
	return m, cmd
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	if m.width < 40 || m.height < 10 {
		return centerText(
			ErrorStyle().Render("Terminal too small")+"\n"+
				MutedStyle().Render("Minimum: 40x10"),
			m.width, m.height,
		)
	}

	tabs := []TabItem{
		{Name: "Chat", Active: m.activeTab == ChatTab},
		{Name: "Graph", Active: m.activeTab == GraphTab},
	}
	tabBar := RenderTabBar(tabs, m.width)

	h := m.height
	if h > 40 {
		h = 40
	}

	contentHeight := h - 4
	if contentHeight < 3 {
		contentHeight = 3
	}

	var content string
	if m.activeTab == ChatTab {
		content = m.renderChatTab(contentHeight)
	} else {
		switch m.viewState {
		case LoadingState:
			content = m.renderLoading(contentHeight)
		case ErrorState:
			content = m.renderError(contentHeight)
		case EmptyState:
			content = m.renderEmpty(contentHeight)
		default:
			content = m.renderGraphTab(contentHeight)
		}
	}

	statusBar := m.renderStatusBar()

	return tabBar + "\n" + content + "\n" + statusBar
}

func (m model) renderLoading(h int) string {
	return centerText(m.spinner.View()+" Loading graph data...", m.width, h)
}

func (m model) renderError(h int) string {
	errText := ErrorStyle().Render(fmt.Sprintf("Error: %s", m.err.Error()))
	hint := MutedStyle().Render("r: Retry")
	content := lipgloss.JoinVertical(lipgloss.Left, " ", errText, " ", hint)
	box := SimpleBox("", content, T().ErrorFg)
	return centerText(box, m.width, h)
}

func (m model) renderEmpty(h int) string {
	title := SubtitleStyle().Render("No relationships found")
	hint := MutedStyle().Render("Submit interactions via the API to build your social graph.")
	content := lipgloss.JoinVertical(lipgloss.Center, " ", title, " ", hint, " ")
	box := SimpleBox("", content, T().Border)
	return centerText(box, m.width, h)
}

func (m model) renderGraphTab(h int) string {
	if m.graphData == nil {
		return ""
	}

	sorted := SortEdges(m.graphData.Edges, m.sortMode)
	if m.selected >= len(sorted) {
		m.selected = max(0, len(sorted)-1)
	}

	highlightNodes := make(map[string]bool)
	if m.selected < len(sorted) {
		highlightNodes[sorted[m.selected].Source] = true
		highlightNodes[sorted[m.selected].Target] = true
	}

	canvasWidth := m.width - 4
	if canvasWidth < 20 {
		canvasWidth = 20
	}
	canvasHeight := h - 2
	if canvasHeight < 5 {
		canvasHeight = 5
	}

	graph := RenderASCIIGraph(m.graphData, canvasWidth-4, canvasHeight/2-2, highlightNodes)

	graphBox := RenderBox(Box{
		Title:       "Graph",
		Content:     graph,
		BorderColor: T().Border,
		Rounded:     true,
		Width:       canvasWidth,
		Height:      canvasHeight / 2,
	})

	personPanel := m.renderPersonPanel(canvasWidth)

	tableBox := RenderBox(Box{
		Title:       fmt.Sprintf("Relationships (sort: %s)", m.sortNames[m.sortMode]),
		Content:     m.relTable.View(),
		BorderColor: T().TableBox,
		Rounded:     true,
		Width:       canvasWidth,
		Height:      canvasHeight / 2,
	})

	graphSection := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, graphBox, personPanel),
	)

	all := lipgloss.JoinVertical(lipgloss.Left, graphSection, tableBox)

	if m.showHelp {
		helpView := m.helpModel.View(m.keys)
		all = lipgloss.JoinVertical(lipgloss.Left, all, helpView)
	}

	return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(all)
}

func (m model) renderPersonPanel(width int) string {
	if m.graphData == nil || len(m.graphData.Nodes) == 0 {
		return SimpleBox("People", MutedStyle().Render("No people found"), T().Border)
	}

	panelWidth := min(30, width/3)
	if panelWidth < 20 {
		panelWidth = 20
	}

	var lines []string
	for _, n := range m.graphData.Nodes {
		name := n.Name
		if len(name) > 10 {
			name = name[:9] + "…"
		}
		role := ""
		if n.Role != "" {
			role = strings.ToUpper(n.Role[:1])
		}
		mood := moodChar(n.Mood)
		energy := energyBar(n.Energy)
		lines = append(lines, fmt.Sprintf(" %s %s %s %s", name, role, mood, energy))
	}

	content := strings.Join(lines, "\n")
	return RenderBox(Box{
		Title:       "People",
		Content:     content,
		BorderColor: T().Border,
		Rounded:     true,
		Width:       panelWidth,
	})
}

func moodChar(mood string) string {
	switch strings.ToLower(mood) {
	case "happy", "grateful":
		return "+"
	case "anxious", "angry":
		return "!"
	case "tired", "sad", "lonely":
		return "-"
	case "energized", "hopeful":
		return "*"
	default:
		return "."
	}
}

func energyBar(energy float64) string {
	if energy <= 0 {
		return "..."
	}
	filled := int(energy * 3)
	if filled > 3 {
		filled = 3
	}
	return strings.Repeat("=", filled) + strings.Repeat(".", 3-filled)
}

func (m model) renderChatTab(h int) string {
	vp := m.chatViewport.View()

	statusStyle := SuccessStyle()
	if !m.chatStatusOK {
		statusStyle = ErrorStyle()
	}
	status := statusStyle.Render(m.chatStatus)
	if m.chatLoading {
		status = MutedStyle().Render(m.spinner.View() + " Analyzing...")
	}

	input := lipgloss.JoinVertical(lipgloss.Left,
		m.chatTextInput.View(),
		status,
	)

	panelWidth := min(70, m.width-6)
	if panelWidth < 30 {
		panelWidth = m.width - 2
	}

	inputBox := RenderBox(Box{
		Title:       "CHAT",
		Content:     input,
		BorderColor: T().Border,
		Rounded:     true,
		Width:       panelWidth,
	})

	all := lipgloss.JoinVertical(lipgloss.Left, vp, "\n", lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(inputBox))
	return lipgloss.NewStyle().Width(m.width).Render(all)
}

func msgLine(msg ChatMessage) string {
	sender := msg.Sender
	if msg.IsUser {
		sender = "You"
	}
	preview := msg.Content
	if len(preview) > 50 {
		preview = preview[:50] + "..."
	}
	return fmt.Sprintf("%s: %s", NormalStyle().Render(sender), preview)
}





func (m model) renderStatusBar() string {
	t := T()

	helpView := m.helpModel.ShortHelpView(m.keys.ShortHelp())

	left := lipgloss.NewStyle().
		Foreground(t.StatusBarFg).
		Background(t.StatusBarBg).
		Render(helpView)

	version := lipgloss.NewStyle().
		Foreground(t.StatusBarFg).
		Background(t.StatusBarBg).
		Render("BubbleTrack")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(version)
	if gap < 1 {
		gap = 1
	}

	spacer := lipgloss.NewStyle().
		Foreground(t.StatusBarFg).
		Background(t.StatusBarBg).
		Width(gap).
		Render(strings.Repeat(" ", gap))

	return left + spacer + version
}

func centerText(text string, width, height int) string {
	textWidth := lipgloss.Width(text)
	xPad := (width - textWidth) / 2
	if xPad < 0 {
		xPad = 0
	}

	lines := strings.Split(text, "\n")
	yPad := (height - len(lines)) / 2
	if yPad < 0 {
		yPad = 0
	}

	var b strings.Builder
	for i := 0; i < yPad; i++ {
		b.WriteString("\n")
	}
	for _, line := range lines {
		b.WriteString(strings.Repeat(" ", xPad))
		b.WriteString(line)
		b.WriteString("\n")
	}

	result := b.String()
	remaining := height - yPad - len(lines)
	for i := 0; i < remaining; i++ {
		b.WriteString("\n")
	}

	return result
}
