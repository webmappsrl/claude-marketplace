# Verifica di una citazione dalle call

Letto da `wm-plan` (`reverse-interaction`) e `wm-tag` (`tag-creation`). È il controllo esterno
sulle citazioni, indipendente da chi le ha prodotte: si fa prima di usare una citazione.

1. Prendi una frase di cinque-otto parole **dentro una sola battuta** della citazione: una battuta
   può andare a capo, e una frase che attraversa l'a capo o unisce due battute non si trova.
2. Cerca con `search_files` di Drive:
   `parentId = '<cartella della fonte>' and fullText contains '"<frase>"'`, con
   `excludeContentSnippets: true`. Un apostrofo nella frase va scritto `\'` (`un\'idea`): senza, la
   ricerca risponde `Invalid query`.

   **La cartella dello scrum non ha un id fisso**: Meet ne crea una nuova con lo stesso titolo ogni
   volta che la serie ricorrente viene ricreata, e le call di prima restano in quella vecchia.
   Risolvile con una ricerca, come fa `wm-transcript-research`:

   `parentId = '1cqWMB7tCkX1IcF2hOwgBIdqad__WUEPT' and mimeType = 'application/vnd.google-apps.folder' and title contains 'scr-umwe-bmapp'`

   Ordinate per `createdTime`, ciascuna copre dalla propria creazione a quella della successiva.
   La citazione che stai verificando ha la data nel titolo della sua fonte: cerca la frase **nella
   sola cartella che copre quella data**. Per una fonte fuori da queste cartelle ometti `parentId`
   e controlla l'id fra i risultati.
3. Esiti:
   - **l'id della fonte citata è fra i risultati:** la citazione è attribuita bene;
   - **compare un altro id:** la citazione è attribuita alla call sbagliata; non usarla così com'è,
     dillo al dev con la call giusta;
   - **nessun risultato su una call di oggi:** Drive può non averla ancora indicizzata; è «verifica
     non possibile», non una citazione falsa;
   - **nessun risultato su una call passata:** la citazione non si trova; non usarla, e dillo al dev.

Costa poche centinaia di token a citazione e non legge il documento.
