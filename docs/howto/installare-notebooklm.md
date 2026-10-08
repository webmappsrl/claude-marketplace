# Installare NotebookLM per wm-skills

`wm-transcript-research` fa leggere le call a NotebookLM tramite il server MCP `notebooklm`, che il
plugin `wm-skills` dichiara da sé. Sulla macchina serve il programma che lo esegue, e un accesso con
l'account di lavoro.

## Il modo più semplice: farlo fare a Claude

Incolla questo prompt in Claude Code. Claude fa i passi che può fare da sé, e ti dice di lanciare
tu quelli che richiedono il browser o un comando `/`:

```text
Installa NotebookLM per il plugin wm-skills su questa macchina. Fai tu i passi da shell, e fermati
a dirmi cosa lanciare quando serve una mia azione. In ordine:

1. Controlla se c'è `uv` (`which uv`). Se manca, installalo con `brew install uv`.
2. Installa il client: `uv tool install notebooklm-mcp-cli`. Poi controlla che `which notebooklm-mcp`
   lo trovi; se non lo trova, dimmi come aggiungere la cartella dei tool di uv al PATH.
3. Controlla con `claude mcp list` se c'è un server NotebookLM configurato a mano (per esempio
   `notebooklm-mcp`). Se c'è, toglilo con `claude mcp remove <nome> -s user`: il plugin wm-skills
   dichiara il suo, e quello manuale lo nasconde.
4. Dimmi di lanciare io `nlm login` da un terminale normale, fuori da Claude Code (con `!` la
   domanda su dove salvare le credenziali non riceve risposta): scelgo 1 (Protected) e accedo con
   l'account @webmapp.it nella finestra di Chrome che si apre. Quando ho finito, verifica con
   `nlm notebook list` che l'accesso funzioni davvero e con `nlm login --check` che l'account sia
   @webmapp.it; se è un account personale, fammi rifare l'accesso.
5. Dimmi di aggiornare il plugin con `/plugin marketplace update`, di riavviare Claude Code e di
   controllare con `/mcp` che `plugin:wm-skills:notebooklm` risulti collegato.

Non toccare nient'altro della configurazione di Claude Code.
```

## I passi, se preferisci farli a mano

1. Installa il client (una volta). Serve `uv`: se `which uv` non trova niente, installalo prima
   con `brew install uv`.

   ```bash
   uv tool install notebooklm-mcp-cli
   ```

   Verifica che sia nel `PATH` con cui parte Claude Code: `which notebooklm-mcp`.

2. Accedi con l'account **`@webmapp.it`** (le call stanno nel Drive di lavoro), da un terminale
   normale e non con `!`: il login fa una domanda a cui con `!` non si può rispondere.

   ```bash
   nlm login
   ```

   Alla domanda su dove salvare le credenziali scegli **1 (Protected)**, e accedi nella finestra di
   Chrome che si apre: è un profilo separato dal tuo browser. L'uscita deve riportare
   `Account: <nome>@webmapp.it`, e `nlm notebook list` deve restituire l'elenco dei notebook. Se riporta un account personale, rifai
   l'accesso scegliendo quello di lavoro nel browser.

3. Se avevi già configurato NotebookLM a mano (per esempio un server `notebooklm-mcp` in
   `~/.claude.json`), toglilo: il plugin dichiara il suo, e finché c'è quello manuale il server del
   plugin non compare.

   ```bash
   claude mcp remove notebooklm-mcp -s user
   ```

4. Riavvia Claude Code e controlla con `/mcp` che `plugin:wm-skills:notebooklm` risulti collegato.

**Dopo una modifica al `.mcp.json` del plugin** (per chi lavora sul repo `claude-marketplace`): i
server MCP si leggono dalla copia installata del plugin, non dal repo. Reinstalla il plugin
(`/plugin uninstall wm-skills@wm-marketplace` e `/plugin install wm-skills@wm-marketplace`) e
riavvia Claude Code.

**Quando manca l'installazione o le credenziali scadono** l'agente risponde `RICERCA FALLITA` e
`wm-plan` te lo mostra in modo evidente, con il comando da lanciare (`! nlm login`): rifai
l'accesso, non ignorare l'avviso. Nel frattempo `wm-plan` prosegue senza le call.

**Se `nlm login` dice che l'accesso è valido ma le ricerche falliscono ancora**, segui la
procedura in `plugins/wm-skills/shared/notebooklm-login.md`: è la stessa che `wm-plan` usa per
guidarti. In breve: aggiorna `notebooklm-mcp-cli`, cancella il profilo, rifai il login da un
terminale normale e ricollega il server con `/mcp`.
