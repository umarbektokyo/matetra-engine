package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
	"github.com/umarbektokyo/matetra-engine/model"
)

// ━━ Splash ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const splash = `                         $$\                $$\
                         $$ |               $$ |
$$$$$$\$$$$\   $$$$$$\ $$$$$$\    $$$$$$\ $$$$$$\    $$$$$$\  $$$$$$\
$$  _$$  _$$\  \____$$\\_$$  _|  $$  __$$\\_$$  _|  $$  __$$\ \____$$\
$$ / $$ / $$ | $$$$$$$ | $$ |    $$$$$$$$ | $$ |    $$ |  \__|$$$$$$$ |
$$ | $$ | $$ |$$  __$$ | $$ |$$\ $$   ____| $$ |$$\ $$ |     $$  __$$ |
$$ | $$ | $$ |\$$$$$$$ | \$$$$  |\$$$$$$$\  \$$$$  |$$ |     \$$$$$$$ |
\__| \__| \__| \_______|  \____/  \_______|  \____/ \__|      \_______|`

// ━━ Palette ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

var (
	fg    = lipgloss.Color("#c0c0c0") // main text
	dim   = lipgloss.Color("#606060") // secondary
	vdim  = lipgloss.Color("#383838") // very dim
	acc   = lipgloss.Color("#a0a0a0") // accent
	hi    = lipgloss.Color("#ffffff") // highlight
	err_  = lipgloss.Color("#cc6666") // error
	ok_   = lipgloss.Color("#8fbc8f") // ok
	warn_ = lipgloss.Color("#d4aa50") // warning
	hl    = lipgloss.Color("#e8d44d") // selection highlight
	imm   = lipgloss.Color("#6b9f6b") // immune
	fib   = lipgloss.Color("#9b7db8") // fibonacci
)

// ━━ Styles ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

var (
	sSplash = lipgloss.NewStyle().Foreground(dim)
	sLabel  = lipgloss.NewStyle().Foreground(acc)
	sDim    = lipgloss.NewStyle().Foreground(dim)
	sVDim   = lipgloss.NewStyle().Foreground(vdim)
	sHi     = lipgloss.NewStyle().Foreground(hi).Bold(true)
	sErr    = lipgloss.NewStyle().Foreground(err_)
	sOk     = lipgloss.NewStyle().Foreground(ok_)
	sWarn   = lipgloss.NewStyle().Foreground(warn_)
	sTitle  = lipgloss.NewStyle().Foreground(acc).Bold(true)

	sNum    = lipgloss.NewStyle().Foreground(hi).Bold(true)
	sNul    = lipgloss.NewStyle().Foreground(vdim)
	sImm    = lipgloss.NewStyle().Foreground(imm).Bold(true)
	sFib    = lipgloss.NewStyle().Foreground(fib).Bold(true)
	sNumHl  = lipgloss.NewStyle().Background(lipgloss.Color("#444")).Foreground(hl).Bold(true)

	sBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333")).
		Padding(0, 1)

	sBoxActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666")).
			Padding(0, 1)

	sBoxHl = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(hl).
		Padding(0, 1)

	sCard = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333")).
		Padding(0, 1).
		Width(28).
		Height(3)

	sCardSel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#888")).
			Padding(0, 1).
			Width(28).
			Height(3)

	sField = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444")).
		Padding(0, 1)

	sFieldActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#888")).
			Padding(0, 1)

	sBar = lipgloss.NewStyle().
		Foreground(fg).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1)
)

// ━━ Wire types ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type wsMsg struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
type srvMsg struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
type replyMsg struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
type wsRx srvMsg
type wsErrMsg struct{ err error }
type authOkMsg struct {
	conn     *websocket.Conn
	username string
	token    string
	first    srvMsg // first message from server (WELCOME or AUTO_REJOINED)
}

// ━━ Screens / Modes ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type scr int

const (
	scrLogin scr = iota
	scrRoom
	scrCreate
	scrLobby
	scrGame
)

type gmode int

const (
	gmNormal  gmode = iota
	gmPlay          // interactive card input
	gmDice          // dice
	gmCommand       // : command
)

// ━━ Model ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type M struct {
	scr      scr
	addr     string
	tok      string
	user     string
	conn     *websocket.Conn
	gs       *model.GameState
	pid      int
	room     string
	errMsg   string
	status   string
	logs     []string
	w, h     int

	// login
	lIn  []textinput.Model
	lF   int

	// room
	rIn textinput.Model

	// create
	cIn    []textinput.Model
	cF     int  // 0=win cond picker, 1=hand size, 2=turn limit
	cWin   int

	// game
	gm      gmode
	sel     int // index into hand
	cmd     textinput.Model
	dIn     textinput.Model

	// play mode
	pCard    int               // card index in gs.Cards
	pFields  []textinput.Model // only user-fillable fields
	pFocus   int
	pReq     string            // full InputsReq
	pUserIdx []int             // which positions in pReq are user-fillable
	pAuto    []int             // auto-filled values (full input array)
}

func newM(addr string) M {
	u := textinput.New(); u.Placeholder = "username"; u.Focus(); u.CharLimit = 20; u.Width = 30
	p := textinput.New(); p.Placeholder = "password"; p.EchoMode = textinput.EchoPassword; p.CharLimit = 50; p.Width = 30
	r := textinput.New(); r.Placeholder = "room code (empty=create)"; r.CharLimit = 6; r.Width = 30
	h := textinput.New(); h.Placeholder = "hand size (6)"; h.CharLimit = 3; h.Width = 20
	t := textinput.New(); t.Placeholder = "turn limit"; t.CharLimit = 5; t.Width = 20
	c := textinput.New(); c.Placeholder = "command"; c.CharLimit = 60; c.Width = 40
	d := textinput.New(); d.Placeholder = "1-6"; d.CharLimit = 1; d.Width = 6

	return M{
		scr: scrLogin, addr: addr, pid: -1,
		lIn: []textinput.Model{u, p},
		rIn: r, cIn: []textinput.Model{h, t},
		cmd: c, dIn: d, w: 100, h: 40,
	}
}

func (m M) Init() tea.Cmd { return textinput.Blink }

// ━━ Update ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m M) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.conn != nil { m.conn.Close() }
			return m, tea.Quit
		}
	case authOkMsg:
		m.conn, m.user, m.tok = msg.conn, msg.username, msg.token
		// Process the first server message (could be WELCOME or AUTO_REJOINED)
		return m.onSrv(wsRx(msg.first))
	case wsRx:
		return m.onSrv(msg)
	case wsErrMsg:
		m.errMsg = msg.err.Error()
		return m, nil
	}

	switch m.scr {
	case scrLogin:  return m.uLogin(msg)
	case scrRoom:   return m.uRoom(msg)
	case scrCreate: return m.uCreate(msg)
	case scrLobby:  return m.uLobby(msg)
	case scrGame:   return m.uGame(msg)
	}
	return m, nil
}

// ── Login ───────────────────────────────────────────────────────────────

func (m M) uLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "tab", "shift+tab":
			m.lF = (m.lF + 1) % 2
			for i := range m.lIn { if i == m.lF { m.lIn[i].Focus() } else { m.lIn[i].Blur() } }
			return m, nil
		case "enter":
			u, p := m.lIn[0].Value(), m.lIn[1].Value()
			if u == "" || p == "" { m.errMsg = "need username + password"; return m, nil }
			m.errMsg = ""; m.status = "connecting..."
			return m, m.doAuth(u, p)
		}
	}
	var cmds []tea.Cmd
	for i := range m.lIn { var c tea.Cmd; m.lIn[i], c = m.lIn[i].Update(msg); cmds = append(cmds, c) }
	return m, tea.Batch(cmds...)
}

// ── Room ────────────────────────────────────────────────────────────────

func (m M) uRoom(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "enter" {
		code := strings.TrimSpace(strings.ToUpper(m.rIn.Value()))
		if code == "" { m.scr = scrCreate; m.cIn[0].Focus(); return m, nil }
		m.tx("JOIN_ROOM", map[string]string{"code": code}); m.status = "joining..."
		return m, nil
	}
	var c tea.Cmd; m.rIn, c = m.rIn.Update(msg); return m, c
}

// ── Create ──────────────────────────────────────────────────────────────

func (m M) uCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "tab":
			// Cycle: win picker -> hand size -> turn limit -> win picker
			maxF := 1; if m.cWin > 0 { maxF = 2 }
			m.cF = (m.cF + 1) % (maxF + 1)
			m.cIn[0].Blur(); m.cIn[1].Blur()
			if m.cF == 1 { m.cIn[0].Focus() }
			if m.cF == 2 { m.cIn[1].Focus() }
			return m, nil
		case "shift+tab":
			maxF := 1; if m.cWin > 0 { maxF = 2 }
			m.cF--; if m.cF < 0 { m.cF = maxF }
			m.cIn[0].Blur(); m.cIn[1].Blur()
			if m.cF == 1 { m.cIn[0].Focus() }
			if m.cF == 2 { m.cIn[1].Focus() }
			return m, nil
		case "up":
			if m.cF == 0 && m.cWin > 0 { m.cWin-- }
			return m, nil
		case "down":
			if m.cF == 0 && m.cWin < 2 { m.cWin++ }
			return m, nil
		case "enter":
			hs := 6; if v := m.cIn[0].Value(); v != "" { if n, e := strconv.Atoi(v); e == nil && n > 0 { hs = n } }
			tl := 0; if v := m.cIn[1].Value(); v != "" { if n, e := strconv.Atoi(v); e == nil && n > 0 { tl = n } }
			m.tx("CREATE_ROOM", map[string]any{"winCondition": m.cWin, "turnLimit": tl, "handSize": hs})
			return m, nil
		case "esc":
			m.scr = scrRoom; m.rIn.Focus(); return m, nil
		}
	}
	// Only pass input to the focused text field
	if m.cF == 1 {
		var c tea.Cmd; m.cIn[0], c = m.cIn[0].Update(msg); return m, c
	}
	if m.cF == 2 {
		var c tea.Cmd; m.cIn[1], c = m.cIn[1].Update(msg); return m, c
	}
	return m, nil
}

// ── Lobby ───────────────────────────────────────────────────────────────

func (m M) uLobby(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "s": m.tx("START_GAME", nil); return m, nil
		case "j": m.tx("ADD_PLAYER", nil); return m, nil
		}
	}
	return m, nil
}

// ── Game ────────────────────────────────────────────────────────────────

func (m M) uGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.gm {
	case gmNormal:  return m.uNormal(msg)
	case gmPlay:    return m.uPlay(msg)
	case gmDice:    return m.uDice(msg)
	case gmCommand: return m.uCmd(msg)
	}
	return m, nil
}

func (m M) uNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok { return m, nil }
	hand := m.hand()
	switch k.String() {
	case "left", "h":
		if m.sel > 0 { m.sel-- }
	case "right", "l":
		if m.sel < len(hand)-1 { m.sel++ }
	case "enter":
		if len(hand) > 0 && m.sel >= 0 && m.sel < len(hand) {
			m.startPlay(hand[m.sel])
		}
	case "d":
		m.gm = gmDice; m.dIn.SetValue(""); m.dIn.Focus()
		m.status = "dice: enter=roll one  a=fill all  esc=cancel"
	case "e":
		m.tx("END_TURN", nil); m.status = ""
	case "f":
		m.tx("FINISH_DICE", nil); m.status = ""
	case ":":
		m.gm = gmCommand; m.cmd.SetValue(""); m.cmd.Focus()
	}
	return m, nil
}

func (m *M) startPlay(cardIdx int) {
	if m.gs == nil { return }
	card := m.gs.Cards[cardIdx]
	req := card.InputsReq

	m.gm = gmPlay
	m.pCard = cardIdx
	m.pReq = req
	m.pFocus = 0
	m.pFields = nil
	m.pUserIdx = nil
	m.errMsg = ""

	// Build the full auto array and figure out which fields need user input
	cp := m.gs.Turn % len(m.gs.Players)
	m.pAuto = make([]int, len(req))

	for i, c := range req {
		switch byte(c) {
		case 'A':
			m.pAuto[i] = cp // auto: defending player
		case 'U':
			m.pAuto[i] = m.pid // auto: card owner
		case 'd':
			m.pAuto[i] = rand.Intn(6) + 1 // auto: random dice
		default:
			// 'n', 'p', 'c', 'X', 'Y', 'i' — user must fill
			m.pAuto[i] = -999 // sentinel
			m.pUserIdx = append(m.pUserIdx, i)
		}
	}

	if len(m.pUserIdx) == 0 {
		// No user input needed, fire now
		m.txPlay(cardIdx, m.pAuto)
		m.gm = gmNormal
		return
	}

	// Create text inputs only for user-fillable fields
	for _, idx := range m.pUserIdx {
		ti := textinput.New()
		ti.CharLimit = 4
		ti.Width = 4
		ti.Placeholder = reqPlaceholder(req[idx])
		m.pFields = append(m.pFields, ti)
	}
	m.pFields[0].Focus()
}

func (m M) uPlay(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc":
			m.gm = gmNormal; m.errMsg = ""; return m, nil
		case "tab", "right":
			if m.pFocus < len(m.pFields)-1 {
				m.pFields[m.pFocus].Blur(); m.pFocus++; m.pFields[m.pFocus].Focus()
			}
			return m, nil
		case "shift+tab", "left":
			if m.pFocus > 0 {
				m.pFields[m.pFocus].Blur(); m.pFocus--; m.pFields[m.pFocus].Focus()
			}
			return m, nil
		case "enter":
			// Collect inputs
			full := make([]int, len(m.pAuto))
			copy(full, m.pAuto)
			for fi, ri := range m.pUserIdx {
				v, err := strconv.Atoi(strings.TrimSpace(m.pFields[fi].Value()))
				if err != nil {
					m.errMsg = fmt.Sprintf("field %d: need a number", fi+1)
					return m, nil
				}
				full[ri] = v
			}
			m.txPlay(m.pCard, full)
			m.gm = gmNormal; m.errMsg = ""
			return m, nil
		}
	}
	if m.pFocus < len(m.pFields) {
		var c tea.Cmd; m.pFields[m.pFocus], c = m.pFields[m.pFocus].Update(msg); return m, c
	}
	return m, nil
}

func (m M) uDice(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc": m.gm = gmNormal; return m, nil
		case "a":
			var vals []int
			if m.gs != nil && m.pid >= 0 {
				for _, n := range m.gs.Numbers[m.pid] {
					if n.IsNull() { vals = append(vals, rand.Intn(6)+1) }
				}
			}
			if len(vals) > 0 { m.tx("ROLL_DICE", map[string]any{"values": vals, "single": false}) }
			m.gm = gmNormal; return m, nil
		case "enter":
			v, err := strconv.Atoi(strings.TrimSpace(m.dIn.Value()))
			if err != nil || v < 1 || v > 6 { m.errMsg = "1-6"; m.dIn.SetValue(""); return m, nil }
			m.tx("ROLL_DICE", map[string]any{"values": []int{v}, "single": true})
			m.dIn.SetValue(""); return m, nil
		}
	}
	var c tea.Cmd; m.dIn, c = m.dIn.Update(msg); return m, c
}

func (m M) uCmd(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc": m.gm = gmNormal; return m, nil
		case "enter":
			c := strings.TrimSpace(m.cmd.Value()); m.cmd.SetValue(""); m.gm = gmNormal
			if c != "" { return m.exec(c) }
			return m, nil
		}
	}
	var c tea.Cmd; m.cmd, c = m.cmd.Update(msg); return m, c
}

func (m M) exec(s string) (tea.Model, tea.Cmd) {
	switch strings.ToLower(strings.Fields(s)[0]) {
	case "end": m.tx("END_TURN", nil)
	case "stopdice","sd": m.tx("FINISH_DICE", nil)
	case "state": m.tx("GET_STATE", nil)
	default: m.errMsg = "? " + s
	}
	return m, nil
}

// ━━ Server ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m *M) onSrv(msg wsRx) (tea.Model, tea.Cmd) {
	m.errMsg = ""
	switch msg.Type {
	case "WELCOME":
		m.scr = scrRoom; m.rIn.Focus(); m.status = ""

	case "AUTO_REJOINED":
		var d map[string]any; json.Unmarshal(msg.Payload, &d)
		if code, ok := d["code"].(string); ok { m.room = code }
		if pid, ok := d["playerID"].(float64); ok { m.pid = int(pid) }
		m.scr = scrGame; m.status = ""

	case "ROOM_CREATED":
		var d map[string]string; json.Unmarshal(msg.Payload, &d)
		m.room = d["code"]; m.scr = scrLobby; m.status = ""
		m.tx("ADD_PLAYER", nil)

	case "JOINED":
		var d map[string]string; json.Unmarshal(msg.Payload, &d)
		m.room = d["code"]; m.scr = scrLobby; m.status = ""
		m.tx("ADD_PLAYER", nil)

	case "PLAYER_ADDED":
		var d map[string]any; json.Unmarshal(msg.Payload, &d)
		if v, ok := d["playerID"].(float64); ok { m.pid = int(v) }
		m.status = ""

	case "RECONNECTED":
		var d map[string]any; json.Unmarshal(msg.Payload, &d)
		if v, ok := d["playerID"].(float64); ok { m.pid = int(v) }
		m.status = ""

	case "STATE_UPDATE":
		// This is the ONLY place we update game state — keeps all clients in sync
		var gs model.GameState
		if json.Unmarshal(msg.Payload, &gs) == nil {
			m.gs = &gs
			if gs.Started { m.scr = scrGame }
			if len(gs.Log) > 0 { m.logs = gs.Log }
			if m.pid == -1 {
				for i, p := range gs.Players { if p.Name == m.user { m.pid = i } }
			}
		}

	case "REPLY":
		var r replyMsg
		if json.Unmarshal(msg.Payload, &r) == nil {
			if !r.Success { m.errMsg = r.Message }
			// Don't show success messages — state update speaks for itself
		}

	case "ERROR":
		var d map[string]string; json.Unmarshal(msg.Payload, &d)
		m.errMsg = d["message"]
	}
	return m, listen(m.conn)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  VIEW
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m M) View() string {
	switch m.scr {
	case scrLogin:  return m.vLogin()
	case scrRoom:   return m.vRoom()
	case scrCreate: return m.vCreate()
	case scrLobby:  return m.vLobby()
	case scrGame:   return m.vGame()
	}
	return ""
}

func (m M) banner() string { return sSplash.Render(splash) }

func (m M) foot() string {
	if m.errMsg != "" { return "\n" + sBar.Render(sErr.Render("err: " + m.errMsg)) + "\n" }
	if m.status != "" { return "\n" + sBar.Render(sDim.Render(m.status)) + "\n" }
	return "\n"
}

// ── Login ───────────────────────────────────────────────────────────────

func (m M) vLogin() string {
	b := m.banner() + "\n\n"
	b += sTitle.Render("  login") + "\n\n"
	b += "  " + m.lIn[0].View() + "\n"
	b += "  " + m.lIn[1].View() + "\n\n"
	b += sVDim.Render("  tab: switch   enter: connect   ctrl+c: quit")
	return b + m.foot()
}

// ── Room ────────────────────────────────────────────────────────────────

func (m M) vRoom() string {
	b := m.banner() + "\n"
	b += sDim.Render("  @"+m.user) + "\n\n"
	b += sTitle.Render("  join or create") + "\n\n"
	b += "  " + m.rIn.View() + "\n\n"
	b += sVDim.Render("  enter code, or empty = new room")
	return b + m.foot()
}

// ── Create ──────────────────────────────────────────────────────────────

func (m M) vCreate() string {
	b := m.banner() + "\n\n"
	b += sTitle.Render("  new room") + "\n\n"

	// Win condition picker
	modeLabel := "  win condition"
	if m.cF == 0 { modeLabel = sLabel.Render("  win condition") + sVDim.Render("  (up/down)") }
	b += modeLabel + "\n"
	modes := []string{"limit (3pts to win)", "largest after N turns", "smallest after N turns"}
	for i, s := range modes {
		prefix := "    "
		if i == m.cWin {
			if m.cF == 0 {
				prefix = sHi.Render("  > ")
				s = sHi.Render(s)
			} else {
				prefix = sLabel.Render("  > ")
				s = sLabel.Render(s)
			}
		} else {
			s = sDim.Render(s)
		}
		b += prefix + s + "\n"
	}

	// Fields
	b += "\n"
	hsLabel := "  hand size  "
	if m.cF == 1 { hsLabel = sLabel.Render("  hand size  ") }
	b += hsLabel + m.cIn[0].View() + "\n"
	if m.cWin > 0 {
		tlLabel := "  turns      "
		if m.cF == 2 { tlLabel = sLabel.Render("  turns      ") }
		b += tlLabel + m.cIn[1].View() + "\n"
	}
	b += "\n" + sVDim.Render("  tab: next section  up/down: pick mode  enter: create  esc: back")
	return b + m.foot()
}

// ── Lobby ───────────────────────────────────────────────────────────────

func (m M) vLobby() string {
	b := m.banner() + "\n"
	b += sLabel.Render("  room ") + sHi.Render(m.room) + "\n\n"
	if m.gs != nil {
		for i, p := range m.gs.Players {
			mk := "  "; if i == m.pid { mk = "> " }
			st := sOk.Render("on"); if !p.Online { st = sDim.Render("off") }
			b += fmt.Sprintf("  %s@%-14s %s\n", mk, p.Name, st)
		}
		b += "\n"
		b += sDim.Render(fmt.Sprintf("  hand: %d", m.gs.Settings.HandSize))
		switch m.gs.Settings.WinCondition {
		case model.WinLimit: b += sDim.Render("  mode: limit")
		case model.WinLargest: b += sDim.Render(fmt.Sprintf("  mode: largest (%d turns)", m.gs.Settings.TurnLimit))
		case model.WinSmallest: b += sDim.Render(fmt.Sprintf("  mode: smallest (%d turns)", m.gs.Settings.TurnLimit))
		}
	}
	b += "\n\n" + sVDim.Render("  [j] join   [s] start")
	return b + m.foot()
}

// ── Game ────────────────────────────────────────────────────────────────

func (m M) vGame() string {
	gs := m.gs
	if gs == nil { return "waiting..." }

	var b strings.Builder
	cp := gs.Turn % len(gs.Players)

	// ── header ──
	b.WriteString(sSplash.Render(splash) + "\n")
	info := sLabel.Render("  room ") + sDim.Render(gs.GameID) +
		sLabel.Render("  turn ") + sDim.Render(fmt.Sprintf("%d", gs.Turn)) +
		sLabel.Render("  current ") + sHi.Render("@"+gs.Players[cp].Name)
	switch gs.Settings.WinCondition {
	case model.WinLimit:
		info += sLabel.Render("  limit ") + sDim.Render(model.NumFromFloat(gs.Settings.Limit).Display())
	case model.WinLargest:
		info += sLabel.Render("  ") + sDim.Render(fmt.Sprintf("largest, %d left", gs.Settings.TurnLimit-gs.Turn))
	case model.WinSmallest:
		info += sLabel.Render("  ") + sDim.Render(fmt.Sprintf("smallest, %d left", gs.Settings.TurnLimit-gs.Turn))
	}
	b.WriteString(info + "\n")

	if gs.Finished && gs.Winner >= 0 {
		overlay := m.vVictoryOverlay(gs)
		return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceChars("."),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#1a1a1a")))
	}

	// ── numbers ──
	hlP, hlN := -1, -1
	if m.gm == gmPlay { hlP, hlN = m.highlight() }

	b.WriteString("\n")
	var boxes []string
	for i, p := range gs.Players {
		me := i == m.pid
		cur := i == cp

		mk := " "
		if me { mk = ">" }

		// Count cards in hand
		handCnt := 0
		for _, c := range gs.Cards { if c.Owner == i { handCnt++ } }

		label := fmt.Sprintf(" %s @%s", mk, p.Name)
		label += sDim.Render(fmt.Sprintf(" [%d cards]", handCnt))
		if gs.Done[i] { label += sDim.Render(" done") }
		if !p.Online { label += sDim.Render(" off") }
		if gs.Settings.WinCondition == model.WinLimit && p.Points > 0 {
			label += sWarn.Render(fmt.Sprintf(" %dpt", p.Points))
		}

		var ns []string
		for j, num := range gs.Numbers[i] {
			hl := (hlP == i && hlN == -1) || (hlP == i && hlN == j)
			ns = append(ns, rNum(num, j, hl))
		}

		content := label + "\n  " + strings.Join(ns, " ")
		st := sBox
		if cur { st = sBoxActive }
		if hlP == i { st = sBoxHl }
		boxes = append(boxes, st.Render(content))
	}

	// Grid: 2 per row if > 3 players
	if len(boxes) <= 3 {
		for _, bx := range boxes { b.WriteString(bx + "\n") }
	} else {
		for i := 0; i < len(boxes); i += 2 {
			if i+1 < len(boxes) {
				b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, boxes[i], " ", boxes[i+1]) + "\n")
			} else {
				b.WriteString(boxes[i] + "\n")
			}
		}
	}

	// ── cards (horizontal) ──
	hand := m.hand()
	b.WriteString("\n " + sTitle.Render("hand") + "\n")
	if len(hand) == 0 {
		b.WriteString(sDim.Render("  (empty)") + "\n")
	} else {
		perRow := max((m.w-2)/30, 1)
		for row := 0; row < len(hand); row += perRow {
			end := row + perRow; if end > len(hand) { end = len(hand) }
			var rc []string
			for idx := row; idx < end; idx++ {
				ci := hand[idx]
				rc = append(rc, rCard(gs.Cards[ci], ci, idx == m.sel && m.gm == gmNormal))
			}
			b.WriteString(" " + lipgloss.JoinHorizontal(lipgloss.Top, rc...) + "\n")
		}
	}

	// ── queue ──
	if len(gs.Queue) > 0 {
		b.WriteString("\n " + sTitle.Render("queue") + "\n")
		for _, ci := range gs.Queue {
			if ci >= 0 && ci < len(gs.Cards) {
				c := gs.Cards[ci]
				who := "?"
				if c.Owner >= 0 && c.Owner < len(gs.Players) {
					who = "@" + gs.Players[c.Owner].Name
				}
				desc := c.Description; if len(desc) > 40 { desc = desc[:37]+"..." }
				b.WriteString(fmt.Sprintf("   %s %s  %s\n", sDim.Render(who), sLabel.Render(c.Name), sVDim.Render(desc)))
			}
		}
	}

	// ── log ──
	if len(m.logs) > 0 {
		b.WriteString("\n")
		start := len(m.logs) - 3; if start < 0 { start = 0 }
		for _, l := range m.logs[start:] {
			b.WriteString(sVDim.Render("  " + l) + "\n")
		}
	}

	// ── bottom panel ──
	b.WriteString("\n")
	switch m.gm {
	case gmNormal:
		b.WriteString(m.vNormal())
	case gmPlay:
		b.WriteString(m.vPlayPanel())
	case gmDice:
		b.WriteString("  " + sWarn.Render("dice") + " " + m.dIn.View() + "  " + sVDim.Render("enter=roll  a=all  esc=cancel") + "\n")
	case gmCommand:
		b.WriteString("  " + sLabel.Render(":") + " " + m.cmd.View() + "\n")
	}

	b.WriteString(m.foot())
	return b.String()
}

func (m M) vNormal() string {
	var s string
	hand := m.hand()
	if len(hand) > 0 && m.sel >= 0 && m.sel < len(hand) && m.gs != nil {
		ci := hand[m.sel]
		c := m.gs.Cards[ci]
		formula := cardFormulaStatic(c.Method, c.InputsReq)
		detail := sHi.Render(c.Name) + " " + sVDim.Render(fmt.Sprintf("#%d", ci)) + "\n" +
			sDim.Render(c.Description)
		if formula != "" {
			detail += "\n" + sLabel.Render(formula)
		}
		userIn := userInputDesc(c.InputsReq)
		if userIn != "(auto)" && userIn != "(no input)" {
			detail += "\n" + sVDim.Render("you pick: "+userIn)
		}
		s += " " + sBoxActive.Render(detail) + "\n"
	}
	s += sVDim.Render("  left/right: select   enter: play   d: dice   e: end turn   f: done rolling   :: cmd") + "\n"
	return s
}

func (m M) vPlayPanel() string {
	if m.gs == nil { return "" }
	card := m.gs.Cards[m.pCard]

	var s string
	s += "  " + sOk.Render("play") + " " + sHi.Render(card.Name) + "\n"

	// Build formula visualization
	formula := cardFormula(card.Method, m.pReq, m.pAuto, m.pUserIdx, m.pFields, m.pFocus, m.gs)
	if formula != "" {
		s += "  " + formula + "\n"
	}

	// Render user-fillable input fields
	if len(m.pFields) > 0 {
		var fields []string
		for fi, ti := range m.pFields {
			ri := m.pUserIdx[fi]
			// Build contextual label
			label := slotLabel(m.pReq, ri, m.pAuto, m.gs)

			st := sField; if fi == m.pFocus { st = sFieldActive }
			content := sLabel.Render(label) + " " + ti.View()
			fields = append(fields, st.Render(content))
		}
		s += "  " + lipgloss.JoinHorizontal(lipgloss.Top, fields...) + "\n"
	}

	s += "  " + sVDim.Render("tab: next   enter: confirm   esc: cancel") + "\n"
	return s
}

// ━━ Victory Overlay ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m M) vVictoryOverlay(gs *model.GameState) string {
	w := gs.Winner
	winner := gs.Players[w]

	// Big winner name
	winnerArt := fmt.Sprintf(`
    *  *  *  *  *  *  *  *  *
    *                       *
    *   @%-18s  *
    *                       *
    *  *  *  *  *  *  *  *  *`, winner.Name)

	// Compute awards
	type award struct {
		title  string
		player string
		detail string
	}
	var awards []award

	// Most Aggressive — most attacks dealt
	bestAtk, bestAtkP := 0, ""
	for _, p := range gs.Players {
		if p.Stats.AttacksDealt > bestAtk { bestAtk = p.Stats.AttacksDealt; bestAtkP = p.Name }
	}
	if bestAtkP != "" && bestAtk > 0 {
		awards = append(awards, award{"Most Aggressive", bestAtkP, fmt.Sprintf("%d attacks", bestAtk)})
	}

	// Card Maniac — most cards played
	bestCards, bestCardsP := 0, ""
	for _, p := range gs.Players {
		if p.Stats.CardsPlayed > bestCards { bestCards = p.Stats.CardsPlayed; bestCardsP = p.Name }
	}
	if bestCardsP != "" {
		awards = append(awards, award{"Card Maniac", bestCardsP, fmt.Sprintf("%d cards", bestCards)})
	}

	// Card Hoarder — fewest cards played (among players who played at least 1)
	leastCards, leastCardsP := -1, ""
	for _, p := range gs.Players {
		if p.Stats.CardsPlayed > 0 && (leastCards == -1 || p.Stats.CardsPlayed < leastCards) {
			leastCards = p.Stats.CardsPlayed; leastCardsP = p.Name
		}
	}
	if leastCardsP != "" && leastCardsP != bestCardsP {
		awards = append(awards, award{"Card Hoarder", leastCardsP, fmt.Sprintf("only %d cards", leastCards)})
	}

	// Lucky Roller — highest dice total
	bestDice, bestDiceP := 0, ""
	for _, p := range gs.Players {
		if p.Stats.DiceTotal > bestDice { bestDice = p.Stats.DiceTotal; bestDiceP = p.Name }
	}
	if bestDiceP != "" && bestDice > 0 {
		awards = append(awards, award{"Lucky Roller", bestDiceP, fmt.Sprintf("total: %d", bestDice)})
	}

	// Big Number — largest number achieved
	var bigNum model.Number
	bigP := ""
	for _, p := range gs.Players {
		n := model.Number{Value: p.Stats.BiggestNumber, Base: p.Stats.BiggestBase}
		if bigP == "" || n.Cmp(bigNum) > 0 { bigNum = n; bigP = p.Name }
	}
	if bigP != "" && !bigNum.IsZero() {
		awards = append(awards, award{"Biggest Number", bigP, bigNum.Display()})
	}

	// Closest Rival — second place
	switch gs.Settings.WinCondition {
	case model.WinLimit:
		bestPts, rivalP := -1, ""
		for i, p := range gs.Players {
			if i == w { continue }
			if p.Points > bestPts { bestPts = p.Points; rivalP = p.Name }
		}
		if rivalP != "" {
			awards = append(awards, award{"Closest Rival", rivalP, fmt.Sprintf("%d points", bestPts)})
		}
	case model.WinLargest, model.WinSmallest:
		var rivalNum model.Number
		rivalP := ""
		largest := gs.Settings.WinCondition == model.WinLargest
		for i, p := range gs.Players {
			if i == w { continue }
			for _, num := range gs.Numbers[i] {
				if num.IsNull() { continue }
				n := model.Number{Value: num.Value, Base: num.Base}
				if rivalP == "" {
					rivalNum = n; rivalP = p.Name
				} else {
					cmp := n.Cmp(rivalNum)
					if (largest && cmp > 0) || (!largest && cmp < 0) {
						rivalNum = n; rivalP = p.Name
					}
				}
			}
		}
		if rivalP != "" {
			awards = append(awards, award{"Closest Rival", rivalP, rivalNum.Display()})
		}
	}

	// Build the overlay
	var content strings.Builder

	content.WriteString(sHi.Render(winnerArt) + "\n\n")
	content.WriteString(sTitle.Render("    GAME OVER") + "\n\n")

	// Player summary table
	content.WriteString(sLabel.Render("  final standings") + "\n")
	for i, p := range gs.Players {
		mk := "  "; if i == w { mk = sHi.Render(">>") }
		pts := ""
		if gs.Settings.WinCondition == model.WinLimit {
			pts = fmt.Sprintf(" %dpts", p.Points)
		}
		content.WriteString(fmt.Sprintf("  %s @%-14s %s %s\n", mk, p.Name, sDim.Render(fmt.Sprintf("%d cards played", p.Stats.CardsPlayed)), sWarn.Render(pts)))
	}

	// Awards
	if len(awards) > 0 {
		content.WriteString("\n" + sLabel.Render("  awards") + "\n")
		for _, a := range awards {
			content.WriteString(fmt.Sprintf("  %-20s @%-14s %s\n",
				sWarn.Render(a.title),
				sHi.Render(a.player),
				sDim.Render(a.detail)))
		}
	}

	content.WriteString("\n" + sVDim.Render("  ctrl+c to exit"))

	// Wrap in a styled box
	popup := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#666")).
		Padding(1, 3).
		Width(52).
		Background(lipgloss.Color("#111")).
		Render(content.String())

	return popup
}

// ━━ Renderers ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func rNum(n model.Number, idx int, hl bool) string {
	if n.IsNull() {
		s := fmt.Sprintf("%d:---", idx)
		if hl { return sNumHl.Render("["+s+"]") }
		return sNul.Render("[" + s + "]")
	}
	s := fmt.Sprintf("%d:%s", idx, n.Display())
	switch n.Mark {
	case "I":
		if hl { return sNumHl.Render("["+s+"]") }
		return sImm.Render("[" + s + " I]")
	case "F":
		if hl { return sNumHl.Render("["+s+"]") }
		return sFib.Render("[" + s + " F]")
	default:
		if hl { return sNumHl.Render("["+s+"]") }
		return sNum.Render("[" + s + "]")
	}
}

func rCard(c model.Card, ci int, sel bool) string {
	tag := sDim.Render("fn")
	switch c.Type {
	case "Constant": tag = sDim.Render("co")
	case "Theorem":  tag = sDim.Render("th")
	}

	name := sHi.Render(c.Name)
	id := sVDim.Render(fmt.Sprintf("#%d", ci))

	desc := c.Description
	if len(desc) > 24 { desc = desc[:21] + "..." }

	inp := userInputShort(c.InputsReq)

	content := fmt.Sprintf("%s %s %s\n%s\n%s", tag, name, id, sDim.Render(desc), sVDim.Render(inp))
	if sel { return sCardSel.Render(content) }
	return sCard.Render(content)
}

// ━━ Input helpers ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func reqLabel(c byte) string {
	switch c {
	case 'n': return "slot"
	case 'p': return "player"
	case 'c': return "card"
	case 'X': return "min"
	case 'Y': return "max"
	case 'i': return "val"
	}
	return "?"
}

func reqPlaceholder(c byte) string {
	switch c {
	case 'n': return "0-4"
	case 'p': return "id"
	case 'X': return "min"
	case 'Y': return "max"
	case 'i': return "val"
	}
	return "?"
}

// userInputDesc returns a human description of what the user actually needs to type
func userInputDesc(req string) string {
	if req == "" { return "(no input)" }
	var parts []string
	for i := 0; i < len(req); i++ {
		switch req[i] {
		case 'A': // auto
		case 'U': // auto
		case 'd': // auto
		case 'n':
			// Figure out context from previous char
			if i > 0 {
				switch req[i-1] {
				case 'A': parts = append(parts, "target's slot (0-4)")
				case 'U': parts = append(parts, "your slot (0-4)")
				default:  parts = append(parts, "slot (0-4)")
				}
			} else {
				parts = append(parts, "slot (0-4)")
			}
		case 'p': parts = append(parts, "player id")
		}
	}
	if len(parts) == 0 { return "(auto)" }
	return strings.Join(parts, ", ")
}

// userInputShort returns a short tag line for what's needed
func userInputShort(req string) string {
	if req == "" { return "no input" }
	var parts []string
	for _, c := range req {
		switch c {
		case 'n': parts = append(parts, "slot")
		case 'p': parts = append(parts, "plr")
		// A, U, d are auto — don't show
		}
	}
	if len(parts) == 0 { return "auto" }
	return strings.Join(parts, " ")
}

func (m M) highlight() (int, int) {
	if m.gs == nil || len(m.pUserIdx) == 0 || m.pFocus >= len(m.pUserIdx) { return -1, -1 }

	ri := m.pUserIdx[m.pFocus]
	c := m.pReq[ri]

	if c == 'n' {
		// Find which player this 'n' belongs to by looking at the preceding A or U
		for prev := ri - 1; prev >= 0; prev-- {
			if m.pReq[prev] == 'A' || m.pReq[prev] == 'U' || m.pReq[prev] == 'p' {
				playerVal := m.pAuto[prev]
				if playerVal >= 0 && playerVal < len(m.gs.Players) {
					// If user typed a specific slot, highlight it
					val, err := strconv.Atoi(strings.TrimSpace(m.pFields[m.pFocus].Value()))
					if err == nil && val >= 0 && val <= 4 {
						return playerVal, val
					}
					return playerVal, -1 // highlight whole player
				}
				break
			}
		}
	}
	return -1, -1
}

// ━━ Formulas ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// cardFormulaStatic returns a static formula string for card detail view
func cardFormulaStatic(method, req string) string {
	switch method {
	case "ADD":          return "target[a] = target[a] + yours[b]"
	case "SUBTRACT":     return "target[a] = target[a] - yours[b]"
	case "MULTIPLY":     return "target[a] = target[a] * yours[b]"
	case "DIVIDE":       return "target[a] = target[a] / yours[b]"
	case "ABSOLUTEVALUE":return "target[a] = |target[a]|"
	case "INVERSE":      return "target[a] = 1 / target[a]"
	case "NEGATIVE":     return "target[a] = -target[a]"
	case "POSITIVE":     return "target[a] = +target[a]"
	case "SQRT":         return "target[a] = sqrt(target[a])"
	case "SQUARE":       return "target[a] = target[a]^2"
	case "COSMOD":       return "target[a] = target[a] * cos(dice)"
	case "SINMOD":       return "target[a] = target[a] * sin(dice)"
	case "TANMOD":       return "target[a] = target[a] * tan(dice)"
	case "LOG10":        return "target[a] = log10(target[a])"
	case "EXPONENTIAL":  return "target[a] = e^target[a]"
	case "NATLOG":       return "target[a] = ln(target[a])"
	case "LOGORHYTHM":   return "target[a] = log_dice(target[a])"
	case "ROOTBASE":     return "target[a] = target[a]^(1/dice)"
	case "EXPONENTBASE": return "target[a] = target[a]^dice"
	case "POLYNOMIAL1":  return "target[a] = target[a]*dice + yours[b]"
	case "POLYNOMIAL2":  return "target[a] = target[a]*d^2 + yours[b]*d + yours[c]"
	case "SIGMANOTATION":     return "sum all numbers of a player into one"
	case "PRODUCTNOTATION":   return "multiply all numbers of a player into one"
	case "PYTHAGOREANTHEOREM":return "target[a] = sqrt(target[a]^2 + yours[b]^2)"
	case "ELEMENTCOMMUTATIVE":return "swap target[a] with target[b]"
	case "ELEMENTCLOSURE":    return "target[a] becomes immune for 1 turn"
	case "ELEMENTDISTRIBUTIVE":return "copy target[a] to all other players"
	case "ELEMENTIDENTITY":   return "does nothing (adds 0 or multiplies by 1)"
	case "PASCALTRIANGLE":    return "collapse adjacent numbers into their sum"
	case "FUNDAMENTALTHEOREMOFARITHMETIC": return "decompose target[a] into prime factors"
	case "FACTORIAL":    return "place dice! into your set"
	}
	if req == "" { return "place constant into your set" }
	return ""
}

// cardFormula returns a live formula with current input values filled in during play mode
func cardFormula(method, req string, auto []int, _ []int, fields []textinput.Model, focus int, gs *model.GameState) string {
	// Build value display for each position
	vals := make([]string, len(req))
	fieldIdx := 0
	for i, c := range req {
		switch byte(c) {
		case 'A':
			name := fmt.Sprintf("p%d", auto[i])
			if auto[i] >= 0 && auto[i] < len(gs.Players) {
				name = "@" + gs.Players[auto[i]].Name
			}
			vals[i] = sDim.Render(name)
		case 'U':
			vals[i] = sDim.Render("you")
		case 'd':
			vals[i] = sDim.Render(fmt.Sprintf("d%d", auto[i]))
		case 'n':
			if fieldIdx < len(fields) {
				v := strings.TrimSpace(fields[fieldIdx].Value())
				if v == "" {
					if fieldIdx == focus {
						vals[i] = sHi.Render("[_]")
					} else {
						vals[i] = sVDim.Render("[_]")
					}
				} else {
					if fieldIdx == focus {
						vals[i] = sHi.Render("["+v+"]")
					} else {
						vals[i] = sLabel.Render("["+v+"]")
					}
				}
				fieldIdx++
			}
		default:
			if fieldIdx < len(fields) {
				v := strings.TrimSpace(fields[fieldIdx].Value())
				if v == "" {
					vals[i] = sVDim.Render("[?]")
				} else {
					vals[i] = sLabel.Render("["+v+"]")
				}
				fieldIdx++
			}
		}
	}

	// Now build the formula string using the method
	a := func(i int) string { if i < len(vals) { return vals[i] }; return "?" }

	switch method {
	case "ADD":          return fmt.Sprintf("%s[%s] = %s[%s] + %s[%s]", a(0), a(1), a(0), a(1), a(2), a(3))
	case "SUBTRACT":     return fmt.Sprintf("%s[%s] = %s[%s] - %s[%s]", a(0), a(1), a(0), a(1), a(2), a(3))
	case "MULTIPLY":     return fmt.Sprintf("%s[%s] = %s[%s] * %s[%s]", a(0), a(1), a(0), a(1), a(2), a(3))
	case "DIVIDE":       return fmt.Sprintf("%s[%s] = %s[%s] / %s[%s]", a(0), a(1), a(0), a(1), a(2), a(3))
	case "ABSOLUTEVALUE":return fmt.Sprintf("%s[%s] = |%s[%s]|", a(0), a(1), a(0), a(1))
	case "INVERSE":      return fmt.Sprintf("%s[%s] = 1 / %s[%s]", a(0), a(1), a(0), a(1))
	case "NEGATIVE":     return fmt.Sprintf("%s[%s] = -%s[%s]", a(0), a(1), a(0), a(1))
	case "SQRT":         return fmt.Sprintf("%s[%s] = sqrt(%s[%s])", a(0), a(1), a(0), a(1))
	case "SQUARE":       return fmt.Sprintf("%s[%s] = %s[%s]^2", a(0), a(1), a(0), a(1))
	case "COSMOD":       return fmt.Sprintf("%s[%s] = %s[%s] * cos(%s)", a(0), a(1), a(0), a(1), a(2))
	case "SINMOD":       return fmt.Sprintf("%s[%s] = %s[%s] * sin(%s)", a(0), a(1), a(0), a(1), a(2))
	case "TANMOD":       return fmt.Sprintf("%s[%s] = %s[%s] * tan(%s)", a(0), a(1), a(0), a(1), a(2))
	case "LOG10":        return fmt.Sprintf("%s[%s] = log10(%s[%s])", a(0), a(1), a(0), a(1))
	case "EXPONENTIAL":  return fmt.Sprintf("%s[%s] = e^%s[%s]", a(0), a(1), a(0), a(1))
	case "NATLOG":       return fmt.Sprintf("%s[%s] = ln(%s[%s])", a(0), a(1), a(0), a(1))
	case "LOGORHYTHM":   return fmt.Sprintf("%s[%s] = log_%s(%s[%s])", a(0), a(1), a(2), a(0), a(1))
	case "ROOTBASE":     return fmt.Sprintf("%s[%s] = %s[%s]^(1/%s)", a(0), a(1), a(0), a(1), a(2))
	case "EXPONENTBASE": return fmt.Sprintf("%s[%s] = %s[%s]^%s", a(0), a(1), a(0), a(1), a(2))
	case "POLYNOMIAL1":  return fmt.Sprintf("%s[%s] = %s[%s]*%s + %s[%s]", a(0), a(1), a(0), a(1), a(4), a(2), a(3))
	case "POLYNOMIAL2":  return fmt.Sprintf("%s[%s] = %s[%s]*%s^2 + %s[%s]*%s + %s[%s]", a(0), a(1), a(0), a(1), a(6), a(2), a(3), a(6), a(4), a(5))
	case "PYTHAGOREANTHEOREM": return fmt.Sprintf("%s[%s] = sqrt(%s[%s]^2 + %s[%s]^2)", a(0), a(1), a(0), a(1), a(2), a(3))
	case "ELEMENTCOMMUTATIVE": return fmt.Sprintf("swap %s[%s] <-> %s[%s]", a(0), a(1), a(2), a(3))
	case "SIGMANOTATION":      return fmt.Sprintf("sum all of %s's numbers", a(0))
	case "PRODUCTNOTATION":    return fmt.Sprintf("multiply all of %s's numbers", a(0))
	case "ELEMENTCLOSURE":     return fmt.Sprintf("%s[%s] becomes immune", a(0), a(1))
	case "ELEMENTDISTRIBUTIVE":return fmt.Sprintf("copy %s[%s] to everyone", a(0), a(1))
	case "PASCALTRIANGLE":     return fmt.Sprintf("collapse island around %s[%s]", a(0), a(1))
	case "FUNDAMENTALTHEOREMOFARITHMETIC": return fmt.Sprintf("factor %s[%s] into primes", a(0), a(1))
	}
	return ""
}

// slotLabel gives a contextual label for a user-fillable field
func slotLabel(req string, ri int, auto []int, gs *model.GameState) string {
	if req[ri] != 'n' {
		return reqLabel(req[ri])
	}
	// Find the preceding player (A or U) to give context
	for prev := ri - 1; prev >= 0; prev-- {
		if req[prev] == 'A' {
			if auto[prev] >= 0 && auto[prev] < len(gs.Players) {
				return "@" + gs.Players[auto[prev]].Name + " slot"
			}
			return "target slot"
		}
		if req[prev] == 'U' {
			return "your slot"
		}
	}
	return "slot"
}

// ━━ Helpers ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m M) hand() []int {
	if m.gs == nil || m.pid < 0 { return nil }
	var out []int
	for i, c := range m.gs.Cards { if c.Owner == m.pid { out = append(out, i) } }
	return out
}

func (m *M) tx(t string, p any) {
	if m.conn == nil { m.errMsg = "not connected"; return }
	m.conn.WriteJSON(wsMsg{Type: t, Payload: p})
}

func (m *M) txPlay(ci int, inputs []int) {
	m.tx("PLAY_CARD", map[string]any{"cardIndex": ci, "inputs": inputs})
	m.status = ""
}

func (m M) doAuth(user, pass string) tea.Cmd {
	return func() tea.Msg {
		u := m.addr; if !strings.HasPrefix(u, "http") { u = "http://" + u }
		body, _ := json.Marshal(map[string]string{"name": user, "password": pass})
		resp, err := http.Post(u+"/auth", "application/json", bytes.NewReader(body))
		if err != nil { return wsErrMsg{fmt.Errorf("auth: %v", err)} }
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			buf := make([]byte, 256); n, _ := resp.Body.Read(buf)
			return wsErrMsg{fmt.Errorf("auth: %s", string(buf[:n]))}
		}
		var ar struct{ Token string `json:"token"` }
		json.NewDecoder(resp.Body).Decode(&ar)

		wu := strings.Replace(u, "http", "ws", 1)
		pu, _ := url.Parse(wu); pu.Path = "/ws"
		q := pu.Query(); q.Set("token", ar.Token); pu.RawQuery = q.Encode()
		conn, _, err := websocket.DefaultDialer.Dial(pu.String(), nil)
		if err != nil { return wsErrMsg{fmt.Errorf("ws: %v", err)} }

		var first srvMsg
		if err := conn.ReadJSON(&first); err != nil { conn.Close(); return wsErrMsg{fmt.Errorf("connect: %v", err)} }
		return authOkMsg{conn: conn, username: user, token: ar.Token, first: first}
	}
}

func listen(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		var msg srvMsg
		if err := conn.ReadJSON(&msg); err != nil { return wsErrMsg{err} }
		return wsRx(msg)
	}
}

// ━━ Main ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func main() {
	addr := "localhost:1729"
	if len(os.Args) > 1 { addr = os.Args[1] }
	p := tea.NewProgram(newM(addr), tea.WithAltScreen())
	if _, err := p.Run(); err != nil { fmt.Fprintf(os.Stderr, "error: %v\n", err); os.Exit(1) }
}
