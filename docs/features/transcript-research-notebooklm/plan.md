# wm-transcript-research su NotebookLM — piano di implementazione

> **Per chi esegue:** sotto-skill `superpowers:executing-plans`, task per task. I passi usano le
> checkbox (`- [ ]`).
>
> **In questo repo i commit sono vietati durante l'esecuzione.** Niente `git add`, `git commit`,
> `git push`, niente branch. I blocchi «Commit» sono istruzioni per il dev, che committa dopo aver
> letto il diff.

**Obiettivo:** far leggere le trascrizioni a NotebookLM, organizzato in un notebook per giorno di
scrum e uno per tag, con citazioni attribuite dal sistema e verificate, per `wm-plan` e `wm-tag`.

**Architettura:** nessun codice. Il plugin dichiara il server `notebooklm` (client
`notebooklm-mcp-cli`). `wm-transcript-research` gestisce i notebook `scrum AAAA-MM-GG` e
`tag <nome del tag>` nell'account NotebookLM del dev, applica il template del team alla creazione,
interroga i notebook che servono in parallelo e mette insieme le risposte in ordine di data.
`wm-plan` prepara il notebook del giorno all'avvio, lo usa nel caso B e passa finestra e tag in
`reverse-interaction`; `wm-plan` e `wm-tag` verificano le citazioni con una regola condivisa.

**Tecnologie:** prompt Markdown di Claude Code, connettore Google Drive (`search_files`), server MCP
`notebooklm-mcp`.

**Spec:** [overview.md](overview.md)

## Vincoli globali

- Italiano corrente, senza modi di dire inglesi tradotti; termini tecnici in inglese.
- Nessuna alternativa scartata nei documenti del lavoro.
- Frontmatter degli agenti: `name`, `description`, `model`, `tools`.
- Nomi dei tool NotebookLM: `mcp__plugin_wm-skills_notebooklm__<tool>`.
- Nomi dei notebook: `scrum AAAA-MM-GG`, `tag <nome del tag>`.
- Fogli Google, siti web e GitHub non diventano fonti.
- Il template detta regole, **mai un formato**.
- Nessuna cancellazione di notebook senza la conferma del dev.
- Finestra provvisoria, fino a oc:8636: dai giorni prima della creazione del ticket a oggi, solo i
  giorni con almeno una call.
- Prima di dichiarare pronto: `claude plugin validate .` e `./.github/scripts/verifica-diagramma.sh`.
- Agenti e skill si caricano all'avvio della sessione; i server MCP del plugin dalla copia installata
  (reinstallare dopo una modifica al `.mcp.json`). Le prove si fanno aprendo la sessione nel repo del
  prodotto, non in `claude-marketplace` (claude-mem vi inietta il contesto del lavoro sulle skill).

## Punti da controllare in review

1. **Attribuzione da memoria:** titolo e id Drive di una citazione devono venire da
   `source_list_drive`, la data dal notebook. Controllo: prova 2.
2. **Elenco di Drive fermo alla prima pagina:** tutte le pagine, `pageSize: 100`. Controllo: prova 2,
   `Copertura` con le call dell'elenco accanto a quelle caricate per ogni notebook.
3. **Merge fra giorni:** una decisione ripresa o ribaltata in un giorno successivo va segnalata.
   Controllo: prova 2 (15/09 e 16/09).
4. **Domande una per chiamata:** tutte in una domanda per notebook, notebook in parallelo con
   `notebook_query_start` e `notebook_query_status` (le chiamate bloccanti allo stesso server passano
   una alla volta).
   Controllo sulle tool call delle prove.
5. **«Non determinabile» senza seconda domanda.** Controllo: prova 1.

---

### Task 1: server `notebooklm` nel plugin

**File:** `plugins/wm-skills/.mcp.json`

- [x] Aggiungere `"notebooklm": { "command": "notebooklm-mcp" }` accanto a `orchestrator`.
- [x] **Controllo:** `jq -c '.mcpServers | keys' plugins/wm-skills/.mcp.json` → `["notebooklm","orchestrator"]`.

---

### Task 2: regola condivisa di verifica delle citazioni

**File:** `plugins/wm-skills/shared/verifica-citazioni.md` (nuovo)

- [x] Procedura: frase di cinque-otto parole da una sola battuta; `search_files` con
  `fullText contains '"<frase>"'` ed `excludeContentSnippets: true` (con `parentId` della cartella
  per lo scrum, senza per le fonti fuori cartella); apostrofo scritto `\'`; quattro esiti
  (attribuita bene, call sbagliata, verifica non possibile per le call di oggi, non trovata).
