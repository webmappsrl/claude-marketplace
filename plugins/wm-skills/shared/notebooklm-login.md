# Rifare l'accesso a NotebookLM

Da leggere quando `wm-transcript-research` risponde `RICERCA FALLITA` per credenziali scadute
(«Authentication expired»). Guida il dev un passo per volta: ogni comando lo lancia lui, e dopo
ogni passo la verifica la fai tu.

## La verifica è una sola: `nlm notebook list`

**Non fidarti dell'uscita di `nlm login`.** Se esiste già un profilo salvato, `nlm login` non rifà
l'accesso: controlla i cookie salvati e, quando non riesce a verificarli con Google, li dichiara
comunque buoni:

```
✓ Authentication valid!
  Notebook count unavailable (network slow); credentials are valid
```

Con quell'uscita i cookie possono essere scaduti lo stesso, e ogni chiamata successiva fallisce con
«Authentication expired». Lo stesso vale per il tool `refresh_auth` del server MCP: ricarica dal
disco i cookie che trova, e risponde `success` anche se sono scaduti.

L'unico controllo affidabile è una chiamata vera:

```bash
nlm notebook list
```

Se restituisce l'elenco dei notebook l'accesso funziona. Se restituisce «Authentication Error»,
non funziona, qualunque cosa abbia detto `nlm login`.

## Procedura

1. **Controlla la versione** con `nlm --version`. Se non è l'ultima, l'aggiornamento è il primo
   passo e spesso è anche la soluzione: con la 0.9.4 il login risultava valido e Google rifiutava i
   cookie; con la 0.15.4 ha funzionato al primo tentativo. Il dev lancia:

   ```
   uv tool upgrade notebooklm-mcp-cli
   ```

2. **Cancella il profilo salvato**, altrimenti `nlm login` si limita a ricontrollarlo. Con il
   prefisso `!` serve `-y`: la domanda di conferma non riceve risposta e il comando si interrompe
   («Aborted»).

   ```
   ! nlm login profile delete default -y
   ```

3. **Il dev rifà l'accesso da un terminale normale, fuori da Claude Code:**

   ```bash
   nlm login
   ```

   Due motivi per non usare `!`. Le versioni recenti chiedono dove salvare le credenziali, e quella
   domanda con `!` non riceve risposta. Inoltre, lanciato con `!`, il login ha chiuso la finestra
   di Chrome prima che si potesse entrare.

   Cosa dire al dev:
   - alla domanda «Where should your saved login live?» scegliere **1 (Protected)**: le
     credenziali restano cifrate, con la chiave nel Portachiavi di macOS. Se poi macOS chiede il
     permesso di accedere al Portachiavi, rispondere «Consenti sempre»;
   - l'accesso va fatto **nella finestra di Chrome che `nlm` apre**, con l'account @webmapp.it.
     `nlm` usa un profilo Chrome suo (`~/.notebooklm-mcp-cli/chrome-profiles`), separato dal
     browser di tutti i giorni: essere collegati a Google nel proprio Chrome non serve.

   Un login riuscito estrae decine di cookie (`Cookies: 56 extracted`). **`Cookies: 1 extracted`
   vuol dire che l'accesso a Google non è avvenuto**, anche se la riga prima dice «Successfully
   authenticated».

4. **Verifica tu** con `nlm notebook list`.

5. **Ricollega il server MCP.** Dopo un aggiornamento della versione `refresh_auth` non basta: il
   server in esecuzione è ancora quello vecchio, non sa leggere le credenziali nel nuovo formato e
   risponde `not_configured`. Il dev lancia `/mcp`, seleziona `plugin:wm-skills:notebooklm` e
   sceglie **Reconnect**. Senza aggiornamento basta chiamare `refresh_auth`.

6. **Verifica il server** chiamando il tool `notebook_list`. Solo quando risponde, rilancia
   `wm-transcript-research` («prepara», o la ricerca che era fallita).

## Quando fermarsi

Se dopo l'aggiornamento e un login da terminale normale `nlm notebook list` fallisce ancora, non
insistere: proponi al dev di proseguire senza le trascrizioni. Resta la modalità manuale
(`nlm login --manual -f <file>`, con i cookie copiati dal proprio Chrome), da fare fuori dal
workflow.
