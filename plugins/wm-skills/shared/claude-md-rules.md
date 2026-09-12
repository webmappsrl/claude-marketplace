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

- **Due assi, non uno.** `docs/features/<slug>/` è il **cantiere**: per lavoro, immutabile,
  racconta cosa fu fatto e quando. `docs/knowledge/<argomento>.md` è la **conoscenza**: per
  argomento, mutabile, dice cosa vale oggi. Il primo si legge per capire com'è nata una cosa,
  il secondo per non rifare un ragionamento già fatto. Nell'indice del `CLAUDE.md` compare la
  conoscenza, non il cantiere.

- **Le pagine di conoscenza sono per argomento, non per ticket.** Nessuno cerca per data:
  chi tocca l'header di una skill non si chiede «cosa fu deciso nel ticket X», si chiede «come
  funziona l'header e cosa è già stato provato». Tre decisioni successive sullo stesso tema non
  sono tre voci con rimandi incrociati: sono **una voce che è cambiata tre volte**.

- **Un argomento nasce al secondo lavoro che lo tocca, non al primo.** Finché un tema è stato
  affrontato una volta sola, la pagina porta il nome di quel lavoro; quando arriva il secondo,
  i due si fondono in una pagina di argomento. Così la struttura emerge da ciò che accade
  davvero, invece di richiedere una tassonomia decisa a tavolino che poi nessuno rispetta.

- **Struttura di una pagina di argomento**: lo **stato attuale in cima** — cosa vale oggi — e
  sotto *come ci siamo arrivati*, con le versioni precedenti e il motivo per cui sono cadute.
  Chi apre la pagina sa subito cosa fare; chi deve cambiare qualcosa sa cosa è già stato
  provato, e non ripropone fra sei mesi un'idea già scartata. Ogni voce porta il ticket da cui
  proviene, così si può sempre risalire al cantiere.

- **Aggiornare una pagina esistente è riscrittura, non aggiunta.** Si apre, si stabilisce cosa
  è ancora valido, si riscrive lo stato e si sposta in fondo ciò che è caduto. È lavoro di
  giudizio e può essere fatto male: non va eseguito in autonomia, e il cantiere resta la fonte
  da cui recuperare se qualcosa va perso.

- **Un solo indice.** Le vecchie `## Feature disponibili` e `## Decisioni architetturali`
  avevano la stessa granularità — una voce per lavoro — quindi erano due elenchi dello stesso
  insieme, compilati in due momenti diversi: si ripetevano per costruzione e divergevano alla
  prima modifica di una sola delle due. Un `CLAUDE.md` che le ha ancora entrambe è un candidato
  alla riorganizzazione, non un errore da correggere di nascosto.

- **Altrove si cita l'identificativo e basta.** `oc:8278` è già un riferimento risolvibile:
  ripetere di cosa parlava crea la seconda copia che poi diverge.

- **Cosa resta nel corpo del `CLAUDE.md`.** Il criterio è: *è nato da un lavoro?* Se sì sta in
  una pagina di conoscenza. Se no — struttura delle cartelle, convenzioni di naming, come si
  valida prima del commit, come si testa in locale — non è la decisione di una feature, è il
  repo: resta una sezione normale, fuori dall'indice.

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
