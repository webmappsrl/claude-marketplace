# wm-tag e la tag-mode di wm-plan

## Stato attuale

`wm-tag` trasforma una trascrizione o un brief cliente in un tag Orchestrator con i suoi ticket
figli. Per ogni ticket invoca `wm-plan` in **tag-mode**: `wm-plan` salta `write-plan`,
`execution`, `notes` e `update-context`, e l'overview finisce nella description del ticket
invece che sul filesystem.

I tag si chiamano `[RDO][CLIENTE][ANNO]N`, con **N calcolato contando i tag esistenti** per lo
stesso cliente e anno su Orchestrator: nessuna coordinazione manuale, nessun conflitto. È però un
**default proposto, non imposto**: il nome del tag resta una scelta del dev.

La descrizione del tag non è prosa libera: ogni macro area si scrive con lo schema **Cosa / Come /
Esiste**, e il «Cosa» cita la fonte — la frase della trascrizione o il riferimento puntuale al
documento. Accanto alle macro aree il tag ospita anche ciò che **non** diventa un ticket:
*Situazioni aperte* (si aspetta un responso da fuori) e *In standby* (manca un prerequisito
nostro).

**Il tag vale da solo.** Dopo la creazione si chiede al dev se vuole procedere con i ticket: sono
due decisioni distinte, e la seconda non è il seguito automatico della prima.

I candidati a ticket si vagliano **uno per volta**, ciascuno presentato in forma autoconsistente, e
**nessun ticket si crea senza un sì esplicito** su quel singolo candidato. La descrizione del tag si
aggiorna una volta sola a fine sessione (`tag-update`): durante il vaglio i rimandi ai ticket non
sono ancora noti.

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
- **Il vaglio uno per volta e il sì esplicito** (2026-09-16): il mapping approvato in blocco faceva
  passare ticket che il dev non aveva davvero letto, e i candidati presentati con rimandi al testo
  precedente («quello del custode») erano illeggibili fuori dal momento in cui erano stati scritti.
- **Situazioni aperte e standby stanno nel tag, non nei ticket** (2026-09-16): un punto bloccato da
  una risposta esterna aperto come ticket resta fermo in board e sparisce dal contesto della call.
