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

### Conoscenza e innesco non sono la stessa cosa

Una pagina di conoscenza si apre **quando serve**: prima di riprogettare qualcosa, per sapere
cosa è già stato provato. Un **innesco** è diverso: è una verifica che deve scattare ogni volta
che si fa una certa cosa, e chi non sa che esiste non la cerca.

Spostare un innesco in `docs/knowledge/` e lasciare nel `CLAUDE.md` solo la riga d'indice lo
archivia fra le cose da leggere di rado: il dettaglio è al posto giusto, ma l'obbligo di
guardarlo è sparito. **Il dettaglio sta nella pagina, l'obbligo resta fra le regole**, in una
riga che dice quando scatta:

> Quando modifichi X, controlla se ha un coupling con altri e aggiornali nello stesso lavoro:
> il dettaglio è in `docs/knowledge/…`.

Come si riconosce un innesco: la domanda è «cosa succede se nessuno lo guarda?». Se la risposta
è «qualcosa si rompe più tardi, e nessun controllo automatico se ne accorge», allora è un
innesco e serve la riga fra le regole — a maggior ragione dove non esiste una rete in CI che
raccolga l'errore.

### La residenza vale anche fra il CLAUDE.md e i file che rimanda

Spostare una procedura in `docs/howto/` non basta se nel `CLAUDE.md` ne resta un pezzo: «la
versione si aggiorna insieme al tag sul commit» è il processo di release raccontato una seconda
volta, in un punto che parla d'altro. Il controllo non si ferma ai confini del file: se un fatto
ha residenza in un howto o in una pagina di conoscenza, nel `CLAUDE.md` resta **il rimando**,
non una versione abbreviata.

Attenzione al caso in cui la versione abbreviata perde una qualifica che l'originale aveva: se
nell'howto quel passo è marcato «(dev)» e nel `CLAUDE.md` no, il riassunto descrive un'azione
che la regola in cima vieta — ed è peggio dell'omissione, perché sembra completo.

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

### Un fatto che sta nelle regole in cima non si ripete altrove

Le regole che precedono tutte le altre esistono perché vanno lette **prima** di qualunque cosa.
Ripeterne il contenuto per esteso in una sezione tematica crea due residenze dello stesso fatto,
e le due copie divergono alla prima riscrittura — con l'aggravante che una delle due parla di un
effetto irreversibile.

Dalla sezione tematica si **rimanda**, in una riga: «è la regola in cima, non una sfumatura».
Chi arriva lì sa che esiste e dove leggerla per intero; chi arriva dall'alto l'ha già letta.

### Un ID di ticket citato dev'essere verificato, non ricordato

Un riferimento sbagliato è **peggio di nessun riferimento**: chi lo segue apre un ticket che
parla d'altro e conclude che la nota è vecchia, quando invece è buona. E il costo di sbagliarlo
è basso proprio perché la verifica è banale — il cantiere sta sul disco:

```bash
ls -d "$REPO"/docs/features/<ID>-*/ 2>/dev/null     # lo slug dice di cosa parla il ticket
```

Se lo slug non c'entra con la riga che stai marcando, l'ID è sbagliato: cercalo con un `grep`
sul contenuto invece di dedurlo dal contesto. **Un ticket produce spesso trappole in domini
diversi** — le due cose scoperte mentre se ne faceva una terza — quindi non dedurre l'argomento
di un ticket da dove compare più spesso: guarda lo slug del suo cantiere.

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
  procedure che dovrebbero stare in un howto, oppure trappole che vogliono una sezione propria.

