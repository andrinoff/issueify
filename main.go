package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// --- Data Structures ---

// Issue represents a single task or issue.
type Issue struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"` // "open" or "closed"
	Labels    []string  `json:"labels"`
	CreatedAt time.Time `json:"created_at"`
}

// LabelPattern defines a regex pattern to automatically assign a label.
type LabelPattern struct {
	Pattern *regexp.Regexp
	Label   string
}

// --- Constants and Configuration ---

const (
	dbFileName    = ".issue_tracker.json"
	statusOpen    = "open"
	statusClosed  = "closed"
	markdownTpl   = `# Project Issues

{{range .}}
- **[{{.Status}}]** {{.Title}} ` + "`[ID: {{.ID}}]`" + `
  - **Labels**: {{join .Labels ", "}}
  - **Created**: {{.CreatedAt.Format "2006-01-02"}}
{{end}}`
)

// defaultLabelPatterns defines the default regex rules for auto-labeling.
var defaultLabelPatterns = []LabelPattern{
	{Pattern: regexp.MustCompile(`(?i)^(BUG|FIX|BUGFIX):`), Label: "bug"},
	{Pattern: regexp.MustCompile(`(?i)^(FEAT|FEATURE):`), Label: "feature"},
	{Pattern: regexp.MustCompile(`(?i)^(DOCS|DOCUMENTATION):`), Label: "documentation"},
	{Pattern: regexp.MustCompile(`(?i)^(REFACTOR):`), Label: "refactor"},
	{Pattern: regexp.MustCompile(`(?i)^(TEST|TESTS):`), Label: "testing"},
	{Pattern: regexp.MustCompile(`(?i)^(CHORE):`), Label: "chore"},
}

// --- Core Logic ---

func getRepoRoot() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
			return path, nil
		}

		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf("not a git repository")
		}
		path = parent
	}
}

func getDBPath() (string, error) {
	root, err := getRepoRoot()
	if err != nil {
		return "", fmt.Errorf("could not find repository root: %w", err)
	}
	return filepath.Join(root, dbFileName), nil
}

func loadIssues() ([]Issue, error) {
	path, err := getDBPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Issue{}, nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read database file: %w", err)
	}

	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("could not parse database file: %w", err)
	}
	return issues, nil
}

func saveIssues(issues []Issue) error {
	path, err := getDBPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize issues to JSON: %w", err)
	}

	return ioutil.WriteFile(path, data, 0644)
}

func autoLabel(issue *Issue) {
	labelSet := make(map[string]bool)
	for _, l := range issue.Labels {
		labelSet[l] = true
	}

	for _, lp := range defaultLabelPatterns {
		if lp.Pattern.MatchString(issue.Title) {
			if !labelSet[lp.Label] {
				issue.Labels = append(issue.Labels, lp.Label)
				labelSet[lp.Label] = true
			}
		}
	}
	sort.Strings(issue.Labels)
}

// --- Command Functions ---

func addIssue(title string) {
	if title == "" {
		log.Fatal("Error: Issue title cannot be empty.")
	}

	issues, err := loadIssues()
	if err != nil {
		log.Fatalf("Error loading issues: %v", err)
	}

	maxID := 0
	for _, issue := range issues {
		if issue.ID > maxID {
			maxID = issue.ID
		}
	}

	newIssue := Issue{
		ID:        maxID + 1,
		Title:     title,
		Status:    statusOpen,
		Labels:    []string{},
		CreatedAt: time.Now(),
	}

	autoLabel(&newIssue)
	issues = append(issues, newIssue)

	if err := saveIssues(issues); err != nil {
		log.Fatalf("Error saving new issue: %v", err)
	}

	fmt.Printf("Successfully added issue #%d: %s\n", newIssue.ID, newIssue.Title)
	fmt.Printf("Labels: %s\n", strings.Join(newIssue.Labels, ", "))
}

