# Delega ad agenti — meccanismo condiviso

Letto da: `wm-plan`, `wm-review-ticket`, `wm-tag` e dagli agenti in `plugins/wm-skills/agents/`.
Non duplicare questo contenuto in una skill: citarlo.

## Nessun commit, mai, senza l'approvazione del dev

**Vale sempre e ovunque in questo progetto: nessuna operazione git** — `git commit`, `git add`,
`git push`, `git checkout`, la creazione di un branch — viene eseguita durante un lavoro. Si
scrivono i file e basta.

Non è una regola di una fase di `wm-plan`: vale per ogni skill, ogni agente e ogni operazione
avviata da uno di essi, compresa la riorganizzazione di un `CLAUDE.md` proposta da
`wm-context-doctor`. Se stai eseguendo qualcosa in un repo Webmapp, questa regola ti riguarda,
qualunque strada tu abbia preso per arrivarci.

Il commit è un atto del dev, che lo fa dopo aver **letto il diff**. Un lavoro committato prima
di quella lettura gli toglie il momento in cui può dire di no — e restituirglielo poi costa un
`reset`, con la fiducia già spesa.

Se un'istruzione che stai seguendo prevede un commit, quella istruzione è sbagliata: scrivi i
file, dichiara che il lavoro è pronto e fermati.

## Tipi di delega

Due tipi, con proprietà opposte. Non vanno unificati: passare contesto a un agente
cieco ne distrugge la ragione d'essere.

| | Delega cieca | Delega informata |
|---|---|---|
| Perché | indipendenza di giudizio | volume di lettura |
| Riceve | solo percorsi e istruzioni | il contesto utile al compito |
| Non riceve mai | alcun riassunto della conversazione | — |
| Casi | `challenge`, `review-gate`, `wm-estimate` | `wm-codebase-research`, `wm-env-detect`, `wm-context-guard`, `wm-context-doctor` |
| Revisione col dev | mai per `challenge` e `review-gate`; ammessa per `wm-estimate` | ammessa, tranne `wm-env-detect` |

La riga sulla revisione non segue il taglio cieca/informata, perché dipende da
un'altra cosa: se l'esito è un **giudizio** o un **fatto misurato**. Vedi
`## Revisione con l'agente`.

Chi ha condotto un ragionamento tende a confermarlo invece di valutarlo: è il motivo
per cui la delega cieca esiste. Se un giorno un agente cieco sembra "mancare di
contesto", la risposta non è dargliene — è accettare che il suo giudizio nasca altrove.

## Quando si delega

Le deleghe sono **statiche, decise per fase** in ogni skill. Non esiste attivazione
automatica basata su una misura del context: vedi `## Misura del context`.

Una fase è delegabile quando ha **input piccolo, output piccolo, lavoro intermedio
grande** e non richiede interazione diretta con il dev — un agente non può dialogare
con l'utente.

## Contratto di ritorno

Ogni agente restituisce **solo conclusioni**, mai il materiale da cui derivano.
Il formato esatto è definito nel prompt del singolo agente.

Chi riceve un'affermazione senza prova non ha modo di verificarla: è esattamente il
materiale che si è scelto di non far entrare nel context. Dove la correttezza
dell'affermazione conta (`wm-codebase-research`), la prova è obbligatoria e
verificabile a macchina.

## Revisione con l'agente

Quando il dev contesta l'esito di un agente, l'esito non si corregge nel context
principale e non si sostituisce a mano: le obiezioni si rimandano all'agente che
lo ha prodotto.

**Si riprende la stessa sessione dell'agente, non se ne apre una nuova.** L'agente
che ha prodotto l'esito ha ancora nel proprio context il materiale che ha letto:
risponde avendo davanti ciò su cui aveva ragionato, e il giro costa poco. Un agente
nuovo ripartirebbe da zero, rileggerebbe tutto, potrebbe arrivare a un risultato
diverso per ragioni che non c'entrano con l'obiezione, e pagherebbe di nuovo
l'intero costo di lettura — cioè proprio quello che la delega serviva a evitare.
Per questo l'identità dell'agente va conservata finché la sua fase non è chiusa.