- [x] **Controllo:** `grep -c 'Invalid query' plugins/wm-skills/shared/verifica-citazioni.md` → `1`.

---

### Task 3: agente `wm-transcript-research`

**File:** `plugins/wm-skills/agents/wm-transcript-research.md` (riscritto)

- [x] Frontmatter: `search_files` di Drive, `notebook_list`, `notebook_create`, `chat_configure`,
  `source_add`, `source_list_drive`, `notebook_query_start`, `notebook_query_status`; niente `read_file_content`, niente `Bash`.
- [x] `## I notebook`: `scrum AAAA-MM-GG` e `tag <nome del tag>`, trovati con `notebook_list`, creati
  se mancano, template (blocchi **Base** e **Parte tag**) applicato alla creazione; pulizia dei
  notebook `scrum …` con più di 90 giorni solo segnalata.
- [x] `## Cosa ti chiedono`: richieste **prepara**, **domande su una finestra** (più i tag),
  **domande sul giorno di oggi**; la finestra comprende sempre oggi.
- [x] `## Le fonti`: call di un giorno elencate per data nel titolo, `pageSize: 100`, tutte le
  pagine; un giorno senza call non ha notebook; si caricano tutte le call del giorno; fonti dei tag
  per id; notebook del tag creato solo con «crea se manca», altrimenti saltato e dichiarato; Fogli,
  siti, GitHub esclusi e dichiarati.
- [x] `## Le domande`: tutte le domande in una chiamata per notebook, notebook in parallelo, seconda
  chiamata solo per i punti scoperti; domanda più stretta se la risposta supera il limite; seconda
  domanda prima di un «non determinabile»; merge in ordine di data con i ribaltamenti segnalati.
- [x] `## Le attribuzioni`: `source_list_drive` su ogni notebook interrogato, controllo sul
  `cited_text`, seconda domanda sul punto quando una citazione non combacia, battute interrotte
  segnalate.
