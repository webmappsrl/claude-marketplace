# Verifica di una citazione dalle call

Letto da `wm-plan` (`reverse-interaction`) e `wm-tag` (`tag-creation`). È il controllo esterno
sulle citazioni, indipendente da chi le ha prodotte: si fa prima di usare una citazione.

1. Prendi una frase di cinque-otto parole **dentro una sola battuta** della citazione: una battuta
   può andare a capo, e una frase che attraversa l'a capo o unisce due battute non si trova.
2. Cerca con `search_files` di Drive:
   `parentId = '<cartella della fonte>' and fullText contains '"<frase>"'`, con
   `excludeContentSnippets: true`. Un apostrofo nella frase va scritto `\'` (`un\'idea`): senza, la
   ricerca risponde `Invalid query`.

   **Le cartelle dello scrum non hanno un id fisso**, e non è detto che siano una sola: Meet ne crea
   una nuova con lo stesso titolo ogni volta che la serie ricorrente viene ricreata, e le call
   vecchie restano in quella di prima. Risolvile con una ricerca, come fa
   `wm-transcript-research`:

   `parentId = '1cqWMB7tCkX1IcF2hOwgBIdqad__WUEPT' and mimeType = 'application/vnd.google-apps.folder' and title contains 'scr-umwe-bmapp'`

   poi cerca la frase in tutte quelle che tornano:
   `(parentId = '<id 1>' or parentId = '<id 2>' …) and fullText contains '"<frase>"'`. Cercando in
   una cartella sola, una citazione vera presa da una call che sta nell'altra risulta «non trovata».
   Per una fonte fuori da queste cartelle ometti `parentId` e controlla l'id fra i risultati.
3. Esiti:
   - **l'id della fonte citata è fra i risultati:** la citazione è attribuita bene;
   - **compare un altro id:** la citazione è attribuita alla call sbagliata; non usarla così com'è,
     dillo al dev con la call giusta;
   - **nessun risultato su una call di oggi:** Drive può non averla ancora indicizzata; è «verifica
     non possibile», non una citazione falsa;
   - **nessun risultato su una call passata:** la citazione non si trova; non usarla, e dillo al dev.

Costa poche centinaia di token a citazione e non legge il documento.
