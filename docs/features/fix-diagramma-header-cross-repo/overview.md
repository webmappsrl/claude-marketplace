# Fix: lettura cross-repo del link Artifact diagramma nell'header wm-plan

## Cosa cambia
La sotto-fase `### header: diagramma` di `wm-plan` non legge più la sezione `## Diagramma di flusso wm-plan` dal `CLAUDE.md` del repo target (dove `wm-plan` viene invocato), ma sempre dal `CLAUDE.md` del repo `webmappsrl/claude-marketplace`, tramite fetch diretto dell'URL raw:

```
https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/CLAUDE.md
```

Da questo contenuto viene estratta la sezione `## Diagramma di flusso wm-plan` e il relativo URL Artifact, con lo stesso formato di output già esistente (`📊 Diagramma di flusso: <URL>`).

## Perché
L'URL Artifact è documentato solo nel `CLAUDE.md` di `claude-marketplace` (repo della skill). Quando `wm-plan` viene invocato da un repo target diverso (es. `camminiditalia`), quel repo non ha (e non deve avere) questa sezione nel proprio `CLAUDE.md`, quindi l'header mostrava erroneamente `📊 Diagramma di flusso: non ancora pubblicato`, facendo pensare che l'Artifact non fosse mai stato pubblicato quando in realtà lo è.

## Requisiti
- [ ] `### header: diagramma` esegue **sempre** fetch remoto di `https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/CLAUDE.md` invece di leggere il `CLAUDE.md` del repo target — anche quando `wm-plan` viene invocato all'interno dello stesso repo `claude-marketplace` (nessun fallback ibrido locale, scelto per semplicità)
- [ ] Estrae la sezione `## Diagramma di flusso wm-plan` e il campo `URL Artifact` da quel contenuto
- [ ] **Fetch HTTP fallito** (rete assente, timeout, status non-2xx): mostra `⚠️ Diagramma di flusso non verificabile` e prosegue senza bloccare (fail-soft, coerente con `### header: versione`)
- [ ] **Fetch riuscito ma sezione/URL assente o non valido nel contenuto**: mostra `📊 Diagramma di flusso: non ancora pubblicato` (comportamento invariato)
- [ ] Se la sezione/URL è presente e valida: mostra `📊 Diagramma di flusso: <URL>` (comportamento invariato)

## Rischi
- Dipendenza aggiuntiva dalla rete per una parte dell'header che prima (quando invocato nello stesso repo `claude-marketplace`) poteva essere letta da filesystem locale — mitigato dal fatto che l'header già dipende dalla rete per il check versione (GitHub API) e per Orchestrator.
- Se il branch `main` di `claude-marketplace` è in ritardo rispetto a un aggiornamento locale dell'URL Artifact (caso raro, l'utente lavora quasi sempre da lì), l'header potrebbe mostrare temporaneamente un URL non aggiornato — accettato come rischio minore, coerente col fatto che l'header stesso già mostra "modalità sviluppo locale" quando ci sono modifiche non committate.

## Out of scope
- Persistenza di un file sorgente HTML dell'Artifact nel repo (proposto ma scartato/non necessario secondo l'utente).
- Modifiche ad altre skill (`wm-tag`, `wm-review-ticket`) — nessuna di esse legge questa sezione.
- Cambiamenti al meccanismo di redeploy dell'Artifact stesso (regola di rigenerazione in `CLAUDE.md`, invariata).

## Moduli toccati
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — sezione `### header: diagramma`
