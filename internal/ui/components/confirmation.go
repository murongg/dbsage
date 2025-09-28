package components

import (
	"fmt"
	"io"
	"strings"

	"dbsage/internal/models"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationItem represents an item in the confirmation list
type ConfirmationItem struct {
	title       string
	description string
	action      string
}

func (i ConfirmationItem) Title() string       { return i.title }
func (i ConfirmationItem) Description() string { return i.description }
func (i ConfirmationItem) FilterValue() string { return i.title }

// Custom styles for confirmation list - minimal padding
var (
	confirmationItemStyle         = lipgloss.NewStyle().PaddingLeft(1)
	confirmationSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
)

// confirmationDelegate implements list.ItemDelegate for confirmation items
type confirmationDelegate struct{}

func NewConfirmationDelegate() confirmationDelegate {
	return confirmationDelegate{}
}

func (d confirmationDelegate) Height() int                             { return 1 }
func (d confirmationDelegate) Spacing() int                            { return 0 }
func (d confirmationDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d confirmationDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ConfirmationItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := confirmationItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return confirmationSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// CreateConfirmationList creates a new confirmation list with custom styling
func CreateConfirmationList() list.Model {
	confirmationList := list.New([]list.Item{}, NewConfirmationDelegate(), 0, 0)
	confirmationList.Title = ""
	confirmationList.SetShowStatusBar(false)
	confirmationList.SetFilteringEnabled(false)
	confirmationList.SetShowHelp(false)
	confirmationList.SetShowTitle(true)

	// Set custom styles - minimal spacing
	confirmationList.Styles.Title = lipgloss.NewStyle().MarginLeft(0)
	confirmationList.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(0)
	confirmationList.Styles.HelpStyle = lipgloss.NewStyle().PaddingLeft(0)

	return confirmationList
}

// SetupConfirmationList populates the confirmation list with tool info
func SetupConfirmationList(confirmationList *list.Model, toolInfo *models.ToolConfirmationInfo) {
	var items []list.Item
	if toolInfo != nil {
		for _, option := range toolInfo.Options {
			items = append(items, ConfirmationItem{
				title:       option.Label,
				description: option.Description,
				action:      option.Action,
			})
		}
	}

	confirmationList.SetItems(items)
	// Set simple title without duplicating SQL
	if toolInfo != nil {
		confirmationList.Title = "Choose an action:"
	}
}
