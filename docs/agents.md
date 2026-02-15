# Agents

AIOS manages skills across multiple AI agent clients. The agent registry detects which clients are installed and manages skill synchronization.

## Supported Agents

| Agent | Display Name | Universal | Skills Directory |
|-------|--------------|-----------|------------------|
| opencode | OpenCode | ✓ | .agents/skills |
| claude-code | Claude Code | | .claude/skills |
| cursor | Cursor | | .cursor/skills |
| codex | Codex | ✓ | .agents/skills |
| gemini-cli | Gemini CLI | ✓ | .agents/skills |
| github-copilot | GitHub Copilot | ✓ | .agents/skills |
| goose | Goose | | .goose/skills |
| windsurf | Windsurf | | .windsurf/skills |
| cline | Cline | | .cline/skills |

## Universal vs Non-Universal

- **Universal agents** share a common `.agents/skills` directory
- **Non-universal agents** have their own skills directory (e.g., `.cursor/skills`)

## Detection

AIOS automatically detects which agents are installed by checking their detect paths:

```bash
aios list-clients
```

Output shows only installed agents:

```json
{
  "opencode": {"installed": true, "path": ".agents/skills", "skills": ["ddd-expert"]},
  "claude-code": {"installed": true, "path": ".claude/skills", "skills": ["ddd-expert"]},
  "codex": {"installed": true, "path": ".agents/skills", "skills": ["ddd-expert"]}
}
```

## Agent Configuration

Agents are defined in `internal/agents/agents.json`:

```json
{
  "name": "opencode",
  "displayName": "OpenCode",
  "skillsDir": ".agents/skills",
  "altSkillsDirs": [".opencode/skills"],
  "globalSkillsDir": "$XDG_CONFIG/opencode/skills",
  "detectPaths": ["$XDG_CONFIG/opencode"],
  "universal": true
}
```

### Fields

- **name** - Internal identifier
- **displayName** - Human-readable name
- **skillsDir** - Project-local skills directory
- **altSkillsDirs** - Alternative skill directories
- **globalSkillsDir** - Global skills location (supports $VAR and ~)
- **detectPaths** - Paths to check for installation
- **universal** - Whether using shared .agents/skills

## Adding New Agents

1. Add entry to `internal/agents/agents.json`
2. Ensure detectPaths exist on your system
3. Run `aios list-clients` to verify detection
