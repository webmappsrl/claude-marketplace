# Un pane che esplode i riferimenti di wm-plan

## Intento, e perché

`wm-plan` parla per sigle: `R3`, `oc:8700`, `review-gate`, un link a un'àncora di `notes.md`.
Sono comode per chi scrive e opache per chi legge: per sapere cosa sia `R3` il dev deve aprire
`overview.md` e cercare l'àncora, e lo deve rifare ogni volta. La frase del dev che ha aperto
questo lavoro: «corregge quel problema che ogni volta non definisci mai quello che richiami».

La mod apre un pane accanto alla chat che, **ogni volta che un riferimento viene citato in un
messaggio, lo mostra lì per esteso**. Non è una legenda ferma di tutto l'overview: nel pane c'è
solo ciò di cui si sta parlando adesso, il più recente in cima.

## Goal

- Ogni riferimento che compare in un messaggio dell'assistant viene risolto nel pane, con il suo
  testo completo e la sua provenienza.
- **Il pane si riapre da solo** al primo riferimento citato dopo che e stato chiuso. Non e una
  comodita: provando il prototipo il pane e sparito premendo le frecce, e senza riapertura
  automatica non torna piu, perche nessuno si ricorda il comando che lo riapre.
- Il pane dice sempre da quale file sta leggendo, così il dev distingue un dato vero da un campione.
- Zero obblighi nuovi per la skill: `wm-plan` non cambia di una riga.

## Non-goal

- Non sostituisce l'header di sessione, che resta a testo.
- Non mostra lo stato del workflow (fase corrente, gate passati): a testo si legge già.
- Non interviene su nessuna tool call: non nega, non riscrive, non blocca niente.
- Non è una mod di controllo delle regole del repo (divieto di commit, scritture su Orchestrator).
  È una cosa diversa, discussa e rimandata: vedi «Fuori da questo lavoro».

## Anti-goal

Il modo in cui questo lavoro fallisce pur funzionando: **il pane mostra righe vuote e non dice
perché**. Il formato dei requisiti con àncore è entrato in `af492d6` e nessun `overview.md` del repo
lo usa ancora; se il formato cambia di nuovo, un parser muto fa credere al dev che non ci siano
requisiti. Il pane deve sempre distinguere «non ho trovato requisiti» da «non ho capito il file».

## Vincoli

- Il modulo gira senza DOM e senza Node: tutto passa da `$`.
- Una lettura di `$.fs.read` è limitata a 4 MiB; gli overview sono molto sotto.
- Un pane aperto senza che nessuno lo chieda si posiziona da 144 colonne di terminale; uno aperto
  su richiesta (`/wm-anchors`) sta a qualsiasi larghezza.
- Il team lavora tutto da terminale: nessuna superficie diversa da coprire.
- `plugins/wm-skills/hooks/hooks.json` usa oggi la chiave `hooks` per lo script di lint di
  `wm-context-doctor`; una mod si dichiara sotto `modules`. **Verificato: le due chiavi convivono**
  nello stesso file — il plugin ufficiale `code-modernization` ha gia entrambe, cinque hook da
  settings piu il suo modulo. La mod si aggiunge al file esistente, senza riorganizzare il plugin.

## Approcci considerati

**Come il pane si alimenta.**

1. *Timer* — `$.clock.every` controlla la data di modifica dell'overview e rilegge. Vede ogni
   modifica, anche quelle fatte a mano, ma tiene un timer acceso per tutta la sessione e rilegge
   anche quando non serve.
2. *Hook sulle tool call* — intercetta `Write`/`Edit` sui path che finiscono per `overview.md`.
   Immediato e gratuito, ma cieco a ciò che non passa dall'assistant: il dev che corregge in vim,
   un `git checkout` che cambia branch.
3. *Hook su `session.append`* — l'evento che il motore solleva per ogni riga conservata. Era la
   scelta iniziale, **ed è stata abbandonata provandola**: il motore ha segnalato che l'hook veniva
   saltato, perché un altro plugin installato (`cc-plugin-sec-default`) gira prima nella catena e
   passa l'evento oltre senza proseguirla. Un approccio che dipende da un evento che un altro
   plugin può togliere di mezzo non è affidabile, e il modo in cui fallisce è silenzioso.
