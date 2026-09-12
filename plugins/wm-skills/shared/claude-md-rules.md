# Scrivere nei CLAUDE.md — regole condivise

Letto da `wm-context-guard` e `wm-context-doctor`. Un solo corpo di regole per entrambi:
due copie divergerebbero alla prima modifica, e il sintomo peggiore sarebbe
`wm-context-doctor` che riordina un repo in una forma che `wm-context-guard` non riconosce.

## Il principio

Il `CLAUDE.md` risponde a una sola domanda: **cosa devo sapere di questo repo prima di
toccarlo.** È pagato da ogni sessione, anche da quella aperta per un typo.

Il dettaglio operativo — una procedura passo-passo, un contratto di campo, una tabella di
valori — serve solo a chi sta facendo quella cosa specifica: va in un file dedicato, con
il rimando nel `CLAUDE.md`. Il repo usa già questo pattern in
`plugins/wm-skills/shared/orchestrator-fallback.md`.

Quando coerenza e brevità confliggono, l'informazione **non si sacrifica: si sposta.**

## I quattro rilievi

| Rilievo | Cos'è | Cosa proporre |
|---|---|---|
| **ripetizione** | la stessa cosa già detta altrove con parole diverse | fondere nella voce esistente |
| **contraddizione** | due voci che dicono il contrario | marcare la vecchia come superata o rimuoverla — **richiede conferma del dev** |
| **duplicato dal codice** | già leggibile da un file, da un test o dalla git history | non scrivere affatto |
| **fuori posto** | procedura o dettaglio che non serve a chi apre il repo | spostare in un file dedicato, lasciare il rimando |

## Dove va cosa

- **La documentazione segue il codice.** Ciò che riguarda un submodule va nel `CLAUDE.md`
  del submodule, non in quello del repo principale — stesso principio già valido per gli
  `overview.md` in `wm-plan`.

- **Un lavoro completato: una riga nell'indice, il resto in una pagina sua.** Il `CLAUDE.md`
  conserva una riga sotto `## Lavori` (titolo, ticket, un gancio di una frase, il link); cosa
  è stato deciso e perché sta in `docs/decisions/<slug>.md`.

- **L'identificativo è il nome della cartella degli artefatti**, `docs/features/<slug>/`, e la
  pagina porta lo stesso nome. Va **letto**, non ricalcolato da ID e titolo: se il titolo del
  ticket cambia, ri-derivarlo produce un nome diverso da quello reale e il legame si rompe
  senza che nessuno se ne accorga. Il controllo è meccanico: ogni pagina ha una cartella
  omonima e viceversa.

- **Un solo indice, non due.** Le vecchie `## Feature disponibili` e
  `## Decisioni architetturali` avevano la stessa granularità — una voce per lavoro — quindi
  erano due indici dello stesso insieme, compilati in due momenti diversi dello stesso
  workflow: si ripetevano per costruzione e divergevano alla prima modifica di una sola delle
  due. Un `CLAUDE.md` che le ha ancora entrambe è un candidato alla riorganizzazione, non un
  errore da correggere di nascosto.

- **Altrove si cita l'identificativo e basta.** `oc:8278` è già un riferimento risolvibile:
  ripetere di cosa parlava crea la seconda copia che poi diverge.

- **Cosa resta nel corpo del `CLAUDE.md`.** Il criterio è: *è nato da un lavoro?* Se sì ha una
  pagina e una riga d'indice. Se no — struttura delle cartelle, convenzioni di naming, come si
  valida prima del commit, come si testa in locale — non è la decisione di una feature, è il
  repo: resta una sezione normale, fuori dall'indice.

- **Il cantiere non è ciò che resta.** `docs/features/<slug>/` racconta com'è andata (overview,
  piano, note, deviazioni) e si legge per capire come è nata una feature.
  `docs/decisions/<slug>.md` è ciò che vale ancora, e si legge per non rifare un ragionamento
  già fatto. Un rilievo che confonde le due è un rilievo sbagliato.

## Cosa non si scrive

- Ciò che il codice dice già da sé (struttura, nomi, firme)
- Ciò che la git history racconta meglio (com'è stato corretto un bug)
- Il resoconto di come si è arrivati a una decisione: si scrive la decisione e il perché,
  non il percorso
- Una voce sproporzionata rispetto alla feature: otto punti per una modifica da due file

## Rimandi

I rimandi sono **link Markdown o percorsi citati in prosa**. Mai la forma
`@percorso/file.md`: in un `CLAUDE.md` quella sintassi non è un link ma un import, e il
file viene caricato automaticamente all'avvio — annullando il motivo stesso di averlo
separato.

## Limite di competenza

Queste regole riguardano **come** si scrive, mai **cosa**. Il merito di una decisione lo
può giudicare solo chi ha assistito al lavoro. Un agente che applica queste regole
segnala e propone: non scrive contenuto e non cancella nulla di propria iniziativa.
