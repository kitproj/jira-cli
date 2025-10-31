# Jira CLI & MCP Server

A Jira CLI and MCP server that allows you and your coding agents to get issue information from Jira. Inspired by the GitHub CLI, it aims to provide a simple and efficient way to humans and agents interact with Jira from the command line.

Being both a CLI and an MCP server means you get the best of both worlds. Agents can be directed to perform specific commands (e.g. `Put the Jira "In progress" by running jira update-issue-status ABC-123 "In Progress"`, or knowing they will do it correctly.

Like `jq`, it is a single tiny (10Mb) binary, without the overhead of installing a Node runtime, and without the need to put your Jira token in plain text file (it uses the system key-ring).

## Installation

### Supported Platforms

Binaries are available for:
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

### Download and Install

Download the binary for your platform from the [release page](https://github.com/kitproj/jira-cli/releases), e.g. for linux/arm64:

```bash
sudo curl -fsL  -o /usr/local/bin/jira https://github.com/kitproj/jira-cli/releases/download/v0.0.9/jira_v0.0.9_linux_arm64
sudo chmod +x /usr/local/bin/jira
```

For macOS (Apple Silicon):
```bash
sudo curl -fsL  -o /usr/local/bin/jira https://github.com/kitproj/jira-cli/releases/download/v0.0.9/jira_v0.0.9_darwin_arm64
sudo chmod +x /usr/local/bin/jira
```

Verify the installation:
```bash
jira -h
```

## Usage

### Configuration

#### Getting a Jira API Token

Before configuring, you'll need to create a Jira API token:

1. Visit your Jira instance: `https://your-domain.atlassian.net/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens`
2. Click "Create API token"
3. Give it a label (e.g., "jira-cli")
4. Copy the generated token (you won't be able to see it again)

#### Configure the CLI

The `jira` CLI can be configured in two ways:

1. **Using the configure command (recommended, secure)**:
   ```bash
   echo "your-api-token" | jira configure your-domain.atlassian.net
   ```
   This stores the host in `~/.config/jira-cli/config.json` and the token securely in your system's keyring.

2. **Using environment variables**:
   ```bash
   export JIRA_HOST=your-domain.atlassian.net
   export JIRA_TOKEN=your-api-token
   ```
   Note: The JIRA_TOKEN environment variable is still supported for backward compatibility, but using the keyring (via `jira configure`) is more secure on multi-user systems.

## Usage

### Direct CLI Usage

```bash
Usage:
  jira configure <host> - Configure JIRA host and token (reads token from stdin)
  jira create-issue <project> <description> [assignee] - Create a new JIRA issue
  jira get-issue <issue-key> - Get details of the specified JIRA issue
  jira list-issues - List issues assigned to the current user
  jira update-issue-status <issue-key> <status> - Update the status of the specified JIRA issue
  jira get-comments <issue-key> - Get comments of the specified JIRA issue
  jira add-comment <issue-key> <comment> - Add a comment to the specified JIRA issue
  jira mcp-server - Start MCP server (Model Context Protocol)
```

#### Examples

**Get issue details:**
```bash
jira get-issue PROJ-123
```

**List your current issues:**
```bash
jira list-issues
# Output:
# Found 3 issue(s) in the last 14 days:
# 
# PROJ-123        In Progress          Implement new feature
# PROJ-124        To Do                Fix critical bug
# PROJ-125        In Review            Update documentation
```

**Create a new issue:**
```bash
jira create-issue PROJ "Fix login bug"
# With assignee:
jira create-issue PROJ "Add dark mode" john.doe
```

**Update issue status:**
```bash
jira update-issue-status PROJ-123 "In Progress"
# Note: Status names must match your Jira workflow (e.g., "To Do", "In Progress", "Done")
```

**Add a comment:**
```bash
jira add-comment PROJ-123 "Working on this now"
```

**Get all comments:**
```bash
jira get-comments PROJ-123
```

### MCP Server Mode

The MCP (Model Context Protocol) server allows AI assistants and other tools to interact with JIRA through a standardized JSON-RPC protocol over stdio. This enables seamless integration with AI coding assistants and other automation tools.

Learn more about MCP: https://modelcontextprotocol.io

**Setup:**

1. First, configure your JIRA host and token (stored securely in the system keyring):
   ```bash
   echo "your-api-token" | jira configure your-domain.atlassian.net
   ```

2. Add the MCP server configuration to your MCP client (e.g., Claude Desktop, Cline):
   ```json
   {
     "mcpServers": {
       "jira": {
         "command": "jira",
         "args": ["mcp-server"]
       }
     }
   }
   ```

   For **Claude Desktop**, add this to:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

The server exposes the following tools:
- `get_issue` - Get details of a JIRA issue (e.g., status, summary, reporter, description)
- `update_issue_status` - Update the status of a JIRA issue using transitions
- `add_comment` - Add a comment to a JIRA issue
- `get_comments` - Get all comments on a JIRA issue
- `create_issue` - Create a new JIRA issue with specified project, description, and optional assignee
- `list_issues` - List issues assigned to the current user that are unresolved and updated in the last 14 days

**Example usage from an AI assistant:**
> "Get the details of issue PROJ-123 and add a comment saying the work is in progress."



## Git Hook Integration

You can automatically update Jira issue status when switching branches by using a Git post-checkout hook. This hook detects the Jira issue key in your branch name (e.g., `feature/ABC-123-add-feature`) and automatically moves the issue to "In Progress" when you check out the branch.

### What is this?

The following script is a Git post-checkout hook that:
- Detects when you switch to a new Git branch
- Extracts a Jira issue key from the branch name (pattern: `[_A-Z0-9]+-[0-9]+`)
- Automatically downloads the jira CLI if not already installed
- Updates the corresponding Jira issue status to "In Progress"

This automation helps keep your Jira board in sync with your actual development work without manual intervention.

### Setup

1. Create the post-checkout hook file:
   ```bash
   nano .git/hooks/post-checkout
   ```

2. Add the following script:
   ```bash
   #!/bin/bash
   set -euo pipefail

   old_branch=$(git name-rev --name-only "$1")
   new_branch=$(git branch --show-current)

   if [ "$old_branch" = "$new_branch" ]; then
     exit 0
   fi

   # grep the branch name for a jira ticket pattern (e.g. ABC-123)
   jira_issue_key=$(echo $new_branch | grep -oE '[_A-Z0-9]+-[0-9]+')

   if [ ! -n "$jira_issue_key" ]; then
       exit
   fi

   if [ ! -e ~/bin/jira ]; then
       platform=$(uname -s | tr '[:upper:]' '[:lower:]')
       arch=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
       curl -fsL -o ~/bin/jira https://github.com/kitproj/jira-cli/releases/download/v0.0.6/jira_v0.0.6_${platform}_${arch}
       mkdir -p ~/bin
       chmod +x ~/bin/jira
   fi

   ~/bin/jira -h jira.intuit.com update-issue-status $jira_issue_key "In Progress"
   ```

3. Make the hook executable:
   ```bash
   chmod +x .git/hooks/post-checkout
   ```

4. Ensure `~/bin` is in your PATH:
   ```bash
   export PATH="$HOME/bin:$PATH"
   ```

### How it works

- **Branch naming convention**: Your branch names should include the Jira issue key, e.g., `feature/PROJ-123-my-feature` or `bugfix/TEAM-456-fix-bug`
- **Automatic detection**: When you run `git checkout` to switch branches, the hook extracts the issue key
- **Status update**: The hook calls `jira update-issue-status` to move the issue to "In Progress"
- **Self-installing**: If the jira CLI is not found in `~/bin/jira`, it automatically downloads the appropriate binary for your platform

**Note**: Update the Jira host (`jira.intuit.com` in the example) to match your organization's Jira instance.

## Troubleshooting

### Common Issues

**"JIRA host must be configured" error**
- Make sure you've run `jira configure <host>` or set the `JIRA_HOST` environment variable
- Check that the config file exists: `cat ~/.config/jira-cli/config.json`

**"Failed to get issue" or authentication errors**
- Verify your API token is still valid (tokens can expire)
- Re-run the configure command to update the token: `echo "new-token" | jira configure your-domain.atlassian.net`
- Make sure your Jira user has permission to access the issue

**"No transition found to status" error**
- Status names must exactly match your Jira workflow (case-sensitive)
- Different issue types may have different workflows
- Use `jira get-issue <key>` to see the current status, then check your Jira board for valid transitions

**Keyring issues on Linux**
- Some Linux systems may not have a keyring service installed
- Install `gnome-keyring` or `kwallet` for your desktop environment
- Alternatively, use environment variables: `export JIRA_TOKEN=your-token`

**MCP server not appearing in Claude Desktop**
- Restart Claude Desktop after editing the config file
- Check the config file syntax is valid JSON
- Verify the `jira` binary is in your PATH: `which jira`

### Getting Help

- Report issues: https://github.com/kitproj/jira-cli/issues
- Check existing issues for solutions and workarounds
