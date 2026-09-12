# Delega ad agenti nelle skill wm-skills

## Stato attuale

Le skill non leggono più tutto nel context della sessione principale: le fasi che macinano molto
testo per restituire poco sono affidate ad **agenti** in `plugins/wm-skills/agents/`, che leggono
nel proprio context e restituiscono solo conclusioni.

| Agente | Fase | Cosa restituisce |
|---|---|---|
| `wm-codebase-research` | `reverse-interaction` | risposte con **prova verbatim**: percorso, righe, estratto |
| `wm-env-detect` | `environment-setup` | i flag d'ambiente già risolti, con contratto nominale |
| `wm-estimate` | `estimation` | la stima, **cieco**: legge solo overview e piano |
| `wm-context-guard` | `update-context` | rilievi di forma sul `CLAUDE.md`, mai sul merito |
| `wm-context-doctor` | invocato dal dev | un piano di riordino del `CLAUDE.md` |

**Due tipi di delega, da non unificare.** *Cieca*, per indipendenza di giudizio (`challenge`,
`review-gate`, `wm-estimate`): riceve solo percorsi, mai un riassunto della conversazione.
*Informata*, per volume di lettura (gli altri): riceve il contesto utile al compito. Passare
contesto a un agente cieco ne distrugge la ragione d'essere.

**Le deleghe sono statiche per fase**, decise in ogni skill: nessuna soglia, nessuna attivazione
automatica. La misura del context resta solo informativa nell'header.

**Con gli agenti si dialoga.** Se il dev contesta l'esito di un agente che produce un giudizio,
le obiezioni tornano **alla stessa sessione** di quell'agente — che ha ancora in context il
materiale letto — mai a un agente nuovo. Tre vie sempre: accettare, sostituire, far rifare.
Massimo due giri. Mai per `challenge` e `review-gate`: un revisore che si corregge quando il
criticato ribatte è l'opposto di ciò che serve.

**Nessun agente scrive contenuto.** `notes.md` e il `CLAUDE.md` registrano decisioni e
responsabilità che esistono solo nel dialogo: un agente arrivato alla fine non ha assistito ai
fatti. Il meccanismo completo — contratto di ritorno, tetto all'output, fallback — è in
`plugins/wm-skills/shared/agent-delegation.md`.

Proviene da oc:8527.

## Come ci siamo arrivati

- **Attivare la delega oltre una soglia di context** (40% attenzione, 65% automatica) — caduta:
  il transcript espone i token occupati ma non la dimensione della finestra, quindi ogni
  percentuale avrebbe avuto un denominatore indovinato, e il modello può cambiare a metà
  sessione.
- **Delegare anche `phpstan-check`** — caduta: è un hard-block di sicurezza, e un agente che
  fallisce restituendo «nessun errore» lo trasformerebbe in un pass silenzioso. Il comando resta
  nel principale, con l'output filtrato.
- **Un agente «redattore» che scrivesse `notes.md` e il `CLAUDE.md`** — caduta per la stessa
  ragione per cui `overview` non è delegata: quel materiale esiste solo nel dialogo.
- **Garanzia «propone, non esegue» ottenuta togliendo `Write` dai tool** — falsa: con `Bash` si
  scrive comunque, e `Bash` serve per le misure. È diventata un divieto comportamentale
  esplicito nel prompt, con la protezione affidata al permesso chiesto all'utente.
