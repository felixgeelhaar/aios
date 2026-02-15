# Connectors

Connectors enable AIOS to integrate with external services like Google Drive, GitHub, Slack, etc.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        AIOS CLI/TUI                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│                  (Onboarding Service)                       │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
    ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
    │   Token     │  │   Tray      │  │  Connector  │
    │   Store     │  │   State     │  │   Port      │
    └─────────────┘  └─────────────┘  └─────────────┘
                                              │
                                              ▼
                              ┌───────────────────────────────┐
                              │      Connector Adapters        │
                              │  ┌─────────────────────────┐   │
                              │  │  GoogleDriveAdapter    │   │
                              │  │  GitHubAdapter         │   │
                              │  │  SlackAdapter         │   │
                              │  └─────────────────────────┘   │
                              └───────────────────────────────┘
                                              │
                                              ▼
                              ┌───────────────────────────────┐
                              │     External Services         │
                              │  • Google Drive API           │
                              │  • GitHub API                 │
                              │  • Slack API                  │
                              └───────────────────────────────┘
```

## Current Implementation

### Google Drive Connector

**Current State:**
- Stores OAuth tokens in Keychain
- Simple token-based authentication
- No actual Drive API integration

**Missing:**
- OAuth 2.0 flow with proper scopes
- Drive file operations (list, read, write)
- Sync between local skills and Drive

### Token Store

AIOS uses Keychain on macOS, Credential Manager on Windows, or libsecret on Linux for secure token storage.

```go
// Token store interface
type TokenStore interface {
    Put(ctx context.Context, key, token string) error
    Get(ctx context.Context, key string) (string, error)
    Delete(ctx context.Context, key string) error
}
```

## Proposed Connector Interface

```go
// ConnectorPort defines the interface for all connectors
type ConnectorPort interface {
    // Connect establishes connection using credentials
    Connect(ctx context.Context, credentials map[string]string) error
    
    // Disconnect removes credentials
    Disconnect(ctx context.Context) error
    
    // IsConnected checks if connector is active
    IsConnected(ctx context.Context) (bool, error)
    
    // ListFiles lists files in the connector's storage
    ListFiles(ctx context.Context, path string) ([]FileInfo, error)
    
    // ReadFile reads a file from connector
    ReadFile(ctx context.Context, path string) ([]byte, error)
    
    // WriteFile writes a file to connector
    WriteFile(ctx context.Context, path string, data []byte) error
}
```

## Google Drive Connector (Proposed)

### Features

1. **OAuth 2.0 Authentication**
   - Local OAuth callback server
   - Secure token refresh
   - Scopes: `drive.readonly`, `drive.file`

2. **File Operations**
   - List skills in Drive
   - Download skill packages
   - Upload local skills
   - Sync detection (local vs remote)

3. **Skill Sync**
   - Push local skills to Drive
   - Pull skills from Drive
   - Conflict resolution

### Configuration

```yaml
connectors:
  google_drive:
    enabled: true
    scopes:
      - https://www.googleapis.com/auth/drive.readonly
      - https://www.googleapis.com/auth/drive.file
    folder_id: ""  # Optional: specific folder for skills
```

### Usage

```bash
# Connect Google Drive
aios connect google-drive

# List Drive skills
aios connectors google-drive list

# Sync skills to Drive
aios connectors google-drive sync --push

# Sync skills from Drive
aios connectors google-drive sync --pull
```

## Adding New Connectors

1. **Define Domain**: Create `domain/connectors/<name>/`
2. **Implement Port**: Create `internal/connectors/<name>/adapter.go`
3. **Register**: Add to onboarding service
4. **CLI Commands**: Add connect/list/sync commands

## Security

- All tokens stored in OS keychain
- OAuth tokens never logged
- Token refresh handled automatically
- Scope-limited access tokens
