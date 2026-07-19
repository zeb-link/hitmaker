package tui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/zeb-link/hitmaker/v2/internal/config"
	"github.com/zeb-link/hitmaker/v2/internal/ui/theme"
)

type configAction int

const (
	configActionNone configAction = iota
	configActionClose
	configActionApply
	configActionSaveGlobal
	configActionSaveLocal
	configActionDefaults
)

type editorPane int

const (
	paneFields editorPane = iota
	paneParams
	panePayloads
	paneConfirmApply
)

type editorField struct {
	label string
	kind  string
	key   string
	group string
	min   float64
	max   float64
	step  float64
}

type selectOption struct {
	label string
	value string
}

type fieldGuide struct {
	title   string
	summary string
	details []string
}

type configEditor struct {
	cfg          config.Config
	fields       []editorField
	helpWidth    int
	focus        int
	pane         editorPane
	paramFocus   int
	payloadFocus int
	editing      bool
	input        textinput.Model
	status       string
	typingKey    string
	typingValue  string
}

func newConfigEditor(cfg config.Config) configEditor {
	input := textinput.New()
	st := textinput.DefaultDarkStyles()
	st.Cursor.Color = theme.Accent
	input.SetStyles(st)
	input.Prompt = theme.Focus.Render("⣿ ")
	input.CharLimit = 512
	return configEditor{
		cfg: cfg,
		fields: []editorField{
			{group: "TRAFFIC", label: "Min hits/min", kind: "number", key: "minRate", min: 0, max: 1000, step: 1},
			{group: "TRAFFIC", label: "Max hits/min", kind: "number", key: "maxRate", min: 0, max: 1000, step: 1},
			{group: "TRAFFIC", label: "Workers/target", kind: "number", key: "concurrent", min: 1, max: 64, step: 1},
			{group: "TRAFFIC", label: "Timeout", kind: "number", key: "timeout", min: 100, max: 60000, step: 500},
			{group: "IDENTITY", label: "Method", kind: "select", key: "method"},
			{group: "IDENTITY", label: "Bot traffic %", kind: "slider", key: "unknown", min: 0, max: 100, step: 5},
			{group: "IDENTITY", label: "Bot pool", kind: "select", key: "botpool"},
			{group: "IDENTITY", label: "Desktop share %", kind: "slider", key: "device", min: 0, max: 100, step: 5},
			{group: "IDENTITY", label: "Unique IP odds", kind: "slider", key: "unique", min: 0, max: 1, step: 0.05},
			{group: "SCHEDULE", label: "Active min", kind: "number", key: "minActive", min: 0, max: 120, step: 1},
			{group: "SCHEDULE", label: "Active max", kind: "number", key: "maxActive", min: 0, max: 120, step: 1},
			{group: "SCHEDULE", label: "Idle odds", kind: "slider", key: "idleOdds", min: 0, max: 1, step: 0.05},
			{group: "SCHEDULE", label: "Idle min", kind: "number", key: "minIdle", min: 0, max: 120, step: 1},
			{group: "SCHEDULE", label: "Idle max", kind: "number", key: "maxIdle", min: 0, max: 120, step: 1},
			{group: "ENTROPY", label: "Level", kind: "select", key: "entropy"},
			{group: "ENTROPY", label: "Audience spread", kind: "slider", key: "entDevice", min: 0, max: 100, step: 5},
			{group: "ENTROPY", label: "Breakout intensity", kind: "slider", key: "entBreakout", min: 0, max: 100, step: 5},
			{group: "ENTROPY", label: "Viral links %", kind: "slider", key: "entViral", min: 0, max: 20, step: 1},
			{group: "ORIGIN", label: "Origin mode", kind: "select", key: "mode"},
			{group: "ORIGIN", label: "Proxy service", kind: "text", key: "provider"},
			{group: "ORIGIN", label: "IPRoyal endpoint", kind: "secret", key: "iproyal"},
			{group: "URL PARAMS", label: "Rules & payloads", kind: "open", key: "params"},
		},
		input: input,
	}
}

func (e configEditor) WithHelpWidth(width int) configEditor {
	e.helpWidth = width
	return e
}

func (e configEditor) Update(msg tea.KeyMsg) (configEditor, configAction, tea.Cmd) {
	key := normalizedKey(msg)
	if e.editing {
		switch key {
		case "enter":
			e.commitInput()
			e.editing = false
			return e, configActionNone, nil
		case "esc":
			e.editing = false
			return e, configActionNone, nil
		}
		var cmd tea.Cmd
		e.input, cmd = e.input.Update(msg)
		return e, configActionNone, cmd
	}
	switch key {
	case "esc":
		if e.pane == paneFields {
			return e, configActionClose, nil
		}
		e.pane = paneFields
	case "q":
		return e, configActionClose, nil
	case "a":
		e.pane = paneConfirmApply
		e.status = "Review the preview, then Enter saves locally and closes."
		return e, configActionNone, nil
	case "g":
		e.status = "Saved globally."
		return e, configActionSaveGlobal, nil
	case "l":
		e.status = "Saved locally."
		return e, configActionSaveLocal, nil
	case "d":
		return e, configActionDefaults, nil
	}

	switch e.pane {
	case paneFields:
		return e.updateFields(msg)
	case paneParams:
		return e.updateParams(msg)
	case panePayloads:
		return e.updatePayloads(msg)
	case paneConfirmApply:
		return e.updateConfirmApply(msg)
	default:
		return e, configActionNone, nil
	}
}

func (e configEditor) updateFields(msg tea.KeyMsg) (configEditor, configAction, tea.Cmd) {
	key := normalizedKey(msg)
	if e.handleInlineInput(msg) {
		return e, configActionNone, nil
	}
	switch key {
	case "up", "k":
		e.clearTyping()
		if e.focus > 0 {
			e.focus--
		}
	case "down", "j":
		e.clearTyping()
		if e.focus < len(e.fields)-1 {
			e.focus++
		}
	case "tab":
		e.clearTyping()
		e.focus = nextEnabledField(e.fields, e.focus, 1, e.cfg)
	case "shift+tab":
		e.clearTyping()
		e.focus = nextEnabledField(e.fields, e.focus, -1, e.cfg)
	case "left", "h":
		e.clearTyping()
		e.adjust(-1)
	case "right", "n":
		e.clearTyping()
		e.adjust(1)
	case "enter":
		field := e.fields[e.focus]
		if field.key == "params" {
			e.pane = paneParams
			return e, configActionNone, nil
		}
		if field.kind == "select" {
			e.focus = nextEnabledField(e.fields, e.focus, 1, e.cfg)
			return e, configActionNone, nil
		}
		if field.kind == "text" || field.kind == "secret" {
			e.input.SetValue(e.rawValue(field.key))
			e.input.Focus()
			e.editing = true
		}
	}
	return e, configActionNone, nil
}

func (e configEditor) updateConfirmApply(msg tea.KeyMsg) (configEditor, configAction, tea.Cmd) {
	switch normalizedKey(msg) {
	case "enter", "a":
		e.pane = paneFields
		e.status = "Saved."
		return e, configActionApply, nil
	case "esc", "q":
		e.pane = paneFields
		e.status = "Cancelled."
	}
	return e, configActionNone, nil
}