- [x] Formato: `Copertura` per notebook (call dell'elenco e caricate), link dei notebook, notebook
  vecchi e doppi (dopo ogni creazione si rilegge l'elenco e si usa il più vecchio); `RICERCA FALLITA`
  con `! nlm login`.
- [x] **Controllo:**

```bash
f=plugins/wm-skills/agents/wm-transcript-research.md
grep -c 'read_file_content\|Bash' $f          # atteso 0
grep -c 'scrum AAAA-MM-GG' $f                  # atteso almeno 2
grep -c 'nextPageToken' $f                     # atteso almeno 1
grep -c 'wm-plan <slug>\|wm-transcript <' $f   # atteso 0
```

---

### Task 4: `wm-plan`

**File:** `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [x] `## Fase: ticket`: dopo `planning_start_at`, `wm-transcript-research` in background con
  «prepara»; le altre richieste all'agente aspettano la sua risposta; avviso evidente se fallisce.
- [x] `### ticket: caso-b`: dopo la descrizione del dev, domanda al notebook del giorno prima della
  bozza, citazioni verificate; altri giorni o tag solo su richiesta.
- [x] `reverse-interaction`: all'agente numero e titolo del ticket, finestra di giorni, domande, e
  per ogni tag il nome e gli id dei Google Doc (estratti con `jq` se `get_tag` finisce su file),
  notebook del tag solo se esiste; senza ticket la finestra è oggi; domande dirette su altri giorni o
  tag con «crea se manca»; tag senza notebook segnalati al dev; verifica di ogni citazione;
  link dei notebook al dev; `Copertura` per notebook; avviso evidente su `RICERCA FALLITA`.
- [x] `## Checklist di completamento`: proposta di cancellare i notebook del giorno segnalati come
  vecchi o doppi, con conferma.
- [x] **Controllo:** `grep -n 'prepara\|verifica-citazioni.md\|Notebook vecchi' plugins/wm-skills/skills/wm-plan/SKILL.md`
  → almeno tre righe; `./.github/scripts/verifica-diagramma.sh` → allineato.

---

### Task 5: `wm-tag`

**File:** `plugins/wm-skills/skills/wm-tag/SKILL.md`

- [x] Riga `**Notebook:**` nella struttura della descrizione del tag.
- [x] `## Fase: tag-creation`: verifica delle citazioni del «Cosa», poi notebook `tag <nome del tag>`
  tramite l'agente, link nella riga `**Notebook:**`; se l'agente fallisce, tag senza quella riga.
- [x] `## Fase: candidate-review`: domande sulla call tramite l'agente sul notebook del tag.
- [x] **Controllo:** `grep -n 'tag <nome del tag>\|verifica-citazioni.md' plugins/wm-skills/skills/wm-tag/SKILL.md`
  → almeno due righe.

---

### Task 6: documentazione

**File:** `docs/howto/installare-notebooklm.md` (nuovo), `docs/howto/provare-wm-transcript-research.md`,
`docs/knowledge/wm-transcript-research.md`, `CLAUDE.md`

- [x] Howto di installazione: `uv tool install notebooklm-mcp-cli`, `nlm login` con l'account
  `@webmapp.it`, rimozione del server configurato a mano, controllo con `/mcp`.
- [x] Howto delle prove: prove 1–4 con la risposta attesa; la prova 3 attende i notebook `scrum …`
  della finestra e il notebook del tag 678.
- [x] Conoscenza: chi legge le call, template, attribuzioni, `wm-tag`; in «Perché così» NotebookLM e
  i notebook per giorno e per tag; «Connettore Drive, non un MCP» fra le scelte superate.
- [x] `CLAUDE.md`: riga della procedura di installazione, riga dei coupling con i nomi dei notebook,
  riga dell'indice della conoscenza.
- [x] Conoscenza: `wm-skills-delega-agentica.md` (dove si usa l'agente) e `wm-tag-e-tag-mode.md`
  (verifica delle citazioni e notebook del tag); tabella delle deleghe in `wm-plan/SKILL.md`.
- [x] Diagramma `docs/guide/wm-plan-diagramma/index.html`, solo contenuto: nodo dell'agente, freccia
  da `Fase: ticket`, paragrafi di `Fase: ticket` e `reverse-interaction`.
  **Controllo:** `./.github/scripts/verifica-diagramma.sh` → allineato.
- [x] Howto di installazione: aggiungere che dopo una modifica al `.mcp.json` il plugin va
  reinstallato (`/plugin uninstall` e `/plugin install`) e Claude Code riavviato.

---

### Task 7: prove con l'agente vero

- [x] Riavviare Claude Code, aprendo la sessione nel repo del prodotto.
- [x] Prova 2 (caso del 23/09): notebook `scrum 2026-09-14` … `scrum 2026-09-18`, citazioni del
  15/09 e del 16/09 attribuite e verificate su Drive, merge in ordine di data.
- [x] Prova 3 (oc:8543), come domanda diretta con «crea se manca»: notebook `scrum …` dei giorni con
  call dal 2026/09/12 a oggi più `tag [CALL][FORESTAS][2026] excel registro sentieri` creato; Saba
  `[deciso]`. Superata nel contenuto; ha rivelato i doppioni (corretti).
- [x] Ripetere prova 2 dopo il riavvio: Piccioli al 15/09 recuperato dalla seconda domanda sul
  `cited_text`, nessun doppione nuovo.
- [x] Prova 1: «non determinabile» dopo la seconda domanda.
- [x] Dopo il riavvio, una ricerca su più notebook: le domande partono tutte con
  `notebook_query_start` e il tempo delle domande resta quello di una sola.
- [x] `wm-plan` su oc:8543 fino alla prima domanda di `reverse-interaction`: notebook del giorno
  preparato all'avvio, verifica delle citazioni, link dei notebook.
- [x] Esiti nel howto, divergenze in [notes.md](notes.md).

---

**Commit (istruzione per il dev), un solo commit a lavoro finito:**

```bash
git add plugins/wm-skills/.mcp.json plugins/wm-skills/shared/verifica-citazioni.md \
  plugins/wm-skills/agents/wm-transcript-research.md plugins/wm-skills/skills/wm-plan/SKILL.md \
  plugins/wm-skills/skills/wm-tag/SKILL.md docs/howto/installare-notebooklm.md \
  docs/howto/provare-wm-transcript-research.md docs/knowledge/wm-transcript-research.md CLAUDE.md \
  docs/knowledge/wm-skills-delega-agentica.md docs/knowledge/wm-tag-e-tag-mode.md \
  docs/guide/wm-plan-diagramma/index.html \
  docs/features/transcript-research-notebooklm/
git commit -m "feat(transcript-research-notebooklm): wm-transcript-research legge le call tramite NotebookLM"
```