- **Le trappole non sono regole del repo, e hanno sezioni proprie.** Una regola vale sempre
  («la documentazione è in italiano»); una **trappola** scatta solo quando tocchi una certa cosa
  («sui campi Media usa `->singleMediaRules()`, `->rules()` blocca ogni salvataggio»). Sono la
  parte più preziosa di un `CLAUDE.md`, perché non si deducono leggendo il codice — e sono anche
  la più numerosa, quindi mescolarle alle regole trasforma quella sezione in un contenitore
  indifferenziato.

  Vivono in **regole path-scoped**, un file per soggetto sotto `.claude/rules/`, con il
  frontmatter `paths:` che dice quando si caricano:

  ```markdown
  ---
  paths:
    - "src/Nova/**"
  ---

  # Trappole: Nova

  - <cosa fa sbagliare>: <la conseguenza> — <cosa fare invece> (oc:<ID>)
  ```

  Il `CLAUDE.md` si carica a **ogni** avvio di sessione: tenerci dentro decine di trappole
  significa pagarle sempre, anche lavorando su tutt'altro. Una regola con `paths:` si carica
  **solo quando si tocca un file corrispondente** — cioè nel momento in cui serve. È la
  differenza fra cercare e ricevere: una trappola non si può cercare, perché chi sta per
  caderci non sospetta che esista.

  Nel `CLAUDE.md` resta **una sezione `## Trappole` di soli rimandi**, una riga per soggetto:
  serve perché il soggetto sia noto anche quando si crea un file nuovo senza averne letto
  nessuno.

  **I soggetti e i `paths:` nascono dal repo**, non da una tassonomia decisa prima. Su un
  backend Laravel saranno `src/Nova/**`, `src/Models/**`, `src/Jobs/**`, `tests/**`; su un
  frontend Angular o Ionic saranno le cartelle dei componenti, dello stato, dei servizi HTTP,
  della build (`*.config.ts`, `package.json`) — e i soggetti diventano «componenti», «store»,
  «chiamate API», «build». Come per le pagine di conoscenza, **un soggetto nasce al secondo
  lavoro che lo tocca**.

  **Un `paths:` che non corrisponde a nulla è una regola che non si carica mai, e nessuno lo
  segnala.** Verifica che ogni cartella citata esista davvero prima di scrivere il file.

  Ogni trappola resta **una riga**: cosa fa sbagliare, la conseguenza, cosa fare invece, e
  **l'ID del ticket a fine riga, mai come titolo di sezione**.

  **Una trappola con un effetto verso l'esterno non sta in quelle sezioni: va in cima al file**,
  fra le regole che precedono tutte le altre — una suite che cancella un database, un test che
  scrive su un registro condiviso, una chiamata che notifica un cliente. Chi arriva al file dal
  fondo non la leggerebbe in tempo, e quel tipo di errore non si annulla.

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

## Un cantiere è un diario, non una fotografia del risultato

`docs/features/<slug>/` si scrive **mentre** il lavoro è in corso: contiene intenzioni,
ipotesi e nomi decisi a tavolino che l'implementazione ha poi cambiato. Una classe annunciata
nell'`overview.md` può non essere mai nata; una procedura descritta nel `notes.md` può essere
stata superata il giorno dopo.

Quindi **un fatto preso da un cantiere non entra nella conoscenza senza essere verificato nel
codice**. È la fonte più ricca che esiste — nessun altro documento spiega perché una scelta è
stata presa — ma parla per costruzione del passato.

**Attenzione a quale dei due è più vecchio: spesso è il file principale.** Il cantiere si
aggiorna mentre il lavoro cambia direzione, il `CLAUDE.md` si aggiorna a fine lavoro, quando la
decisione presa in review è già di giorni prima — e a volte non si aggiorna affatto. Un caso
reale: una classe annunciata nel piano e poi eliminata in review era registrata come eliminata
nel `notes.md` del cantiere, mentre il `CLAUDE.md` continuava a descriverla come esistente. Il
sospetto non va quindi puntato sulla fonte più informale, ma su **entrambe**: l'arbitro è il
codice, e chi dei due gli dà ragione lo si scopre solo guardando.

**Il cantiere non si riscrive: si annota.** Quando si scopre che un cantiere afferma qualcosa
di falso o superato, correggerlo cancellerebbe la registrazione di com'è andata davvero. Si
appende invece in fondo al file un blocco di errata, datato, che non tocca nulla di ciò che
c'era:

```markdown
> **Errata (2026-09-12)** — questo documento afferma «<citazione>».
> Verificato: <file>:<righe> — <cosa dice il codice oggi>.
> Lo stato attuale è in [docs/knowledge/<pagina>.md](../../knowledge/<pagina>.md).
```

Serve a una cosa sola: impedire che lo stesso errore venga ripescato da chi legge quel cantiere
fra sei mesi, o dall'agente che ci pesca dentro alla prossima migrazione. Un errore corretto
solo a valle torna, perché la fonte da cui è arrivato è ancora lì e nessuno sa che è sbagliata.

## Spostare un fatto lo promuove: va verificato mentre lo si sposta

Un riordino non è neutro. Una frase falsa in fondo a un archivio di quattrocento righe non la
legge nessuno; la stessa frase in un file di cento righe, o in una pagina indicizzata, la legge
ogni sessione. **Chi migra un contenuto se ne assume la verità**: non lo sta ricopiando, lo sta
ripubblicando con più visibilità di prima.

