# Jira MCP

Give your AI agent full Jira access with just 3 tools.

Most Jira MCPs dump their entire API surface into the model's context. The agent wastes tokens picking between `jira_get_issue`, `jira_fetch_issue`, `jira_issue_get`... and still gets it wrong.

jira-mcp gives the model exactly what it needs:

| Tool | What it does |
|---|---|
| `jira_read` | Fetch issues by key, search by JQL, list projects/boards/sprints |
| `jira_write` | Create, update, delete, transition, comment — accepts Markdown, supports `dry_run` |
| `jira_schema` | Discover fields, transitions, and allowed values |

Three tools that compose naturally: schema to discover, read to find, write to change. Less surface area means fewer wrong picks, fewer redundant calls, more context for your actual work.

**Your credentials stay on your machine.** jira-mcp runs as a local process over stdio — no server, no proxy, nothing between your agent and Atlassian.

## Compared to [mcp-atlassian](https://github.com/sooperset/mcp-atlassian)

mcp-atlassian is a full Atlassian suite — 72 tools, Confluence, OAuth, SSE transport. That's powerful, and it's the right pick if you need all of it.

jira-mcp does one thing: Jira. And it does it with as little friction as possible.

**Why that matters for your agent:**
- **3 tools, not 72** — less context burned, sharper focus, fewer hallucinated tool calls
- **Zero runtime dependencies** — single Go binary, no Python, no venv, no pip
- **Works out of the box** — Basic Auth, stdio transport, ship it

Use mcp-atlassian if you need Confluence, OAuth, or SSE. Use jira-mcp if you want Jira to just work.

## Compared to [acli](https://developer.atlassian.com/cloud/acli/guides/introduction/)

acli is great for humans typing commands in a terminal. jira-mcp is built for AI agents — and that difference matters.

