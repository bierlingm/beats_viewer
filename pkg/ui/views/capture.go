package views

import (
	"strings"

	"beats_viewer/pkg/model"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	captureTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4"))

	captureLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	captureSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#73F59F"))
)

type CaptureView struct {
	textarea textarea.Model
	width    int
	height   int

	channel      model.Channel
	source       model.Source
	focusField   int // 0=channel, 1=source, 2=textarea
	submitted    bool
	cancelled    bool
}

func NewCaptureView(width, height int) *CaptureView {
	ta := textarea.New()
	ta.Placeholder = "Your insight here..."
	ta.Focus()
	ta.SetWidth(width - 6)
	ta.SetHeight(5)
	ta.ShowLineNumbers = false

	return &CaptureView{
		textarea:   ta,
		width:      width,
		height:     height,
		channel:    model.ChannelReflection,
		source:     model.SourceInternal,
		focusField: 2,
	}
}

func (cv *CaptureView) SetSize(width, height int) {
	cv.width = width
	cv.height = height
	cv.textarea.SetWidth(width - 6)
}

func (cv *CaptureView) SetChannel(ch model.Channel) {
	cv.channel = ch
}

func (cv *CaptureView) SetSource(src model.Source) {
	cv.source = src
}

func (cv *CaptureView) Reset() {
	cv.textarea.Reset()
	cv.submitted = false
	cv.cancelled = false
	cv.focusField = 2
	cv.textarea.Focus()
}

func (cv *CaptureView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			cv.focusField = (cv.focusField + 1) % 3
			if cv.focusField == 2 {
				cv.textarea.Focus()
			} else {
				cv.textarea.Blur()
			}
			return nil
		case "shift+tab":
			cv.focusField = (cv.focusField + 2) % 3
			if cv.focusField == 2 {
				cv.textarea.Focus()
			} else {
				cv.textarea.Blur()
			}
			return nil
		case "enter":
			if cv.focusField == 0 {
				cv.cycleChannel()
				return nil
			} else if cv.focusField == 1 {
				cv.cycleSource()
				return nil
			} else if msg.Alt {
				cv.submitted = true
				return nil
			}
		case "ctrl+s":
			cv.submitted = true
			return nil
		case "esc":
			cv.cancelled = true
			return nil
		case "left":
			if cv.focusField == 0 {
				cv.cycleChannelBack()
				return nil
			} else if cv.focusField == 1 {
				cv.cycleSourceBack()
				return nil
			}
		case "right":
			if cv.focusField == 0 {
				cv.cycleChannel()
				return nil
			} else if cv.focusField == 1 {
				cv.cycleSource()
				return nil
			}
		}

		if cv.focusField == 2 {
			var cmd tea.Cmd
			cv.textarea, cmd = cv.textarea.Update(msg)
			return cmd
		}
	}

	return nil
}

func (cv *CaptureView) cycleChannel() {
	channels := model.AllChannels()
	for i, ch := range channels {
		if ch == cv.channel {
			cv.channel = channels[(i+1)%len(channels)]
			return
		}
	}
	cv.channel = channels[0]
}

func (cv *CaptureView) cycleChannelBack() {
	channels := model.AllChannels()
	for i, ch := range channels {
		if ch == cv.channel {
			cv.channel = channels[(i+len(channels)-1)%len(channels)]
			return
		}
	}
	cv.channel = channels[0]
}

func (cv *CaptureView) cycleSource() {
	sources := model.AllSources()
	for i, src := range sources {
		if src == cv.source {
			cv.source = sources[(i+1)%len(sources)]
			return
		}
	}
	cv.source = sources[0]
}

func (cv *CaptureView) cycleSourceBack() {
	sources := model.AllSources()
	for i, src := range sources {
		if src == cv.source {
			cv.source = sources[(i+len(sources)-1)%len(sources)]
			return
		}
	}
	cv.source = sources[0]
}

func (cv *CaptureView) IsSubmitted() bool {
	return cv.submitted
}

func (cv *CaptureView) IsCancelled() bool {
	return cv.cancelled
}

func (cv *CaptureView) GetContent() string {
	return cv.textarea.Value()
}

func (cv *CaptureView) GetChannel() model.Channel {
	return cv.channel
}

func (cv *CaptureView) GetSource() model.Source {
	return cv.source
}

func (cv *CaptureView) View() string {
	var sb strings.Builder

	sb.WriteString(captureTitleStyle.Render("btv capture"))
	sb.WriteString("\n\n")

	channelLabel := captureLabelStyle.Render("Channel:")
	channelValue := cv.channel.String()
	if cv.focusField == 0 {
		channelValue = captureSelectedStyle.Render("[" + channelValue + " ▼]")
	} else {
		channelValue = "[" + channelValue + " ▼]"
	}

	sourceLabel := captureLabelStyle.Render("Source:")
	sourceValue := cv.source.String()
	if cv.focusField == 1 {
		sourceValue = captureSelectedStyle.Render("[" + sourceValue + " ▼]")
	} else {
		sourceValue = "[" + sourceValue + " ▼]"
	}

	sb.WriteString(channelLabel + " " + channelValue + "    " + sourceLabel + " " + sourceValue)
	sb.WriteString("\n\n")

	textareaStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#383838"))
	if cv.focusField == 2 {
		textareaStyle = textareaStyle.BorderForeground(lipgloss.Color("#7D56F4"))
	}
	sb.WriteString(textareaStyle.Render(cv.textarea.View()))
	sb.WriteString("\n\n")

	sb.WriteString(captureLabelStyle.Render("Ctrl+S: Save    Tab: Next field    Esc: Cancel"))

	return lipgloss.NewStyle().
		Width(cv.width).
		Height(cv.height).
		Padding(1, 2).
		Render(sb.String())
}
