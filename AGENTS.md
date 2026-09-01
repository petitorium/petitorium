# Petitorium - Agent Guidelines

This document provides guidelines for AI agents working on Petitorium, a Terminal API Client written in Go.

## Build, Lint & Test Commands

Petitorium uses standard Go tooling. Note: A Makefile is not present; use `go` commands directly.

```bash
# Build the application
go build -o petitorium .

# Install to $GOPATH/bin
go install .

# Run the application
go run main.go

# Format and Vet
go fmt ./...
go vet ./...

# Tests (Note: Add *_test.go files as needed)
go test ./...              # Run all tests
go test -v ./...           # Verbose mode
go test -v -run TestName ./path/to/pkg # Run a single test
go test -cover ./...       # Coverage report

# Dependency management
go mod tidy
```

## Project Structure

- `cmd/`: Core application logic, TUI setup (`ui.go`, `ui_setup.go`), and event handling.
- `workspace/`: Data models (`models.go`) and persistence logic (`manager.go`).
- `config/`: Configuration management using Viper.
- `plugins/`: Plugin system implementation.
- `main.go`: Entry point.

## Code Style Guidelines

### Imports

Group imports into three blocks separated by newlines:

1. Standard library
2. Third-party libraries (e.g., `github.com/rivo/tview`)
3. Internal project packages (`github.com/petitorium/petitorium/...`)

### Naming Conventions

- **Exported**: `PascalCase`
- **Unexported**: `camelCase`
- **Acronyms**: Keep consistent case (e.g., `URL`, `HTTP`, `JSON`)
- **Receivers**: Short (1-3 chars) representing the type (e.g., `func (cb *CustomButton)`)

### Error Handling

- Return `error` as the last return value.
- Use `fmt.Errorf` with `%w` for wrapping errors.
- Avoid ignoring errors (`_ = ...`); handle them or log them to the UI status.
- Use `panic` only for unrecoverable initialization failures in `root.go`.

### Struct Tags and Persistence

- Use `yaml` tags for all fields in `workspace/models.go` as data is persisted in YAML.
- Use `omitempty` for optional collections or strings to keep files clean.

## UI & UX Guidelines (tview/tcell)

### Consistent Styling

- **ColorManager**: Use `ColorManager` (from `cmd/colors.go`) for all component colors.
- **Helper Functions**: Prefer using helper functions in `cmd/ui.go` (e.g., `createPanel`, `createInputField`, `createThemedButton`) to maintain consistency.

### Focus Management

- **Focus Cycles**: New components must be added to the appropriate focus cycle in `cmd/cycles.go`.
- **Tab/Shift+Tab**: Ensure `Tab` cycles forward and `Shift+Tab` cycles backward through all interactive elements.
- **Focus Restoration**: When opening modals (especially save/download dialogs or tag/command-runner modals), capture the current focus with `app.GetFocus()` before opening the modal. Use `showSuccessModalWithFocus`/`showErrorModalWithFocus` to restore focus when the modal closes, or call `app.SetFocus(previousFocus)` directly in the modal's `closeModalFunc`. Failing to restore focus causes UI corruption (e.g., j/k keys move the entire app inside the terminal).

  Example pattern for tag/command-runner modals:
  ```go
  previousFocus := app.GetFocus()
  closeModalFunc := func() {
      pages.RemovePage("modalName")
      pages.SwitchToPage("main")
      if previousFocus != nil {
          app.SetFocus(previousFocus)
      }
  }
  ```

### Vim Bindings

- Support Vim-style navigation (`h/j/k/l`, `g/G`, `d/u`) for read-only panels like response preview or collections tree.

### Performance

- **QueueUpdateDraw**: Use `app.QueueUpdateDraw` for any UI updates triggered from background goroutines (like HTTP requests) to avoid race conditions and ensure thread safety.
- **Never use app.Draw()**: Calling `app.Draw()` directly causes the application to hang (deadlock in the tview event/draw loop). Always rely on the normal draw cycle or `QueueUpdateDraw`.

## Common Patterns

### Creating a New Modal

Use `createModal` from `cmd/ui.go` to wrap forms or text views. Modals should be added to the `Pages` primitive.

### Standard Modal Sizes

All modals must use one of the standard sizes defined in `cmd/modals.go` via `createSizedModal(p tview.Primitive, size modalSize, backgroundColor tcell.Color)`:

| Size | Dimensions | Used for |
| --- | --- | --- |
| `modalSizeConfirm` | 50x8 | yes/no confirmations, deletions, short notices, progress |
| `modalSizeForm` | 50x10 | single-field forms (rename, duplicate, create item, row editors) |
| `modalSizeEditor` | 60x15 | multi-field forms (new request, duplicate request, move item) |
| `modalSizeSearch` | 100x20 | search overlays with a results list |
| `modalSizeLarge` | 80x25 | complex editors (tag editor, command runner, file pickers) |
| `modalSizeFullscreen` | 120x40 | split-panel modals (environments, workspaces, marketplace) |

Only fall back to `createModal` with explicit dimensions when the height must adapt to the content (e.g. a list sized to its number of entries, as in the tag picker and plugin version picker).

### External Editor Integration

Refer to `cmd/event_handlers.go` for how external editors are launched. It typically involves stopping the TUI, running the editor, and restarting the TUI.

## Key Dependencies

- `github.com/rivo/tview`: Primary UI framework.
- `github.com/gdamore/tcell/v2`: Terminal handling.
- `github.com/spf13/cobra`: CLI commands.
- `github.com/spf13/viper`: Configuration.

## AI Agent Workflow

1. **Explore**: Use `grep` and `read` to understand how components are linked (e.g., `UI` struct in `cmd/ui_orchestrator.go`).
2. **Implement**: Match the pattern of existing helper functions.
3. **Verify**: Always run `go build .` to check for compilation errors before submitting.
4. **Style**: Strictly follow `go fmt` output.

## Debugging

For debugging output that should not appear in the UI, write to `/tmp/petitorium_debug.log` using a `logToFile()` helper function rather than `fmt.Printf()`. This avoids corrupting the terminal display.

Example pattern:
```go
func logToFile(msg string) {
    f, err := os.OpenFile("/tmp/petitorium_debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err == nil {
        f.WriteString(msg)
        f.Close()
    }
}
```

When adding debug statements, always use `logToFile()` instead of `fmt.Printf()` to avoid interfering with the TUI rendering.

---

_Note: This file is intended for agentic coding agents. Update it as the project conventions evolve._
