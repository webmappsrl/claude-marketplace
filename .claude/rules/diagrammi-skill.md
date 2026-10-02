---
paths:
  - "docs/guide/wm-plan-diagramma/**"
  - "docs/guide/wm-geohub-import-check-diagramma/**"
  - "plugins/wm-skills/skills/wm-plan/SKILL.md"
  - "plugins/wm-skills/skills/wm-geohub-import-check/SKILL.md"
  - ".github/scripts/verifica-diagramma.sh"
---

# Diagrammi di flusso delle skill

Due skill hanno un diagramma pubblicato, con la stessa struttura (due colonne di pari altezza,
legenda a piena larghezza sotto) e lo stesso controllo in CI:

| Skill | Sorgente | Pagina |
|---|---|---|
| `wm-plan` | `docs/guide/wm-plan-diagramma/index.html` | <https://webmappsrl.github.io/claude-marketplace/wm-plan-diagramma/> |
| `wm-geohub-import-check` | `docs/guide/wm-geohub-import-check-diagramma/index.html` | <https://webmappsrl.github.io/claude-marketplace/wm-geohub-import-check-diagramma/> |

Una fase si dichiara nello `SKILL.md` come `## Fase: <nome>`, con un nome senza spazi: il
controllo in CI riconosce solo lettere minuscole e trattini.

**Lo stile è diverso.** Il diagramma di `wm-plan` ha la palette congelata di oc:8283. Quello di
`wm-geohub-import-check` usa il design system Webmapp (artifact «Webmapp», `tokens.json`): un solo
tema chiaro, titoli in Montserrat Medium `petrolio` #005485, testo in Roboto, spigoli vivi e niente
ombre; nodi `petrolio` per le fasi, `logo-arancio` con testo nero per i blocchi, `logo-malva` per
l'agente. Il blocco di CSS in fondo al `<style>` e i colori di Mermaid nello script vanno
cambiati insieme. `logo-arancio` non va mai usato come colore di testo (2,3:1 sul bianco).

## Il diagramma di wm-plan

Il diagramma del workflow è una pagina pubblicata su GitHub Pages:
<https://webmappsrl.github.io/claude-marketplace/wm-plan-diagramma/>

Il sorgente è `docs/guide/wm-plan-diagramma/index.html`, pubblicato dal workflow
`.github/workflows/pages.yml` — che pubblica **solo `docs/guide/`**, perché il resto di `docs/`
è interno.

## Dopo una modifica al workflow di una delle due skill, aggiorna il sorgente

La pagina si ripubblica da sola al push su `main`: nessun redeploy manuale, nessun URL da
riscrivere. Prima viveva come Artifact Claude e ogni aggiornamento era un gesto a mano.

**Il template grafico è congelato** in entrambe le pagine: due colonne di pari altezza, legenda
a piena larghezza sotto, e lo stile di ciascuna — la palette di oc:8283 per `wm-plan`, il design
system Webmapp per `wm-geohub-import-check`. Si aggiorna il **contenuto** — nodi e paragrafi,
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
nello `SKILL.md` di ciascuna skill con i nodi del suo diagramma e fallisce se non coincidono.

Confronta **i nomi**: se una fase cambia comportamento mantenendo il titolo, il controllo passa
e il diagramma resta vecchio. Quella parte resta responsabilità di chi modifica la skill.
