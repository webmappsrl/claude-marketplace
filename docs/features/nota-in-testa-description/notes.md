# Notes — Aggiungere informazione alla description senza riscriverla

## Deviazioni dal piano

- **Nessun commit e nessun branch durante l'esecuzione**: il `CLAUDE.md` del repo li vieta
  dentro il lavoro. Il lavoro è rimasto sul working tree di `develop`, per scelta del dev.
- **Task 4 e 5 eseguiti direttamente, senza agente né review dedicata; la review del Task 3 è
  confluita in quella finale**: il dev ha chiesto di accorciare i tempi. La review finale ha
  coperto tutto il lavoro.
- **`compose.Apply` va oltre il piano** (review del Task 1): le frasi non si cercano dentro un
  `prepend` già presente in testa, un testo è «già presente» solo se dopo di lui c'è fine stringa
  o `<`, e il contesto dell'anteprima non spezza i caratteri multi-byte.
- **`preview.Readable` va oltre il piano** (review del Task 2): considera tag solo i nomi HTML
  noti (un testo come `n<5 e x>10` o `List<String>` resta visibile), conserva spazi e a capo
  dentro `<pre>`, e tiene sulla stessa riga del trattino un titolo dentro un `<li>`.
- **Correzioni dalla review finale**:
  - `Diff` confronta i valori grezzi: se cambia solo il markup lo dice, invece di scrivere
    «Nessuna modifica» e poi mandare la PATCH;
  - la resa come testo leggibile vale solo per i campi HTML delle story: la `description` dei
    tag, che è Markdown, resta com'è;
  - dopo la conferma la risposta è una sintesi breve, senza ripetere l'anteprima;
  - un avviso sotto `customer_request` ricorda che Orchestrator lo mette in testa e notifica il
    cliente;
  - tutte le anteprime dei tool di scrittura sono testo semplice con a capo veri.

## Decisioni

- **Niente intestazione di autore e data** aggiunta dal tool: l'aveva proposta Claude, il dev
  l'ha tolta perché nessuno l'aveva chiesta.
- **Un solo tool, `update_story`, che sceglie in base ai parametri**: `description` sostituisce,
  `prepend`/`annotations` aggiungono, i due insieme vengono rifiutati. Scelta del dev, dopo aver
  scartato un tool separato e una bozza tenuta in memoria dal server.
- **Nessun controllo nelle skill per un binario vecchio**: skill e binario viaggiano nello stesso
  plugin, il disallineamento capita solo aggiornando a sessione aperta, e si risolve riavviando.
- **Vince chi scrive per ultimo**: nessuna protezione dalle scritture contemporanee.
- **Nessuna verifica periodica del comportamento di Orchestrator** (che la PATCH sostituisca):
  scartata come rischio ipotetico.

- **La challenge di `wm-plan` distingue rischi reali e ipotetici** (aggiunta a fine lavoro, su
  richiesta del dev): in questa sessione scenari improbabili portati come domande hanno fatto
  cambiare il progetto più volte. Il revisore ora classifica ogni punto, e al dev arrivano come
  domande solo i rischi reali; gli ipotetici una riga ciascuno, da ignorare salvo sua richiesta.

- **`wm-plan` ha una regola «una domanda per messaggio» valida in tutto il workflow** (aggiunta
  a fine lavoro, su richiesta del dev): prima valeva solo in `reverse-interaction`, e in questa
  sessione i messaggi con più cose da decidere hanno fatto leggere come deleghe i silenzi del dev.

## Follow-up

Minori rimandati dalle review, nessuno blocca:

- messaggio fuorviante quando si riapplicano annotazioni sovrapposte (non scrive nulla, ma dice
  «non compare» invece di «già presente»);
- il contesto di un'annotazione nell'anteprima può essere un blocco non bilanciato o includere il
  `prepend`;
- `insideTag` non riconosce un'entità spezzata, un `>` dentro un attributo o un commento HTML;
- le tabelle sono rese con una riga vuota fra una riga e l'altra;
- l'avviso sul testo perso conta i caratteri sull'HTML grezzo, non sul testo reso;
- dentro `<pre>` tre a capo di fila diventano due;
- tag sconosciuti e commenti HTML compaiono letteralmente nell'anteprima;
- la Checklist di `wm-plan` chiede due approvazioni di fila (bozze, poi anteprima): era già così.
