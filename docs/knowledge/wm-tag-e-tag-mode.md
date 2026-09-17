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
aggiorna una volta sola a fine sessione (`tag-update`), per non chiedere al dev una preview e una
conferma a ogni candidato vagliato.

**Rilanciare `wm-tag` su un tag esistente è un lavoro diverso dal primo giro**: non riprende un
elenco lasciato a metà, ma verifica se i blocchi sono caduti. Le voci *In standby* si controllano
nel codice (`wm-codebase-research`), quelle in *Situazioni aperte* si chiedono al dev, perché la
risposta viene da fuori. Quelle sbloccate tornano candidati e passano dallo stesso vaglio. Un punto
che ha già un ticket non è più un candidato: le informazioni nuove si portano sul ticket.

L'alternativa — aprire subito un ticket per ogni punto bloccato e lasciarlo in backlog — è stata
scartata: un ticket in backlog con informazioni che nessuno aggiorna è peggio dell'assenza del
ticket, perché chi lo prende in mano mesi dopo non sa che la descrizione è ferma.

**I ticket aperti dal tag non si elencano nella descrizione**: l'associazione è una relazione vera
su Orchestrator, che la pagina del tag mostra già. Una tabella scritta a mano sarebbe una seconda
copia dello stesso dato, e divergerebbe alla prima associazione fatta fuori dalla skill.

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
- **La tabella dei ticket nella descrizione del tag** (2026-09-17, superata): la prima stesura del
  vaglio la prevedeva, per ritrovare dal tag i ticket che ne erano nati. Caduta perché quel dato
  Orchestrator ce l'ha già come relazione e lo mostra nella pagina del tag: la tabella era una
  copia mantenuta a mano, cieca a ogni associazione fatta dall'interfaccia.
