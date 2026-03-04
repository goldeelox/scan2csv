package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	textInput textinput.Model
	table     table.Model
	wg        *sync.WaitGroup
	config    Config
	status    string
	width     int
	height    int
}

func New(config Config) model {
	ti := textinput.New()
	ti.Placeholder = "Scan barcode..."
	ti.SetWidth(24)
	ti.Focus()

	filename := fmt.Sprintf("%s/scans_%s.csv", config.OutputDir, time.Now().Format(time.DateOnly))
	rows := loadRows(filename)
	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "Timestamp", Width: 22},
			{Title: "Barcode", Width: 50},
		}),
		table.WithRows(rows),
		table.WithHeight(11),
		table.WithWidth(76),
	)
	tbl.GotoBottom()

	tblStyle := table.DefaultStyles()
	tblStyle.Header = tblStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	tblStyle.Selected = tblStyle.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	tbl.SetStyles(tblStyle)

	return model{
		textInput: ti,
		table:     tbl,
		wg:        &sync.WaitGroup{},
		config:    config,
	}
}

func (m model) writeToFile(msg string) tea.Cmd {
	return func() tea.Msg {
		defer m.wg.Done()

		now := time.Now()
		filename := fmt.Sprintf("%s/scans_%s.csv", m.config.OutputDir, now.Format(time.DateOnly))

		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = fmt.Fprintf(f, "%s,%s\n", time.Now().Format(time.DateTime), msg); err != nil {
			return err
		}
		return nil
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var inputCmd, tableCmd tea.Cmd

	switch msg := msg.(type) {
	case error:
		m.status = statusMsg(msg.Error())
		return m, nil
	case StatusOk:
		m.status = statusMsg(msg.String())
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.table.Focused() {
				m.table.GotoBottom()
				m.table.Blur()
				m.textInput.Focus()
			} else {
				m.textInput.Blur()
				m.table.Focus()
			}
		case "enter":
			if !m.textInput.Focused() {
				m.table.GotoBottom()
				m.table.Blur()
				m.textInput.Focus()
				return m, nil
			}

			input := m.textInput.Value()
			m.textInput.Reset()
			if len(input) == 0 {
				return m, nil
			}
			now := time.Now().Format(time.DateTime)
			rows := append(m.table.Rows(), table.Row{now, input})
			m.table.SetRows(rows)
			m.table.GotoBottom()
			m.wg.Add(1)
			return m, m.writeToFile(input)
		}
	}

	m.textInput, inputCmd = m.textInput.Update(msg)
	m.table, tableCmd = m.table.Update(msg)
	return m, tea.Batch(inputCmd, tableCmd)
}

func (m model) View() tea.View {
	textInputStyle := lipgloss.NewStyle().
		PaddingTop(1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		PaddingTop(1).
		Width(72).
		Faint(true)

	content := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.table.View(),
			textInputStyle.Render(m.textInput.View()),
			statusStyle.Render(m.status),
		),
	)
	v := tea.NewView(content)
	return v
}

type StatusOk string

func (s StatusOk) String() string {
	return string(s)
}

func statusMsg(msg string) string {
	return fmt.Sprintf("%s %s", time.Now().Format(time.DateTime), msg)
}