func (e configEditor) updateParams(msg tea.KeyMsg) (configEditor, configAction, tea.Cmd) {
	params := e.cfg.Requests.URLParams
	switch normalizedKey(msg) {
	case "up", "k":
		if e.paramFocus > 0 {
			e.paramFocus--
		}
	case "down", "j", "tab":
		if e.paramFocus < len(params)-1 {
			e.paramFocus++
		}
	case "shift+tab":
		if e.paramFocus > 0 {
			e.paramFocus--
		}
	case "n", "+":
		e.cfg.Requests.URLParams = append(e.cfg.Requests.URLParams, config.URLParam{Key: "param", Value: "1", Probability: 100})
		e.paramFocus = len(e.cfg.Requests.URLParams) - 1
	case "x", "delete", "backspace":
		if len(params) > 0 {
			e.cfg.Requests.URLParams = append(params[:e.paramFocus], params[e.paramFocus+1:]...)
			if e.paramFocus >= len(e.cfg.Requests.URLParams) && e.paramFocus > 0 {
				e.paramFocus--
			}
		}
	case "p", "enter":
		if len(params) > 0 {
			e.pane = panePayloads
			e.payloadFocus = 0
		}
	case "e":
		if len(params) > 0 {
			e.input.SetValue(paramToLine(params[e.paramFocus]))
			e.input.Focus()
			e.editing = true
		}
	}
	return e, configActionNone, nil
}

func (e configEditor) updatePayloads(msg tea.KeyMsg) (configEditor, configAction, tea.Cmd) {
	if len(e.cfg.Requests.URLParams) == 0 {
		e.pane = paneParams
		return e, configActionNone, nil
	}
	param := &e.cfg.Requests.URLParams[e.paramFocus]
	switch normalizedKey(msg) {
	case "up", "k":
		if e.payloadFocus > 0 {
			e.payloadFocus--
		}
	case "down", "j", "tab":
		if e.payloadFocus < len(param.Payloads)-1 {
			e.payloadFocus++
		}
	case "shift+tab":
		if e.payloadFocus > 0 {
			e.payloadFocus--
		}
	case "n", "+":
		param.Payloads = append(param.Payloads, config.Payload{Name: "Variant", Weight: 1, KV: map[string]string{"campaign": "demo"}})
		e.payloadFocus = len(param.Payloads) - 1
	case "x", "delete", "backspace":
		if len(param.Payloads) > 0 {
			param.Payloads = append(param.Payloads[:e.payloadFocus], param.Payloads[e.payloadFocus+1:]...)
			if e.payloadFocus >= len(param.Payloads) && e.payloadFocus > 0 {
				e.payloadFocus--
			}
		}
	case "e", "enter":
		if len(param.Payloads) > 0 {
			e.input.SetValue(payloadToLine(param.Payloads[e.payloadFocus]))
			e.input.Focus()
			e.editing = true
		}
	case "esc":
		e.pane = paneParams
	}
	return e, configActionNone, nil
}

func (e *configEditor) commitInput() {
	text := strings.TrimSpace(e.input.Value())
	if e.pane == paneParams && len(e.cfg.Requests.URLParams) > 0 {
		if param, err := lineToParam(text, e.cfg.Requests.URLParams[e.paramFocus]); err == nil {
			e.cfg.Requests.URLParams[e.paramFocus] = param
			e.status = "URL param updated."
		} else {
			e.status = err.Error()
		}
		return
	}
	if e.pane == panePayloads && len(e.cfg.Requests.URLParams) > 0 {
		param := &e.cfg.Requests.URLParams[e.paramFocus]
		if len(param.Payloads) > 0 {
			if payload, err := lineToPayload(text, param.Payloads[e.payloadFocus]); err == nil {
				param.Payloads[e.payloadFocus] = payload
				e.status = "Payload updated."
			} else {
				e.status = err.Error()
			}
		}
		return
	}
	e.setRaw(e.fields[e.focus].key, text)
}

func (e *configEditor) handleInlineInput(msg tea.KeyMsg) bool {
	field := e.fields[e.focus]
	if e.fieldDisabled(field) {
		return false
	}
	if field.kind != "number" && field.kind != "slider" {
		return false
	}
	key := normalizedKey(msg)
	if key == "backspace" || key == "delete" {
		if e.typingKey != field.key {
			e.typingKey = field.key
			e.typingValue = e.rawValue(field.key)
		}
		if len(e.typingValue) > 0 {
			e.typingValue = e.typingValue[:len(e.typingValue)-1]
			if e.typingValue == "" || e.typingValue == "-" || e.typingValue == "." {
				e.status = fmt.Sprintf("%s cleared; type a number", field.label)
				return true
			}
			e.setRaw(field.key, e.typingValue)
		}
		return true
	}
	runes := []rune(msg.Key().Text)
	if len(runes) != 1 {
		return false
	}
	r := runes[0]
	if (r < '0' || r > '9') && r != '.' {
		return false
	}
	if r == '.' && field.kind != "slider" {
		return false
	}
	if e.typingKey != field.key {
		e.typingKey = field.key
		e.typingValue = ""
	}
	if r == '.' && strings.Contains(e.typingValue, ".") {
		return true
	}
	e.typingValue += string(r)
	e.setRaw(field.key, e.typingValue)
	return true
}

func (e *configEditor) clearTyping() {
	e.typingKey = ""
	e.typingValue = ""
}

func (e configEditor) currentSelectValue(key string) string {
	switch key {
	case "method":
		return e.cfg.Requests.Method
	case "mode":
		return string(e.cfg.Origin.Mode)
	case "entropy":
		return string(e.cfg.Entropy.Level)
	case "botpool":
		idx := botPoolIndex(e.cfg.Requests.Bots)
		if idx < 0 {
			return "custom"
		}
		return strconv.Itoa(idx)
	default:
		return ""
	}
}

func (e configEditor) selectDisplayValue(key string) string {
	return e.currentSelectValue(key)
}

func (e configEditor) selectOptions(key string) []selectOption {
	switch key {
	case "method":
		return []selectOption{
			{label: "GET", value: "GET"},
			{label: "HEAD", value: "HEAD"},
			{label: "POST", value: "POST"},
		}
	case "mode":
		return []selectOption{
			{label: "None", value: string(config.ModeNone)},
			{label: "Auto", value: string(config.ModeAuto)},
			{label: "Vercel", value: string(config.ModeVercel)},
			{label: "Proxy", value: string(config.ModeProxy)},
		}
	case "entropy":
		options := []selectOption{
			{label: "Off", value: string(config.EntropyOff)},
			{label: "Calm", value: string(config.EntropyCalm)},
			{label: "Chaos", value: string(config.EntropyChaos)},
			{label: "Mayhem", value: string(config.EntropyMayhem)},
		}
		if e.cfg.Entropy.Level == config.EntropyCustom {
			options = append(options, selectOption{label: "Custom", value: string(config.EntropyCustom)})
		}
		return options
	case "botpool":
		options := make([]selectOption, 0, len(botPoolPresets)+1)
		for i, preset := range botPoolPresets {
			options = append(options, selectOption{label: preset.label, value: strconv.Itoa(i)})
		}
		if botPoolIndex(e.cfg.Requests.Bots) < 0 {
			options = append(options, selectOption{label: botPoolLabel(e.cfg.Requests.Bots), value: "custom"})
		}
		return options
	default:
		return nil
	}
}

