# wm-tag e la tag-mode di wm-plan

## Stato attuale

`wm-tag` trasforma una trascrizione o un brief cliente in un tag Orchestrator con i suoi ticket
figli. Per ogni ticket invoca `wm-plan` in **tag-mode**: `wm-plan` salta `write-plan`,
`execution`, `notes` e `update-context`, e l'overview finisce nella description del ticket
invece che sul filesystem.

I tag si chiamano `[RDO][CLIENTE][ANNO]N`, con **N calcolato contando i tag esistenti** per lo
stesso cliente e anno su Orchestrator: nessuna coordinazione manuale, nessun conflitto.

La navigazione multi-repo usa `~/.config/webmapp/repos.json`, un dizionario persistente
**aggiornato in modo incrementale** e mai riscritto da zero, così i path inseriti a mano
sopravvivono.

La regola di preview più conferma prima di ogni scrittura su Orchestrator vale **anche sui
tag**, non solo sulle story.

## Come ci siamo arrivati

- **`repos.json` non si rigenera** (oc:8157): riscriverlo ad ogni giro avrebbe cancellato i path
  aggiunti a mano per i repo che la scansione automatica non trova.
- **La regola sulle scritture è stata estesa esplicitamente ai tag** (oc:8157): era formulata
  per le story, e senza l'estensione un POST su un tag sarebbe passato senza preview.
