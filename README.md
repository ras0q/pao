# pao

Context-aware async resources for [Bubble Tea](https://github.com/charmbracelet/bubbletea) v2.

## Install

```sh
go get github.com/ras0q/pao
```

## Usage

```go
type model struct {
	user pao.Resource[User]
}

func newModel(ctx context.Context) model {
	return model{
		user: pao.NewResource(ctx, fetchUser),
	}
}

func (m model) Init() tea.Cmd {
	return m.user.Ensure()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "r" {
			return m, m.user.Refresh()
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	return tea.NewView(m.user.Snapshot().String())
}
```

- `State[T]` holds data and fetch status; `Snapshot()` returns a copy for View.
- `LoadedMsg.State` carries the state when a fetch completes.
- `Ensure` fetches missing or invalidated data.
- `Refresh` fetches again while keeping existing data visible.
- `Invalidate` marks data stale and cancels an in-flight fetch.
- Canceling the resource context cancels its in-flight fetch.

See the [example](_examples/main.go) for a runnable program.
