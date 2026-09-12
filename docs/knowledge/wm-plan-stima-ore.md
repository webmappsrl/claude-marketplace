# Stima delle ore in wm-plan

## Stato attuale

`Fase: estimation` gira **solo per le Feature**: i bug non si stimano in ore, perché il costo
sta nella diagnosi e non nella fix, e solo le Feature ricevono `estimated_hours`.

La stima è **per componente**, non con un buffer forfettario. Ogni componente è classificato
come *scrittura pura* (nessuna domanda aperta dopo overview e challenge, buffer 0%) o
*decisioni aperte* (UX, reverse-engineering di legacy, buffer 20-30%). Sopra il totale si
aggiunge un **5% di integrazione trasversale**, che copre il rischio di interazione fra
componenti che nessun componente singolo cattura.

Il tempo di pianificazione è **misurato, non stimato**: `Fase: ticket` registra
`planning_start_at` e il confronto avviene a fine `Fase: estimation`. Al dev si mostra sempre
"Misurato + Stimato = Totale", mai un numero unico fuso.

Ogni stima scritta su Orchestrator porta il marcatore `[stima v2 — per-componente]`.

Se durante l'esecuzione emerge un imprevisto stimabile, `execution: re-estimation` propone al
dev una revisione con conferma esplicita, prima del PATCH di `estimated_hours`.

## Come ci siamo arrivati

- **Il buffer forfettario è caduto** (oc:8278): un'analisi su Orchestrator dei ticket di luglio
  2026 mostrava una sovrastima sistematica proprio sui task ben specificati — quelli in cui il
  buffer non serviva.
- **Il marcatore di versione serve alle calibrazioni future** (oc:8278): senza, i dati storici
  mescolerebbero criterio vecchio e nuovo e nessuna analisi successiva sarebbe attendibile.
- **Il coefficiente di velocità per-dev è rimandato** (oc:8278): Orchestrator non espone oggi un
  endpoint di aggregazione stimato-contro-effettivo per utente, e il campione per dev (8-15
  ticket) è troppo piccolo per un coefficiente affidabile.