4. **Il pane legge la conversazione da sé** *(scelto)* — al momento di disegnarsi chiama
   `$.session.messages()`, prende gli ultimi messaggi dell'assistant, ci cerca i riferimenti e
   risolve ciascuno sull'overview. Non dipende da nessuna catena di eventi: nessun altro plugin
   può interporsi fra il pane e i dati. Il ridisegno è agganciato a `turn.complete`, che serve solo
   a chiedere l'aggiornamento, e se anche quello venisse intercettato il pane resterebbe corretto
   al disegno successivo, non vuoto.

**Dove vive la mod.**

1. *Plugin separato* (`plugins/wm-plan-anchors/`) — version e tag propri, nessun rischio sul plugin
   esistente; ma è un secondo plugin da installare e da ricordarsi di allineare.
2. **Dentro `wm-skills`** *(scelto)* — si installa con la skill, e il contratto sul formato dei
   requisiti sta nello stesso plugin che quel formato lo definisce. Tenerli separati è il modo di
   farli divergere.

## Cosa fa, nel dettaglio deciso

**I riferimenti riconosciuti sono quattro:**

| Riferimento | Cosa mostra il pane | Da dove |
|---|---|---|
| `R1`, `R2`… | il requisito per esteso, la sigla `[UX]` se c'è, se la riga `Verifica` c'è o manca | `overview.md` della feature |
| `oc:<numero>` | titolo, tipo e stato del ticket | Orchestrator, via il server MCP |
| nomi di fase (`review-gate`, `reverse-interaction`, `caso-a-split-detection`) | cosa fa quella fase, in due righe | `SKILL.md` di `wm-plan` |
| link a un'àncora di `notes.md` | la voce puntata, per esteso | `notes.md` della feature |

**Comportamento del pane:** il riferimento citato per ultimo sta in cima, marcato come tale; i
precedenti restano sotto, fino a dodici, poi i più vecchi cadono. Ogni voce porta la sua àncora.
In fondo, sempre, il file da cui si sta leggendo.

**Quale feature:** la cartella sotto `docs/features/` il cui `overview.md` è stato modificato più di
recente, fra quelle che usano le àncore. È la feature su cui si sta lavorando.

**Quando non trova niente:** il pane lo dice, distinguendo i due casi — nessun overview con àncore
nel repo, oppure overview trovato ma nessun requisito riconosciuto.

## Cosa ha insegnato il prototipo

Due cose scoperte provandolo, non deducendole:

- **Il pane si chiude per sbaglio.** Basta premere le frecce. Da qui il requisito della riapertura
  automatica sopra.
- **Piu pane non si affiancano: diventano linguette**, una sola mostrata per volta, e il contratto
  dell'API non dichiara nemmeno quale tasto passa dall'una all'altra. Il contenuto si suddivide
  dentro un pane solo, oppure su superfici diverse che si vedono insieme (pane, band sopra il
  prompt, status line). Niente schede.
- **Un altro plugin può togliere un evento di mano.** `cc-plugin-sec-default` consuma
  `session.append` senza passarlo oltre, e l'hook della mod non viene mai eseguito. Il motore lo
  dice, ma solo a chi guarda le notifiche: da qui la regola di non appoggiare il contenuto del pane
  su un evento che altri possono interrompere.
- **La cartella delle mods di sviluppo cambia con la sessione**, quindi una mod che vive li e usa e
  getta per costruzione: un motivo in piu perche il lavoro vero stia in `plugins/wm-skills/`.

## Deciso provando, non ragionando

- **Le voci stanno in ordine di posizione nel testo**, non in ordine di definizione. Il prototipo
  sbagliava: mostrava in cima l'etichetta definita per prima, non quella nominata per ultima, e
  bastava guardarlo per vederlo.
- **Il pane si limita allo spazio che ha**: le voci piu vecchie cadono fuori invece di restare
  tagliate a meta. L'altezza e automatica, non un numero fisso di voci.
- **Una sigla fra backtick non e una citazione.** Il prototipo esplodeva i propri esempi: scrivendo
  «quando cito `R1`» il pane mostrava R1 come se ci stessimo lavorando. Il segnale era gia nel
  testo — della sigla si parla fra backtick, al requisito ci si riferisce senza — e ora i pezzi fra
  backtick, singoli o in blocco, sono esclusi dalla ricerca. Gli spazi restano al loro posto, cosi
  l'ordine per posizione nel testo non si sposta.
