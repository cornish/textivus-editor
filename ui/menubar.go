package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MenuAction represents an action triggered by a menu item
type MenuAction int

const (
	ActionNone MenuAction = iota
	// File menu
	ActionNew
	ActionOpen
	ActionSave
	ActionSaveAs
	ActionRevert
	ActionExit
	// Edit menu
	ActionUndo
	ActionRedo
	ActionCut
	ActionCopy
	ActionPaste
	ActionSelectAll
	ActionFind
	// View menu
	ActionWordWrap
	ActionLineNumbers
	// Help menu
	ActionAbout
)

// MenuItem represents a single menu option
type MenuItem struct {
	Label    string
	Shortcut string
	Action   MenuAction
	Disabled bool
}

// Menu represents a dropdown menu
type Menu struct {
	Label string
	Items []MenuItem
}

// MenuBar represents the top menu bar
type MenuBar struct {
	menus      []Menu
	activeMenu int  // -1 if no menu is open
	activeItem int  // Index of highlighted item in open menu
	isOpen     bool // Whether a dropdown is open
	width      int
	styles     Styles
}

// NewMenuBar creates a new menu bar with default menus
func NewMenuBar(styles Styles) *MenuBar {
	return &MenuBar{
		menus: []Menu{
			{
				Label: "File",
				Items: []MenuItem{
					{Label: "New", Shortcut: "Ctrl+N", Action: ActionNew},
					{Label: "Open", Shortcut: "Ctrl+O", Action: ActionOpen},
					{Label: "Save", Shortcut: "Ctrl+S", Action: ActionSave},
					{Label: "Revert", Shortcut: "", Action: ActionRevert},
					{Label: "Exit", Shortcut: "Ctrl+Q", Action: ActionExit},
				},
			},
			{
				Label: "Edit",
				Items: []MenuItem{
					{Label: "Undo", Shortcut: "Ctrl+Z", Action: ActionUndo},
					{Label: "Redo", Shortcut: "Ctrl+Y", Action: ActionRedo},
					{Label: "Cut", Shortcut: "Ctrl+X", Action: ActionCut},
					{Label: "Copy", Shortcut: "Ctrl+C", Action: ActionCopy},
					{Label: "Paste", Shortcut: "Ctrl+V", Action: ActionPaste},
					{Label: "Select All", Shortcut: "Ctrl+A", Action: ActionSelectAll},
					{Label: "Find", Shortcut: "Ctrl+F", Action: ActionFind},
				},
			},
			{
				Label: "View",
				Items: []MenuItem{
					{Label: "[ ] Word Wrap", Shortcut: "", Action: ActionWordWrap},
					{Label: "[ ] Line Numbers", Shortcut: "", Action: ActionLineNumbers},
				},
			},
			{
				Label: "Help",
				Items: []MenuItem{
					{Label: "About", Shortcut: "F1", Action: ActionAbout},
				},
			},
		},
		activeMenu: -1,
		activeItem: 0,
		isOpen:     false,
		styles:     styles,
	}
}

// SetWidth sets the width of the menu bar
func (m *MenuBar) SetWidth(width int) {
	m.width = width
}

// IsOpen returns true if a menu dropdown is open
func (m *MenuBar) IsOpen() bool {
	return m.isOpen
}

// OpenMenu opens the menu at the given index
func (m *MenuBar) OpenMenu(index int) {
	if index >= 0 && index < len(m.menus) {
		m.activeMenu = index
		m.activeItem = 0
		m.isOpen = true
	}
}

// Close closes any open menu
func (m *MenuBar) Close() {
	m.isOpen = false
	m.activeMenu = -1
	m.activeItem = 0
}

// Toggle toggles the menu at the given index
func (m *MenuBar) Toggle(index int) {
	if m.isOpen && m.activeMenu == index {
		m.Close()
	} else {
		m.OpenMenu(index)
	}
}

// NextMenu moves to the next menu
func (m *MenuBar) NextMenu() {
	if len(m.menus) == 0 {
		return
	}
	m.activeMenu = (m.activeMenu + 1) % len(m.menus)
	m.activeItem = 0
}

// PrevMenu moves to the previous menu
func (m *MenuBar) PrevMenu() {
	if len(m.menus) == 0 {
		return
	}
	m.activeMenu--
	if m.activeMenu < 0 {
		m.activeMenu = len(m.menus) - 1
	}
	m.activeItem = 0
}

