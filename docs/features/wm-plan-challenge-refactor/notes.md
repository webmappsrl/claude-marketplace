# Notes — wm-plan Challenge refactor

## Deviazioni dal piano
Nessuna deviazione rilevante.

## Bug trovati
- La skill originale non esplicitava l'obbligo di creare documentazione in assenza di ticket — buco scoperto durante la sessione di refactor.

## Decisioni
- Il subagente Challenge legge `overview.md` dal filesystem invece di riceverlo nel prompt: unico modo per garantire isolamento reale del contesto.
- Prompt avversariale esplicito necessario perché senza framing il modello tende a produrre analisi equilibrate invece che critiche.
- Challenge spostata dopo `overview.md` per evitare confirmation bias: Claude ha già investito contesto nel costruire la soluzione e tende a razionalizzare i rischi invece di esporli.

## Follow-up
Nessuno.