- **Il testo nel pane e pulito dal markdown**: i backtick e gli asterischi dei messaggi non
  significano niente li e sono solo rumore.

## Oltre i requisiti: le etichette definite parlando

Lo stesso meccanismo vale per qualunque cosa venga **definita in conversazione** su una riga di
elenco, nella forma `- **NOME** — descrizione`, con NOME in maiuscolo di almeno tre caratteri.
Da li in poi ogni occorrenza di quel nome lo richiama nel pane, come fa un requisito.

Serve perche il problema del dev non riguarda solo `overview.md`: riguarda tutto cio che viene
richiamato per sigla senza essere ridefinito. Una regola del repo, il nome di un modo di
funzionare, un vincolo deciso a meta conversazione. La maiuscola e la forma della riga tengono la
cosa esplicita: una definizione si dichiara, non si indovina.

## In testa al pane: quanto context resta

La prima riga del pane e la percentuale di context della conversazione principale, colorata, con
accanto il modello degli agenti al lavoro quando ce ne sono.

Era stata messa nella band, la riga sopra il prompt, e **spostata nel pane provandola**: li finiva
in mezzo al testo della conversazione e lo sguardo la scavalcava. Nel pane sta ferma in un posto
che si guarda apposta. La status line non era un'alternativa: accetta solo testo, senza colori.

Il modello degli agenti e l'unico pezzo che vale la pena mostrare della lista nativa sotto il
prompt, perche e l'unico che quella lista non da: li si legge nome, durata e token, mai su quale
modello sta girando.

I colori si chiedono al tema come `success`, `warning`, `error`, non come verde, giallo e rosso
crudi: seguono il tema scelto dalla persona e restano leggibili su qualunque sfondo.

**Le soglie sono una scelta, non un dato del motore**, e stanno in una funzione sola per poterle
cambiare: verde sotto il 50%, giallo dal 50%, rosso dal 75% con la scritta che conviene un
compact. Vanno tarate sull'uso reale: se il compact scatta prima o dopo, la soglia rossa va
spostata di conseguenza.

Il dato e `context.percent` di `$.session.usage()`, la stessa cifra della status line. Non e il
numero di `/context`, che e una stima per categorie e puo non coincidere.

## Strade chiuse, provate e abbandonate

Tutte e tre provate davvero, non scartate a tavolino. Valgono piu delle feature: dicono dove la
superficie non arriva.

**Il pane non e interattivo, su questo terminale.** Non esiste un'azione per dare la tastiera a un
pane: `ctrl+x tab` e `abovePrompt:focus` e mette a fuoco la band, non il pane; e le azioni del
contesto `Pane` (`pane:close` su `ctrl+x x`) valgono solo quando il focus c'e gia, cioe mai. Senza
focus, un `Button` nel pane non si preme da tastiera, e il click non ha dato esito. **Quindi il
pane si legge e basta**: niente pulsanti, niente vista di un file espansa dentro il pane, niente
comandi con argomenti per supplire. Tutto il codice che serviva a quello e stato tolto, invece di
restare morto.

**I percorsi premibili dentro i messaggi.** Il contratto lo permette — il sito `AssistantMessage`
passa il testo, e l'elemento `Markdown` ha `onLinkPress` con `pressableLinks`. Ma il click su un
percorso e gia preso da Claude Code, che lo copia negli appunti: si vede `copied 47 chars to
clipboard`. Vincere quella corsa significherebbe togliere una funzione che esiste. In piu quell'hook
ridisegna **ogni** messaggio che contiene un percorso, ed e il pezzo piu invasivo che si possa
scrivere in una mod: molto rischio, nessun guadagno.

**Una scorciatoia propria.** Un plugin non puo portarsi dietro le proprie combinazioni: stanno in
`~/.claude/keybindings.json`, che e personale. La mod puo scriverci lei al primo avvio, prendendo
in prestito un'azione poco usata (`app:cycleDiffBase`, che di default vive solo nel contesto
`DiffPanel`) e legandola in `Global`: **provato, la riga viene scritta correttamente**. Ma non ha
dato esito nella sessione in corso, e soprattutto scrivere nella configurazione personale di chi
installa un plugin e una cosa seria, da dichiarare nel README e non da scoprire. Rimosso.

## Il rischio, e come si contiene

