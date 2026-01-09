# Memento

<img width="1552" height="793" alt="image" src="https://github.com/user-attachments/assets/810e1f04-2f9e-4a5f-a06a-af8029b92b54" />


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

## License

MIT License
