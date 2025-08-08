# Issueify

[![Go Version](https://img.shields.io/badge/go-1.18+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/v/release/YOUR_USERNAME/YOUR_REPO)](https://github.com/YOUR_USERNAME/YOUR_REPO/releases)

A simple, fast, and powerful CLI tool for managing local development issues and publishing them to GitHub. Never lose track of a `TODO` or a bug fix again.

---

## Features

- **Local-First:** Keep a lightweight, local database of your issues without cluttering up your project's official issue tracker.
- **Automatic Labeling:** Automatically tags issues with labels like `bug`, `feature`, or `docs` based on keywords in the title.
- **GitHub Integration:** Publish your local issues to a GitHub repository with a single command.
- **`gh` CLI Support:** Seamlessly uses your existing GitHub CLI (`gh`) authentication and repository context.
- **Flexible Export:** Publish your issue list to Markdown or JSON for documentation and scripting.

---

## Installation

You can install `issueify` by building it from the source. This ensures you have the latest version and makes the command globally available on your system.

**Prerequisites:**

- Go (version 1.18 or newer)
- Git

**Instructions:**

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/andrinoff/issueify.git
    cd issueify
    ```

2.  **Build the binary:**

    ```bash
    go build -o issueify .
    ```

3.  **Move the binary to your bin directory:**
    To make `issueify` accessible from any directory, move the compiled executable to a location in your system's `PATH`. A common choice is `/usr/local/bin` on macOS/Linux.

    ```bash
    # For macOS and Linux
    mv issueify /usr/local/bin/

    # For Windows, move issueify.exe to a folder in your PATH
    ```

4.  **Verify the installation:**
    You should now be able to run the `issueify` command from anywhere.
    ```bash
    issueify help
    ```

---

## Usage

Here are the primary commands for `issueify`.

| Command          | Description                                                                             |
| ---------------- | --------------------------------------------------------------------------------------- |
| `add "<title>"`  | Adds a new issue. Prefix the title with `FEAT:`, `BUG:`, etc., for auto-labeling.       |
| `list`           | Lists all open issues. Use `--label=<l>` to filter or `--all` to include closed issues. |
| `close <id>`     | Closes an issue by its numerical ID.                                                    |
| `publish <fmt>`  | Exports all issues to `markdown` or `json` format.                                      |
| `publish-github` | Publishes all open local issues to your GitHub repository.                              |
| `help`           | Shows the help message.                                                                 |

**Example Workflow:**

```bash
# Add a new feature and a bug
issueify add "FEAT: Implement dark mode"
issueify add "BUG: Submit button is not aligned correctly"

# List your current issues
issueify list

# Publish them to GitHub
issueify publish-github
```

---

## Configuration for GitHub Publishing

The `publish-github` command is designed to work out-of-the-box if you use the official GitHub CLI (`gh`).

### Using `gh` (Recommended)

If you have `gh` installed and authenticated (`gh auth login`), `issueify` will automatically detect your repository and use your credentials. No further setup is needed.

### Using Environment Variables (Fallback)

If `gh` is not available, `issueify` will fall back to using environment variables. You must set the following:

- `GITHUB_TOKEN`: A [Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token) with `repo` scope.
- `GITHUB_OWNER`: The username or organization that owns the repository (e.g., `octocat`).
- `GITHUB_REPO`: The name of the repository (e.g., `Hello-World`).

```bash
export GITHUB_TOKEN="your_github_token"
export GITHUB_OWNER="your_username"
export GITHUB_REPO="your_project_repo"
```

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
