# Project: jira-mcp

MCP server giving AI agents full Jira access via 3 tools (jira_read, jira_write, jira_schema). Runs locally over stdio, uses API token auth, no external dependencies.

Tech stack: Go 1.25, go-jira, modelcontextprotocol/go-sdk, goldmark (Markdown->ADF)

## Branch
Default branch: main

## Workspace Overview
```
cmd/jira-mcp/     - entry point, config from env vars
internal/jira/    - Jira client wrapper with retry logic
internal/jiramcp/ - MCP server + 3 tool handlers
internal/mdconv/  - Markdown to Atlassian Document Format (ADF) converter
```

## Where To Look
- `cmd/jira-mcp/main.go` - startup, env vars (JIRA_URL, JIRA_EMAIL, JIRA_API_TOKEN)
- `internal/jira/client.go` - all Jira API calls with exponential backoff on 429
- `internal/jiramcp/server.go` - tool registration, instructions text
- `internal/jiramcp/tool_read.go` - jira_read handler (keys/JQL/resource modes)
- `internal/jiramcp/tool_write.go` - jira_write handler (create/update/delete/transition/comment)
- `internal/jiramcp/tool_schema.go` - jira_schema handler (fields/transitions/field_options)
- `internal/mdconv/mdconv.go` - Markdown AST to ADF conversion

## Architectural Invariants
- All Jira calls go through `internal/jira/client.go` with retry on rate limits
- Markdown in descriptions/comments is auto-converted to ADF via `mdconv.ToADF()`
- Write operations always have dry_run support
- Single binary, no runtime deps — stdio transport only
- go-jira types re-exported in `internal/jira/types.go` to hide dependency

## Key Subsystems

### Jira Client (internal/jira/client.go)
- Wraps go-jira with call-level retry (429 with Retry-After header, 502, 503)
- Exponential backoff starting at 1s, max 3 retries
- Methods: GetIssue, SearchIssues, DeleteIssue, GetTransitions, DoTransition, AddComment, UpdateComment, GetAllBoards, GetAllSprints, GetSprintIssues, MoveIssuesToSprint, GetAllProjects, GetFields, CreateIssueV3, UpdateIssueV3, GetFieldOptions

### MCP Server (internal/jiramcp/server.go)
- 3 tools registered: readTool, writeTool, schemaTool
- Instructions string tells model the workflow: schema → read → write
- Handlers live in tool_*.go files

### jira_read Tool
- Modes (mutually exclusive): keys (1+ issue keys), jql (search), resource (projects/boards/sprints/sprint_issues)
- Single key → GetIssue (supports expand); 2+ keys → JQL search
- Fields, expand, limit, pagination (start_at, next_page_token)

### jira_write Tool
- Actions: create, update, delete, transition, comment, edit_comment, move_to_sprint
- move_to_sprint batches by sprint_id
- buildIssuePayload() merges WriteItem fields + fields_json escape hatch
- Markdown descriptions/comments converted to ADF

### jira_schema Tool
- resources: fields (all fields), transitions (needs issue_key), field_options (needs field_id)
- Used to discover valid IDs before writing

### Markdown Converter (internal/mdconv/)
- goldmark-based Markdown → ADF document conversion
- Handles: paragraphs, headings, lists, code blocks, blockquotes, inline code/emphasis/links

## Development Practices
- Build: `go build -o bin/jira-mcp ./cmd/jira-mcp`
- Format/lint: `golangci-lint fmt ./... && go mod tidy`
- Lint: `golangci-lint run ./...`
- Test: `go test ./...`
- Taskfile: `task fmt|lint|test|build`
- Release: goreleaser (cross-platform binaries + Docker + Homebrew cask)
