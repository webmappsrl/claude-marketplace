> Ticket: oc:8661

# Giorni del ticket per la ricerca sulle call — piano di implementazione

> **Per chi esegue:** sub-skill `superpowers:executing-plans`. I passi hanno la checkbox (`- [ ]`).
> **Nessun commit durante l'esecuzione:** il commit lo fa il dev alla fine, su `develop`, senza
> branch (regola del `CLAUDE.md` del repo).

**Obiettivo:** `wm-plan` passa a `wm-transcript-research` i giorni del ticket (cambi di stato,
giorni in `progress`, giorni con un commit `(oc:<ID>)`, oggi) calcolati da un tool nuovo del server
MCP, invece di una finestra di date.

**Architettura:** un file Go nuovo, `internal/tools/storydays.go`, con il tool `get_story_days`
registrato nel gruppo `stories`: due GET su `status-history` con `Client.Do`, `git fetch` e
`git log` con `os/exec` sui repository indicati o su quello in cui gira la sessione, unione in un
insieme. Poi testo: prompt dell'agente, `wm-plan`, documentazione.

**Tecnologie:** Go 1.27.1, `github.com/modelcontextprotocol/go-sdk` v1.7.0, `git`.

**Spec:** [overview.md](overview.md)

## Vincoli globali

- Il tool non scrive nulla: su Orchestrator solo GET, in git solo `fetch` e `log`.
- `git` gira con `GIT_TERMINAL_PROMPT=0`, `GIT_SSH_COMMAND=ssh -o BatchMode=yes` e un timeout di
  30 secondi per comando.
- Le date di Orchestrator sono la parte `AAAA-MM-GG` delle stringhe restituite: nessuna
  conversione di fuso.
- Nessun errore fa fallire il tool: va in `warnings`.
- Documentazione e messaggi in italiano; identificatori Go in inglese come nel resto del package.

## Punti da guardare in review

1. **Commit con il numero del ticket fuori dallo scope** (`vedi oc:8661`, `oc:86610`): non devono
   contare. Test nel Task 1.
2. **Periodo che inizia poco dopo mezzanotte italiana** (`00:30+02:00`, in UTC il giorno prima):
   la data è quella italiana. Test nel Task 1.
3. **Sessione aperta fuori da un repository**: solo Orchestrator e un warning, nessun errore. Test
   nel Task 1.
4. **Endpoint che risponde 404 o 401**: i giorni dei commit arrivano lo stesso. Test nel Task 1.
5. **Submodule**: senza `repos`, i commit del submodule contano. Test nel Task 1.

---

### Task 1: tool `get_story_days`

**File:**
- Crea: `plugins/wm-skills/mcp/internal/tools/storydays.go`
- Crea: `plugins/wm-skills/mcp/internal/tools/storydays_test.go`
- Modifica: `plugins/wm-skills/mcp/internal/tools/registry.go:43-45` (registrazione nel gruppo
  `stories`)
- Rigenera: `plugins/wm-skills/bin/orchestrator-mcp`

**Interfacce:**
- Usa: `Deps.Client.Do(ctx, "GET", path, nil) (json.RawMessage, error)`; `testSession`,
  `callTool`, `resultText` di `handlers_test.go`.
- Produce: tool MCP `get_story_days` con input `{story_id int, repos []string (facoltativo)}` e
  output `{story_id, current_status, days []string, warnings []string}`, usato dai Task 2 e 3.

- [ ] **Passo 1: scrivi i test che falliscono** — `storydays_test.go`:

