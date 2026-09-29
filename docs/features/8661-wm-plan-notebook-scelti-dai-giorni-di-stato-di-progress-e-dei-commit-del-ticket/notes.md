> Ticket: oc:8661

# Notes — wm-plan: notebook scelti dai giorni di stato, di progress e dei commit del ticket

## Deviazioni dal piano

- **Nessun branch, nessun commit durante il lavoro:** il `CLAUDE.md` del repo lo vieta e prevale
  sulla fase `execution: branch` di `wm-plan`. Si lavora su `develop`; il commit lo fa il dev.
- **Task 4, howto:** il piano limitava la modifica alla riga 28, la spec chiede che tutte le prove
  passino un elenco di giorni: corrette anche le prove 2 e 3 (trovato dalla review finale).

## Bug trovati

Dalla review finale (revisore senza contesto della conversazione):

- **Il timeout di `git fetch` non teneva** con una rete che scarta i pacchetti: ucciso `git`, il
  processo `ssh` figlio tiene aperta la pipe e `cmd.Output()` aspetta che esca da solo. Misurato
  1m15s per repository verso `10.255.255.1` con timeout di 3 secondi. Corretto con
  `cmd.WaitDelay`, `ConnectTimeout=10` in `GIT_SSH_COMMAND` e il messaggio «tempo scaduto»: ora 5
  secondi. Test `TestStoryDaysFetchBloccatoRispettaIlTimeout`.
- **Submodule non inizializzato senza avviso:** `git submodule foreach` lo salta, e i suoi commit
  mancavano in silenzio (clone senza `--recurse-submodules`). Ora `git submodule status` lo trova e
  il tool lo nomina in `warnings`. Test `TestStoryDaysSubmoduleNonInizializzatoDaUnWarning`.
- **Frase di `reverse-interaction` poco leggibile:** la chiamata a `get_story_days` stava in un
  inciso dentro l'elenco di cosa passare all'agente; ora è un passo a sé, prima.

## Decisioni

- **Tag associati a oc:8661:** `claude-marketplace` (648, esistente, caratteristica repository);
  `wm-plan` (688, creato il 29/09/2026 come etichetta, caratteristica contenuto).

## Follow-up

Minori rimandati dalla review finale:

- con l'endpoint in 404 o 401 il tool scrive due warning con lo stesso messaggio;
- fuori da un repository `wm-plan` potrebbe richiamare il tool con `repos` esplicito;
- manca un test per il 401 (stesso ramo del 404);
- `git fetch` aggiorna i riferimenti dei branch remoti: non scrive sul lavoro locale, ma non è
  «solo lettura» in senso stretto come dice l'overview;
- `git log --all` comprende `refs/stash`: uno stash su un commit del ticket aggiunge un giorno;
- `GIT_SSH_COMMAND` prevale su `core.sshCommand`: chi ha un comando ssh proprio avrà il `fetch` in
  warning.

**Prova del tool dopo il riavvio (29/09):** `get_story_days` su oc:8636 (repo `orchestrator`) →
23, 28, 29/09: creazione e commit del 28 e 29, confermati con `git log`; su oc:8543 (repo
`forestas`) → 14, 16, 21, 22, 23, 24 e 29/09, 7 giorni contro i 12 della finestra, con i commit
del 21–24 del submodule `wm-package`; su oc:8661 → solo oggi. Nessun warning. Con `repos` esplicito
il tool non entra nei submodule dei repository indicati: li aggiunge solo senza `repos`, che è il
caso di `wm-plan`.

**Prova di `wm-plan` su oc:8543 in `forestas` (29/09): superata.** `wm-plan` ha chiamato
`get_story_days` e ha passato all'agente i 7 giorni del ticket come elenco, al posto della
finestra; al dev ha mostrato i link dei notebook di quei giorni (non l'elenco delle date in chiaro),
nessun warning. La prima domanda nasce dalle call del 16/09 e del 21/09, citazioni verificate su
Drive. Differenza dalla skill: ha passato `repos` esplicito (`forestas` e `forestas/wm-package`)
invece di ometterlo; stesso risultato.

- Follow-up, fuori da questo lavoro: per le call del 22/09, del 23/09 e del tag l'agente ha
  dichiarato citazioni «non verificabili» (regola sul `cited_text`): da capire se NotebookLM non ha
  dato riferimenti o se la seconda domanda sul punto non è partita.
