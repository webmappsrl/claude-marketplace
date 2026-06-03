# Notes — wm-plan UX/UI Skill Detection

## Deviazioni dal piano
- La sottosezione UX detection è stata chiamata `2c` invece di essere inserita tra `2a` e `2b` — scelta per mantenere la numerazione esistente senza rinominare `2b`.

## Decisioni
- Il parere UX restituito dalla skill popola direttamente le sezioni "Requisiti" e "Rischi" di `overview.md` con prefisso `[UX]`, senza un passo intermedio di conferma utente — il developer vede comunque tutto all'approvazione dell'overview.
- Non è stata aggiunta logica UX in Fase 5 (writing-plans): i requisiti `[UX]` in `overview.md` vengono già raccolti automaticamente nel briefing.
- Il flusso per la skill non trovata include il comando esatto di installazione (`/plugin install ui-ux-pro-max@wm-marketplace`) per minimizzare l'attrito.

## Follow-up
- Valutare se aggiungere anche `frontend-design` (Anthropic) come skill UX alternativa per progetti React.
- Verificare compatibilità effettiva del plugin `nextlevelbuilder/ui-ux-pro-max-skill` con il sistema di plugin Claude Code dopo installazione reale.