func (e configEditor) selectLabel(key, value string) string {
	for _, option := range e.selectOptions(key) {
		if option.value == value {
			return option.label
		}
	}
	return value
}

func (e *configEditor) adjust(dir int) {
	field := e.fields[e.focus]
	switch field.key {
	case "minRate":
		e.cfg.Traffic.MinPerMin = clampInt(e.cfg.Traffic.MinPerMin+dir*int(field.step), int(field.min), int(field.max))
	case "maxRate":
		e.cfg.Traffic.MaxPerMin = clampInt(e.cfg.Traffic.MaxPerMin+dir*int(field.step), int(field.min), int(field.max))
	case "concurrent":
		e.cfg.Traffic.Concurrent = clampInt(e.cfg.Traffic.Concurrent+dir, 1, 64)
	case "timeout":
		e.cfg.Traffic.TimeoutMs = clampInt(e.cfg.Traffic.TimeoutMs+dir*500, 100, 60000)
	case "method":
		e.cfg.Requests.Method = rotate(e.cfg.Requests.Method, []string{"GET", "HEAD", "POST"}, dir)
	case "device":
		e.cfg.Requests.DeviceRatio = clampInt(e.cfg.Requests.DeviceRatio+dir*5, 0, 100)
	case "unknown":
		e.cfg.Requests.UnknownRatio = clampInt(e.cfg.Requests.UnknownRatio+dir*5, 0, 100)
	case "unique":
		e.cfg.Requests.UniqueIPProb = clampFloat(e.cfg.Requests.UniqueIPProb+float64(dir)*0.05, 0, 1)
	case "minActive":
		e.cfg.Schedule.MinActive = clampInt(e.cfg.Schedule.MinActive+dir, 0, 120)
	case "maxActive":
		e.cfg.Schedule.MaxActive = clampInt(e.cfg.Schedule.MaxActive+dir, 0, 120)
	case "idleOdds":
		e.cfg.Schedule.IdleOdds = clampFloat(e.cfg.Schedule.IdleOdds+float64(dir)*0.05, 0, 1)
	case "minIdle":
		e.cfg.Schedule.MinIdle = clampInt(e.cfg.Schedule.MinIdle+dir, 0, 120)
	case "maxIdle":
		e.cfg.Schedule.MaxIdle = clampInt(e.cfg.Schedule.MaxIdle+dir, 0, 120)
	case "entropy":
		e.cfg.Entropy.Level = config.EntropyLevel(rotate(string(e.cfg.Entropy.Level),
			[]string{"off", "calm", "chaos", "mayhem"}, dir))
	case "entDevice":
		e.beginCustomEntropy()
		e.cfg.Entropy.DeviceSpread = clampInt(e.cfg.Entropy.DeviceSpread+dir*int(field.step), 0, 100)
	case "entBreakout":
		e.beginCustomEntropy()
		e.cfg.Entropy.Breakout = clampInt(e.cfg.Entropy.Breakout+dir*int(field.step), 0, 100)
	case "entViral":
		e.beginCustomEntropy()
		e.cfg.Entropy.ViralPercent = clampInt(e.cfg.Entropy.ViralPercent+dir*int(field.step), 0, 100)
	case "mode":
		e.cfg.Origin.Mode = config.Mode(rotate(string(e.cfg.Origin.Mode), []string{"none", "auto", "vercel", "proxy"}, dir))
		if (e.cfg.Origin.Mode == config.ModeAuto || e.cfg.Origin.Mode == config.ModeProxy) && e.cfg.Origin.Provider == "" {
			e.cfg.Origin.Provider = "iproyal"
		}
	case "botpool":
		idx := botPoolIndex(e.cfg.Requests.Bots)
		if idx < 0 {
			idx = 0
		}
		idx = (idx + dir) % len(botPoolPresets)
		if idx < 0 {
			idx += len(botPoolPresets)
		}
		if spec := botPoolPresets[idx].spec; spec == nil {
			e.cfg.Requests.Bots = nil
		} else {
			e.cfg.Requests.Bots = append([]string(nil), spec...)
		}
	}
	_ = e.cfg.Validate()
}

// beginCustomEntropy snapshots the level's current effective knobs into the
// stored fields and switches to the custom level, so that hand-tuning one knob
// leaves the others where the preset had them.
func (e *configEditor) beginCustomEntropy() {
	if e.cfg.Entropy.Level == config.EntropyCustom {
		return
	}
	spread, breakout, viral := e.cfg.Entropy.EffectiveHuman()
	e.cfg.Entropy.DeviceSpread = spread
	e.cfg.Entropy.Breakout = breakout
	e.cfg.Entropy.ViralPercent = viral
	e.cfg.Entropy.Level = config.EntropyCustom
}

func (e *configEditor) setRaw(key, value string) {
	switch key {
	case "minRate":
		e.cfg.Traffic.MinPerMin = atoi(value, e.cfg.Traffic.MinPerMin)
	case "maxRate":
		e.cfg.Traffic.MaxPerMin = atoi(value, e.cfg.Traffic.MaxPerMin)
	case "concurrent":
		e.cfg.Traffic.Concurrent = atoi(value, e.cfg.Traffic.Concurrent)
	case "timeout":
		e.cfg.Traffic.TimeoutMs = atoi(value, e.cfg.Traffic.TimeoutMs)
	case "method":
		e.cfg.Requests.Method = strings.ToUpper(value)
	case "device":
		e.cfg.Requests.DeviceRatio = atoi(value, e.cfg.Requests.DeviceRatio)
	case "unknown":
		e.cfg.Requests.UnknownRatio = atoi(value, e.cfg.Requests.UnknownRatio)
	case "unique":
		e.cfg.Requests.UniqueIPProb = atof(value, e.cfg.Requests.UniqueIPProb)
	case "minActive":
		e.cfg.Schedule.MinActive = atoi(value, e.cfg.Schedule.MinActive)
	case "maxActive":
		e.cfg.Schedule.MaxActive = atoi(value, e.cfg.Schedule.MaxActive)
	case "idleOdds":
		e.cfg.Schedule.IdleOdds = atof(value, e.cfg.Schedule.IdleOdds)
	case "minIdle":
		e.cfg.Schedule.MinIdle = atoi(value, e.cfg.Schedule.MinIdle)
	case "maxIdle":
		e.cfg.Schedule.MaxIdle = atoi(value, e.cfg.Schedule.MaxIdle)
	case "entDevice":
		e.beginCustomEntropy()
		e.cfg.Entropy.DeviceSpread = clampInt(atoi(value, e.cfg.Entropy.DeviceSpread), 0, 100)
	case "entBreakout":
		e.beginCustomEntropy()
		e.cfg.Entropy.Breakout = clampInt(atoi(value, e.cfg.Entropy.Breakout), 0, 100)
	case "entViral":
		e.beginCustomEntropy()
		e.cfg.Entropy.ViralPercent = clampInt(atoi(value, e.cfg.Entropy.ViralPercent), 0, 100)
	case "mode":
		e.cfg.Origin.Mode = config.Mode(value)
	case "provider":
		e.cfg.Origin.Provider = value
	case "iproyal":
		if e.cfg.Origin.ProviderConfig == nil {
			e.cfg.Origin.ProviderConfig = map[string]string{}
		}
		e.cfg.Origin.Provider = "iproyal"
		e.cfg.Origin.ProviderConfig["url"] = value
	}
	if err := e.cfg.Validate(); err != nil {
		e.status = err.Error()
	} else {
		e.status = "Updated " + keyLabel(key) + "."
	}
}

