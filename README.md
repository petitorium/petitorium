# Petitorium

A powerful TUI (Terminal User Interface) for API interaction and testing.

## Keybindings

### Global Navigation

- `Tab` - Cycle through main panels (Collections → Request → Response)
- `q` / `Q` - Quit application

### Collections Panel

- `h` - Collapse collection / Move to parent collection
- `l` - Expand collection / Select request
- `n` - Create new collection
- `r` - Create new request
- `R` (Shift+R) - Rename selected collection/request
- `m` - Move selected collection/request
- `d` - Delete selected collection/request

### Request Panel

- `Tab` - Cycle through request elements (Method dropdown → URL input → Send button → Request tabs)
- `i` - Enter insert mode for body editing (when in body tab and view mode)
- `Enter` (on Send button) - Send HTTP request

### Body View Panel (Vim-style navigation)

- `h` - Scroll left
- `j` - Scroll down
- `k` - Scroll up
- `l` - Scroll right
- `g` - Go to top
- `G` - Go to bottom
- `w` - Page down (like Ctrl+F in vim)
- `b` - Page up (like Ctrl+B in vim)

### Body Edit Panel

- `Esc` - Exit insert mode back to view mode (saves changes automatically)
- `Ctrl+s` - Save changes without leaving edit mode
- `Ctrl+h` - Move cursor left
- `Ctrl+j` - Move cursor down
- `Ctrl+k` - Move cursor up
- `Ctrl+l` - Move cursor right

### External Editor

- `F4` - Open body in external editor

### Tab Navigation (Commented/Not Working)

- `1` - Switch to Body tab
- `2` - Switch to Auth tab
- `3` - Switch to Query tab
- `4` - Switch to Headers tab
- `Left Arrow` - Previous tab
- `Right Arrow` - Next tab

### Focus Management (Commented/Not Working)

- `Shift+Tab` - Previous element (Backtab)

## Features

- API request management with collections
- Multiple HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
- **Send HTTP requests** with headers and body
- Request body editing with syntax highlighting
- **Automatic save on exit from edit mode**
- **Save without leaving edit mode** (Ctrl+s)
- Vim-style navigation
- External editor integration
- Configurable themes and colors
- Persistent collection expansion state

## Installation

```bash
go build -o petitorium .
```

## Usage

```bash
./petitorium
```

### Sending HTTP Requests

1. Select a request from the collections panel
2. Choose HTTP method from dropdown
3. Enter URL in the input field
4. Add headers in the Headers tab
5. Write request body in the Body tab
6. Press `Enter` on the Send button or navigate to it and press `Enter`

### Body Editing Workflow

- Press `i` to enter insert mode for body editing
- Press `Esc` to exit insert mode and **automatically save changes**
- Press `Ctrl+s` to save changes **without leaving edit mode**
- Use `Ctrl+hjkl` for navigation while in edit mode

## Configuration

Configuration files are stored in `~/.config/petitorium/`:

- `config.yaml` - Application settings and themes
- `collections.yaml` - API collections and requests
- `expansion_state.yaml` - Collection expansion states

