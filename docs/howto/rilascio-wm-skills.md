# Rilasciare una versione di wm-skills

I passi marcati **(dev)** non li esegue Claude: la prima regola del `CLAUDE.md` vieta
all'agente `git commit`, `git add`, `git push` e la creazione di branch. Claude prepara le
modifiche ai file e si ferma; commit, tag e push li fa il dev.

Da eseguire in quest'ordine.

1. Bump di `version` in `plugins/wm-skills/.claude-plugin/plugin.json`. Semver: **patch** per un
   fix, **minor** per una skill o una feature retro-compatibile, **major** per un breaking change
   nel contratto degli artefatti o nel nome di una skill.
2. Allinea la riga `**Versione installata:** v<version>` in
   `plugins/wm-skills/skills/wm-plan/SKILL.md` → `### header: versione` allo stesso valore.
3. Allinea la costante `Version` in `plugins/wm-skills/mcp/internal/version/version.go` e
   ricompila con `plugins/wm-skills/mcp/build.sh`: il binario è versionato nel repo e viaggia nel
   plugin, quindi va aggiornato ad ogni rilascio.
4. `claude plugin validate .` — deve passare prima di dichiarare il lavoro pronto.
5. **(dev)** Commit dei file toccati su `main`.
6. **(dev)** `git tag v<version> && git push origin v<version>`.

Il binario compilato nel repo è per `darwin/arm64`: su Linux o Mac Intel va ricompilato
cambiando `GOOS`/`GOARCH`, altrimenti il server MCP non parte.