func (e configEditor) fieldDisabled(field editorField) bool {
	if field.key == "provider" || field.key == "iproyal" {
		return e.cfg.Origin.Mode != config.ModeProxy && e.cfg.Origin.Mode != config.ModeAuto
	}
	return false
}

func (e configEditor) View(width, height int, err error) string {
	if width < 70 {
		width = 70
	}
	contentWidth := max(1, width-bodyInsetX*2)
	title := e.titleView(contentWidth)
	if e.editing {
		return e.editView(width, height)
	}
	if e.pane == paneConfirmApply {
		return e.applyPreviewView(width, height, err)
	}
	status := e.status
	if err != nil {
		status = err.Error()
	}
	// Top-aligned, full-screen, with the shortcut bar pinned to the bottom — the
	// same frame shape as the live dashboard, so switching between them doesn't
	// jump the content around. Everything is sized from the real terminal height
	// so it adapts to any size.
	commands := e.commandBar(width)
	footer := theme.Subtle.Render(status)
	bottom := lipgloss.JoinVertical(lipgloss.Left, commands, footer)

	fillTo := height - lipgloss.Height(bottom)
	if fillTo < 6 {
		fillTo = 6
	}
	contentHeight := max(5, fillTo-bodyInsetY)
	panelWidth := max(46, contentWidth)
	var body string
	switch e.pane {
	case paneParams, panePayloads:
		body = e.detailView(panelWidth, contentHeight)
	default:
		deckHeight := contentHeight - lipgloss.Height(title) - 2 // title + blank line + breathing room
		if deckHeight < 5 {
			deckHeight = 5
		}
		leftWidth, helpWidth := e.editorColumnWidths(panelWidth)
		if helpWidth > 0 && deckHeight >= 18 {
			deck := e.fieldsView(leftWidth, deckHeight)
			guide := e.fieldGuideView(helpWidth, deckHeight)
			body = lipgloss.JoinHorizontal(lipgloss.Top, deck, " ", guide)
		} else {
			showHint := deckHeight >= 22
			if showHint {
				deckHeight--
			}
			deck := e.fieldsView(panelWidth, deckHeight)
			body = deck
			if showHint {
				hint := lipgloss.NewStyle().Width(panelWidth).Render(e.fieldHintLine(panelWidth))
				body = lipgloss.JoinVertical(lipgloss.Left, deck, hint)
			}
		}
		if body == "" {
			deck := e.fieldsView(panelWidth, deckHeight)
			hint := lipgloss.NewStyle().Width(panelWidth).Render(e.fieldHintLine(panelWidth))
			body = lipgloss.JoinVertical(lipgloss.Left, deck, hint)
		}
	}
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body)
	content = insetBlock(content, bodyInsetX, bodyInsetY)
	content = lipgloss.NewStyle().Width(width).Height(fillTo).MaxHeight(fillTo).Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, content, bottom)
}

func (e configEditor) titleView(width int) string {
	left := lipgloss.JoinHorizontal(lipgloss.Center,
		theme.Logo.Render("HITMAKER CONFIG"),
		"  ",
		theme.Subtle.Render("traffic cockpit"),
	)
	return lipgloss.NewStyle().Width(width).Render(left)
}

func (e configEditor) editorColumnWidths(width int) (int, int) {
	if e.helpWidth > 0 && width > e.helpWidth+1 {
		return width - e.helpWidth - 1, e.helpWidth
	}
	if width < 118 {
		return width, 0
	}
	left := 74
	if width >= 140 {
		left = 82
	}
	right := width - left - 2
	if right < 38 {
		return width, 0
	}
	return left, right
}

func (e configEditor) commandBar(width int) string {
	parts := []string{
		theme.PillHot.Render("SHORTCUTS"),
		theme.Pill.Render("Tab next"),
		theme.Pill.Render("Type numbers"),
		theme.Pill.Render("←/→ nudge"),
		theme.Pill.Render("Enter next/open"),
		theme.PillHot.Render("A save & close"),
		theme.Pill.Render("G save global"),
		theme.Pill.Render("L save local"),
		theme.Pill.Render("D defaults"),
		theme.Pill.Render("Esc back"),
	}
	if width < 96 {
		parts = []string{
			theme.PillHot.Render("KEYS"),
			theme.Pill.Render("Tab"),
			theme.Pill.Render("Type #"),
			theme.Pill.Render("←/→"),
			theme.Pill.Render("Enter next"),
			theme.PillHot.Render("A save+close"),
			theme.Pill.Render("G global"),
			theme.Pill.Render("L local"),
			theme.Pill.Render("Esc"),
		}
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(parts, " "))
}

func (e configEditor) editView(width, height int) string {
	label := "Edit value"
	if e.pane == paneParams {
		label = "Edit param: key=value probability"
	} else if e.pane == panePayloads {
		label = "Edit payload: name weight key=value,key=value"
	}
	box := theme.FocusBorder.Width(min(72, width-6)).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			theme.Title.Render(label),
			"",
			e.input.View(),
			"",
			theme.Subtle.Render("Enter commit  Esc cancel"),
		),
	)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func (e configEditor) fieldsView(width, height int) string {
	// Width is the final outer pane width. theme.Border adds border (2) and
	// horizontal padding (2), so the text/highlight area is width-4.
	innerWidth := max(20, width-4)
	lines := []string{theme.Title.Render("CONTROL DECK")}
	lastGroup := ""
	selectedLine := 1
	for i, field := range e.fields {
		if field.group != lastGroup {
			if lastGroup != "" {
				lines = append(lines, "")
			}
			lines = append(lines, theme.Focus.Render(field.group))
			lastGroup = field.group
		}
		disabled := e.fieldDisabled(field)
		selected := i == e.focus && e.pane == paneFields
		if selected {
			selectedLine = len(lines)
			// Quiet focus: an amber left tick + amber-bold label, keeping the row's
			// own colors. No width-padded background, so it can never wrap.
			label := theme.Focus.Render(fmt.Sprintf("%-18s", field.label))
			lines = append(lines, theme.Tick.Render("▌")+" "+label+" "+e.displayValue(field))
		} else {
			line := fmt.Sprintf("  %-18s %s", field.label, e.displayValue(field))
			if disabled {
				line = theme.Subtle.Render(line)
			}
			lines = append(lines, line)
		}
	}
	lines = clipAround(lines, selectedLine, max(5, height-2))
	return theme.Border.Width(innerWidth).Render(strings.Join(lines, "\n"))
}

