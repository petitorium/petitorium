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
- `D` (Shift+D) - Duplicate selected request
- `R` (Shift+R) - Rename selected collection/request
- `m` - Move selected collection/request
- `d` - Delete selected collection/request

### Request Panel

- `Tab` - Cycle through request elements (Method dropdown → URL input → Send button → Request tabs)
- `i` - Enter insert mode for body editing (when in body tab and view mode) or header value editing (when in headers tab)
- `Enter` (on Send button) - Send HTTP request

### Header Editing (Vim-style)

- `i` - Enter edit mode for header values (when focused on a header value in view mode)
- `Esc` - Exit edit mode and return to view mode (when editing header values)

### Body View Panel (Vim-style navigation)

- `h` - Scroll left
- `j` - Scroll down
- `k` - Scroll up
- `l` - Scroll right
- `g` - Go to top
- `G` - Go to bottom
- `w` - Page down (like Ctrl+F in vim)
- `b` - Page up (like Ctrl+B in vim)

### External Editor

- `i` - Open body in external editor for editing (main interface)
- `F4` - Open body in external editor (main interface or new request form)

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
- Request body editing with external editor integration
- Full syntax highlighting and editing capabilities in external editor
- Vim-style navigation for response viewing
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

- Press `i` to open body in external editor for editing (main interface)
- Press `F4` to open body in external editor (main interface or when creating new requests)
- The external editor provides full support for pasting, complex editing, and syntax highlighting

## Configuration

Configuration files are stored in `~/.config/petitorium/`:

- `config.yaml` - Application settings and themes
- `collections.yaml` - API collections and requests
- `expansion_state.yaml` - Collection expansion states