// NextItem moves to the next item in the current menu
func (m *MenuBar) NextItem() {
	if !m.isOpen || m.activeMenu < 0 || m.activeMenu >= len(m.menus) {
		return
	}
	items := m.menus[m.activeMenu].Items
	if len(items) == 0 {
		return
	}
	m.activeItem = (m.activeItem + 1) % len(items)
}

// PrevItem moves to the previous item in the current menu
func (m *MenuBar) PrevItem() {
	if !m.isOpen || m.activeMenu < 0 || m.activeMenu >= len(m.menus) {
		return
	}
	items := m.menus[m.activeMenu].Items
	if len(items) == 0 {
		return
	}
	m.activeItem--
	if m.activeItem < 0 {
		m.activeItem = len(items) - 1
	}
}

// Select returns the action of the currently selected item and closes the menu
func (m *MenuBar) Select() MenuAction {
	if !m.isOpen || m.activeMenu < 0 || m.activeMenu >= len(m.menus) {
		return ActionNone
	}
	items := m.menus[m.activeMenu].Items
	if m.activeItem < 0 || m.activeItem >= len(items) {
		return ActionNone
	}
	// Don't select disabled items
	if items[m.activeItem].Disabled {
		return ActionNone
	}
	action := items[m.activeItem].Action
	m.Close()
	return action
}

// SetItemDisabled sets the disabled state of a menu item by action
func (m *MenuBar) SetItemDisabled(action MenuAction, disabled bool) {
	for i := range m.menus {
		for j := range m.menus[i].Items {
			if m.menus[i].Items[j].Action == action {
				m.menus[i].Items[j].Disabled = disabled
				return
			}
		}
	}
}

// SetItemLabel sets the label of a menu item by action
func (m *MenuBar) SetItemLabel(action MenuAction, label string) {
	for i := range m.menus {
		for j := range m.menus[i].Items {
			if m.menus[i].Items[j].Action == action {
				m.menus[i].Items[j].Label = label
				return
			}
		}
	}
}

// menuItemWidth returns the display width of a menu item
func (m *MenuBar) menuItemWidth(index int) int {
	if index < 0 || index >= len(m.menus) {
		return 0
	}
	return len(m.menus[index].Label) + 4 // "  " + Label + "  "
}

// HandleClick handles a click at the given x position in the menu bar
// Returns true if the click was handled
func (m *MenuBar) HandleClick(x, y int) (bool, MenuAction) {
	if y == 0 {
		// Click on menu bar - find which menu
		pos := 0
		for i := range m.menus {
			w := m.menuItemWidth(i)
			if x >= pos && x < pos+w {
				if m.isOpen && m.activeMenu == i {
					m.Close()
				} else {
					m.OpenMenu(i)
				}
				return true, ActionNone
			}
			pos += w
		}
		// Clicked elsewhere on menu bar - close menu
		if m.isOpen {
			m.Close()
		}
		return true, ActionNone
	}

	// Click on dropdown
	if m.isOpen && y > 0 {
		// Calculate which item was clicked (accounting for border)
		itemIndex := y - 2 // -1 for menu bar, -1 for top border
		if m.activeMenu >= 0 && m.activeMenu < len(m.menus) {
			items := m.menus[m.activeMenu].Items
			if itemIndex >= 0 && itemIndex < len(items) {
				m.activeItem = itemIndex
				return true, m.Select()
			}
		}
	}

	return false, ActionNone
}

// DropdownHeight returns just the dropdown height (excluding the menu bar)
func (m *MenuBar) DropdownHeight() int {
	if !m.isOpen || m.activeMenu < 0 || m.activeMenu >= len(m.menus) {
		return 0
	}
	return len(m.menus[m.activeMenu].Items) + 2 // items + borders
}

// Height returns the total height (menu bar + dropdown if open)
func (m *MenuBar) Height() int {
	return 1 + m.DropdownHeight()
}

