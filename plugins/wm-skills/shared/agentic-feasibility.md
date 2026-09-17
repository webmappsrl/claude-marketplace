# Quali fasi si delegano, e perché

## Il criterio

Una fase si delega quando ha **input piccolo, output piccolo, lavoro intermedio grande**
e non richiede interazione diretta con il dev.

Due controlli prima di dire sì:
1. **L'agente può concludere da solo?** Se la conclusione dipende da qualcosa che esiste
   solo nel dialogo (una decisione presa, una responsabilità assunta), no.
2. **Cosa succede se sbaglia e nessuno se ne accorge?** Se alimenta una decisione di
   sicurezza o un registro di responsabilità, no.

## wm-plan

| Fase | Delegabile | Motivo |
|---|---|---|
| `ticket` | no | dialogo con il dev, scritture su Orchestrator da confermare |
| `environment-setup` | **sì** (`wm-env-detect`) | molti comandi, esito = pochi flag |
| `docker-check` | no | effetti collaterali sul sistema del dev (`docker compose stop`) |
| `init-context` | no | il `CLAUDE.md` serve tutto e per tutto il workflow: delegarlo = riassumerlo |
| `reverse-interaction` | **sì, la ricerca** (`wm-codebase-research`) | il dialogo resta nel principale, la ricerca no |
| `overview` | no | nasce dal dialogo, la scrive chi l'ha condotto (oc:8282) |
| `challenge` | **già delegata, cieca** | indipendenza di giudizio |
| `estimation` | **sì, cieca** (`wm-estimate`) | chi ha scritto l'overview stima ottimista in modo sistematico |
| `write-plan` | no | delega già a `superpowers:writing-plans`; il piano torna comunque nel principale |
| `execution: implementation` | già delegata | a `superpowers` |
| `review-gate: subagent` | **già delegata, cieca** | indipendenza di giudizio |
| `review-gate: phpstan-check` | **no** | hard-block di sicurezza (oc:8341): un agente in mezzo può trasformare un fallimento in un pass silenzioso |
| `notes` | no | registra decisioni e responsabilità che esistono solo nel dialogo |
| `update-context` | **sì, il controllo** (`wm-context-guard`) | il testo lo scrive il principale, la verifica di forma no |

## wm-tag

| Fase | Delegabile | Motivo |
|---|---|---|
| `input` | no | chiede materiale al dev quando manca — dialogo diretto |
| `repo-map` | delegabile in futuro, nessun agente attuale | meccanica (find + jq), ma il compito non coincide con lo scope dichiarato di `wm-env-detect` (rilevamento ambiente per `wm-plan`, non discovery di repo generici); in caso di dubbio non si stira lo scope di un agente esistente |
| `client-extraction` | no | propone un nome e **attende conferma esplicita** del dev — è un giudizio che nasce nel dialogo |
| `tag-naming` | no | il calcolo di N è minimo (una `list_tags` + un conteggio), non c'è lavoro intermedio grande da spostare fuori; la fase termina comunque con una proposta da confermare |
| `tag-description` | **sì, la voce «Esiste»** (`wm-codebase-research`) | il testo sorgente (trascrizione) è già stato letto nel principale in `input`: non c'è un input piccolo da passare a un agente senza duplicare la lettura; la fase si chiude con approvazione esplicita del dev, incluse correzioni iterative. Fa eccezione la voce «Esiste» di ogni macro area, che è una domanda sul codice e si delega |
| `tag-creation` | no | scrittura su Orchestrator con preview e conferma esplicita — stessa regola delle scritture in `wm-plan: ticket` |
| `tag-resume` | **sì, la verifica delle voci in standby** (`wm-codebase-research`) | il giro di verifica sui blocchi: «il prerequisito esiste ora nel repo?» è una domanda sul codice, con input piccolo (le voci in standby lette dal tag) e risposta breve, quindi si delega in una chiamata sola. Il resto — le situazioni aperte, che si sbloccano solo con una risposta che porta il dev — è dialogo |
| `candidate-list` | no | termina con una richiesta di feedback esplicita al dev; il materiale su cui lavora è lo stesso testo già in context, non un input piccolo isolabile |
| `candidate-review` | no | è il vaglio uno per volta: ogni candidato si chiude con un sì o un no del dev, e senza quel sì non si crea nulla — è dialogo puro |
| `ticket-loop` | no | orchestrazione del dialogo (annuncio, invocazione di `wm-plan`) — non è delegabile un ciclo che deve poter dialogare col dev ad ogni passo |
| `tag-update` | no | scrittura su Orchestrator con preview e conferma esplicita, sulla descrizione mostrata per intero al dev |

