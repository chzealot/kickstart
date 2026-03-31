package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Icons
	iconSuccess = successStyle.Render("✔")
	iconError   = errorStyle.Render("✘")
	iconWarn    = warnStyle.Render("⚠")
	iconInfo    = infoStyle.Render("ℹ")
	iconArrow   = infoStyle.Render("→")
)

// Title prints a styled title.
func Title(text string) {
	fmt.Println(titleStyle.Render(text))
}

// Success prints a success message.
func Success(format string, a ...any) {
	fmt.Printf("%s %s\n", iconSuccess, fmt.Sprintf(format, a...))
}

// Error prints an error message.
func Error(format string, a ...any) {
	fmt.Printf("%s %s\n", iconError, fmt.Sprintf(format, a...))
}

// Warn prints a warning message.
func Warn(format string, a ...any) {
	fmt.Printf("%s %s\n", iconWarn, fmt.Sprintf(format, a...))
}

// Info prints an info message.
func Info(format string, a ...any) {
	fmt.Printf("%s %s\n", iconInfo, fmt.Sprintf(format, a...))
}

// Step prints a step/action message.
func Step(format string, a ...any) {
	fmt.Printf("%s %s\n", iconArrow, fmt.Sprintf(format, a...))
}

// Dim prints dimmed/secondary text.
func Dim(format string, a ...any) {
	fmt.Println(dimStyle.Render(fmt.Sprintf(format, a...)))
}

// Section prints a section header with a divider.
func Section(title string) {
	fmt.Println()
	fmt.Println(titleStyle.Render("── " + title + " ──"))
}

// spinnerModel is a bubbletea model for the spinner.
type spinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
}

type doneMsg struct{}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case doneMsg:
		m.done = true
		return m, tea.Quit
	case tea.KeyMsg:
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

// SpinnerHandle controls a running spinner.
type SpinnerHandle struct {
	program *tea.Program
}

// Stop stops the spinner.
func (h *SpinnerHandle) Stop() {
	h.program.Send(doneMsg{})
	h.program.Wait()
}

// StartSpinner starts a spinner with a message and returns a handle to stop it.
func StartSpinner(message string) *SpinnerHandle {
	s := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("69"))),
	)

	m := spinnerModel{spinner: s, message: message}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))

	go func() {
		if _, err := p.Run(); err != nil {
			// Silently ignore spinner errors
		}
	}()

	// Give the spinner a moment to start rendering
	time.Sleep(50 * time.Millisecond)

	return &SpinnerHandle{program: p}
}