Al dev si offrono sempre tre vie, mai due: accettare l'esito, sostituirlo con una
propria decisione, oppure farlo rifare all'agente con le proprie obiezioni.
**Massimo due giri di revisione**, poi decide il dev: oltre non si converge, e un
agente che continua a correggersi finisce per assecondare invece che ragionare.

Ciò che si rimanda all'agente sono le obiezioni del dev sull'esito, non il
contesto della conversazione — vale anche per `wm-estimate`, che resta cieco su
come si è arrivati all'overview.

**Regola generale, valida anche per agenti futuri**: la revisione si applica a ogni
agente il cui esito è un **giudizio** — una stima, un rilievo, una conclusione tratta
da ciò che ha letto, un piano proposto. Oggi copre gli agenti informati
(`wm-codebase-research`, `wm-context-guard`, `wm-context-doctor`), oltre a
`wm-estimate`. Il piano di `wm-context-doctor` è il caso in cui la revisione serve
di più: il dev accetta tipicamente solo una parte degli interventi proposti, e senza
un giro di revisione l'unica alternativa sarebbe prendere o lasciare l'intero piano.

**Non si applica a chi restituisce fatti verificabili direttamente.**
`wm-env-detect` non produce un giudizio ma flag d'ambiente misurati: se sbaglia non
è un'opinione da discutere, è un errore di lettura, e si verifica rieseguendo il
check nel context principale — costa meno di un giro di revisione.

**Mai per `challenge` e `review-gate`**: lì il punto è avere un giudizio che non si
piega quando chi è stato criticato ribatte. Un revisore che si corregge per
accontentare è l'opposto di ciò che serve, e rimandargli le obiezioni
distruggerebbe la proprietà per cui esiste. Se il dev non è d'accordo con una
critica di quei due, la decisione è sua e si registra, ma il subagente non viene
richiamato.

## Tetto all'output

Ogni agente ha un tetto dichiarato nel proprio prompt. Un agente che restituisce
troppo costa **più** del non delegare: si paga il suo context e in più il suo output
nel principale.

Superato il tetto, il context principale chiede all'agente una sintesi entro il
limite invece di accettare l'output lungo. Un output che resta oltre il tetto al
secondo tentativo va trattato come fallimento (vedi `## Fallback`).

## Fallback

**Nessun agente eredita un comportamento di default.** Ogni prompt dichiara cosa
succede se l'agente fallisce, va in timeout o restituisce un formato inatteso.

Il fail-soft è frequente in questo repo (`environment-setup`, `docker-check`), ma non
è una regola generale: dove un agente alimenta una decisione di sicurezza il fallimento
deve bloccare. Per questo `review-gate: phpstan-check` non è delegato affatto.

## Misura del context

Il transcript di sessione registra i token occupati:

```bash
TRANSCRIPT=~/.claude/projects/<progetto>/<session-id>.jsonl
grep -v '"isSidechain":true' "$TRANSCRIPT" | grep -o '"usage":{[^}]*}' | tail -1
```

Somma di `input_tokens`, `cache_creation_input_tokens` e `cache_read_input_tokens`.

Tre vincoli, tutti obbligatori:

1. **Valore assoluto, mai percentuale.** Il transcript non contiene la dimensione della
   finestra: ogni percentuale richiederebbe un denominatore indovinato (200k o 1M, e il
   modello può cambiare a metà sessione).
2. **Solo informativa.** Non attiva né disattiva nulla. Un numero che non decide non può
   decidere male — ed è il motivo per cui i suoi difetti noti (compattazione che azzera il
   valore, path non contrattuale, formato di terze parti) restano innocui.
3. **Mai un numero inventato.** Transcript non raggiungibile o formato inatteso →
   `⚠️ Misura del context non disponibile.`, e si prosegue.

Il filtro `isSidechain` esclude le righe scritte dai subagenti che condividono il
transcript: senza, si leggerebbe il context dell'agente invece del proprio.
