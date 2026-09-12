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

## Una migrazione si produce dalle regole, mai da una versione precedente

Chi esegue un piano di riordino del `CLAUDE.md` **non riparte da una versione trovata nella
storia di git**, in un backup o in un altro branch — anche se sembra già approvata, anche se
farebbe risparmiare lavoro. La produce da zero, applicando le regole di questo file al
contenuto attuale.

Il motivo: una versione precedente porta con sé gli errori che aveva, e li fa passare per
verificati. Se era stata annullata, di solito è proprio perché qualcosa non andava; e se era
stata corretta a mano, quelle correzioni non le ha fatte il metodo, quindi riprenderle non dice
nulla su quanto il metodo funzioni. Il risultato deve dipendere solo dalle regole: è l'unico
modo per accorgersi di quale regola manca quando esce sbagliato.

## Ogni fatto ha una residenza

Un fatto — un comando, un percorso, un vincolo, un valore — vive in **una sezione sola**.
Altrove, al massimo un rimando di mezza riga. Se `localhost:8099` compare sia nella regola che
lo impone sia nella sezione sull'ambiente, o se «valida prima di dichiarare il lavoro pronto»
sta in due punti con due formulazioni, quello è un doppione anche se le parole non coincidono.

Il doppione letterale è il caso facile. Quello che sfugge è **lo stesso fatto detto in due
modi**: si riconosce chiedendosi, per ogni affermazione, «dove altro nel file questo è già
detto?» — non «dove altro compare questa frase?». Ogni riorganizzazione ne fa rientrare
qualcuno, perché chi sposta una sezione porta con sé i fatti che le servivano senza togliere
quelli rimasti dove stavano.

### Quando un fatto ha due facce legittime

Capita che lo stesso oggetto abbia due aspetti che appartengono a sezioni diverse: un file può
essere insieme **congelato** (vincolo) e **da aggiornare in certi casi** (manutenzione); un
comando può essere insieme uno strumento e un obbligo. La tentazione è metterne una faccia per
sezione — ed è il modo in cui il doppione rientra a ogni riorganizzazione, perché ogni faccia
sembra a casa propria.

**Le due facce vanno nella stessa voce**, in una sezione sola, in una frase che le tiene
insieme: «X è congelato nella struttura, e si aggiorna nel contenuto quando cambia Y». Altrove,
se serve, solo un rimando.

Il segnale che questo caso si sta presentando: la stessa cosa compare in due sezioni **con
verbi opposti** — «non modificare» da una parte, «aggiorna» dall'altra. Non è una
contraddizione da risolvere scegliendo, è un fatto spezzato da ricomporre.

## Nessuna istruzione può contraddire le regole in cima

Le prime sezioni del file sono le regole che valgono più di tutto. Qualsiasi istruzione più
in basso che **implichi un'azione lì vietata** è una contraddizione, e va segnalata come tale
anche quando non nomina l'azione direttamente: «esegui X prima di ogni commit» in un file che
vieta all'agente di committare chiede un passo che non può avvenire, e una regola che si
impara a ignorare insegna che le regole si possono ignorare.

Stessa attenzione al **perimetro**: una regola scritta più larga di quanto il codice
giustifichi («ogni modifica al repo richiede…» quando il vincolo riguarda solo una parte) o
viene seguita alla lettera producendo lavoro inutile, o viene ignorata. E al **lessico**: se
il repo ha smesso di «rigenerare» qualcosa e ora la «aggiorna», la parola vecchia in una regola
nuova è un segnale che la regola è stata copiata, non riletta.

## Un vincolo che protegge da un effetto esterno vive nel file principale

Una regola path-scoped in `.claude/rules/` si carica solo quando si toccano i file che
dichiara. Va bene per i vincoli che riguardano *quei* file. Non va bene per un vincolo che
protegge da un effetto verso l'esterno — pubblicare qualcosa online, scrivere su un sistema
che notifica un cliente, esporre dati — perché il momento in cui serve è quasi sempre un
altro: chi aggiunge un file sotto `docs/` non sta toccando il diagramma, e la rule che dice
«si pubblica solo `docs/guide/`» non si carica.

Quindi: **il divieto sta nel `CLAUDE.md`**, in forma di divieto e non di descrizione
(«pubblica solo X, il resto non va online», non «pubblicazione di X»). La rule path-scoped può
ripeterlo con il dettaglio operativo, ma non sostituirlo. Quando sposti materiale dal
`CLAUDE.md` a una rule, controlla se dentro c'era un vincolo di questo tipo: è la cosa che si
perde più facilmente, perché sembra un dettaglio della procedura e invece è l'unica riga che
conta.

## Una procedura spostata in howto eredita le regole in cima

Le regole in cima al `CLAUDE.md` valgono per tutto il repo, anche per i file che il
`CLAUDE.md` rimanda. Una checklist spostata in `docs/howto/` che elenca «commit» e «tag» come
passi, in un repo la cui prima regola vieta all'agente di committare, è una procedura che
contraddice il file che la cita — e chi la esegue non ha davanti quella prima regola, perché
ha aperto l'howto e non il `CLAUDE.md`.

Quindi: quando una procedura contiene passi che l'agente non può eseguire, **quei passi vanno
marcati esplicitamente come del dev**, nella procedura stessa, non solo nel `CLAUDE.md` che la
rimanda. Il controllo di coerenza con le regole in cima si applica anche ai file in `howto/`,
non solo al `CLAUDE.md`: spostare una procedura non la esonera.

## I quattro rilievi

| Rilievo | Cos'è | Cosa proporre |
|---|---|---|
| **ripetizione** | la stessa cosa già detta altrove con parole diverse | fondere nella voce esistente |
| **contraddizione** | due voci che dicono il contrario | **prima si verifica nel codice quale delle due è vera** — è un fatto, non un'opinione; poi si propone di marcare la falsa come superata o rimuoverla. Al dev resta solo la scelta fra marcare e rimuovere: **mai** la domanda «quale delle due vale oggi», che ha già risposta nel repo |
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

- **«Utente finale» si legge rispetto a cosa produce quel repo.** Non significa sempre «il
  cliente»: significa chi usa la cosa che il repo produce. In un progetto Laravel per un
  cliente è il cliente; in un repo di strumenti interni — un plugin, una libreria, un CLI —
  sono i dev che lo installano e lo usano. La documentazione d'uso di uno strumento interno è
  una guida a tutti gli effetti, anche se nomina concetti tecnici: il suo lettore è tecnico.

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
