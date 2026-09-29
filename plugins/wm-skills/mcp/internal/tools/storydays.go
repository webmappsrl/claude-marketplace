package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type storyDaysInput struct {
	StoryID int      `json:"story_id" jsonschema:"identificatore numerico del ticket, la parte dopo oc:"`
	Repos   []string `json:"repos,omitempty" jsonschema:"percorsi dei repository in cui cercare i commit; senza, il repository in cui è aperta la sessione e i suoi submodule"`
}

type storyDaysOutput struct {
	StoryID       int      `json:"story_id"`
	CurrentStatus string   `json:"current_status,omitempty"`
	Days          []string `json:"days"`
	Warnings      []string `json:"warnings,omitempty"`
}

type statusHistory struct {
	CurrentStatus string `json:"current_status"`
	Intervals     []struct {
		From string `json:"from"`
	} `json:"intervals"`
	Days []struct {
		Date string `json:"date"`
	} `json:"days"`
}

// nowFunc e gitTimeout sono variabili per poterli fissare nei test.
var (
	nowFunc    = time.Now
	gitTimeout = 30 * time.Second
	sshCommand = "ssh -o BatchMode=yes -o ConnectTimeout=10"
)

func registerStoryDays(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_story_days",
		Description: "Restituisce i giorni in cui a un ticket è successo qualcosa: cambi di stato e giorni in progress (da Orchestrator), " +
			"giorni con un commit che ha (oc:<ID>) nel messaggio, e oggi. Date AAAA-MM-GG in ordine, senza doppioni. " +
			"Non fallisce: i problemi incontrati (endpoint, git fetch, repository non trovato) sono in warnings.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in storyDaysInput) (*mcp.CallToolResult, storyDaysOutput, error) {
		return nil, storyDays(ctx, deps, in), nil
	})
}

func storyDays(ctx context.Context, deps Deps, in storyDaysInput) storyDaysOutput {
	out := storyDaysOutput{StoryID: in.StoryID}
	set := map[string]bool{}
	addDate(set, nowFunc().In(time.Local).Format("2006-01-02"))

	base := fmt.Sprintf("/api/stories/%d/status-history", in.StoryID)
	if h, err := fetchHistory(ctx, deps, base); err != nil {
		out.Warnings = append(out.Warnings, "Orchestrator, cambi di stato non letti: "+err.Error())
	} else {
		out.CurrentStatus = h.CurrentStatus
		for _, iv := range h.Intervals {
			addDate(set, iv.From)
		}
	}
	if h, err := fetchHistory(ctx, deps, base+"?status=progress"); err != nil {
		out.Warnings = append(out.Warnings, "Orchestrator, giorni in progress non letti: "+err.Error())
	} else {
		for _, d := range h.Days {
			addDate(set, d.Date)
		}
	}

	repos := in.Repos
	if len(repos) == 0 {
		var warn string
		repos, warn = defaultRepos(ctx)
		if warn != "" {
			out.Warnings = append(out.Warnings, warn)
		}
	}
	for _, repo := range repos {
		dates, warns := commitDates(ctx, repo, in.StoryID)
		out.Warnings = append(out.Warnings, warns...)
		for _, d := range dates {
			addDate(set, d)
		}
	}

	out.Days = make([]string, 0, len(set))
	for d := range set {
		out.Days = append(out.Days, d)
	}
	sort.Strings(out.Days)
	return out
}

func fetchHistory(ctx context.Context, deps Deps, path string) (statusHistory, error) {
	var h statusHistory
	raw, err := deps.Client.Do(ctx, "GET", path, nil)
	if err != nil {
		return h, err
	}
	if err := json.Unmarshal(raw, &h); err != nil {
		return h, fmt.Errorf("risposta non leggibile: %w", err)
	}
	return h, nil
}

// addDate prende i primi dieci caratteri: le date di Orchestrator sono già
// nel fuso Europe/Rome, e convertirle sposterebbe di giorno i periodi che
// iniziano poco dopo mezzanotte.
func addDate(set map[string]bool, s string) {
	if len(s) < 10 {
		return
	}
	if _, err := time.Parse("2006-01-02", s[:10]); err == nil {
		set[s[:10]] = true
	}
}

// runGit non chiede mai credenziali e non resta fermo oltre gitTimeout.
// WaitDelay serve perché, ucciso git, un ssh figlio può tenere aperta la pipe
// di stderr: senza, Output() aspetterebbe che ssh esca da solo (misurato 1m15s
// con una rete che scarta i pacchetti).
func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_SSH_COMMAND="+sshCommand)
	cmd.WaitDelay = 2 * time.Second
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("tempo scaduto dopo %v", gitTimeout)
	}
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// defaultRepos è il repository in cui è aperta la sessione (la directory di
// lavoro del server MCP) più i suoi submodule.
func defaultRepos(ctx context.Context) ([]string, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "commit non cercati: directory di lavoro non leggibile"
	}
	top, err := runGit(ctx, cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Sprintf("commit non cercati: la sessione non è aperta in un repository git (%s)", cwd)
	}
	repos := []string{strings.TrimSpace(top)}
	var warns []string
	if subs, err := runGit(ctx, repos[0], "submodule", "foreach", "--quiet", "--recursive", "pwd"); err != nil {
		warns = append(warns, fmt.Sprintf("submodule non letti (%v)", err))
	} else {
		repos = append(repos, lines(subs)...)
	}
	// «git submodule status» marca con «-» i submodule registrati ma non
	// inizializzati: foreach li salta, e i loro commit mancherebbero in silenzio.
	if st, err := runGit(ctx, repos[0], "submodule", "status", "--recursive"); err == nil {
		for _, l := range lines(st) {
			if f := strings.Fields(l); len(f) >= 2 && strings.HasPrefix(f[0], "-") {
				warns = append(warns, fmt.Sprintf("submodule %s non inizializzato: i suoi commit non sono cercati (git submodule update --init)", f[1]))
			}
		}
	}
	return repos, strings.Join(warns, "; ")
}

func commitDates(ctx context.Context, repo string, storyID int) ([]string, []string) {
	var warns []string
	if _, err := runGit(ctx, repo, "fetch", "--all", "--quiet"); err != nil {
		warns = append(warns, fmt.Sprintf("%s: git fetch non riuscito, uso i branch già presenti (%v)", repo, err))
	}
	out, err := runGit(ctx, repo, "log", "--all", "--fixed-strings",
		"--grep", fmt.Sprintf("(oc:%d)", storyID), "--format=%ad", "--date=format-local:%Y-%m-%d")
	if err != nil {
		return nil, append(warns, fmt.Sprintf("%s: git log non riuscito (%v)", repo, err))
	}
	return lines(out), warns
}
