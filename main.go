package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ethanholz/readwise-go"
	"github.com/joho/godotenv"
)

var (
	docStyle  = lipgloss.NewStyle().Margin(1, 2)
	bookMap   map[string]int
	bookItems []list.Item
	instance  *readwise.Instance
	prev      prevItems
)

type item struct {
	title, desc string
	list.Item
}

type prevItems struct {
	items    []list.Item
	title    string
	prevItem *prevItems
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			itemTest := m.list.SelectedItem()
			switch bookType := itemTest.(type) {
			case item:
				bookID := bookMap[bookType.Title()]
				highlights, err := instance.GetHighlightsForBook(bookID)
				if err != nil {
					fmt.Println(err)
				}
				tempPrev := prev
				prev.prevItem = &tempPrev
				prev.title = bookType.title
				prev.items = adaptHighlights(highlights)
				m.list.SetItems(prev.items)
				m.list.Title = prev.title
				m.list.ResetFilter()
			}
		case "b", "backspace":
			// if _, ok := bookMap[m.list.Title]; ok {
			// 	m.list.SetItems(bookItems)
			// 	m.list.Title = "Books"
			// }
			if prev.prevItem != nil {
				m.list.SetItems(prev.prevItem.items)
				m.list.Title = prev.prevItem.title
				prev = *prev.prevItem
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func adaptBookList(bookList *readwise.BookList) []list.Item {
	items := []list.Item{}
	books := bookList.Results
	for _, v := range books {
		tempItem := item{title: v.Title, desc: v.Author}
		items = append(items, tempItem)
	}
	return items
}

func adaptHighlights(highlightList *readwise.HighlightList) []list.Item {
	items := []list.Item{}
	highlights := highlightList.Results
	for _, v := range highlights {
		tempItem := item{title: v.Text, desc: v.HighlightedAt}
		items = append(items, tempItem)
	}
	return items
}

func generateBookMap(bookList *readwise.BookList) map[string]int {
	bookMap := make(map[string]int)
	books := bookList.Results
	for _, v := range books {
		bookMap[v.Title] = v.ID
	}
	return bookMap
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
	}
	instance = readwise.New()
	bookList, error := instance.GetBookList()
	if error != nil {
		log.Fatal("Failed to get books")
	}
	bookMap = generateBookMap(bookList)
	bookItems = adaptBookList(bookList)
	prev.items = bookItems
	prev.title = "Books"
	prev.prevItem = nil

	m := model{list: list.New(bookItems, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Books"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
