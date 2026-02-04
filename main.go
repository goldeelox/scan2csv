package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	Version  string = "dev"
	Revision string = "unknown"
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
	ti.Width = 24
	ti.Cursor.BlinkSpeed = 500 * time.Millisecond
	ti.PromptStyle = ti.PromptStyle.PaddingTop(1)
	ti.Focus()

	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "Timestamp", Width: 22},
			{Title: "Barcode", Width: 50},
		}),
		table.WithHeight(11),
	)

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

func statusMsg(msg string) string {
	return fmt.Sprintf("%s %s", time.Now().Format(time.DateTime), msg)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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
	case tea.KeyMsg:
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

	m.textInput, cmd = m.textInput.Update(msg)
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		PaddingTop(1).
		Faint(true)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.table.View(),
			m.textInput.View(),
			statusStyle.Render(m.status),
		),
	)
}

type StatusOk string

func (s StatusOk) String() string {
	return string(s)
}

func main() {
	config := ParseConfig()

	if config.Version {
		fmt.Printf("version=%s revision=%s\n", Version, Revision)
		return
	}
	if !config.RemovableDisksSupported {
		fmt.Printf("removable disk support not available\n")
		os.Exit(1)
	}
	if config.RemovableDisksSupported && config.DetectRemovableDisks {
		detectRemovableDisks()
		return
	}

	m := New(config)
	p := tea.NewProgram(m)

	if config.RemovableDisksSupported && len(config.RemovableDiskUUID) > 0 {
		disk := NewRemovableDisk(config.RemovableDiskUUID, config.AutoMount)
		go func() {
			for {
				time.Sleep(2 * time.Second)

				if disk.AutoMount && disk.isAttached() {
					if err := disk.mount(); err != nil {
						p.Send(err)
					}
				}

				if !disk.isMounted() {
					continue
				}

				m.wg.Add(1)
				func() {
					defer m.wg.Done()
					fsys := os.DirFS(m.config.OutputDir)
					matches, _ := fs.Glob(fsys, "scans*.csv")
					for _, match := range matches {
						disk.copyFile(m.config.OutputDir + "/" + match)
					}
					p.Send(StatusOk("backup complete"))

					if err := disk.unmount(); err != nil {
						p.Send(err)
						return
					}

					if err := disk.poweroff(); err != nil {
						p.Send(err)
					}
				}()
			}
		}()
	}

	if _, err := p.Run(); err != nil {
		log.Printf("ut oh %v", err)
	}

	log.Printf("waiting for all jobs to finish...")
	m.wg.Wait()
	log.Printf("bye")
}
