package tools

import "testing"

func TestListTasksPathEncodesSpecialCharacters(t *testing.T) {
	got := listTasksPath("-created_at & status")
	want := "/api/tasks?sort=-created_at+%26+status"
	if got != want {
		t.Fatalf("percorso non codificato correttamente:\n got:  %s\n want: %s", got, want)
	}
}

func TestListTasksPathWithoutSortHasNoQuery(t *testing.T) {
	if got := listTasksPath(""); got != "/api/tasks" {
		t.Fatalf("senza ordinamento non deve comparire la query string: %s", got)
	}
}
