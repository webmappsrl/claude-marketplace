---
paths:
  - "docs/guide/wm-plan-diagramma/**"
  - "plugins/wm-skills/skills/wm-plan/SKILL.md"
  - ".github/scripts/verifica-diagramma.sh"
---

# Diagramma di flusso di wm-plan

Il diagramma del workflow è una pagina pubblicata su GitHub Pages:
<https://webmappsrl.github.io/claude-marketplace/wm-plan-diagramma/>

Il sorgente è `docs/guide/wm-plan-diagramma/index.html`, pubblicato dal workflow
`.github/workflows/pages.yml` — che pubblica **solo `docs/guide/`**, perché il resto di `docs/`
è interno.

## Dopo una modifica al workflow di wm-plan, aggiorna il sorgente

La pagina si ripubblica da sola al push su `main`: nessun redeploy manuale, nessun URL da
riscrivere. Prima viveva come Artifact Claude e ogni aggiornamento era un gesto a mano.

**Il template grafico è congelato** (oc:8283): due colonne di pari altezza, legenda a piena
larghezza sotto, palette e tipografia date. Si aggiorna il **contenuto** — nodi e paragrafi,
quando una fase viene aggiunta, rinominata o rimossa — mai struttura, CSS o stile.

## Due cose che si rompono in silenzio

**Il runtime Mermaid va importato come modulo.** Dalla versione 11 Mermaid distribuisce solo
build ESM, che non definiscono alcun globale: caricato con uno `<script src>` classico non dà
errore e non disegna nulla — il riquadro resta vuoto e sembra un problema di CSS.

**L'SVG non va lasciato adattare al contenitore.** È largo circa 1600 unità dentro un riquadro
da 500: scalandolo per entrare in larghezza viene poi centrato verticalmente, con centinaia di
pixel di vuoto sopra il disegno. `max-width: none` e `height: auto` lo tengono a grandezza
naturale, e lo scorrimento orizzontale fa il resto.

## Il controllo in CI vede solo i nomi

`.github/scripts/verifica-diagramma.sh` (workflow `coerenza.yml`) confronta le fasi dichiarate
in `SKILL.md` con i nodi del diagramma e fallisce se non coincidono.

Confronta **i nomi**: se una fase cambia comportamento mantenendo il titolo, il controllo passa
e il diagramma resta vecchio. Quella parte resta responsabilità di chi modifica la skill.
