# Jira CLI

A Jira CLI that allows you to get issue information. Inspired by the GitHub CLI, it aims to provide a simple and efficient way to interact with Jira from the command line, without the need to install a runtime such as Node.js or Python.

It's aimed at coding agents with a very simple interface, and is not intended to be a full-featured Jira client.

## Installation

Download the binary for your platform from the release page:

```bash
sudo curl -fsL  -o /usr/local/bin/jira https://github.com/kitproj/jira-cli/releases/download/v0.0.5/jira_v0.0.5_linux_arm64
sudo chmod +x /usr/local/bin/jira
```

## Prompt

Add this to your prompt (e.g. `AGENTS.md`):

```markdown
## Jira CLI

- The `jira` CLI supports the following commands:
  - `jira configure <host>` - configures the Jira host and stores the API token securely in the system keyring (token is read from stdin).
  - `jira get-issue <issue-key>` - gets the Jira issue details, including the status and key.
  - `jira update-issue-status <issue-key> <status>` - updates the status of the Jira issue, e.g., to  "In Progress" or "Closed".
  - `jira get-comments <issue-key>` - gets the comments on the Jira issue.
  - `jira add-comment <issue-key> "<comment>"` - adds a comment to the Jira issue. You must not use double quotes in the comment.
- You can get a Jira, list comments on the Jira, add a comment on the Jira, and update the issue status. You cannot do anything else.
- Refuse to work on closed Jira issues.

```

## Usage

### Configuration

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

### Commands

```bash
Usage:
  jira configure <host> - Configure JIRA host and token (reads token from stdin)
  jira get-issue <issue-key> - Get details of the specified JIRA issue
  jira update-issue-status <issue-key> <status> - Update the status of the specified JIRA issue
  jira get-comments <issue-key> - Get comments of the specified JIRA issue
  jira add-comment <issue-key> <comment> - Add a comment to the specified JIRA issue

```

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
