> Ticket: oc:8278

# Rivedere criteri di stima ore in wm-plan

## Cosa cambia

La `Fase: estimation` di `wm-plan` passa da un modello a componenti + buffer forfettario unico (minimo 10%, mai zero) a un modello che distingue, per ogni componente della stima:

- **Scrittura pura** — componenti per cui, al termine di `Fase: overview` + `Fase: challenge`, non restano domande aperte su "come deve comportarsi" (specifica già completa: campi, endpoint, escaping, cache, ecc.). Stimate a ritmo di scrittura assistita da Claude Code, buffer 0%.
- **Decisioni aperte** — componenti che richiedono ancora scelte UX/comportamentali, o reverse-engineering di comportamento legacy non documentato. Stimate a ritmo umano, buffer 20-30%.

Il minimo forfettario per feature multi-file scende da 1h a 0.5h (cuscinetto review/commit/PR, non scrittura).

## Perché

Analisi dei ticket Feature di luglio 2026 (dati Orchestrator, stima vs ore effettive) mostra un bias sistematico di overstima sui task con specifica tecnica già completamente definita (es. oc:8211: stimate 8h, effettive 2.03h — escluso dal computo per probabile tracking non adeguato del progress, ma indicativo del pattern; oc:8276, oc:8155: piccoli ticket comunque overstimati). Al contrario, i due casi di sottostima nel periodo (oc:8259: 2h→6.8h, oc:8252: 8h→10.4h) sono entrambi lavoro di reverse-engineering di un widget legacy non documentato con decisioni UX ancora aperte durante l'esecuzione.

La formula attuale stima "tempo di scrittura per componente" a un ritmo pre-Claude-Code, che non riflette la velocità reale su codice ben specificato, mentre il tempo di decisione/esplorazione umana resta quello che resta — da cui l'asimmetria overstima/sottostima osservata.

## Requisiti

- [ ] Ogni componente della stima viene classificato come "scrittura pura" o "decisioni aperte" in base all'esito di `Fase: overview` + `Fase: challenge` (nessuna domanda aperta residua → scrittura pura)
- [ ] Buffer rischio applicato per componente, non più come valore unico forfettario finale: 0% su scrittura pura, 20-30% su decisioni aperte
- [ ] Il totale della stima è la somma dei componenti classificati, con l'aggiunta di un buffer di integrazione trasversale del 5% sul totale quando la feature coinvolge più di un componente (copre rischio di interazione/adapter tra componenti, non attribuibile a un singolo componente)
- [ ] Minimo forfettario per feature multi-file abbassato da 1h a 0.5h
- [ ] La preview della stima mostrata al dev in `estimation: conferma` (prima dell'accettazione) espone per ogni componente: classificazione (scrittura pura / decisioni aperte), ore stimate e buffer applicato — così il dev vede dove si concentra il rischio e può correggere puntualmente, non solo accettare/rifiutare un totale opaco
- [ ] Il coefficiente per-dev NON viene implementato in questo ciclo (vedi Out of scope)
- [ ] Ogni stima scritta su Orchestrator (`estimation: scrittura su Orchestrator`) include un marcatore di versione del criterio usato (es. `[stima v2 — per-componente]` nella nota/description), così le stime future restano distinguibili da quelle prodotte col criterio precedente per analisi di calibrazione successive
- [ ] `Fase: notes` registra esplicitamente ogni caso in cui un componente classificato "scrittura pura" ha in realtà richiesto decisioni non previste durante l'esecuzione (falso negativo di classificazione) — dato necessario per calibrare il criterio nei cicli successivi
- [ ] Se durante `Fase: execution` emerge un problema non previsto e stimabile in ore, viene proposta al dev una revisione della stima (nuovo totale + motivazione), con conferma esplicita prima di aggiornare `estimated_hours` su Orchestrator — segue la stessa regola generale scritture (preview + conferma) già in vigore per POST/PATCH
- [ ] Il tempo di pianificazione è **misurato, non stimato**: `wm-plan` registra il timestamp reale di inizio (`Fase: ticket`) e lo confronta con il timestamp al momento della conferma di `Fase: estimation`, calcolando la durata effettiva trascorsa. Questo valore viene incluso nel totale `estimated_hours` e appare come **prima voce** della tabella di breakdown (es. "Pianificazione: \<Nh\> (misurata) — dialogo reverse-interaction, overview, challenge, stima"), separata dalle voci di scrittura/implementazione (queste ultime restano stime, essendo lavoro futuro) — permette di leggere il rapporto pianificazione/esecuzione per dev come metrica
- [ ] Il valore proposto al dev in `estimation: conferma` distingue esplicitamente la quota **misurata** (pianificazione, dato reale) dalla quota **stimata** (implementazione, dato previsionale) — es. "Misurato: 0.5h + Stimato: 0.6h = Totale: 1.1h" — mai un unico numero fuso che nasconda quanto è certezza e quanto è previsione

## Rischi

- **Classificazione soggettiva**: la distinzione "scrittura pura" vs "decisioni aperte" dipende dal giudizio di Claude a fine `Fase: challenge` — mitigato dal criterio operativo esplicito (zero domande aperte residue = scrittura pura), non lasciato all'impressione generale.
- **Buffer 0% può sottostimare imprevisti anche su task "semplici"**: mitigato dal fatto che il minimo 0.5h resta comunque presente come cuscinetto minimo di review/commit/PR.
- **Dati di calibrazione limitati**: l'analisi si basa su un solo mese (luglio 2026) e un campione ridotto di ticket per dev (8-15) — il nuovo criterio va validato nei prossimi cicli, non è definitivo.

## Out of scope

- **Coefficiente di velocità per-dev**, aggiornato automaticamente a ogni ticket chiuso. Non implementato in questo ciclo perché: (1) Orchestrator non espone oggi un endpoint di listing/aggregazione per stimare/effettive per utente — solo GET singolo ticket o tag→stories; (2) il campione per dev (8-15 ticket) è troppo ridotto per un coefficiente affidabile. Da rivalutare in un ciclo successivo con più storico.
- Non si modifica il meccanismo di stima per Bug/Task (restano non stimati in ore, invariato).
- Non si tocca la logica di `Fase: challenge` in sé, solo il suo output viene usato come input alla classificazione dei componenti in `Fase: estimation`.

## Moduli toccati

- `plugins/wm-skills/skills/wm-plan/SKILL.md` — sezione `## Fase: estimation` (sottosezioni `estimation: analisi`, `estimation: conferma`, `estimation: scrittura su Orchestrator`)
