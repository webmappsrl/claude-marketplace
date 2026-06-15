> Ticket: oc:8068

# Notes — wm-review-ticket

## Deviazioni dal piano

Nessuna deviazione rilevante.

## Decisioni

- **Contratto artefatti via WebFetch invece di duplicazione**: `wm-review-ticket` non duplica la struttura `docs/features/` ma la legge direttamente da GitHub raw al momento dell'esecuzione. Garantisce che le due skill rimangano sincronizzate senza aggiornamenti manuali.
- **`wm-plan` come fonte autoritativa**: anziché documentare il contratto in entrambe le skill o in un file condiviso (path instabile), `wm-plan` è l'unica fonte e `wm-review-ticket` la referenzia. La tabella coupling in `CLAUDE.md` rende esplicito dove intervenire se la struttura cambia.
- **Review opzionale in Fase 6d di wm-plan**: il suggerimento di invocare `wm-review-ticket` è formulato come domanda esplicita sì/no — non un hint passivo. Deciso durante l'implementazione su richiesta dell'utente.

## Follow-up

- Effort configurabile (low/medium/high) — lasciato fuori scope, da valutare in ticket separato
- Input diretto con URL PR senza ticket — fuori scope, aggiungibile in futuro
- Test smoke su ticket reale con PR associata — da fare dopo il merge