## wm-review-ticket

| Fase | Delegabile | Motivo |
|---|---|---|
| `Fase 1 — Lettura ticket` | no | mostra un riepilogo al dev e, se manca l'URL PR, chiede materiale mancante — dialogo diretto |
| `Fase 2 — Recupero contratto artefatti` | no | una singola `WebFetch` con estrazione di una sezione: non c'è lavoro intermedio grande da spostare, delegarla costerebbe più del beneficio |
| `Fase 3a — Identifica il repo` | no | se il repo non è riconosciuto nel `CLAUDE.md`, chiede il path al dev — dialogo condizionale |
| `Fase 3b — Stash se dirty` | no | operazione meccanica minima (un `git stash` condizionale), nessun lavoro intermedio da isolare |
| `Fase 3c — Checkout branch PR` | delegabile in futuro, nessun agente attuale | tre tentativi con fallback, ma l'ultimo fallback può richiedere input dal dev (hash o nome branch corretto) — non c'è oggi un agente per operazioni di checkout con fallback dialogico, e forzare `wm-env-detect` (scope: rilevamento ambiente di `wm-plan`) sarebbe uno stiramento |
| `Fase 4 — Lettura contesto wm-plan` | **sì, informata** (`wm-codebase-research`) | input piccolo (path del repo, slug del ticket), lavoro intermedio grande (leggere `overview.md`, `plan.md`, `notes.md`, potenzialmente lunghi), output piccolo (sintesi di cosa deve fare il codice e cosa è stato pianificato); nessuna decisione del dev in questa fase, solo raccolta di contesto per la review vera e propria |
| `Fase 5a — Ottieni il diff` | no | uno o due comandi `git diff`/`git show`, nessun lavoro intermedio da isolare |
| `Fase 5b — Finder paralleli` | **già delegata, cieca** (5 finder ad-hoc) | indipendenza di giudizio, stesso pattern di `challenge`/`review-gate` in `wm-plan` — non è però uno dei cinque agenti nominati: sono subagenti dedicati definiti dentro la skill stessa |
| `Fase 5c — Dedup e verifica` | delegabile in futuro, nessun agente attuale | input medio (i candidati dei 5 finder), output piccolo (lista deduplicata e verificata), ma nessuno dei cinque agenti disponibili copre questo compito — inventare una delega qui violerebbe il vincolo sugli agenti esistenti |
| `Fase 5d — Output` | no | il verdetto va presentato e discusso col dev nella stessa conversazione — non è un risultato da restituire "solo conclusioni" a un principale che poi lo rigira as-is, è già l'interfaccia col dev |
| `Fase 6a — Ripristina stash` | no | operazione meccanica minima, simmetrica a 3b |
| `Fase 6b — Aggiorna ticket Orchestrator` | no | scrittura con preview e conferma esplicita che alimenta lo status del ticket (registro di stato/responsabilità) — stessa regola delle scritture in `wm-plan` |

## Classificare una fase nuova

Applicare il criterio e i due controlli. In caso di dubbio, **non delegare**: il costo di
una fase non delegata è context; il costo di una delega sbagliata è un errore invisibile
a valle.