```go
package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

const historyAll = `{"story_id":8636,"timezone":"Europe/Rome","current_status":"progress","intervals":[
 {"status":"new","from":"2026-09-21T08:00:00+02:00","to":"2026-09-21T09:00:00+02:00"},
 {"status":"progress","from":"2026-09-21T09:00:00+02:00","to":"2026-09-21T17:00:00+02:00"},
 {"status":"waiting","from":"2026-09-21T17:00:00+02:00","to":"2026-09-23T09:30:00+02:00"},
 {"status":"progress","from":"2026-09-23T09:30:00+02:00","to":null}],
 "days":[{"date":"2026-09-21","statuses":{}},{"date":"2026-09-22","statuses":{}},{"date":"2026-09-23","statuses":{}}]}`

const historyProgress = `{"story_id":8636,"timezone":"Europe/Rome","current_status":"progress","intervals":[],
 "days":[{"date":"2026-09-21","statuses":{"progress":480}},{"date":"2026-09-23","statuses":{"progress":150}}]}`

// historyServer risponde a status-history con all senza filtro e con progress
// per ?status=progress; con code diverso da 200 risponde solo l'errore.
func historyServer(t *testing.T, code int, all, progress string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if code != http.StatusOK {
			w.WriteHeader(code)
			_, _ = w.Write([]byte(`{"message":"No query results"}`))
			return
		}
		if r.URL.Query().Get("status") == "progress" {
			_, _ = w.Write([]byte(progress))
			return
		}
		_, _ = w.Write([]byte(all))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func gitT(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitT(t, dir, nil, "init", "-q", "-b", "main")
	gitT(t, dir, nil, "config", "user.email", "test@example.com")
	gitT(t, dir, nil, "config", "user.name", "Test")
	return dir
}

func commitAt(t *testing.T, dir, msg, date string) {
	t.Helper()
	gitT(t, dir, []string{"GIT_AUTHOR_DATE=" + date, "GIT_COMMITTER_DATE=" + date},
		"commit", "-q", "--allow-empty", "-m", msg)
}

func fixToday(t *testing.T) {
	t.Helper()
	t.Setenv("TZ", "Europe/Rome")
	old := nowFunc
	nowFunc = func() time.Time { return time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { nowFunc = old })
}

func callStoryDays(t *testing.T, url string, args map[string]any) storyDaysOutput {
	t.Helper()
	session := testSession(t, map[string]bool{"stories": true}, url)
	res := callTool(t, session, "get_story_days", args)
	if res.IsError {
		t.Fatalf("get_story_days non deve fallire: %s", resultText(res))
	}
	var out storyDaysOutput
	if err := json.Unmarshal([]byte(resultText(res)), &out); err != nil {
		t.Fatalf("output non leggibile: %v\n%s", err, resultText(res))
	}
	return out
}

func TestStoryDaysUnisceStatiProgressCommitEOggi(t *testing.T) {
	fixToday(t)
	repo := newRepo(t)
	commitAt(t, repo, "chore: inizio", "2026-09-20T10:00:00+02:00")
	gitT(t, repo, nil, "checkout", "-q", "-b", "feature/oc-8636")
	commitAt(t, repo, "feat(oc:8636): endpoint", "2026-09-25T10:00:00+02:00")
	commitAt(t, repo, "docs: vedi oc:8636 e (oc:86360)", "2026-09-26T10:00:00+02:00")
	gitT(t, repo, nil, "checkout", "-q", "main")

	srv := historyServer(t, http.StatusOK, historyAll, historyProgress)
	out := callStoryDays(t, srv.URL, map[string]any{"story_id": 8636, "repos": []string{repo}})

	want := []string{"2026-09-21", "2026-09-23", "2026-09-25", "2026-09-29"}
	if !reflect.DeepEqual(out.Days, want) {
		t.Fatalf("giorni: atteso %v (il 22 è waiting, il 26 non ha lo scope), ottenuto %v", want, out.Days)
	}
	if out.CurrentStatus != "progress" {
		t.Fatalf("current_status: atteso progress, ottenuto %q", out.CurrentStatus)
	}
}

func TestStoryDaysSenzaRigheDiStatoEMezzanotte(t *testing.T) {
	fixToday(t)
	all := `{"current_status":"todo","intervals":[{"status":"todo","from":"2026-09-22T00:30:00+02:00","to":null}],"days":[]}`
	srv := historyServer(t, http.StatusOK, all, `{"intervals":[],"days":[]}`)
	out := callStoryDays(t, srv.URL, map[string]any{"story_id": 1, "repos": []string{newRepo(t)}})

	want := []string{"2026-09-22", "2026-09-29"}
	if !reflect.DeepEqual(out.Days, want) {
		t.Fatalf("atteso %v (00:30 italiane resta il 22), ottenuto %v", want, out.Days)
	}
}

func TestStoryDaysEndpointInErroreTieneICommit(t *testing.T) {
	fixToday(t)
	repo := newRepo(t)
	commitAt(t, repo, "fix(oc:8636): correzione", "2026-09-24T10:00:00+02:00")
	srv := historyServer(t, http.StatusNotFound, "", "")
	out := callStoryDays(t, srv.URL, map[string]any{"story_id": 8636, "repos": []string{repo}})

	want := []string{"2026-09-24", "2026-09-29"}
	if !reflect.DeepEqual(out.Days, want) {
		t.Fatalf("atteso %v, ottenuto %v", want, out.Days)
	}
	if len(out.Warnings) == 0 {
		t.Fatal("l'errore dell'endpoint deve finire in warnings")
	}
}

func TestStoryDaysSenzaReposUsaIlRepositoryESubmodule(t *testing.T) {
	fixToday(t)
	sub := newRepo(t)
	commitAt(t, sub, "feat(oc:8636): nel submodule", "2026-09-27T10:00:00+02:00")
	main := newRepo(t)
	commitAt(t, main, "chore: inizio", "2026-09-20T10:00:00+02:00")
	gitT(t, main, nil, "-c", "protocol.file.allow=always", "submodule", "add", "-q", sub, "pkg")
	commitAt(t, main, "chore: submodule", "2026-09-20T11:00:00+02:00")
	t.Chdir(main)

	srv := historyServer(t, http.StatusOK, `{"intervals":[],"days":[]}`, `{"intervals":[],"days":[]}`)
	out := callStoryDays(t, srv.URL, map[string]any{"story_id": 8636})

	want := []string{"2026-09-27", "2026-09-29"}
	if !reflect.DeepEqual(out.Days, want) {
		t.Fatalf("atteso %v (commit del submodule), ottenuto %v", want, out.Days)
	}
}

func TestStoryDaysFuoriDaUnRepository(t *testing.T) {
	fixToday(t)
	t.Chdir(t.TempDir())
	srv := historyServer(t, http.StatusOK, historyAll, historyProgress)
	out := callStoryDays(t, srv.URL, map[string]any{"story_id": 8636})

	want := []string{"2026-09-21", "2026-09-23", "2026-09-29"}
	if !reflect.DeepEqual(out.Days, want) {
		t.Fatalf("atteso %v, ottenuto %v", want, out.Days)
	}
	if len(out.Warnings) == 0 {
		t.Fatal("fuori da un repository serve un warning")
	}
}
```