func listIssues(filterLabel string, showClosed bool) {
	issues, err := loadIssues()
	if err != nil {
		log.Fatalf("Error loading issues: %v", err)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("                 Issue Tracker")
	fmt.Println("--------------------------------------------------")

	count := 0
	for _, issue := range issues {
		if !showClosed && issue.Status == statusClosed {
			continue
		}

		if filterLabel != "" {
			found := false
			for _, label := range issue.Labels {
				if label == filterLabel {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		statusMarker := "✅"
		if issue.Status == statusOpen {
			statusMarker = "⚪️"
		}

		fmt.Printf("%s ID: %-3d | %-50s | Labels: %s\n", statusMarker, issue.ID, issue.Title, strings.Join(issue.Labels, ", "))
		count++
	}

	if count == 0 {
		fmt.Println("No issues found.")
	}
	fmt.Println("--------------------------------------------------")
}

func closeIssue(idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatalf("Error: Invalid ID format. Please provide a number.")
	}

	issues, err := loadIssues()
	if err != nil {
		log.Fatalf("Error loading issues: %v", err)
	}

	found := false
	for i := range issues {
		if issues[i].ID == id {
			if issues[i].Status == statusClosed {
				fmt.Printf("Issue #%d is already closed.\n", id)
				return
			}
			issues[i].Status = statusClosed
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("Error: Issue with ID #%d not found.", id)
	}

	if err := saveIssues(issues); err != nil {
		log.Fatalf("Error saving updated issues: %v", err)
	}

	fmt.Printf("Successfully closed issue #%d.\n", id)
}

func publishIssues(format string) {
	issues, err := loadIssues()
	if err != nil {
		log.Fatalf("Error loading issues: %v", err)
	}

	funcMap := template.FuncMap{"join": strings.Join}

	switch format {
	case "markdown":
		t, err := template.New("publish").Funcs(funcMap).Parse(markdownTpl)
		if err != nil {
			log.Fatalf("Error parsing template: %v", err)
		}
		if err := t.Execute(os.Stdout, issues); err != nil {
			log.Fatalf("Error executing template: %v", err)
		}
	case "json":
		data, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			log.Fatalf("Error exporting to JSON: %v", err)
		}
		fmt.Println(string(data))
	default:
		log.Fatalf("Error: Unknown format '%s'. Supported formats: markdown, json.", format)
	}
}

// pushToGithub creates issues in a GitHub repository from local open issues.
// It will try to use the 'gh' CLI for repository info and authentication first.
// If 'gh' is not available, it will fall back to environment variables.
func pushToGithub() {
	var token, owner, repo string

	// Try to get config from 'gh' CLI first.
	ghRepoCmd := exec.Command("gh", "repo", "view", "--json", "name,owner", "--jq", ".owner.login + \"/\" + .name")
	ghRepoOutput, err := ghRepoCmd.Output()
	if err == nil {
		repoParts := strings.Split(strings.TrimSpace(string(ghRepoOutput)), "/")
		if len(repoParts) == 2 {
			owner = repoParts[0]
			repo = repoParts[1]
		}
	}

	ghTokenCmd := exec.Command("gh", "auth", "token")
	ghTokenOutput, err := ghTokenCmd.Output()
	if err == nil {
		token = strings.TrimSpace(string(ghTokenOutput))
	}

	if owner != "" && repo != "" && token != "" {
		fmt.Printf("Detected repository '%s/%s' and using auth token from 'gh' CLI.\n", owner, repo)
	} else {
		fmt.Println("Could not get repository info or token from 'gh' CLI. Falling back to environment variables.")
		token = os.Getenv("GITHUB_TOKEN")
		owner = os.Getenv("GITHUB_OWNER")
		repo = os.Getenv("GITHUB_REPO")

		if token == "" || owner == "" || repo == "" {
			log.Fatal("Error: Please install and authenticate the 'gh' CLI ('gh auth login'), or set GITHUB_TOKEN, GITHUB_OWNER, and GITHUB_REPO environment variables.")
		}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	localIssues, err := loadIssues()
	if err != nil {
		log.Fatalf("Error loading local issues: %v", err)
	}

	fmt.Printf("Publishing open issues to %s/%s...\n", owner, repo)
	count := 0
	for _, issue := range localIssues {
		if issue.Status == statusOpen {
			gitIssue := &github.IssueRequest{
				Title:  &issue.Title,
				Labels: &issue.Labels,
			}

			_, _, err := client.Issues.Create(ctx, owner, repo, gitIssue)
			if err != nil {
				log.Printf("Error creating GitHub issue for local ID #%d: %v", issue.ID, err)
				continue
			}
			fmt.Printf("Successfully created GitHub issue for: \"%s\"\n", issue.Title)
			count++
		}
	}
	fmt.Printf("Finished. Published %d issues to GitHub.\n", count)

	if err := saveIssues([]Issue{}); err != nil {
		log.Fatalf("Error clearing local issues after publishing: %v", err)
	}
	fmt.Println("Successfully cleared all local issues.")
}

// --- Main Function and CLI Handling ---

func printHelp() {
	fmt.Println(`
Issue Tracker - A simple CLI tool for managing development issues.

Usage:
  issue-tracker <command> [arguments]

Commands:
  add "<title>"         Adds a new issue. The title must be in quotes.
                        Prefix with 'BUG:', 'FEAT:', etc., for auto-labeling.

  list [--label=<l>]    Lists all open issues.
                        --label: Filter issues by a specific label.
                        --all: Show closed issues as well.

  close <id>            Closes an issue by its ID.

  publish <format>      Publishes all issues in a specified format (markdown, json).
                        Example: issue-tracker publish markdown > ISSUES.md

  push        Pushes all open issues to a GitHub repository and
                        clears the local issue database.
                        This command will automatically use the official 'gh' CLI
                        for authentication and repository detection if installed.
                        As a fallback, it will use GITHUB_TOKEN, GITHUB_OWNER,
                        and GITHUB_REPO environment variables.

  help                  Shows this help message.`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		if len(args) == 0 {
			log.Fatal("Error: 'add' command requires a title.")
		}
		addIssue(strings.Join(args, " "))
	case "list":
		filterLabel := ""
		showAll := false
		for _, arg := range args {
			if strings.HasPrefix(arg, "--label=") {
				filterLabel = strings.TrimPrefix(arg, "--label=")
			} else if arg == "--all" {
				showAll = true
			}
		}
		listIssues(filterLabel, showAll)
	case "close":
		if len(args) != 1 {
			log.Fatal("Error: 'close' command requires exactly one ID.")
		}
		closeIssue(args[0])
	case "publish":
		if len(args) != 1 {
			log.Fatal("Error: 'publish' command requires a format.")
		}
		publishIssues(args[0])
	case "push":
		pushToGithub()
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
	}
}