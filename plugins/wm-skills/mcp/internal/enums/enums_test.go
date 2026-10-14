package enums

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParsePHPEnumReadsValuesNotCaseNames(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "StoryType.php"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParsePHPEnum(raw)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	want := []string{"Bug", "Feature", "Help desk", "Scrum"}

	if len(got) != len(want) {
		t.Fatalf("attesi %v, ottenuti %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("valore %d: atteso %q, ottenuto %q", i, want[i], got[i])
		}
	}
}

func TestParsePHPEnumIgnoresMethodBody(t *testing.T) {
	// Il corpo del metodo values() contiene la parola "cases" e non deve
	// produrre valori spuri.
	raw := []byte("enum X: string {\n case A = 'a';\n public static function values(): array { return array_column(self::cases(), 'value'); }\n}")
	got, err := ParsePHPEnum(raw)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("atteso un solo valore \"a\", ottenuto %v", got)
	}
}

func TestParsePHPEnumOnGarbage(t *testing.T) {
	got, err := ParsePHPEnum([]byte("non è PHP"))
	if err != nil {
		t.Fatalf("errore inatteso su contenuto senza dichiarazioni case: %v", err)
	}
	if got != nil {
		t.Fatalf("atteso nil su contenuto non riconoscibile, ottenuto %v", got)
	}
}

func TestParsePHPEnumAcceptsDoubleQuotes(t *testing.T) {
	raw := []byte("enum X: string {\n case A = \"a\";\n case B = \"b\";\n}")
	got, err := ParsePHPEnum(raw)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	want := []string{"a", "b"}
	if len(got) != len(want) {
		t.Fatalf("attesi %v, ottenuti %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("valore %d: atteso %q, ottenuto %q", i, want[i], got[i])
		}
	}
}

func TestParsePHPEnumErrorsOnUnrecognizedCaseLine(t *testing.T) {
	// La seconda riga usa una costante invece di un letterale fra apici:
	// l'espressione regolare non la riconosce. Un elenco parziale silenzioso
	// farebbe respingere come inesistente un valore che invece c'è.
	raw := []byte("enum X: string {\n case A = 'a';\n case B = B_CONST;\n}")
	got, err := ParsePHPEnum(raw)
	if err == nil {
		t.Fatalf("atteso un errore per una riga case non riconosciuta, ottenuto elenco %v", got)
	}
}

func TestLoadOrFetchFailsWhenNetworkDownAndCacheCorrupt(t *testing.T) {
	origType, origStatus, origQuote := storyTypeURL, storyStatusURL, quoteStatusURL
	quoteStatusURL = "http://127.0.0.1:1"
	storyTypeURL = "http://127.0.0.1:1"
	storyStatusURL = "http://127.0.0.1:1"
	t.Cleanup(func() { storyTypeURL, storyStatusURL, quoteStatusURL = origType, origStatus, origQuote })

	cache := filepath.Join(t.TempDir(), "enums.json")
	// Copia locale troncata/corrotta: non è JSON valido.
	if err := os.WriteFile(cache, []byte("{non json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadOrFetch(context.Background(), cache)
	if err == nil {
		t.Fatal("atteso un errore quando la rete non risponde e la copia locale è corrotta")
	}
}

func TestFetchFreshMergesFreshWithCacheOnPartialFailure(t *testing.T) {
	// fetchFresh è la funzione sincrona usata sia quando non esiste copia
	// locale, sia in secondo piano per aggiornarla: qui si verifica ancora il
	// comportamento di merce fresco+cache, indipendentemente dal fast path di
	// LoadOrFetch (verificato a parte).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("enum X: string {\n case A = 'nuovo';\n}"))
	}))
	defer srv.Close()

	origType, origStatus, origQuote := storyTypeURL, storyStatusURL, quoteStatusURL
	quoteStatusURL = "http://127.0.0.1:1"
	storyTypeURL = srv.URL
	storyStatusURL = "http://127.0.0.1:1" // non raggiungibile
	t.Cleanup(func() { storyTypeURL, storyStatusURL, quoteStatusURL = origType, origStatus, origQuote })

	cache := filepath.Join(t.TempDir(), "enums.json")
	old := Values{StoryType: []string{"vecchio"}, StoryStatus: []string{"vecchio-status"}}
	encoded, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, encoded, 0o644); err != nil {
		t.Fatal(err)
	}

	v, err := fetchFresh(context.Background(), cache)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if len(v.StoryType) != 1 || v.StoryType[0] != "nuovo" {
		t.Fatalf("atteso il valore appena scaricato per StoryType, ottenuto %v", v.StoryType)
	}
	if len(v.StoryStatus) != 1 || v.StoryStatus[0] != "vecchio-status" {
		t.Fatalf("atteso il valore dalla copia locale per StoryStatus, ottenuto %v", v.StoryStatus)
	}

	updated, err := os.ReadFile(cache)
	if err != nil {
		t.Fatal(err)
	}
	var got Values
	if err := json.Unmarshal(updated, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.StoryType) != 1 || got.StoryType[0] != "nuovo" || len(got.StoryStatus) != 1 || got.StoryStatus[0] != "vecchio-status" {
		t.Fatalf("la copia locale non è stata aggiornata con la combinazione fresco+cache: %+v", got)
	}
}

