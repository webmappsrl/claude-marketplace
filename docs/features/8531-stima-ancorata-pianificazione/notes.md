> Ticket: oc:8531

# Notes

## Deviazioni dal piano

Nessuna: il piano è stato scritto a consuntivo sulla sequenza effettivamente seguita. Il lavoro
è nato dall'analisi del consuntivo di oc:8530, richiesta dal dev a lavoro chiuso, quindi
overview e piano sono stati prodotti dopo che la soluzione era già stata discussa e approvata
nel dialogo. È una deviazione dal workflow di `wm-plan`, non dal piano.

## Bug trovati

**Contraddizione preesistente fra due file**, trovata leggendo: `docs/knowledge/wm-plan-stima-ore.md`
prescriveva il marcatore `[stima v2 — per-componente]` su ogni stima scritta su Orchestrator,
mentre `wm-plan/SKILL.md` lo vieta esplicitamente. Due verità in due file, nessuna delle quali
segnalava l'altra. Sanata: vince il divieto, con il motivo registrato nella storia della pagina.

## Decisioni

**Si è toccata la cecità di un agente cieco, ed è la decisione che pesa di più.** `wm-estimate`
nasce cieco perché chi ha condotto il dialogo stima ottimista. Passargli la durata della
pianificazione è una deroga apparente: si è distinto fra la cecità che protegge il giudizio —
non sapere *come* si è arrivati all'overview, di cosa il dev si è convinto — e l'ignoranza di un
fatto misurato, che non protegge nulla e lascia l'agente a sommare componenti nel vuoto.

La distinzione regge, ma è una porta socchiusa: la prossima richiesta di passare all'agente
«solo un altro dato» va guardata con sospetto, perché ogni singolo dato sembrerà innocuo quanto
questo.

**Il dato è ancora, non addendo.** La quota di pianificazione continua a sommarla il context
principale. Se l'agente la sommasse a sua volta, verrebbe contata due volte.

## Follow-up

- **Verificare all'uso se l'ancora ha sostituito il giudizio.** Se le stime diventassero tutte
  «circa quanto la pianificazione», l'agente avrebbe imparato ad ancorarsi invece che a stimare.
- **Il campione resta uno.** Quando ci saranno tre o quattro lavori con consuntivo misurato,
  vale la pena rileggere questi criteri con i dati veri: potrebbero risultare ancora troppo
  generosi.
- **Nessuna calibrazione sui ticket storici** finché Orchestrator non espone un'aggregazione
  stimato-contro-effettivo (già registrato in oc:8278).