Quindi ogni affermazione controllabile che attraversa la migrazione **si verifica mentre passa**,
contro il codice o contro il sistema che gira. Non serve rileggere tutto: basta che nessun fatto
verificabile arrivi nella nuova struttura senza che qualcuno l'abbia guardato. Un fatto che non
regge non si trasporta: si corregge, e la versione caduta va sotto «come ci siamo arrivati» col
ticket che l'ha superata — oppure sparisce, se non ha mai avuto un motivo.

Il caso che genera questa regola: un `CLAUDE.md` diceva «Laravel 10» mentre `composer.json`
dichiarava `^12.0`, e la frase ha attraversato indenne un riordino completo perché era scritta
nel file di partenza.

## Una frase falsa da sola non è una contraddizione, ed è la più difficile da vedere

Cercare le contraddizioni significa cercare **coppie**: due voci che si contendono la stessa
verità, e allora si va nel codice a stabilire chi ha ragione. Ma una voce falsa **senza gemella**
non contraddice nessuno: sta lì, coerente con tutto il resto, e nessun controllo di coerenza la
tocca. È il punto cieco di qualunque lettura che confronti il file solo con sé stesso.

Per questo **i fatti di identità del repo vanno controllati alla fonte, sempre, anche quando
nessuno li contesta**: la versione del framework e del linguaggio (contro il file delle
dipendenze e contro il runtime), quali cartelle sono davvero submodule (contro `.gitmodules`),
quali servizi esistono, quali comandi rispondono. Sono le righe che aprono il file, quelle che
ogni sessione legge per prime e che nessuno rilegge mai — e invecchiano in silenzio, perché un
aggiornamento di versione non tocca la documentazione.

## Quando un repo monta un package condiviso, la conoscenza si divide in due

Un consumer di `wm-package` — o di una libreria frontend come `wm-core` o `map-core` — non ha
una conoscenza sola: ne ha due, con residenze diverse.

- **Nel package** vive il **dominio**: come funziona il meccanismo, quali tabelle usa, quali
  stati esistono, quali vincoli valgono per chiunque lo monti. Cambia una volta e vale per
  tutti i consumer.
- **Nel repo che lo monta** vive la **customizzazione**: la sorgente dei dati di quel cliente,
  le sue regole di business, cosa è stato abilitato o disabilitato, i suoi valori di
  configurazione. Cambia per un cliente solo.

Un caso reale: il catasto sentieri. Il dominio sta in
`wm-package/docs/resources/TrailRegistry.md`; il repo cliente documenta solo il proprio import
dalla sorgente regionale e apre la voce con **«Fonte di verità: `<percorso nel package>` per il
dominio»**.

**La voce nel consumer non ripete il dominio: lo nomina e rimanda.** Ripeterlo crea due copie
che divergono al primo aggiornamento del package, e il consumer è quello che si accorge per
ultimo di essere rimasto indietro — perché il package avanza per conto suo.

Il segnale che la divisione è saltata: una voce nel consumer che spiega *come funziona* il
meccanismo invece di *come lo usiamo qui*. Quella spiegazione appartiene al package, anche se è
nata lavorando sul cliente.

**Se il fatto riguarda il package e non il consumer, non si documenta nel consumer**: si porta
nel package. Vale anche al contrario — un vincolo che esiste solo per un cliente non sale nel
package, dove varrebbe per tutti senza motivo.

## Un fatto si verifica dove vive, non dove è scritto

Il codice dice cosa il programma *farebbe*. Se le cose stiano davvero così lo dice solo il
sistema che gira. Quando una pagina afferma una quantità o un comportamento osservabile — «il
72% delle Quote non ha owner», «il bundle supera i 2 MB», «quella rotta risponde 403» — non è
**verificabile leggendo il codice**, ed è anche la cosa più facile da sbagliare: nessuno la
rilegge, perché sembra un dettaglio.

Se il repo ha un ambiente locale che si può interrogare, **la verifica si fa lì**, prima di
scrivere la frase. Cosa sia quell'ambiente dipende dallo stack, e la domanda da farsi è sempre
la stessa — *dove vive il fatto che sto affermando?*

- **Backend** (Laravel, container Docker, DB di sviluppo): una `SELECT` per ogni quantità sui
  dati, `artisan route:list` per le rotte dichiarate, la suite per i comportamenti.