Il pane interpreta la forma della riga del requisito
(`- <a id="r1"></a>**R1** — …`), che ha meno di un mese e non ha ancora prodotto un overview reale.

**E il buco e gia aperto oggi, senza la mod.** Un censimento del repo ha trovato che il formato e
riconosciuto da codice in **tre posti gia esistenti**, non uno:

- `plugins/wm-skills/scripts/feature-docs-check.sh`, con otto agganci distinti (la riga del
  requisito, la sottovoce `Verifica`, la sottovoce `Non fatto`, la normalizzazione dei link);
- `plugins/wm-skills/scripts/tests/feature-docs-check.test.sh`, che e un riconoscitore a tutti gli
  effetti: contiene fixture verbatim con le ancore e una `sed` che dipende da `<a id="r2">`;
- il parser della mod, da scrivere.

Nessuno dei primi due e legato a quanto `wm-plan/SKILL.md` prescrive. Il controllo in CI quindi
non e una cautela per il lavoro nuovo: chiude un buco che esiste adesso.

**Il formato e doppio.** Non c'e solo la riga dell'overview: c'e la voce di `notes.md` che la
referenzia (`- **Requisito:** [R3](overview.md#r3)`), e lo script confronta le due una contro
l'altra. Un parser che leggesse solo l'overview coprirebbe meta contratto: per questo fra i
riferimenti che il pane esplode ci sono anche le ancore di `notes.md`.

**Il corpus esistente non e migrato.** `docs/features/` contiene 20 overview e **nessuno** usa il
formato con le ancore: sono tutti nella vecchia forma a checkbox. Il controllo in CI non deve
quindi validare gli overview del repo, o fallirebbe su tutti e venti: valida le righe d'esempio
estratte dallo SKILL.md, che e la forma gia scelta sopra. Se e quando migrare i venti file e una
decisione a se, fuori da questo lavoro.

Perciò:

- una riga nella tabella **Coupling tra skill** del `CLAUDE.md`: `wm-plan` ↔ la mod ↔
  `feature-docs-check.sh`, contratto = la forma della riga del requisito e dell'àncora;
*(Fatto: `.github/scripts/verifica-formato-requisito.sh` e il workflow omonimo.)*

- **un controllo in CI che non confronta regex con regex**, perche due espressioni diverse possono
  essere equivalenti e il confronto direbbe il falso. Prende le righe d'esempio **dallo SKILL.md**,
  che resta la fonte autoritativa, le scrive in un `overview.md` temporaneo e verifica che entrambi
  i parser — quello awk e quello della mod — riconoscano esattamente gli identificativi attesi.
  Insieme alle righe buone passa due contro-esempi che **non** devono essere riconosciuti, la
  vecchia forma `- [ ] R1 — …` e una riga senza àncora: senza quelli, un parser che accetta tutto
  passerebbe il controllo. Se l'estrazione dallo SKILL.md non trova nessuna riga, il controllo
  fallisce invece di passare: un controllo cieco e peggio di nessun controllo.
- Forma e aggancio sul modello dei due esistenti: `.github/scripts/verifica-formato-requisito.sh`
  in bash, con la testata «perche esiste» e il blocco `# LIMITE, da conoscere:`, errori accumulati
  in `ESITO` per vederli tutti in un giro; e un workflow suo,
  `.github/workflows/formato-requisito.yml`, perche il repo tiene un workflow per script e i
  `paths:` qui sono altri. A differenza degli altri due serve `actions/setup-node`, perche il lato
  TypeScript va eseguito: il parser della mod va quindi scritto come funzione pura, invocabile
  fuori dal pane.
- il pane non resta mai muto: quando non riconosce il file lo scrive.

## Cosa resta a chi implementa

- Sistemare, nello stesso lavoro, le virgolette mancanti attorno a `${CLAUDE_PLUGIN_ROOT}` nel
  comando dell'hook di lint gia presente: il validator le chiede, e il warning comparira comunque
  appena si tocca quel file.
