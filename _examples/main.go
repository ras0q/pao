package main

import (
	"context"
	"math/rand/v2"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ras0q/pao"
)

func main() {
	m := newModel(context.Background())

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

type model struct {
	numberResource pao.Resource[int]
}

func newModel(ctx context.Context) model {
	return model{
		numberResource: pao.NewResource(ctx, func(ctx context.Context) (int, error) {
			time.Sleep(1 * time.Second)
			return rand.Int(), nil
		}),
	}
}

var _ tea.Model = model{}

func (m model) Init() tea.Cmd {
	return m.numberResource.Ensure()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pao.LoadedMsg[int]:
		if msg.Resource == m.numberResource && msg.State.Err != nil {
			return m, tea.Printf("fetch error: %v", msg.State.Err)
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "r":
			return m, m.numberResource.Refresh()
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	view := tea.NewView(
		m.numberResource.Snapshot().String() + "\n\nr: refresh  q: quit",
	)
	view.AltScreen = false

	return view
}