// View renders the menu bar (just the bar, not the dropdown)
func (m *MenuBar) View() string {
	// Calculate total width of menu items
	currentWidth := 0
	for i := range m.menus {
		currentWidth += m.menuItemWidth(i)
	}

	// Calculate padding needed
	paddingWidth := 0
	if currentWidth < m.width {
		paddingWidth = m.width - currentWidth
	}

	// Build the menu bar with a continuous dark blue background
	// and switching to cyan for the active item (classic DOS EDIT style)
	var sb strings.Builder

	// Start with dark blue background, white text
	sb.WriteString("\033[44;97m") // Dark blue bg, bright white text

	for i, menu := range m.menus {
		itemText := "  " + menu.Label + "  "
		if m.isOpen && i == m.activeMenu {
			// Switch to cyan background for active item
			sb.WriteString("\033[46;30;1m") // Cyan bg, black bold
			sb.WriteString(itemText)
			sb.WriteString("\033[44;97m") // Back to dark blue
		} else {
			sb.WriteString(itemText)
		}
	}

	// Add padding to fill width
	if paddingWidth > 0 {
		sb.WriteString(strings.Repeat(" ", paddingWidth))
	}

	// Reset at end of line
	sb.WriteString("\033[0m")

	return sb.String()
}

// RenderDropdown renders the dropdown menu as separate lines for overlay
// Returns the lines and the horizontal offset where the dropdown starts
func (m *MenuBar) RenderDropdown() ([]string, int) {
	if !m.isOpen || m.activeMenu < 0 || m.activeMenu >= len(m.menus) {
		return nil, 0
	}

	// Calculate offset
	offset := 0
	for i := 0; i < m.activeMenu; i++ {
		offset += m.menuItemWidth(i)
	}

	dropdown := m.renderDropdownContent() // Get dropdown without offset padding
	return strings.Split(dropdown, "\n"), offset
}

// renderDropdownContent renders just the dropdown box without horizontal offset
func (m *MenuBar) renderDropdownContent() string {
	if m.activeMenu < 0 || m.activeMenu >= len(m.menus) {
		return ""
	}

	menu := m.menus[m.activeMenu]

	// Find max width for items
	maxWidth := 0
	for _, item := range menu.Items {
		w := len(item.Label)
		if item.Shortcut != "" {
			w += 2 + len(item.Shortcut)
		}
		if w > maxWidth {
			maxWidth = w
		}
	}

	// Render items
	var items []string
	for i, item := range menu.Items {
		var style lipgloss.Style
		if item.Disabled {
			style = m.styles.MenuOptionDisabled
		} else if i == m.activeItem {
			style = m.styles.MenuOptionActive
		} else {
			style = m.styles.MenuOption
		}

		// Format: "Label    Shortcut"
		line := item.Label
		if item.Shortcut != "" {
			spaces := maxWidth - len(item.Label) - len(item.Shortcut)
			if spaces < 2 {
				spaces = 2
			}
			line += strings.Repeat(" ", spaces) + item.Shortcut
		} else {
			line += strings.Repeat(" ", maxWidth-len(item.Label))
		}

		items = append(items, style.Render(line))
	}

	// Join items vertically
	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Apply dropdown style (with border)
	return m.styles.MenuDropdown.Render(content)
}

// renderDropdown renders the dropdown menu
func (m *MenuBar) renderDropdown() string {
	if m.activeMenu < 0 || m.activeMenu >= len(m.menus) {
		return ""
	}

	menu := m.menus[m.activeMenu]

	// Calculate dropdown position (horizontal offset)
	offset := 0
	for i := 0; i < m.activeMenu; i++ {
		offset += m.menuItemWidth(i)
	}

	// Find max width for items
	maxWidth := 0
	for _, item := range menu.Items {
		w := len(item.Label)
		if item.Shortcut != "" {
			w += 2 + len(item.Shortcut)
		}
		if w > maxWidth {
			maxWidth = w
		}
	}

	// Render items
	var items []string
	for i, item := range menu.Items {
		var style lipgloss.Style
		if i == m.activeItem {
			style = m.styles.MenuOptionActive
		} else {
			style = m.styles.MenuOption
		}

		// Format: "Label    Shortcut"
		line := item.Label
		if item.Shortcut != "" {
			spaces := maxWidth - len(item.Label) - len(item.Shortcut)
			if spaces < 2 {
				spaces = 2
			}
			line += strings.Repeat(" ", spaces) + item.Shortcut
		} else {
			line += strings.Repeat(" ", maxWidth-len(item.Label))
		}

		items = append(items, style.Render(line))
	}

	// Join items vertically
	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Apply dropdown style (with border)
	dropdown := m.styles.MenuDropdown.Render(content)

	// Add horizontal offset
	if offset > 0 {
		lines := strings.Split(dropdown, "\n")
		padding := strings.Repeat(" ", offset)
		for i, line := range lines {
			lines[i] = padding + line
		}
		dropdown = strings.Join(lines, "\n")
	}

	return dropdown
}
