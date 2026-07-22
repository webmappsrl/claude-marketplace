# Notes — Fix diagramma header cross-repo

## Deviazioni dal piano
- Il piano (Task 1, Step 2) prevedeva `curl -s --max-time 5 ...` e trattava "risposta HTTP non 2xx" come caso di fallimento. In verifica (Step 4) è emerso che `curl -s` da solo **non fallisce** su risposte non-2xx: su un 404 restituisce il body dell'errore (`404: Not Found`) con exit code 0, invece di un output vuoto/fallimento. Corretto usando `curl -sf` (flag `-f`, fail-fast su HTTP error), che restituisce effettivamente output vuoto ed exit code non-zero su 404 — verificato con exit code 56.

## Bug trovati
Nessuno oltre alla deviazione sopra.

## Decisioni
Nessuna decisione on-the-fly oltre alla correzione del flag curl.

## Follow-up
Nessuno.
