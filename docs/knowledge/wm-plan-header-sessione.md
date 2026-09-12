# Header di sessione di wm-plan

## Stato attuale

`wm-plan` apre ogni invocazione con un banner ASCII seguito da poche righe di stato
(`### header: versione`, `### header: diagramma`, `### header: context`). L'header è puramente
informativo: non attiva e non disattiva nulla.

**La versione installata è un valore statico scritto dentro `SKILL.md`** (`**Versione
installata:** v<version>`), non letto a runtime da `plugin.json`, dalla cache dei plugin o da
git. Si aggiorna a mano ad ogni release — è un passo della procedura in
[docs/howto/rilascio-wm-skills.md](../howto/rilascio-wm-skills.md). Viaggiando col resto della
skill, resta allineato ad ogni `/plugin marketplace update` senza risoluzione di path.

**Il check di aggiornamento disponibile** confronta l'hash di HEAD del repo installato con
l'ultimo commit su `main` via API GitHub. Il repo si individua a partire dalla cache dei plugin
(`find ~/.claude/plugins/cache … -path '*/wm-skills/*/skills/wm-plan/SKILL.md'`), quindi
funziona da qualsiasi cwd. Se il branch non è `main` o `SKILL.md` ha modifiche non committate,
il check è sostituito da `🔧 modalità sviluppo locale`: mai disattivato in silenzio.

**Il link al diagramma è un URL GitHub Pages fisso**
(`https://webmappsrl.github.io/claude-marketplace/wm-plan-diagramma/`). Non c'è nessun fetch da
fare per leggerlo, nessun redeploy da eseguire e nessun URL da riscrivere: la pagina si
ripubblica dal push che ne modifica il sorgente. I vincoli sul sorgente stanno in
`.claude/rules/wm-plan-diagramma.md`.

## Come ci siamo arrivati

- **L'header doveva comparire solo alla prima invocazione di sessione** (oc:8283): nessun
  meccanismo di stato, solo una regola in prosa, perché le skill non hanno un tracciamento di
  sessione garantito. Caduta: la regola comportamentale non teneva, e `SKILL.md` oggi impone
  l'header **ad ogni** invocazione, senza eccezioni.
- **Nessun file di stato locale per il check versione** (oc:8283): sarebbe stato sovrascritto ad
  ogni `marketplace update` e partirebbe senza baseline su una nuova installazione. Il confronto
  fra hash è rimasto.
- **Il path della skill si risolveva relativo alla cwd** (`realpath plugins/wm-skills/…`):
  funzionava solo invocando `wm-plan` da dentro `claude-marketplace`, da ogni altro repo falliva
  in silenzio e il check finiva sempre in `⚠️ Check versione non disponibile`. Caduta a favore
  della ricerca nella cache dei plugin.
- **Il diagramma era un Artifact Claude, con URL versionato nel `CLAUDE.md`** (oc:8283) e un
  redeploy manuale dopo ogni modifica al repo. Due meccaniche si sono succedute per leggere quel
  link, entrambe cadute quando l'Artifact è stato sostituito da GitHub Pages:
  - **fetch remoto del `CLAUDE.md` di `claude-marketplace` via `curl -sf`** su
    `raw.githubusercontent.com` — introdotto perché leggere il `CLAUDE.md` del repo *target*
    dava un falso "non ancora pubblicato" quando `wm-plan` girava altrove; caduto perché faceva
    dipendere l'header dalla rete;
  - **URL statico scritto in `SKILL.md`** — caduto nella sua motivazione originale (un Artifact
    ripubblicato a mano richiedeva di riscrivere l'URL a ogni cambio). L'URL sta ancora in
    `SKILL.md`, ma ora è un indirizzo stabile che nessuno riscrive.
- **Il redeploy dell'Artifact scattava su qualsiasi modifica al repo**, non solo su quelle a
  `wm-plan` (oc:8283, scelta esplicita del dev), e si tentava senza check preventivo
  dell'account, perché due account Claude possono condividere la stessa email e non esiste un
  segnale affidabile per distinguerli: l'unico modo verificato era il fallimento del redeploy.
  Tutto questo è caduto con il passaggio a GitHub Pages.
- **Il drift fra diagramma pubblicato e stato reale di `SKILL.md`** era accettato come rischio
  consapevole (oc:8283, deciso in Fase: challenge). Oggi è coperto in parte da
  `.github/scripts/verifica-diagramma.sh`, che confronta però solo i nomi delle fasi.
