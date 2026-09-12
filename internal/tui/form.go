package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type fieldSpec struct {
	label       string
	placeholder string
	password    bool
}

type formConfig[T any] struct {
	addTitle  string
	editTitle string
	fields    []fieldSpec
	id        func(T) int
	load      func(orig T, in []textinput.Model)
	build     func(id int, in []textinput.Model) T
	save      func(T) tea.Cmd
}

type formScreen[T any] struct {
	cfg      formConfig[T]
	inputs   []textinput.Model
	focus    int
	original T
	err      error
}

func newFormScreen[T any](cfg formConfig[T], original T) *formScreen[T] {
	s := &formScreen[T]{cfg: cfg, original: original}
	s.inputs = make([]textinput.Model, len(cfg.fields))

	for i, f := range cfg.fields {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 128
		ti.SetWidth(40)
		ti.Placeholder = f.placeholder
		if f.password {
			ti.EchoMode = textinput.EchoPassword
		}
		s.inputs[i] = ti
	}

	cfg.load(original, s.inputs)
	s.inputs[0].Focus()

	return s
}

func (fs *formScreen[T]) Title() string {
	if fs.cfg.id(fs.original) == 0 {
		return fs.cfg.addTitle
	}

	return fs.cfg.editTitle
}

func (fs *formScreen[T]) Init() tea.Cmd {
	return textinput.Blink
}

func (fs *formScreen[T]) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case savedMsg:
		if msg.err != nil {
			fs.err = msg.err
			return fs, nil
		}
		return fs, pop

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return fs, pop
		case "tab", "down":
			return fs, fs.focusOn(fs.focus + 1)
		case "shift+tab", "up":
			return fs, fs.focusOn(fs.focus - 1)
		case "enter":
			if fs.focus < len(fs.inputs)-1 {
				return fs, fs.focusOn(fs.focus + 1)
			}

			v := fs.cfg.build(fs.cfg.id(fs.original), fs.inputs)
			fs.err = nil
			return fs, fs.cfg.save(v)
		}
	}

	var cmd tea.Cmd
	fs.inputs[fs.focus], cmd = fs.inputs[fs.focus].Update(msg)
	return fs, cmd
}

func (fs *formScreen[T]) View() string {
	var b strings.Builder

	for i := range fs.inputs {
		style, marker := blurredStyle, "  "
		if i == fs.focus {
			style, marker = focusedStyle, "> "
		}
		b.WriteString(marker)
		b.WriteString(style.Render(labelStyle.Render(fs.cfg.fields[i].label)))
		b.WriteString(fs.inputs[i].View())
		b.WriteString("\n")
	}

	if fs.err != nil {
		b.WriteString("\n" + errorStyle.Render(fs.err.Error()) + "\n")
	}

	b.WriteString("\ntab/shift+tab move · enter next or save · esc cancel")
	return b.String()
}

func (fs *formScreen[T]) capturesInput() bool {
	return true
}

func (fs *formScreen[T]) focusOn(i int) tea.Cmd {
	fc := len(fs.cfg.fields)
	fs.focus = ((i % fc) + fc) % fc
	var cmd tea.Cmd

	for j := range fs.inputs {
		if j == fs.focus {
			cmd = fs.inputs[j].Focus()
		} else {
			fs.inputs[j].Blur()
		}
	}

	return cmd
}

func parseTags(raw string) []string {
	var tags []string
	for _, t := range strings.Split(raw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
