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

- **Quattro destinazioni, quattro domande diverse.** Prima di scrivere qualsiasi cosa,
  stabilisci a quale domanda risponde:

  | Dove | Domanda | Per chi | Cambia quando |
  |---|---|---|---|
  | `docs/features/<slug>/` | com'è andato quel lavoro | chi indaga il passato | mai: è il cantiere, immutabile |
  | `docs/knowledge/<argomento>.md` | perché funziona così, cosa è già stato provato | chi deve cambiarlo | cambia la decisione |
  | `docs/howto/<procedura>.md` | come si fa | chi lavora sul repo, o Claude che esegue | cambia il codice |
  | `docs/guide/<argomento>/` | cosa fa il prodotto e come si usa | **l'utente finale** | cambia ciò che l'utente vede |

  Nell'indice del `CLAUDE.md` compare la **conoscenza**, con i rimandi a howto e guide dove
  servono. Il cantiere non va nell'indice.

- **`howto` e `guide` non sono la stessa cosa, anche se sono entrambi istruzioni.** Li separa
  il destinatario, non la forma. Un howto può contenere percorsi, comandi, nomi di file e
  branch; una guida cliente **non deve contenerne nessuno** — stessa regola che `wm-plan` già
  applica alla risposta al cliente. Tenerli insieme porta prima o poi a una guida con dentro un
  percorso di file, o a un testo scritto per il pubblico sbagliato. C'è anche un rischio meno
  visibile: Claude legge un howto come istruzione operativa, mentre una guida descrive un
  prodotto e non dice cosa fare.

- **Le guide cliente portano i propri asset.** Screenshot e immagini vanno in
  `docs/guide/<argomento>/` accanto al testo, mai in una cartella immagini comune: spostare o
  cancellare una guida non deve lasciare file orfani.

- **Le guide sono pubblicate, il resto no.** `docs/guide/` è l'unica parte della
  documentazione destinata a diventare un sito. Tutto ciò che le sta accanto — il cantiere in
  `docs/features/`, i perché in `docs/knowledge/`, le procedure in `docs/howto/` — è interno e
  **non deve finire online**: su un repo cliente significherebbe pubblicare le note di sviluppo
  di un progetto altrui.

  Conseguenza pratica: la pubblicazione va fatta con un workflow che dichiari esplicitamente
  di pubblicare `docs/guide/` e nient'altro. GitHub Pages configurato da interfaccia accetta
  come sorgente solo la root o l'intera `docs/`, quindi quella strada pubblicherebbe anche il
  resto. Un workflow versionato nel repo è verificabile e lascia traccia quando cambia; una
  tendina nelle impostazioni no.

  Scrivendo una guida, dai per scontato che sia **leggibile da chiunque**: nessun percorso
  interno, nessun nome di branch, nessun dato di un altro cliente.

- **Una procedura non ha la struttura di una pagina di conoscenza.** Non ha uno «stato
  attuale» e delle «versioni cadute»: ha dei passi. Se ti trovi a forzare una procedura nella
  struttura della conoscenza, è nel posto sbagliato.

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

- **`## Regole del repo`: un asse diverso, non una zona protetta.** Nel corpo del `CLAUDE.md`
  c'è una sezione per le regole che valgono perché *questo* repo è fatto così: struttura delle
  cartelle, convenzioni di naming, come si valida prima del commit, come si testa in locale,
  procedure che qui sono obbligatorie e altrove non esistono.

  Gli agenti la trattano **come tutto il resto** per forma e coerenza: se due regole si
  contraddicono lo dicono, se una è diventata falsa rispetto al codice lo dicono, se una è
  cresciuta al punto da meritare un file dedicato lo propongono.

  Quello che **non** devono fare è pretendere che abbia le proprietà delle pagine di
  conoscenza. Una regola del repo non ha un ticket da citare, non ha un cantiere in
  `docs/features/`, non ha una riga nell'indice della conoscenza. Misurarla con quel metro
  significherebbe segnalare come difetto una cosa perfettamente a posto.

- **Anche fra le regole del repo vale la separazione fra obbligo e procedura.** Che una regola
  non si misuri col metro della conoscenza non significa che possa contenere qualsiasi cosa:
  nel `CLAUDE.md` resta **l'obbligo**, in una o due righe imperative, e i passi per eseguirlo
  vanno in `docs/howto/` con il rimando.

  Esempio: fra le regole sta «prima di ogni release esegui la checklist di rilascio», con il
  link; i cinque passi del rilascio stanno nell'howto. Chi apre il repo deve sapere *che* la
  checklist esiste ed è obbligatoria; *quali* siano i passi gli serve solo nel momento in cui
  rilascia.

  Senza questo criterio la sezione delle regole diventa il nuovo posto in cui il file ricresce
  — lo stesso problema di prima, spostato di una sezione. Il segnale è misurabile: se
  `## Regole del repo` è la sezione più pesante del `CLAUDE.md`, quasi certamente contiene
  procedure che dovrebbero stare altrove.

- **Una regola del repo può nascere da un lavoro, e non è una contraddizione.** Quello che
  conta è dove vive dopo. La *decisione* di avere quella regola — con il perché e le
  alternative scartate — sta nella pagina di conoscenza, col suo ticket. La *regola* in sé,
  quella da seguire ogni volta che si tocca il repo, sta fra le regole del repo. Sono la stessa
  cosa vista da due lati: **perché è così** contro **cosa devi fare**.

  Esempio: la rigenerazione del diagramma in `claude-marketplace` è stata introdotta da un
  ticket, ma è una regola che vale solo in quel repo e va seguita sempre; il ragionamento che
  l'ha prodotta sta nella conoscenza, l'obbligo sta fra le regole.

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