- Come il pane interroga Orchestrator per `oc:<numero>` senza rallentare il disegno (la lettura
  dell'overview è locale, quella del ticket no): cache per sessione, e il pane che mostra il
  numero anche prima di avere il titolo.
- Aggiornare in `CLAUDE.md` la tabella dei comandi e la frase «In CI girano…», che elencano i
  controlli attivi.
- ~~I test~~ *(fatto: `plugins/wm-skills/mods/wm-plan-anchors/parse.test.ts`, 12 casi)*, con
  `claude plugin test <cartella>`, che gira ogni `*.test.ts`/`*.test.tsx` sotto la mod
  ed esce 1 al primo fallimento, quindi si aggancia alla CI cosi com'e. Il kit si importa da
  `claude-code/testing`. Da coprire: il parser su un overview col formato nuovo, su uno col vecchio
  formato a checkbox, su uno senza sezione Requisiti, e su spazzatura — un parser non deve mai
  lanciare. Poi il disegno del pane, montato su piu di una superficie, perche non dipenda dal
  terminale.
- **Nei test il filesystem non esiste**: `$.fs.*` non sta sul `$` del test. Si risponde dal basso
  registrando `on('fs.read')`, `on('fs.list')`, `on('fs.stat')` con `{ value }` o `{ deny }`, e si
  esercita il codice facendo partire un hook. Attenzione: i path arrivano **assoluti**, risolti
  sulla working directory, quindi vanno normalizzati o le asserzioni diventano fragili.
- Un `{ deny }` non gestito fa saltare l'hook, e il test fallisce dicendolo. E la rete che
  garantisce che i casi di errore del pane siano scritti davvero, non solo dichiarati.

## Fuori da questo lavoro

**La sezione con gli agenti al lavoro.** Era stata costruita nel prototipo — modello, percentuale
di context consumata, cosa sta facendo ciascuno, e una riga premibile che apre la sua trascrizione
con `$.ui.view({ agentId })`. E stata tolta guardandola: **Claude Code mostra gia nativamente**,
sotto il prompt, l'elenco con la conversazione principale, gli agenti, la durata e i token. Il pane
duplicava quella lista e sarebbe invecchiato per conto suo. Quello che la vista nativa non da — il
modello e la percentuale di context per agente — non vale una seconda lista da mantenere.

Nella stessa conversazione è emersa una seconda mod, più importante di questa e tenuta separata:
gli hook su `tool.call` che **fanno rispettare** i divieti oggi scritti in prosa nel `CLAUDE.md` —
niente `git commit` dentro un lavoro, nessuna scrittura verso `orchestrator.maphub.it`, i binari di
`bin/` non toccati a mano. Un hook nega, non raccomanda: quelle regole smettono di dipendere dalla
buona volontà del modello. Il dev ha scelto di partire dal pane; questa resta da aprire a parte.

## Com'e fatto il controllo in CI, alla fine

Non esegue i parser: li **legge**. Estrae dallo `SKILL.md` le righe d'esempio dei requisiti, estrae
la regex awk da `feature-docs-check.sh` e la costante `RIGA_REQUISITO` da `parse.ts`, traduce la
seconda in espressione regolare POSIX e verifica che entrambe riconoscano quelle righe.

Eseguirli sarebbe stato piu diretto, ma il lato TypeScript richiede Node 22 per girare senza
compilazione, e il repo non ha Node fra le sue dipendenze: sarebbe stata una dipendenza nuova in
CI per un controllo di tre righe. Cosi bastano bash, grep e sed, come per gli altri due controlli
del repo.

Insieme alle righe buone passa tre contro-esempi che non devono essere riconosciuti — il vecchio
formato a checkbox, un identificativo senza ancora, un'ancora che non comincia per `r` — e se
dallo `SKILL.md` non estrae nessuna riga fallisce invece di passare.

**Provato a fallire, non solo a passare:** cambiando la regex di `parse.ts` il controllo segnala
le tre righe non riconosciute; rendendola permissiva segnala i tre contro-esempi accettati.

Il limite, scritto nella testata dello script: verifica che i due parser riconoscano le stesse
righe, non che le interpretino allo stesso modo. Quello lo coprono i test del parser.

## Stato

Esiste un prototipo buttabile, fuori dal repo, in
`/Users/bongiu/.claude/dev-mods/396ef7bd-f27e-4061-9e68-fdcd0fb1714c/wm-plan-anchors/`:
copre i soli requisiti, legge un `overview.md` di esempio che si porta dentro, è stato caricato e
visto funzionare. Non è il lavoro: il lavoro è portarlo in `wm-skills` con i quattro riferimenti, i
test e il controllo in CI.