When you give an AI shell access, it can do anything: delete users, change org settings, trigger [Rovodev](https://www.atlassian.com/software/rovo/dev). jira-mcp limits the blast radius to Jira. Structured tool calls instead of shell execution means no injection risk, no accidental admin actions, and a model that stays in its lane.

**jira-mcp gives your AI agent:**
- **Safety by default** — no shell injection risk, Jira-only blast radius
- **Native Markdown** — write comments and descriptions in Markdown, it converts automatically
- **Built-in dry run** — preview every write before it happens
- **Lean context** — 3 tools vs. the full CLI surface; your agent stays focused
- **One-line read-only mode** — just instruct the model to use `jira_read` only, no extra tokens needed

Use acli when a human is at the keyboard or when you need Admin/Rovodev operations. Use jira-mcp when an AI agent is driving.

## Quick start

**Prerequisites:** Jira Server/Data Center 7.x (REST API v2) with Basic Auth enabled. Jira Cloud is not supported.

### 1. Prepare Jira credentials

Use the credentials accepted by your Jira deployment's [Basic authentication](https://developer.atlassian.com/server/jira/platform/basic-authentication/):

- `JIRA_URL`: your Jira base URL, including its context path if applicable.
- `JIRA_EMAIL`: your Jira **username**, despite the variable's legacy name.
- `JIRA_API_TOKEN`: the secret sent as the Basic Auth **password**, despite the variable's name. Use a password or a token only if your deployment accepts it through Basic Auth; Bearer PATs are not interchangeable.

### 2. Configure OpenCode with `.env`

Download the binary for your platform from the [releases page](https://github.com/mmatczuk/jira-mcp/releases) and install it, for example in `~/.local/bin/jira-mcp`.

jira-mcp loads `.env` from its **process working directory**, not from the directory containing the binary. It does not search parent directories or load `.env.local` automatically.

#### `.env` example

Create or update `.env` in a directory of your choice without overwriting existing credentials:

```dotenv
JIRA_URL=https://jira.example.com
JIRA_EMAIL=jira-username
JIRA_API_TOKEN='your-basic-auth-secret'
```

Keep the file private (for example, `chmod 600 /absolute/path/to/jira-config/.env`) and out of version control. This repository ignores `.env` and `.env.*`, except `.env.example`.

#### `opencode.jsonc` example

Create `opencode.jsonc` in your workspace, or merge this configuration into your existing project or global `~/.config/opencode/opencode.jsonc`, preserving other settings:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "jira": {
      "type": "local",
      "command": ["/absolute/path/to/.local/bin/jira-mcp"],
      "cwd": "/absolute/path/to/jira-config",
      "enabled": true
    }
  }
}
```

Replace both paths with real absolute paths; `cwd` must contain the `.env` shown above. The [OpenCode local MCP configuration](https://opencode.ai/docs/mcp-servers/#local) supports `cwd`; without it, the server runs from the workspace directory. jira-mcp reads the file itself: no shell wrapper, `source .env`, or additional jira-mcp flag is needed.

**Environment variables take precedence over `.env`, even when set to an empty value.** Do not add `JIRA_*` entries to `mcp.jira.environment` when the file should provide them. In particular, `{env:JIRA_API_TOKEN}` is resolved by OpenCode before jira-mcp reads `.env` and can override the file with an empty or stale value.

### 3. Verify OpenCode

Quit OpenCode completely, then restart without inherited Jira variables if `.env` is the intended source:

```bash
env -u JIRA_URL -u JIRA_EMAIL -u JIRA_API_TOKEN opencode
```

Check the connection from the same workspace:

```bash
opencode mcp list
```

Then ask OpenCode to list Jira projects using `jira_read` with `resource: "projects"`. Connected confirms MCP connectivity, not Jira authentication. Public projects may be available anonymously; to verify authenticated access, read a known non-public issue that your account can access.

If startup reports a missing variable, check `cwd`, file readability/syntax, and empty ENV overrides. For compatibility, errors loading the automatic `.env` are currently ignored; required empty variables still stop startup. A Jira 401 after connection is not proof that `.env` was skipped: check inherited ENV, the Jira URL, username, and whether the secret is accepted through Basic Auth. Never paste credentials into logs or bug reports.

<details>
<summary>Claude Code and other MCP clients (Homebrew, Docker, binary)</summary>

### Install and add to Claude Code

Pick one path:

**Homebrew**

```bash
brew tap mmatczuk/jira-mcp https://github.com/mmatczuk/jira-mcp
brew install jira-mcp
```

```bash
claude mcp add-json jira '{
  "command": "jira-mcp",
  "env": {
    "JIRA_URL": "https://jira.example.com",
    "JIRA_EMAIL": "jira-username",
    "JIRA_API_TOKEN": "your-basic-auth-secret"
  }
}'
```

**Docker**

No install needed. The `-e VAR` flags (without a value) forward each variable from `env` into the container:

```bash
claude mcp add-json jira '{
  "command": "docker",
  "args": [
    "run", "-i", "--rm",
    "-e", "JIRA_URL",
    "-e", "JIRA_EMAIL",
    "-e", "JIRA_API_TOKEN",
    "mmatczuk/jira-mcp"
  ],
  "env": {
    "JIRA_URL": "https://jira.example.com",
    "JIRA_EMAIL": "jira-username",
    "JIRA_API_TOKEN": "your-basic-auth-secret"
  }
}'
```

**Binary**

Download the binary for your platform from the [releases page](https://github.com/mmatczuk/jira-mcp/releases) and put it on your `PATH`, then:

```bash
claude mcp add-json jira '{
  "command": "jira-mcp",
  "env": {
    "JIRA_URL": "https://jira.example.com",
    "JIRA_EMAIL": "jira-username",
    "JIRA_API_TOKEN": "your-basic-auth-secret"
  }
}'
```

### Verify Claude Code

First, confirm Claude Code picked up the server:

```bash
claude mcp list
```

You should see:

```
Checking MCP server health...

jira: jira-mcp - ✓ Connected
```

Then open a Claude Code session and ask: *"List my Jira projects"*:

```
❯ List my Jira projects

⏺ jira - jira_read (MCP)(resource: "projects")
  ⎿  Found 3 project(s)

⏺ 3 projects:

  - ACME — Acme Corp
  - PLAT — Platform
  - OPS — Operations
```

To verify authenticated access, use the non-public issue check described above. If a read fails, check the server logs and verify your credentials.

### Other MCP clients

Use the same binary and env vars. The server speaks standard MCP over stdio.

</details>

## License

MIT
