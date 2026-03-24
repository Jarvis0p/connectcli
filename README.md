# ConnectCLI

A command-line interface tool for managing the Connecteam application.

## Setup

1. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```

2. **Build the CLI tool:**
   ```bash
   go build -o connectcli
   ```

3. **Set up credentials:**
   Create a `.connectcli/credentials` file in your home directory with the following format:
   ```
   session=<your_session_cookie>
   csrf=<your_csrf_token>
   ```

   Example:
   ```
   session=2|1:0|10:1751652781|7:session|48:ZWFjYzc5ZTEtMDVkMC00YmM1LTk5ZDMtYTAzN2Y3NmRjYWUz|c5b18a9f8a9f941af746a938848bf7b608295bbd0dbd259efc4d30a6e97181ad
   csrf=your_csrf_token_here
   jira=krishna@securify.llc:your_atlassian_token_here
   slack_webhook=https://hooks.slack.com/services/...
   slack_user_token=xoxp-your-user-token
   slack_bot_token=xoxb-your-bot-token
   ```

   - **`slack_webhook`** — incoming webhook for channel messages (punch in/out notifications).
   - **`slack_user_token`** or **`slack_bot_token`** (optional) — Bearer token for Slack **`users.profile.set`**. On punch-in, your status text is set to the **client name**; punch-out clears it. Use a token whose scopes allow profile updates (often **`users.profile:write`** on a user token; bot tokens may only update the bot’s profile depending on workspace settings). If both keys are present, the **last** one in the file wins.

4. **Configuration (auto-generated):**
   The tool will automatically create a `~/.connectcli/config` file with:
   ```
   punchclock.objectId=9229216
   ```

## Usage

### Validate Session
Check if your saved session is still valid:
```bash
./connectcli validate-session
```

This command will:
1. Load the session cookie and CSRF token from `~/.connectcli/credentials`
2. Make a request to the Connecteam API to validate the session
3. Display whether the session is valid or not

### Fetch Content Structure
Fetch the content structure and save the punch clock object ID:
```bash
./connectcli fetch-content
```

This command will:
1. Load the session cookie from `~/.connectcli/credentials`
2. Fetch the content structure from the Connecteam API
3. Extract the punch clock object ID (9229216 in your case)
4. Save it to `~/.connectcli/config` for future use

### Fetch Timesheet Data
Fetch timesheet data for a specific date or date range:
```bash
# Single date
./connectcli fetch timesheet 01/07/25

# Date range
./connectcli fetch timesheet 29/06/25-01/07/25

# With verbose output (full employee notes)
./connectcli fetch timesheet -v 01/07/25
./connectcli fetch timesheet 01/07/25 -v

### Fetch Jira Tickets
Fetch Jira tickets from the TECH project:
```bash
# Fetch first 500 tickets
./connectcli fetch jira

# Fetch next 100 tickets
./connectcli fetch jira --more
```

This command will:
1. Load the session cookie and CSRF token from `~/.connectcli/credentials`
2. Load the punch clock object ID from `~/.connectcli/config`
3. Parse the date(s) in dd/mm/yy format
4. Fetch timesheet data from the Connecteam API
5. Display the timesheet in a clean table format with:
   - Date (Mon 1/2 format)
   - Project/Type (e.g., Securify - Internal, BlueAlly)
   - Start and End times (12-hour format)
   - Duration (HH:MM format)
   - Employee notes (truncated if too long)
   - Summary statistics

This command will:
1. Load the Jira token from `~/.connectcli/credentials`
2. Fetch tickets from the TECH project via Jira API
3. Save tickets to `jira_tickets.json` in the project directory
4. Prevent duplicate entries based on ticket ID
5. Support pagination with `--more` flag for additional batches

## Commands

- `validate-session` - Validate the saved session credentials
- `fetch-content` - Fetch content structure and save punch clock object ID
- `fetch timesheet` - Fetch timesheet data for specific dates
- `fetch jira` - Fetch Jira tickets from TECH project

## Development

This tool is built using:
- Go 1.21+
- Cobra CLI framework
- Standard library HTTP client

The tool makes HTTP/2 requests to `app.connecteam.com` to validate sessions and will be extended to support data fetching and insertion operations. 