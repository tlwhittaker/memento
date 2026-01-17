# Memento

<img width="1552" height="793" alt="image" src="https://github.com/user-attachments/assets/810e1f04-2f9e-4a5f-a06a-af8029b92b54" />

A terminal user interface (TUI) for [Memos](https://github.com/usememos/memos) - a lightweight, self-hosted memo hub.

Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Features

### Core
- View, create, edit, and delete memos
- Pin, archive, and visibility controls
- Fuzzy search with `/` key
- Calendar view with memo indicators
- Tags browser

### Editor
- **Vim-style keybindings** (default) with normal/insert/visual modes
- **Simple editor mode** for users unfamiliar with vim
- Undo/redo support
- System clipboard integration

### Navigation
- **Command palette** (`:`) with fuzzy search
- **Help overlay** (`?`) with context-aware keybindings
- Quick navigation: `gg` (top), `G` (bottom), `gt` (tags), `gc` (calendar)
- Mouse support (click, scroll)

### Search & Filtering
- Fuzzy text search
- Advanced filters: `#tag`, `pinned`, `archived`, `v:public`, `date:today`
- Saved filter shortcuts (1-9 keys)

### Display
- **Markdown rendering** toggle in detail view
- 11 built-in color themes including high-contrast options
- Relative or absolute date formats
- Compact, comfortable, or expanded view density

### Export
- Export memos to Markdown, JSON, or plain text
- Single memo or bulk export

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

## Usage

Run the TUI:

```bash
memento
```

## Keybindings

### Global
| Key | Action |
|-----|--------|
| `?` | Toggle help overlay |
| `:` | Open command palette |
| `q` | Quit / Go back |

### List View
| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down |
| `Enter` | View memo details |
| `n` | Create new memo |
| `e` | Edit memo |
| `d` | Delete memo |
| `p` | Toggle pin |
| `a` | Toggle archive |
| `v` | Cycle visibility |
| `/` | Search |
| `c` | Calendar view |
| `gg` | Jump to top |
| `G` | Jump to bottom |

### Editor (Vim mode)
| Key | Action |
|-----|--------|
| `i` | Insert mode |
| `Esc` | Normal mode |
| `:w` | Save |
| `:q` | Quit |
| `:wq` | Save and quit |
| `Ctrl+s` | Save |

### Editor (Simple mode)
| Key | Action |
|-----|--------|
| `Ctrl+s` | Save |
| `Esc` | Cancel |

Switch between editor modes with `:editor vim` or `:editor simple` in the command palette.

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
- `high-contrast-dark`
- `high-contrast-light`
- `custom` (define your own colors)

Set theme in config.yaml or switch via command palette (`:theme <name>`).

```yaml
theme: dracula

# Or define custom colors:
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

## Editor Mode

By default, Memento uses vim-style keybindings in the editor. If you prefer a simpler editing experience, you can switch to simple mode:

```yaml
editor:
  mode: simple  # or "vim" (default)
```

Or toggle at runtime via the command palette: `:editor simple` / `:editor vim`

## License

MIT License