func (e configEditor) detailView(width, height int) string {
	switch e.pane {
	case paneParams:
		return e.paramsView(width)
	case panePayloads:
		return e.payloadsView(width)
	default:
		return e.fieldHelpView(width)
	}
}

// fieldHintLine renders a compact one-line contextual hint for the focused
// field, kept to a single row so the control deck can use the full height.
func (e configEditor) fieldHintLine(width int) string {
	field := e.fields[e.focus]
	hint := field.group + " · " + field.label + " — " + e.fieldInstructions(field)
	if e.fieldDisabled(field) {
		hint = "Disabled until Origin mode is Auto or Proxy service. " + e.fieldInstructions(field)
	}
	return theme.Subtle.Render(trim(hint, max(10, width-1)))
}

func (e configEditor) fieldHelpView(width int) string {
	field := e.fields[e.focus]
	lines := []string{
		theme.Title.Render("FIELD"),
		theme.Focus.Render(field.group),
		field.label,
		"",
		e.fieldInstructions(field),
	}
	if e.fieldDisabled(field) {
		lines = append(lines, "", theme.Subtle.Render("Disabled until Origin mode is Auto or Proxy service."))
	}
	return theme.FocusBorder.Width(width).Render(strings.Join(lines, "\n"))
}

func (e configEditor) fieldGuideView(width, height int) string {
	field := e.fields[e.focus]
	guide := e.fieldGuide(field)
	innerWidth := max(20, width-4)
	lines := []string{
		theme.Title.Render("FIELD GUIDE"),
		theme.Focus.Render(field.group + " / " + field.label),
		"",
	}
	if current := e.guideCurrentValue(field); current != "" {
		lines = append(lines, theme.Subtle.Render("Current"), current, "")
	}
	lines = append(lines, wrapStyled(guide.summary, innerWidth)...)
	if len(guide.details) > 0 {
		lines = append(lines, "")
		for _, detail := range guide.details {
			lines = append(lines, wrapBullet(detail, innerWidth)...)
		}
	}
	if e.fieldDisabled(field) {
		lines = append(lines, "", theme.Warn.Render("Disabled"), theme.Subtle.Render("Enable Auto or Proxy service origin mode to edit this field."))
	}
	lines = append(lines, "", theme.Subtle.Render(e.fieldInstructions(field)))
	maxLines := max(4, height-2)
	if len(lines) > maxLines {
		lines = append(lines[:maxLines-1], theme.Subtle.Render("…"))
	}
	body := padLines(lines, max(0, height-2))
	return theme.Border.Width(innerWidth).Render(body)
}

func (e configEditor) fieldInstructions(field editorField) string {
	switch field.kind {
	case "number", "slider":
		switch field.key {
		case "unknown":
			return "Percent of all hits sent as bots/agents. Pick which bots in Bot pool below."
		case "device":
			return "Of the non-bot (human) hits, the percent that are desktop vs mobile."
		case "unique":
			return "Odds each hit uses a fresh IP. Lower means more returning-visitor repeats."
		case "entDevice":
			return "How far each link's desktop share drifts from the base. Switches level to Custom."
		case "entBreakout":
			return "How big the busiest links get. Higher makes a few links dominate. Switches to Custom."
		case "entViral":
			return "Share of links that become breakouts — near-max traffic, rarely idle. Switches to Custom."
		}
		return "Type numbers to replace the value immediately. Backspace edits. Left/right nudges."
	case "select":
		if field.key == "mode" {
			return "Left/right changes the option. Enter moves to the next row."
		}
		if field.key == "entropy" {
			return "Left/right dials entropy. Off = every link identical; Mayhem = wild variation."
		}
		if field.key == "botpool" {
			return "Left/right picks which bots the Bot traffic % draws from. Enter moves to the next row."
		}
		return "Left/right changes the option. Enter moves to the next row."
	case "text", "secret":
		return "Press Enter to edit this text field."
	case "open":
		return "Press Enter to edit URL parameter rules and payloads."
	default:
		return "Use Tab to move, A to apply, G/L to save."
	}
}

func (e configEditor) guideCurrentValue(field editorField) string {
	switch field.key {
	case "iproyal":
		if e.rawValue(field.key) == "" || e.fieldDisabled(field) {
			return "not set"
		}
		return "configured"
	case "params":
		return fmt.Sprintf("%d rules, %d payload variants", len(e.cfg.Requests.URLParams), countPayloads(e.cfg.Requests.URLParams))
	default:
		value := e.displayValuePlain(field)
		if strings.TrimSpace(value) == "" {
			return ""
		}
		return value
	}
}

