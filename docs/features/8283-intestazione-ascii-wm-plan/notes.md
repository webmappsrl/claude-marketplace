> Ticket: oc:8283

# Notes — Intestazione ASCII per skill wm-plan con info versione/aggiornamenti e diagramma di flusso

## Deviazioni dal piano

I 6 task (1, 2, 3, 4, 4b, 5) sono stati eseguiti come pianificato in `plan.md`. Un solo aggiustamento post-esecuzione riportato sotto in "Bug trovati".

## Bug trovati

- **Banner ASCII "figlet-style" del Task 1 non leggibile**: il primo test in locale (`/plugin marketplace add`) ha mostrato che il banner scritto a mano per "wm-plan" (art in stile figlet, 6 righe) era trascritto male e mostrava un testo illeggibile/sbagliato invece della scritta attesa. Corretto sostituendolo con un banner box-drawing più semplice e verificabile carattere per carattere (`┌─ W M · P L A N ─┐`), meno rischioso di un font ASCII complesso generato a mano. Il diagramma Mermaid dell'Artifact non era invece soggetto allo stesso rischio, perché il rendering veniva verificato visivamente ad ogni redeploy.
- **Narrazione dell'agente inframezzata all'header nel primo test**: sempre dal test in locale, l'header (banner → check versione → link diagramma) usciva intervallato da commenti conversazionali ("ora leggo il CLAUDE.md", "faccio i check di versione") che rompevano l'effetto di blocco unico e immediato voluto per l'header. Non previsto in overview/plan perché non era emerso finché non si è visto il comportamento reale a runtime. Corretto aggiungendo un vincolo esplicito in `Header di sessione`: nessuna narrazione prima, durante o dopo il blocco banner+check+link.

## Decisioni

- **Task 5 eseguito con conferma esplicita account prima della pubblicazione**: durante l'esecuzione l'utente ha rilevato di essere loggato con l'account **net7** invece che **webmapp**, e ha effettuato lo switch account (`/login`) prima di procedere alla pubblicazione dell'Artifact — conferma il pattern "tentativo/fallimento" deciso in `Fase: challenge` dell'overview: la verifica reale è avvenuta empiricamente (l'utente ha controllato da sé), non tramite un check automatico nella skill.
- **Stima accettata dal dev a 1h** invece della proposta iniziale (3.5h) in Fase: estimation — nessuna motivazione tecnica richiesta, applicata la regola "la stima finale è sempre quella approvata dal dev".

## Follow-up

- Il rischio di **drift documentale silenzioso** tra l'Artifact e lo stato reale di `SKILL.md` resta accettato consapevolmente (vedi overview.md → Rischi) — nessun meccanismo di verifica automatica implementato in questo ciclo. Task futuro se il criterio va rivisto.
- L'istruzione di rigenerazione in `CLAUDE.md` scatta su qualsiasi modifica al repo, non solo su modifiche a `wm-plan` — comportamento voluto, ma da monitorare per capire se in pratica genera troppi redeploy non necessari.
