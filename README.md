# Memento

A (heavily still-in-progress) terminal user interface (TUI) for [Memos](https://github.com/usememos/memos) - a lightweight, self-hosted memo hub.

Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Features

- View all your memos in a scrollable list
- Create new memos with multi-line support
- Edit existing memos
- View full memo details with scrolling
- Delete memos with confirmation
- **Vim-style keybindings** with normal/insert/visual modes
- **Fuzzy search** with `/` key
- **Pin, archive, and visibility controls**
- **9 built-in color themes** (Dracula, Nord, Solarized, Gruvbox, etc.)
- **YAML configuration** for customization
- **Undo/redo** in editor
- **System clipboard** integration
- **Mouse support** (click, scroll)
- Debug mode for troubleshooting

## Installation

### Prerequisites

- Go 1.21 or higher
- A running Memos instance with API access

### Build from source

```bash
git clone https://github.com/tlwhittaker/memento.git
cd memento
make build
```

### Install system-wide

```bash
sudo make install
```

This installs the binary to `/usr/local/bin` and the man page to `/usr/local/share/man/man1`.

## Configuration

### API Credentials

Create a `.env` file in `~/.config/memento/` (or set environment variables):

```bash
mkdir -p ~/.config/memento
cp .env.example ~/.config/memento/.env
```

Edit `.env` with your Memos instance details:

```
MEMOS_API_URL=https://your-memos-instance.com/api/v1
MEMOS_AUTH_TOKEN=your_access_token
```

#### Getting an Access Token

1. Log in to your Memos instance
2. Go to Settings > My Account > Access Tokens
3. Create a new access token
4. Copy the token to your `.env` file

### User Settings

Create a config file with:

```bash
memento --init-config
```

This creates `~/.config/memento/config.yaml`:

```yaml
theme: dracula

date_format: relative

sort_by: display_time
sort_order: desc

editor:
  mode: vim
  tab_width: 4
  word_wrap: true

keybindings:
  quit: "q"
  new: "n"
  edit: "e"
  delete: "d"
  search: "/"
  refresh: "r"

debug: false
```

## Usage

Run the TUI:

```bash
memento
```

With debug logging:

```bash
memento --debug
```

## Keybindings

### List Screen

| Key | Action |
|-----|--------|
| `j` / `Down` | Move down |
| `k` / `Up` | Move up |
| `Enter` | View memo details |
| `n` | Create new memo |
| `e` | Edit selected memo |
| `d` | Delete selected memo |
| `p` | Toggle pin status |
| `a` | Toggle archive status |
| `v` | Cycle visibility |
| `/` | Search (fuzzy matching) |
| `r` | Refresh memo list |
| `q` | Quit |

### Detail Screen

| Key | Action |
|-----|--------|
| `j` / `Down` | Scroll down |
| `k` / `Up` | Scroll up |
| `e` | Edit memo |
| `d` | Delete memo |
| `p` | Toggle pin |
| `a` | Toggle archive |
| `v` | Cycle visibility |
| `Esc` / `q` | Back to list |

### Editor (Vim Mode)

#### Normal Mode

| Key | Action |
|-----|--------|
| `i` | Enter insert mode |
| `a` | Append (insert after cursor) |
| `o` | Open line below |
| `O` | Open line above |
| `v` | Visual selection mode |
| `h/j/k/l` | Move cursor |
| `w` | Next word |
| `b` | Previous word |
| `e` | End of word |
| `0` | Line start |
| `$` | Line end |
| `gg` | Document start |
| `G` | Document end |
| `dd` | Delete line |
| `dw` | Delete word |
| `D` | Delete to end of line |
| `yy` | Yank line |
| `yw` | Yank word |
| `p` | Paste after |
| `P` | Paste before |
| `u` | Undo |
| `Ctrl+r` | Redo |
| `x` | Delete character |
| `Ctrl+s` | Save |
| `Esc` | Cancel/exit |

#### Insert Mode

| Key | Action |
|-----|--------|
| `Esc` | Return to normal mode |
| `Ctrl+v` | Paste from system clipboard |
| Arrow keys | Navigate cursor |
| `Enter` | New line |

#### Visual Mode

| Key | Action |
|-----|--------|
| Movement keys | Extend selection |
| `d` / `x` | Delete selection |
| `y` | Yank selection |
| `c` | Change selection |
| `Esc` | Exit visual mode |

### Dialogs

| Key | Action |
|-----|--------|
| `y` | Confirm (delete/discard) |
| `n` / `Esc` | Cancel |

### Mouse

| Action | Effect |
|--------|--------|
| Click | Select memo in list |
| Click again | Open selected memo |
| Scroll wheel | Navigate lists |
| Right-click | Go back (in detail view) |

## Themes

Built-in themes:
- `dracula` (default)
- `nord`
- `solarized-dark`
- `solarized-light`
- `gruvbox`
- `one-dark`
- `tokyo-night`
- `catppuccin-mocha`
- `monokai`
- `custom` (define your own colors)

Set theme in config.yaml or create a custom theme:

```yaml
theme: custom
colors:
  primary: "#7D56F4"
  secondary: "#00D9FF"
  error: "#FF5555"
  success: "#50FA7B"
  warning: "#FFB86C"
  muted: "#6272A4"
  text: "#F8F8F2"
  background: "#282A36"
  selected: "#44475A"
```

## Project Structure

```
memento/
├── cmd/memento/main.go      # Application entry point
├── internal/
│   ├── api/                 # Memos API client
│   │   ├── client.go        # HTTP client
│   │   ├── memos.go         # Memo endpoints (CRUD, pin, archive, visibility)
│   │   └── types.go         # API types
│   ├── config/              # Configuration
│   │   ├── config.go        # Env var loading
│   │   └── settings.go      # YAML config
│   ├── debug/               # Debug logging
│   │   └── debug.go
│   ├── models/              # Domain models
│   │   └── memo.go          # Memo model
│   └── ui/                  # TUI components
│       ├── model.go         # Root Bubbletea model
│       ├── update.go        # Input handling
│       ├── view.go          # Rendering
│       ├── styles.go        # Lipgloss styles
│       ├── themes.go        # Color themes
│       ├── vim.go           # Vim movement functions
│       ├── history.go       # Undo/redo system
│       └── clipboard.go     # System clipboard
├── docs/
│   └── memento.1            # Man page
├── .env.example             # Example configuration
├── Makefile                 # Build automation
└── README.md
```

## Make Targets

```bash
make build        # Build the binary
make install      # Install to /usr/local
make uninstall    # Remove installed files
make run          # Build and run
make run-debug    # Run with debug logging
make init-config  # Create default config file
make test         # Run tests
make lint         # Run linters
make dist         # Cross-compile for multiple platforms
make help         # Show all targets
```

## License

MIT License