func (e configEditor) fieldGuide(field editorField) fieldGuide {
	switch field.key {
	case "minRate":
		return fieldGuide{summary: "The lowest hit rate a worker will choose during an active phase.", details: []string{"Each worker picks a rate between min and max.", "Keep this below max hits/min."}}
	case "maxRate":
		return fieldGuide{summary: "The highest hit rate a worker can choose during an active phase.", details: []string{"The actual rate varies per worker and phase.", "Raise this to create burstier traffic."}}
	case "concurrent":
		return fieldGuide{summary: "How many worker goroutines run per target URL.", details: []string{"More workers create more overlapping traffic.", "The app still respects the global worker cap."}}
	case "timeout":
		return fieldGuide{summary: "Request timeout in milliseconds.", details: []string{"Slow or blocked requests fail after this window.", "Higher values can make shutdown wait on stuck networks longer."}}
	case "method":
		return fieldGuide{summary: "HTTP method used for generated traffic.", details: []string{"GET is the normal redirect-testing path.", "HEAD checks reachability with less body traffic.", "POST sends an empty body when you need POST-shaped traffic."}}
	case "unknown":
		return fieldGuide{summary: "Percent of traffic sent as known bots or non-human agents.", details: []string{"The Bot pool field controls which bot identities are eligible.", "Set to 0 for human-only identity rotation."}}
	case "botpool":
		return fieldGuide{summary: "Which bot identities can be used when Bot traffic is above 0%.", details: []string{"All bots includes crawlers, AI agents, fetchers, CLIs, and libraries.", "Use narrower pools to exercise specific analytics buckets."}}
	case "device":
		return fieldGuide{summary: "Desktop share for non-bot traffic.", details: []string{"The rest of human traffic is mobile.", "Bot traffic ignores this ratio and uses each bot's real user agent."}}
	case "unique":
		return fieldGuide{summary: "Chance that a hit receives a fresh synthetic visitor IP.", details: []string{"Lower values create more repeat-visitor behavior.", "This affects identity headers and display stats, not paid proxy exit selection."}}
	case "minActive":
		return fieldGuide{summary: "Shortest active phase length in minutes.", details: []string{"Workers send traffic during active phases.", "Set min and max close together for steadier runs."}}
	case "maxActive":
		return fieldGuide{summary: "Longest active phase length in minutes.", details: []string{"Each worker chooses a duration between active min and max.", "Longer phases make traffic look less stop-start."}}
	case "idleOdds":
		return fieldGuide{summary: "Chance a worker idles after an active phase.", details: []string{"Idle phases create natural gaps between bursts.", "Set to 0 for continuous traffic."}}
	case "minIdle":
		return fieldGuide{summary: "Shortest idle phase length in minutes.", details: []string{"Used only when the worker decides to idle.", "Idle min can be 0 for quick pauses."}}
	case "maxIdle":
		return fieldGuide{summary: "Longest idle phase length in minutes.", details: []string{"Large values create long quiet periods.", "Keep this near min idle for tighter traffic rhythm."}}
	case "entropy":
		return fieldGuide{summary: "How much personality each link gets. Without it, every link converges to the same profile and analytics look flat.", details: []string{
			"Off: every link uses the same device mix, rate, and idle rhythm.",
			"Calm: gentle per-link variation.",
			"Chaos (default): clear differences and the occasional breakout link.",
			"Mayhem: wild variation — a few links dominate hard.",
			"Editing a knob below switches the level to Custom.",
		}}
	case "entDevice":
		return fieldGuide{summary: "How far each link's desktop/mobile mix drifts from the base share.", details: []string{"0 keeps every link at the base device ratio.", "Higher makes some links desktop-heavy and others mobile-heavy."}}
	case "entBreakout":
		return fieldGuide{summary: "How large the busiest links get, on a long-tailed curve.", details: []string{"Most links stay near the base rate; a few climb well above it.", "Higher widens the gap between quiet and busy links."}}
	case "entViral":
		return fieldGuide{summary: "Share of links that become breakout 'viral' links.", details: []string{"Viral links hug the top of the rate range and rarely go idle.", "Set to 0 for no forced breakouts (the tail can still produce busy links)."}}
	case "mode":
		return fieldGuide{summary: "Controls where requests appear to come from. Identity, bot selection, method, and URL params still rotate in every mode.", details: []string{
			"None: direct requests with no geo/IP spoofing headers.",
			"Auto: public domains with valid TLDs use the paid proxy; localhost, .local, IP literals, and internal names stay direct with Vercel geo headers.",
			"Vercel: direct requests with x-forwarded-for, x-real-ip, and x-vercel-ip-* geo headers.",
			"Proxy: every request routes through the configured paid proxy provider; geo spoofing headers are disabled.",
		}}
	case "provider":
		return fieldGuide{summary: "Paid proxy provider used by Proxy mode and by Auto mode for public-domain targets.", details: []string{"Currently supported provider: iproyal.", "Credentials are redacted in config output."}}
	case "iproyal":
		return fieldGuide{summary: "IPRoyal proxy endpoint URL.", details: []string{"Expected format: http://user:pass@host:port.", "Required when Proxy mode is used, or when Auto mode targets public domains."}}
	case "params":
		return fieldGuide{summary: "Probabilistic URL parameter rules and payload variants.", details: []string{"Use this to add campaign, QR, or test params to some hits.", "Payload variants let one rule choose between weighted parameter sets."}}
	default:
		return fieldGuide{summary: e.fieldInstructions(field)}
	}
}

func wrapStyled(text string, width int) []string {
	if width < 12 {
		width = 12
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{}
	line := words[0]
	for _, word := range words[1:] {
		if lipgloss.Width(line)+1+lipgloss.Width(word) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		line += " " + word
	}
	lines = append(lines, line)
	return lines
}

func wrapBullet(text string, width int) []string {
	if width < 12 {
		width = 12
	}
	wrapped := wrapStyled(text, max(8, width-2))
	for i := range wrapped {
		if i == 0 {
			wrapped[i] = theme.Focus.Render("• ") + wrapped[i]
		} else {
			wrapped[i] = "  " + wrapped[i]
		}
	}
	return wrapped
}

func (e configEditor) applyPreviewView(width, height int, err error) string {
	boxWidth := min(92, max(56, width-8))
	lines := []string{
		theme.Title.Render("SAVE & CLOSE"),
		theme.Subtle.Render("Review the settings below. Enter saves to ./.hitmaker.json and closes."),
		"",
	}
	lines = append(lines, e.previewLines()...)
	if err != nil {
		lines = append(lines, "", theme.Bad.Render(err.Error()))
	}
	lines = append(lines, "", theme.PillHot.Render("Enter save & close")+" "+theme.Pill.Render("Esc back"))
	box := theme.FocusBorder.Width(boxWidth).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func (e configEditor) previewLines() []string {
	cfg := e.cfg
	return []string{
		theme.Focus.Render("TRAFFIC"),
		fmt.Sprintf("  %s  %d-%d hits/min across %d worker(s)", meter(float64(cfg.Traffic.MaxPerMin), 100), cfg.Traffic.MinPerMin, cfg.Traffic.MaxPerMin, cfg.Traffic.Concurrent),
		"",
		theme.Focus.Render("IDENTITY"),
		fmt.Sprintf("  Bot traffic %s %d%%", meter(float64(cfg.Requests.UnknownRatio), 100), cfg.Requests.UnknownRatio),
		fmt.Sprintf("  Bot pool    %s", botPoolLabel(cfg.Requests.Bots)),
		fmt.Sprintf("  Desktop     %s %d%% of human hits", meter(float64(cfg.Requests.DeviceRatio), 100), cfg.Requests.DeviceRatio),
		fmt.Sprintf("  Unique IP   %s %.0f%%", meter(cfg.Requests.UniqueIPProb*100, 100), cfg.Requests.UniqueIPProb*100),
		"",
		theme.Focus.Render("SCHEDULE"),
		fmt.Sprintf("  Active %d-%d min, idle %.0f%% for %d-%d min", cfg.Schedule.MinActive, cfg.Schedule.MaxActive, cfg.Schedule.IdleOdds*100, cfg.Schedule.MinIdle, cfg.Schedule.MaxIdle),
		"",
		theme.Focus.Render("ENTROPY"),
		entropyPreviewLine(cfg.Entropy),
		"",
		theme.Focus.Render("ORIGIN"),
		modeLabel(cfg.Origin.Mode),
		"",
		theme.Focus.Render("URL PARAMS"),
		fmt.Sprintf("%d parameter rules, %d payload variants", len(cfg.Requests.URLParams), countPayloads(cfg.Requests.URLParams)),
	}
}

func entropyLevelLabel(level config.EntropyLevel) string {
	switch level {
	case config.EntropyOff:
		return "Off"
	case config.EntropyCalm:
		return "Calm"
	case config.EntropyChaos:
		return "Chaos"
	case config.EntropyMayhem:
		return "Mayhem"
	case config.EntropyCustom:
		return "Custom"
	default:
		return string(level)
	}
}

func entropyPreviewLine(e config.EntropyConfig) string {
	label := entropyLevelLabel(e.Level)
	if e.Level == config.EntropyOff {
		return fmt.Sprintf("  %s — every link identical", label)
	}
	spread, breakout, viral := e.EffectiveHuman()
	return fmt.Sprintf("  %s — audience ±%d%%, breakout %d%%, %d%% viral links", label, spread, breakout, viral)
}

func (e configEditor) paramsView(width int) string {
	lines := []string{theme.Title.Render("URL PARAMS"), theme.Subtle.Render("N add  X delete  E edit  Enter payloads  Esc back")}
	for i, param := range e.cfg.Requests.URLParams {
		cursor := " "
		style := lipgloss.NewStyle()
		if i == e.paramFocus {
			cursor = theme.Focus.Render("▸")
			style = style.Background(theme.Panel)
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s %-12s = %-12s %5.0f%%  payloads %d",
			cursor, param.Key, param.Value, param.Probability, len(param.Payloads))))
	}
	if len(e.cfg.Requests.URLParams) == 0 {
		lines = append(lines, theme.Subtle.Render("No params. Press N to add one."))
	}
	return theme.FocusBorder.Width(width).Render(strings.Join(lines, "\n"))
}

