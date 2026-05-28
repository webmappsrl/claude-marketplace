> Ticket: oc:7973

# Notes — Ask user to set ticket status to progress when work begins in wm-plan skill

## Deviazioni dal piano
Nessuna deviazione rilevante.

## Bug trovati
Nessuno.

## Decisioni

- **Modifiche a posteriori incluse nello stesso commit:** durante l'implementazione sono state richieste tre modifiche non previste nel piano originale:
  1. Spostare la scrittura di `notes.md` e l'aggiornamento di `CLAUDE.md` (Fase 7 e 8) **prima** del commit finale, così tutti i file vengono committati insieme.
  2. Nella bozza `description` del ticket (note dev), il link alla cartella docs deve essere un tag HTML `<a href="...">` cliccabile, ricavando l'URL del repo da `git remote get-url origin` con branch fisso `main`.
  3. Le modifiche richieste a posteriori (dopo approvazione del piano) devono essere registrate nella sezione "Decisioni" di `notes.md` — aggiunta come regola esplicita in Fase 7.

## Follow-up
Nessuno.