func TestLoadOrFetchReturnsCacheImmediatelyWithoutWaitingForNetwork(t *testing.T) {
	// Se la copia locale esiste ed è utilizzabile, LoadOrFetch deve
	// restituirla subito, senza attendere la rete: il server deve essere
	// pronto immediatamente a ogni sessione. Punta a un indirizzo che non
	// risponde mai (127.0.0.1:1): se LoadOrFetch attendesse la rete, il test
	// scadrebbe per timeout.
	origType, origStatus, origQuote := storyTypeURL, storyStatusURL, quoteStatusURL
	quoteStatusURL = "http://127.0.0.1:1"
	storyTypeURL = "http://127.0.0.1:1"
	storyStatusURL = "http://127.0.0.1:1"
	t.Cleanup(func() { storyTypeURL, storyStatusURL, quoteStatusURL = origType, origStatus, origQuote })

	cache := filepath.Join(t.TempDir(), "enums.json")
	old := Values{StoryType: []string{"vecchio"}, StoryStatus: []string{"vecchio-status"}}
	encoded, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, encoded, 0o644); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	v, err := LoadOrFetch(context.Background(), cache)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("LoadOrFetch non deve attendere la rete quando la copia locale esiste: %v", elapsed)
	}
	if len(v.StoryType) != 1 || v.StoryType[0] != "vecchio" {
		t.Fatalf("atteso il valore dalla copia locale, ottenuto %v", v.StoryType)
	}
	if len(v.StoryStatus) != 1 || v.StoryStatus[0] != "vecchio-status" {
		t.Fatalf("atteso il valore dalla copia locale, ottenuto %v", v.StoryStatus)
	}
}

func TestFetchFreshReadsQuoteStatusAndKeepsOldCacheUsable(t *testing.T) {
	// Lo stato dei preventivi arriva da QuoteStatus.php come gli stati dei
	// ticket. Una copia locale salvata prima che esistesse quote_status deve
	// restare utilizzabile, e il nuovo elenco deve finire nella copia.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("enum QuoteStatus: string {\n case New = 'new';\n case On_Hold = 'on hold';\n}"))
	}))
	defer srv.Close()

	origType, origStatus, origQuote := storyTypeURL, storyStatusURL, quoteStatusURL
	storyTypeURL = "http://127.0.0.1:1"
	storyStatusURL = "http://127.0.0.1:1"
	quoteStatusURL = srv.URL
	t.Cleanup(func() { storyTypeURL, storyStatusURL, quoteStatusURL = origType, origStatus, origQuote })

	cache := filepath.Join(t.TempDir(), "enums.json")
	if err := os.WriteFile(cache, []byte(`{"story_type":["Bug"],"story_status":["new"]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	v, err := fetchFresh(context.Background(), cache)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if len(v.QuoteStatus) != 2 || v.QuoteStatus[1] != "on hold" {
		t.Fatalf("atteso il valore letterale dei case, non il nome: %v", v.QuoteStatus)
	}
	if len(v.StoryType) != 1 || v.StoryType[0] != "Bug" {
		t.Fatalf("atteso StoryType dalla copia locale, ottenuto %v", v.StoryType)
	}

	var got Values
	updated, err := os.ReadFile(cache)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(updated, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.QuoteStatus) != 2 {
		t.Fatalf("la copia locale non contiene quote_status: %+v", got)
	}
}
