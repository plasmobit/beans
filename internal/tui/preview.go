package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/pkg/bean"
	"github.com/hmans/beans/internal/ui"
)

// previewModel is a read-only detail preview for the two-column layout.
// It has no focus; the only interaction is scrolling the body.
type previewModel struct {
	bean   *bean.Bean
	width  int
	height int
	scroll int // first visible body line
}

func newPreviewModel(b *bean.Bean, width, height int) previewModel {
	return previewModel{
		bean:   b,
		width:  width,
		height: height,
	}
}

func (m previewModel) View() string {
	if m.bean == nil {
		return m.renderEmpty()
	}
	return m.renderBean()
}

func (m previewModel) renderEmpty() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ui.ColorMuted)

	return style.Render("No bean selected")
}

// previewScrollStep is the number of body lines one mouse wheel notch scrolls.
const previewScrollStep = 3

func (m previewModel) renderBean() string {
	header := m.renderHeader()
	bodyLines := m.bodyLines()
	window := m.bodyWindow(header)
	scroll := min(max(0, m.scroll), max(0, len(bodyLines)-window))
	visible := bodyLines[scroll:min(len(bodyLines), scroll+window)]
	if scroll+window < len(bodyLines) && len(visible) > 0 {
		visible[len(visible)-1] = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("...")
	}

	content := header + "\n" + strings.Join(visible, "\n")

	// Truncate content to fit within available height
	// Border takes 2 lines (top + bottom), padding takes 0 vertical
	innerHeight := m.height - 2
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > innerHeight {
		contentLines = contentLines[:innerHeight]
	}
	content = strings.Join(contentLines, "\n")

	// Border - use exact height to prevent overflow
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorMuted).
		Padding(0, 1).
		Width(m.width - 2).
		Height(innerHeight)

	result := borderStyle.Render(content)

	// Ensure output is exactly m.height lines
	// When truncating, preserve the bottom border (last line)
	resultLines := strings.Split(result, "\n")
	if len(resultLines) > m.height {
		// Keep first (m.height-1) lines + the last line (bottom border)
		bottomBorder := resultLines[len(resultLines)-1]
		resultLines = resultLines[:m.height-1]
		resultLines = append(resultLines, bottomBorder)
		result = strings.Join(resultLines, "\n")
	}

	return result
}

// scrollBy moves the body window by delta lines, clamped to the body.
func (m *previewModel) scrollBy(delta int) {
	if m.bean == nil {
		return
	}
	maxScroll := max(0, len(m.bodyLines())-m.bodyWindow(m.renderHeader()))
	m.scroll = min(max(0, m.scroll+delta), maxScroll)
}

// wrap breaks s into the lines the bordered, padded pane displays, so that
// line counts match the screen.
func (m previewModel) wrap(s string) string {
	return lipgloss.NewStyle().Width(max(1, m.width-4)).Render(s)
}

// renderHeader renders ID, title, metadata and tags, ending in a blank line.
func (m previewModel) renderHeader() string {
	// Header: ID and Title
	idStyle := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	header := idStyle.Render(m.bean.ID) + "\n" + titleStyle.Render(m.bean.Title)

	// Metadata: Status, Type, Priority
	metaStyle := lipgloss.NewStyle().Foreground(ui.ColorMuted)
	meta := metaStyle.Render("Status: " + m.bean.Status + "  Type: " + m.bean.Type)
	if m.bean.Priority != "" && m.bean.Priority != "normal" {
		meta += metaStyle.Render("  Priority: " + m.bean.Priority)
	}

	// Tags
	var tagsLine string
	if len(m.bean.Tags) > 0 {
		tagsLine = ui.RenderTags(m.bean.Tags)
	}

	// Compose
	var parts []string
	parts = append(parts, header)
	parts = append(parts, "")
	parts = append(parts, meta)
	if tagsLine != "" {
		parts = append(parts, tagsLine)
	}
	parts = append(parts, "")

	return m.wrap(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

// bodyWindow returns how many body lines fit below the given header.
func (m previewModel) bodyWindow(header string) int {
	return max(1, m.height-2-lipgloss.Height(header))
}

// bodyLines returns the rendered body, wrapped to the pane width.
func (m previewModel) bodyLines() []string {
	return strings.Split(m.wrap(m.renderBody()), "\n")
}

func (m previewModel) renderBody() string {
	if m.bean.Body == "" {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No description")
	}

	// Render markdown (reuse existing glamour renderer from detail.go)
	renderer := getGlamourRenderer()
	if renderer == nil {
		return m.bean.Body
	}

	rendered, err := renderer.Render(m.bean.Body)
	if err != nil {
		return m.bean.Body
	}

	return strings.TrimSpace(rendered)
}