func (e configEditor) payloadsView(width int) string {
	param := e.cfg.Requests.URLParams[e.paramFocus]
	lines := []string{theme.Title.Render("PAYLOADS for " + param.Key), theme.Subtle.Render("N add  X delete  E edit  Esc back")}
	for i, payload := range param.Payloads {
		cursor := " "
		style := lipgloss.NewStyle()
		if i == e.payloadFocus {
			cursor = theme.Focus.Render("▸")
			style = style.Background(theme.Panel)
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s %-16s weight %-5.1f %s", cursor, payload.Name, payload.Weight, kvPreview(payload.KV))))
	}
	if len(param.Payloads) == 0 {
		lines = append(lines, theme.Subtle.Render("No payloads. Press N to add one."))
	}
	return theme.FocusBorder.Width(width).Render(strings.Join(lines, "\n"))
}

func (e configEditor) displayValue(field editorField) string {
	switch field.kind {
	case "slider":
		return slider(e.numberValue(field.key), field.min, field.max)
	case "select":
		if field.key == "mode" {
			return radioSegment(e.selectDisplayValue(field.key), e.selectOptions(field.key))
		}
		if field.key == "botpool" {
			return theme.PillHot.Render("● " + e.selectLabel(field.key, e.selectDisplayValue(field.key)))
		}
		return radioSegment(e.selectDisplayValue(field.key), e.selectOptions(field.key))
	case "secret":
		if e.fieldDisabled(field) {
			return theme.Subtle.Render("disabled")
		}
		value := e.rawValue(field.key)
		if value == "" {
			return theme.Subtle.Render("not set")
		}
		return "••••••••"
	case "open":
		return fmt.Sprintf("%d rules", len(e.cfg.Requests.URLParams))
	default:
		if e.fieldDisabled(field) {
			return theme.Subtle.Render(e.rawValue(field.key))
		}
		return e.rawValue(field.key)
	}
}

// displayValuePlain mirrors displayValue but without any ANSI styling, for use
// inside a full-row highlight where nested color codes would break the fill.
func (e configEditor) displayValuePlain(field editorField) string {
	switch field.kind {
	case "slider":
		return sliderPlain(e.numberValue(field.key), field.min, field.max)
	case "select":
		if field.key == "mode" {
			return radioSegmentPlain(e.currentSelectValue(field.key), e.selectOptions(field.key))
		}
		if field.key == "botpool" {
			return "● " + botPoolLabel(e.cfg.Requests.Bots)
		}
		return radioSegmentPlain(e.currentSelectValue(field.key), e.selectOptions(field.key))
	case "secret":
		if e.fieldDisabled(field) {
			return "disabled"
		}
		if e.rawValue(field.key) == "" {
			return "not set"
		}
		return "••••••••"
	case "open":
		return fmt.Sprintf("%d rules", len(e.cfg.Requests.URLParams))
	default:
		return e.rawValue(field.key)
	}
}

func sliderPlain(value, minValue, maxValue float64) string {
	const cells = 16
	ratio := 0.0
	if maxValue > minValue {
		ratio = (value - minValue) / (maxValue - minValue)
	}
	filled := clampInt(int(ratio*cells+0.5), 0, cells)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", cells-filled)
	if maxValue == 1 {
		return fmt.Sprintf("%s %.0f%%", bar, value*100)
	}
	return fmt.Sprintf("%s %.0f", bar, value)
}

func radioSegmentPlain(active string, options []selectOption) string {
	parts := make([]string, 0, len(options))
	for _, option := range options {
		if option.value == active {
			parts = append(parts, "● "+option.label)
		} else {
			parts = append(parts, "○ "+option.label)
		}
	}
	return strings.Join(parts, " ")
}

func (e configEditor) rawValue(key string) string {
	switch key {
	case "minRate":
		return strconv.Itoa(e.cfg.Traffic.MinPerMin)
	case "maxRate":
		return strconv.Itoa(e.cfg.Traffic.MaxPerMin)
	case "concurrent":
		return strconv.Itoa(e.cfg.Traffic.Concurrent)
	case "timeout":
		return strconv.Itoa(e.cfg.Traffic.TimeoutMs)
	case "method":
		return e.cfg.Requests.Method
	case "device":
		return strconv.Itoa(e.cfg.Requests.DeviceRatio)
	case "unknown":
		return strconv.Itoa(e.cfg.Requests.UnknownRatio)
	case "unique":
		return fmt.Sprintf("%.2f", e.cfg.Requests.UniqueIPProb)
	case "minActive":
		return strconv.Itoa(e.cfg.Schedule.MinActive)
	case "maxActive":
		return strconv.Itoa(e.cfg.Schedule.MaxActive)
	case "idleOdds":
		return fmt.Sprintf("%.2f", e.cfg.Schedule.IdleOdds)
	case "minIdle":
		return strconv.Itoa(e.cfg.Schedule.MinIdle)
	case "maxIdle":
		return strconv.Itoa(e.cfg.Schedule.MaxIdle)
	case "entropy":
		return string(e.cfg.Entropy.Level)
	case "entDevice":
		spread, _, _ := e.cfg.Entropy.EffectiveHuman()
		return strconv.Itoa(spread)
	case "entBreakout":
		_, breakout, _ := e.cfg.Entropy.EffectiveHuman()
		return strconv.Itoa(breakout)
	case "entViral":
		_, _, viral := e.cfg.Entropy.EffectiveHuman()
		return strconv.Itoa(viral)
	case "mode":
		return string(e.cfg.Origin.Mode)
	case "botpool":
		return botPoolLabel(e.cfg.Requests.Bots)
	case "provider":
		return e.cfg.Origin.Provider
	case "iproyal":
		if e.cfg.Origin.ProviderConfig == nil {
			return ""
		}
		return e.cfg.Origin.ProviderConfig["url"]
	default:
		return ""
	}
}

func (e configEditor) numberValue(key string) float64 {
	value, _ := strconv.ParseFloat(e.rawValue(key), 64)
	return value
}

func slider(value, minValue, maxValue float64) string {
	const cells = 16
	ratio := 0.0
	if maxValue > minValue {
		ratio = (value - minValue) / (maxValue - minValue)
	}
	filled := clampInt(int(ratio*cells+0.5), 0, cells)
	// Blocky retro meter — a nod to the old block-graphics look.
	bar := theme.Focus.Render(strings.Repeat("█", filled)) + theme.Subtle.Render(strings.Repeat("░", cells-filled))
	if maxValue == 1 {
		return fmt.Sprintf("%s %.0f%%", bar, value*100)
	}
	return fmt.Sprintf("%s %.0f", bar, value)
}

func meter(value, maxValue float64) string {
	const cells = 12
	ratio := 0.0
	if maxValue > 0 {
		ratio = value / maxValue
	}
	filled := clampInt(int(ratio*cells+0.5), 0, cells)
	return theme.Focus.Render(strings.Repeat("▰", filled)) + theme.Subtle.Render(strings.Repeat("▱", cells-filled))
}

func segment(active string, values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if strings.EqualFold(value, active) {
			parts = append(parts, theme.PillHot.Render(value))
		} else {
			parts = append(parts, theme.Pill.Render(value))
		}
	}
	return strings.Join(parts, " ")
}

