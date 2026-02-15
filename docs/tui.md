# TUI - Terminal User Interface

AIOS includes an interactive Terminal User Interface (TUI) built with Bubble Tea and Lip Gloss.

## Launching the TUI

```bash
aios tui
```

The TUI automatically detects whether it's running in a terminal and falls back to a text-based script mode if not.

## Features

### Main Menu

1. **Projects** - List tracked projects
2. **Workspace Validate** - Check workspace symlink health
3. **Workspace Repair** - Fix broken symlinks
4. **Skills** - Manage skills
5. **Quit** - Exit the TUI

### Skills Menu

1. **Init skill** - Create a new skill scaffold
2. **Sync skill** - Sync skill to agents
3. **Back** - Return to main menu

## Navigation

- **Arrow keys** (↑/↓) or **k/j** - Navigate menu
- **Enter** - Select menu item
- **1-9** - Quick select by number
- **q** or **ctrl+c** - Quit

## Visual Styling

The TUI uses Lip Gloss for styling:

- **Headers** - Bold cyan text
- **Selected items** - Highlighted with bold styling
- **Success messages** - Green text
- **Error messages** - Red text
- **Input fields** - Cyan text

## Non-Interactive Mode

When not running in a terminal (e.g., in CI/CD), the TUI falls back to a text-based script mode that works identically but without the visual styling.

```bash
# This automatically uses script mode when piped
aios tui < input.txt
```

## Architecture

The TUI is implemented in `internal/core/tui.go` using:

- **Bubble Tea** - Elm-inspired TUI framework
- **Lip Gloss** - Style definition library
- **Automatic terminal detection** - Falls back to text mode when needed

## Use Cases

The TUI is ideal for:

- Interactive skill management
- Quick workspace diagnostics
- Learning the CLI without reading docs
- Environments without full CLI familiarity