- [ ] **Passo 2: verifica che falliscano**

Esegui: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -run StoryDays`
Atteso: errore di compilazione, `storyDaysOutput` e `nowFunc` non definiti.

- [ ] **Passo 3: implementa** — `storydays.go`:

```go
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
func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_SSH_COMMAND=ssh -o BatchMode=yes")
	out, err := cmd.Output()
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
	if subs, err := runGit(ctx, repos[0], "submodule", "foreach", "--quiet", "--recursive", "pwd"); err == nil {
		repos = append(repos, lines(subs)...)
	}
	return repos, ""
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
```

In `registry.go`, dentro `if groups["stories"] {`, dopo `registerStories(server, deps)`:

```go
		registerStoryDays(server, deps)
```

- [ ] **Passo 4: verifica che passino, poi tutta la suite**

Esegui: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -run StoryDays -v && go test ./...`
Atteso: PASS. Se `registry_test.go` conta i tool del gruppo `stories`, aggiorna il conteggio.

- [ ] **Passo 5: ricompila il binario**

Esegui: `plugins/wm-skills/mcp/build.sh`
Atteso: `Compilato: …/bin/orchestrator-mcp`.

---

### Task 2: agente `wm-transcript-research`

**File:** Modifica `plugins/wm-skills/agents/wm-transcript-research.md` (righe 71-72, 84-89,
100-101, 196).

**Interfacce:** consuma l'elenco `days` del Task 1, passato da `wm-plan` come «giorni del ticket».

- [ ] **Passo 1: pulizia per data di creazione** — sostituisci il paragrafo **Pulizia** con:

```markdown
**Pulizia.** Guarda nell'elenco i notebook `scrum …` **creati** più di 90 giorni fa (la data di
creazione di `notebook_list`, non il giorno del nome) e riportali nella risposta sotto
`Notebook vecchi:`. I notebook che hai interrogato in questa ricerca non li riporti mai: servono
al ticket, anche se il loro giorno è vecchio. Non li cancelli mai: decide il dev.
```

- [ ] **Passo 2: giorni del ticket al posto della finestra** — in `## Cosa ti chiedono` sostituisci
la voce della finestra e la riga sotto l'elenco con:

```markdown
- **domande sui giorni del ticket** — un elenco di date, non un intervallo: i giorni in cui al
  ticket è successo qualcosa, calcolati da chi ti chiama; più eventuali **tag** del ticket, da
  interrogare solo se il loro notebook esiste, salvo «crea se manca».
- **domande sul giorno di oggi** — per esempio prima di formulare un ticket nuovo.

I giorni **comprendono sempre oggi**. Senza indicazioni, sono solo oggi.
```

- [ ] **Passo 3: ricerca per date sparse** — in `## Le fonti` sostituisci il paragrafo «Per una
finestra di più giorni…» con:

```markdown
Per più giorni elenca tutte le date in una ricerca (`title contains '2026/09/15' or title contains
'2026/09/22' …`), una clausola per data, seguendo le pagine: poi raggruppa le call per giorno.
```

- [ ] **Passo 4: `Copertura`** — nel formato, la riga `- finestra: <data>–<data>; giorni con
almeno una call: <elenco>` diventa:

```markdown
- giorni del ticket: <elenco>; con almeno una call: <elenco>
```

- [ ] **Passo 5: controllo**

Esegui: `grep -n "finestra" plugins/wm-skills/agents/wm-transcript-research.md`
Atteso: nessuna riga.

---

### Task 3: skill `wm-plan`

**File:** Modifica `plugins/wm-skills/skills/wm-plan/SKILL.md` (tabella `## Orchestrator`,
`caso-a-split-execution`, `reverse-interaction` righe 732-760).

**Interfacce:** consuma `get_story_days` (Task 1) e la richiesta «domande sui giorni del ticket»
(Task 2).

- [ ] **Passo 1: tabella `## Orchestrator`** — aggiungi dopo la riga di `get_story`:

```markdown
| Giorni in cui al ticket è successo qualcosa (stati, progress, commit, oggi) | `get_story_days` |
```

- [ ] **Passo 2: `reverse-interaction`** — nel punto «Non chiedere ciò che il team ha già deciso in
call», sostituisci la parentesi sulla finestra (da «la **finestra di giorni**» a «almeno una
call)») con:

```markdown
i **giorni del ticket** — chiedili a `get_story_days` con l'id del ticket, senza `repos` (vale
il repository in cui lavori, con i submodule): cambi di stato, giorni in `progress`, giorni con
un commit `(oc:<ID>)`, oggi. Mostra al dev i giorni e, **in modo evidente**, ogni riga di
`warnings`: un endpoint o un `git fetch` non riusciti vogliono dire giorni mancanti —
```

- [ ] **Passo 3: senza ticket e domande dirette** — sostituisci la riga «**Senza ticket** la
finestra è «oggi»…» e il paragrafo **Domande dirette** con:

```markdown
  **Senza ticket** i giorni sono solo oggi: le call di oggi ci sono sempre, con o senza ticket.

  **Domande dirette** del dev sulle call di quel ticket («cosa si è deciso su questo ticket?»,
  «guarda cosa si è deciso in call»): richiama `get_story_days` e passa all'agente i giorni del
  ticket, più i giorni che il dev indica («guarda anche il 10/09»), e i tag con «crea se manca». Se
  la `Copertura` riporta tag senza notebook, dillo al dev: può chiederti di crearli.
```

- [ ] **Passo 4: split** — in `caso-a-split-execution`, punto 4, dopo «prosegui il workflow in
`Fase: init-context` usando i dati già noti di quel ticket (senza rifare la GET).» aggiungi:

```markdown
  Se è uno dei ticket nuovi, in `reverse-interaction` aggiungi ai suoi giorni quelli del ticket
  originale (`get_story_days` con l'id originale): il ticket nuovo nasce oggi e non ha storia,
  mentre le sue richieste sono state discusse nei giorni del ticket da cui viene.
```

- [ ] **Passo 5: controllo**

Esegui: `grep -n "finestra di giorni\|qualche giorno prima" plugins/wm-skills/skills/wm-plan/SKILL.md`
Atteso: nessuna riga. Poi `./.github/scripts/verifica-diagramma.sh` → allineato; se il paragrafo di
`reverse-interaction` nel diagramma parla di finestra, aggiornane il testo (solo contenuto).

---

### Task 4: documentazione, controlli e prova vera

**File:** Modifica `docs/knowledge/wm-transcript-research.md` (righe 20-23, 71 circa sulla
pulizia, 103-105), `docs/howto/provare-wm-transcript-research.md` (riga 28).

- [ ] **Passo 1: conoscenza** — righe 20-23: al posto di «dei giorni della finestra (da qualche
giorno prima … oc:8636)» scrivi «dei **giorni del ticket** (cambi di stato, giorni in `progress`,
giorni con un commit `(oc:<ID>)`, oggi: li calcola il tool `get_story_days`, oc:8661)». Righe
103-105: togli la frase «La finestra andrà calcolata … (oc:8636).» e al suo posto scrivi che i
giorni si scelgono per eventi del ticket, non per date, perché di un ticket si parla quando gli
succede qualcosa. Se la pagina descrive la pulizia a 90 giorni, precisa «per data di creazione del
notebook».

- [ ] **Passo 2: howto** — riga 28: «finestra degli ultimi trenta giorni» diventa «giorni del
ticket calcolati con `get_story_days`».

- [ ] **Passo 3: controlli**

Esegui: `claude plugin validate . && ./.github/scripts/verifica-diagramma.sh && (cd plugins/wm-skills/mcp && go test ./...)`
Atteso: tutto passa.

- [x] **Passo 4: prova vera (dopo il riavvio della sessione)** — chiama `get_story_days` su
oc:8543 e oc:8636 senza `repos`: i giorni coincidono con i cambi di stato visibili nella tab Logs
di Nova e con i commit `(oc:<ID>)` del repo; poi `wm-plan` su un ticket fino alla prima domanda di
`reverse-interaction`, controllando che all'agente arrivino i giorni e non una finestra. Esiti in
[notes.md](notes.md).

---

**Commit (istruzione per il dev), uno solo a lavoro finito, su `develop`:**

```bash
git add plugins/wm-skills/mcp/internal/tools/storydays.go plugins/wm-skills/mcp/internal/tools/storydays_test.go \
  plugins/wm-skills/mcp/internal/tools/registry.go plugins/wm-skills/bin/orchestrator-mcp \
  plugins/wm-skills/agents/wm-transcript-research.md plugins/wm-skills/skills/wm-plan/SKILL.md \
  docs/knowledge/wm-transcript-research.md docs/howto/provare-wm-transcript-research.md \
  docs/guide/wm-plan-diagramma/index.html \
  docs/features/8661-wm-plan-notebook-scelti-dai-giorni-di-stato-di-progress-e-dei-commit-del-ticket/
git commit -m "feat(oc:8661): wm-plan sceglie i notebook dai giorni di stato, di progress e dei commit del ticket"
```