- **Frontend** (Angular, Ionic, librerie): la build e il suo output per dimensioni e chunk, il
  `package.json` e il lockfile per le versioni realmente installate, i test e il dev server per
  il comportamento, la console del browser per gli errori a runtime.
- **Ovunque**: se una cosa si può eseguire invece che dedurre, si esegue.

**Lo stack non si indovina: lo si chiede a `wm-env-detect`.** È l'agente che rileva stack,
Docker, submodule e strumenti del repo e restituisce **solo i valori risolti**, senza portare
nel context l'output dei comandi di rilevamento. Dedurre l'ambiente dai nomi dei file è il modo
più rapido per cercare un container che non esiste o dare per assente una suite che c'è.

Il costo è un comando; l'alternativa è un'affermazione che nessuno saprà mai se è vera. Se
l'ambiente non è avviabile, non si inventa: si scrive la frase senza la cifra, oppure si dice
che non è stata verificata.

**Solo letture, e mai contro la produzione.** `SELECT` e comandi di sola lettura: nessuna
scrittura, nessuna migration, nessun comando che tocchi i dati. Se l'unico ambiente
raggiungibile è la produzione, non si verifica: si scrive la frase senza la cifra, o non la si
scrive affatto.

**Ogni cifra porta la data del rilevamento** — «(produzione, 2026-09-02)» — perché un dato
misurato invecchia mentre il codice resta, che sia il conteggio di una tabella o il peso di un
bundle. Con la data, chi legge fra sei mesi sa quanto
fidarsi e può rifare la query; senza, un numero vecchio si traveste da fatto stabile. Uno
scarto di poche unità su un dato datato non è un errore: è il tempo che passa.

Vale anche al contrario, per chi controlla: una cifra scritta in una pagina si ricontrolla
dove il dato vive, non si giudica a naso. Una frase impeccabile può essere falsa, e leggerla
non è verificarla.

## I criteri ufficiali, riportati qui

Estratti da <https://code.claude.com/docs/en/memory> e
<https://code.claude.com/docs/en/best-practices>. Stanno qui, e non dietro una chiamata di
rete, perché il confronto deve poter avvenire sempre: un controllo che dipende dalla
connessione è un controllo che salta proprio quando serve.

- **Sotto le 200 righe** per file. Oltre, il file consuma più contesto e **l'aderenza cala**:
  non è una questione estetica, è che le istruzioni vengono seguite meno.
- **Un `CLAUDE.md` è contesto, non configurazione applicata.** Nessuna riga è garantita:
  ciò che deve valere sempre, a prescindere dal giudizio, si scrive come hook, non come frase.
- **Istruzioni concrete abbastanza da verificarle**: «indentazione a 2 spazi», non «formatta
  bene»; «esegui `npm test` prima del commit», non «testa le modifiche».
- **Due istruzioni in conflitto fanno scegliere a caso.** È la ragione tecnica per cui le
  contraddizioni sono il primo dei quattro rilievi.
- **Gli import `@percorso` non alleggeriscono**: il file importato si carica comunque
  all'avvio. L'unico meccanismo che alleggerisce davvero è la regola con `paths:` in
  `.claude/rules/`, che si carica solo quando si toccano i file corrispondenti. I `CLAUDE.md`
  annidati in sottocartelle si caricano solo lavorando lì dentro.
- **Una procedura a più passi, o valida per una sola parte del repo, non sta nel file
  principale**: diventa una skill o una regola con `paths:`.
- **I commenti HTML a blocco** (`<!-- ... -->`) vengono rimossi prima che il file entri nel
  contesto: costano zero e servono a chi mantiene il file.
- **Si taglia ciò che è derivabile dal codice** — alberi di cartelle, elenchi di dipendenze,
  panoramiche di architettura — e **si tiene ciò che non lo è**: le trappole, il perché di una
  scelta, le convenzioni che si discostano dal comportamento predefinito degli strumenti.

**Le regole del team vincono in caso di conflitto**, ma il conflitto va dichiarato al dev, non
risolto in silenzio. Un caso reale ricorrente: il criterio «l'architettura è derivabile dal
codice» porta a proporre il taglio di una sezione che il team ha scritto apposta per dire
*quale* parte guardare. Quel taglio non si propone: si segnala che i due criteri divergono e
si lascia decidere.

## Limite di competenza

Queste regole riguardano **come** si scrive, mai **cosa**. Il merito di una decisione lo
può giudicare solo chi ha assistito al lavoro. Un agente che applica queste regole
segnala e propone: non scrive contenuto e non cancella nulla di propria iniziativa.
