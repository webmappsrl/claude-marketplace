package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
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

func TestStoryDaysFetchBloccatoRispettaIlTimeout(t *testing.T) {
	fixToday(t)
	oldT, oldS := gitTimeout, sshCommand
	gitTimeout, sshCommand = time.Second, "sleep 30 #"
	t.Cleanup(func() { gitTimeout, sshCommand = oldT, oldS })
	repo := newRepo(t)
	commitAt(t, repo, "feat(oc:8636): x", "2026-09-24T10:00:00+02:00")
	gitT(t, repo, nil, "remote", "add", "origin", "ssh://git@example.invalid/x.git")

	start := time.Now()
	dates, warns := commitDates(context.Background(), repo, 8636)
	if d := time.Since(start); d > 10*time.Second {
		t.Fatalf("il fetch bloccato deve tornare entro pochi secondi, ci ha messo %v", d)
	}
	if !reflect.DeepEqual(dates, []string{"2026-09-24"}) {
		t.Fatalf("i commit locali vanno trovati comunque, ottenuto %v", dates)
	}
	if len(warns) == 0 || !strings.Contains(warns[0], "tempo") {
		t.Fatalf("il warning deve dire che è scaduto il tempo: %v", warns)
	}
}

func TestStoryDaysSubmoduleNonInizializzatoDaUnWarning(t *testing.T) {
	fixToday(t)
	sub := newRepo(t)
	commitAt(t, sub, "feat(oc:8636): nel submodule", "2026-09-27T10:00:00+02:00")
	main := newRepo(t)
	commitAt(t, main, "chore: inizio", "2026-09-20T10:00:00+02:00")
	gitT(t, main, nil, "-c", "protocol.file.allow=always", "submodule", "add", "-q", sub, "pkg")
	commitAt(t, main, "chore: submodule", "2026-09-20T11:00:00+02:00")
	clone := filepath.Join(t.TempDir(), "clone")
	gitT(t, main, nil, "clone", "-q", main, clone)
	t.Chdir(clone)

	_, warn := defaultRepos(context.Background())
	if !strings.Contains(warn, "pkg") {
		t.Fatalf("un submodule non inizializzato va segnalato per nome, warning: %q", warn)
	}
}
