# Installare NotebookLM per wm-skills

`wm-transcript-research` fa leggere le call a NotebookLM tramite il server MCP `notebooklm`, che il
plugin `wm-skills` dichiara da sé. Sulla macchina serve il programma che lo esegue, e un accesso con
l'account di lavoro.

1. Installa il client (una volta):

   ```bash
   uv tool install notebooklm-mcp-cli
   ```

   Verifica che sia nel `PATH` con cui parte Claude Code: `which notebooklm-mcp`.

2. Accedi con l'account **`@webmapp.it`** (le call stanno nel Drive di lavoro):

   ```
   ! nlm login
   ```

   L'uscita deve riportare `Account: <nome>@webmapp.it`. Se riporta un account personale, rifai
   l'accesso scegliendo quello di lavoro nel browser.

3. Se avevi già configurato NotebookLM a mano (per esempio un server `notebooklm-mcp` in
   `~/.claude.json`), toglilo: il plugin dichiara il suo.

4. Riavvia Claude Code e controlla con `/mcp` che `plugin:wm-skills:notebooklm` risulti collegato.

**Dopo una modifica al `.mcp.json` del plugin** (per chi lavora sul repo `claude-marketplace`): i
server MCP si leggono dalla copia installata del plugin, non dal repo. Reinstalla il plugin
(`/plugin uninstall wm-skills@wm-marketplace` e `/plugin install wm-skills@wm-marketplace`) e
riavvia Claude Code.

**Quando le credenziali scadono** `wm-plan` avvisa «ricerca sulle trascrizioni non disponibile» e
indica `! nlm login`: rifallo, non ignorare l'avviso.