func radioSegment(active string, options []selectOption) string {
	parts := make([]string, 0, len(options))
	for _, option := range options {
		label := "○ " + option.label
		if option.value == active {
			label = "● " + option.label
			parts = append(parts, theme.PillHot.Render(label))
		} else {
			parts = append(parts, theme.Pill.Render(label))
		}
	}
	return strings.Join(parts, " ")
}

func modeSegment(active config.Mode) string {
	values := []config.Mode{config.ModeNone, config.ModeAuto, config.ModeVercel, config.ModeProxy}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		label := modeLabel(value)
		if value == active {
			parts = append(parts, theme.PillHot.Render(label))
		} else {
			parts = append(parts, theme.Pill.Render(label))
		}
	}
	return strings.Join(parts, " ")
}

func modeLabel(mode config.Mode) string {
	switch mode {
	case config.ModeNone:
		return "None"
	case config.ModeAuto:
		return "Auto"
	case config.ModeVercel:
		return "Vercel geo headers (spoofing)"
	case config.ModeProxy:
		return "Proxy service"
	default:
		return string(mode)
	}
}

// botPoolPresets are the one-key bot-pool choices offered in the editor. spec is
// the token list written to config.Requests.Bots (nil = the whole catalog).
var botPoolPresets = []struct {
	label string
	spec  []string
}{
	{"All bots", nil},
	{"AI (crawlers + assistants)", []string{"ai"}},
	{"AI crawlers", []string{"ai_crawler"}},
	{"AI assistants", []string{"ai_assistant"}},
	{"Search crawlers", []string{"crawler"}},
	{"Social fetchers", []string{"fetcher"}},
	{"CLI clients", []string{"cli"}},
	{"Libraries", []string{"library"}},
}

// botPoolIndex returns the preset index for the current spec, or -1 if the spec
// is a custom name/category list set outside the editor (e.g. via `config set`).
func botPoolIndex(bots []string) int {
	if len(bots) == 0 {
		return 0
	}
	if len(bots) == 1 {
		for i, preset := range botPoolPresets {
			if len(preset.spec) == 1 && strings.EqualFold(preset.spec[0], bots[0]) {
				return i
			}
		}
	}
	return -1
}

func botPoolLabel(bots []string) string {
	if idx := botPoolIndex(bots); idx >= 0 {
		return botPoolPresets[idx].label
	}
	return fmt.Sprintf("Custom (%d)", len(bots))
}

func rotate(current string, values []string, dir int) string {
	idx := 0
	for i, value := range values {
		if strings.EqualFold(value, current) {
			idx = i
			break
		}
	}
	idx = (idx + dir) % len(values)
	if idx < 0 {
		idx += len(values)
	}
	return values[idx]
}

func nextEnabledField(fields []editorField, current, dir int, cfg config.Config) int {
	if len(fields) == 0 {
		return current
	}
	idx := current
	probe := configEditor{cfg: cfg}
	for i := 0; i < len(fields); i++ {
		idx = (idx + dir) % len(fields)
		if idx < 0 {
			idx += len(fields)
		}
		if !probe.fieldDisabled(fields[idx]) {
			return idx
		}
	}
	return current
}

func keyLabel(key string) string {
	switch key {
	case "minRate":
		return "min hits/min"
	case "maxRate":
		return "max hits/min"
	case "concurrent":
		return "workers/target"
	case "timeout":
		return "timeout"
	case "device":
		return "desktop share"
	case "unknown":
		return "bot traffic"
	case "unique":
		return "unique IP odds"
	case "idleOdds":
		return "idle odds"
	case "entropy":
		return "entropy level"
	case "entDevice":
		return "audience spread"
	case "entBreakout":
		return "breakout intensity"
	case "entViral":
		return "viral links"
	default:
		return key
	}
}

func paramToLine(param config.URLParam) string {
	return fmt.Sprintf("%s=%s %.0f", param.Key, param.Value, param.Probability)
}

func lineToParam(line string, fallback config.URLParam) (config.URLParam, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return fallback, fmt.Errorf("param cannot be empty")
	}
	keyValue := strings.SplitN(parts[0], "=", 2)
	fallback.Key = keyValue[0]
	fallback.Value = ""
	if len(keyValue) == 2 {
		fallback.Value = keyValue[1]
	}
	if len(parts) > 1 {
		p, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return fallback, err
		}
		fallback.Probability = p
	}
	return fallback, nil
}

func payloadToLine(payload config.Payload) string {
	return fmt.Sprintf("%s %.1f %s", payload.Name, payload.Weight, kvPreview(payload.KV))
}

func lineToPayload(line string, fallback config.Payload) (config.Payload, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return fallback, fmt.Errorf("payload format: name weight key=value,key=value")
	}
	fallback.Name = parts[0]
	weight, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fallback, err
	}
	fallback.Weight = weight
	if len(parts) > 2 {
		fallback.KV = map[string]string{}
		for _, pair := range strings.Split(parts[2], ",") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 && kv[0] != "" {
				fallback.KV[kv[0]] = kv[1]
			}
		}
	}
	return fallback, nil
}

func kvPreview(kv map[string]string) string {
	if len(kv) == 0 {
		return theme.Subtle.Render("no kv")
	}
	parts := []string{}
	for key, value := range kv {
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, ",")
}

func countPayloads(params []config.URLParam) int {
	total := 0
	for _, param := range params {
		total += len(param.Payloads)
	}
	return total
}

func clipAround(lines []string, focus, maxLines int) []string {
	if maxLines <= 0 || len(lines) <= maxLines {
		return lines
	}
	if maxLines < 4 {
		maxLines = 4
	}
	start := focus - maxLines/2
	if start < 0 {
		start = 0
	}
	end := start + maxLines
	if end > len(lines) {
		end = len(lines)
		start = max(0, end-maxLines)
	}
	out := append([]string(nil), lines[start:end]...)
	if start > 0 {
		out[0] = theme.Subtle.Render("↑ more")
	}
	if end < len(lines) {
		out[len(out)-1] = theme.Subtle.Render("↓ more")
	}
	return out
}

func atoi(value string, fallback int) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func atof(value string, fallback float64) float64 {
	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return n
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func clampFloat(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func normalizedKey(msg tea.KeyMsg) string {
	// v2: printable keys carry their characters in Key().Text; special keys
	// (enter, esc, tab, …) report an empty Text and are named by String().
	if text := msg.Key().Text; text != "" {
		if r := []rune(text); len(r) == 1 {
			return strings.ToLower(string(r[0]))
		}
	}
	return strings.ToLower(msg.String())
}
