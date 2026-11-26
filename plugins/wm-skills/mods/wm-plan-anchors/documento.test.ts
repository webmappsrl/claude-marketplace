import { describe, expect, test } from 'claude-code/testing'

import { indicizza, indicizzaTutti, riferimenti, slug } from './documento.ts'
import { testaOverview } from './parse.ts'

/**
 * Le prime righe dell'overview di oc:9056, copiate come stanno.
 *
 * È l'overview su cui il pane è rimasto vuoto: i requisiti sono heading di
 * terzo livello, non voci di elenco. Sta qui perché il caso che ha rotto la
 * mod resti provato, e non ci si torni sopra.
 */
const HEADING = [
  '# oc:9056 — Il fascicolo si compila in più riprese',
  '',
  '> Stato: approvata il 26/11/2026',
  '',
  '## Requisiti',
  '',
  '### <a id="r1"></a>**R1** — La creazione chiede tre campi',
  '',
  'Il SUS apre una domanda mandando `reference`, `name` e `geometry`, senza dover avere già',
  'compilato il resto del modulo.',
  '',
  '**Verifica:** una creazione con i soli `reference`, `name` e `geometry` risponde `201`.',
  '',
  '### <a id="r2"></a>**R2** — Il riferimento del SUS resta legato alla domanda',
  '',
  '`reference` è una colonna della tabella, indicizzata e unica.',
  '',
  '## Rischi',
  '',
  'Qui non ci sono requisiti.',
].join('\n')

/** Lo stesso contenuto nella forma a elenco, quella che SKILL.md prescrive. */
const ELENCO = [
  '## Requisiti',
  '',
  '- <a id="r1"></a>**R1** — La creazione chiede tre campi',
  '  - Verifica: una creazione con i soli tre campi risponde 201.',
  '- <a id="r2"></a>**R2** — Il riferimento del SUS resta legato alla domanda',
  '',
  '## Rischi',
].join('\n')

describe('indicizza', () => {
  test('riconosce i requisiti scritti come heading', () => {
    const indice = indicizza(HEADING, 'overview.md')

    expect(indice.has('r1')).toBe(true)
    expect(indice.get('R1')?.titolo).toBe('La creazione chiede tre campi')
    expect(indice.get('R1')?.sigla).toBe('R1')
    expect(indice.get('R1')?.livello).toBe(3)
  })

  test('riconosce i requisiti scritti come voci di elenco', () => {
    const indice = indicizza(ELENCO, 'overview.md')

    expect(indice.get('R1')?.titolo).toBe('La creazione chiede tre campi')
    expect(indice.get('R1')?.livello).toBe(0)
  })

  test('le due forme danno lo stesso indice', () => {
    const a = indicizza(HEADING, 'overview.md')
    const b = indicizza(ELENCO, 'overview.md')

    expect(a.get('R1')?.titolo).toBe(b.get('R1')?.titolo)
    expect(a.get('R2')?.titolo).toBe(b.get('R2')?.titolo)
  })

  test("il corpo di un heading arriva fino al successivo di pari rango", () => {
    const corpo = indicizza(HEADING, 'overview.md').get('R1')?.corpo ?? ''

    expect(corpo).toContain('senza dover avere già')
    expect(corpo).toContain('**Verifica:**')
    expect(corpo).not.toContain('Il riferimento del SUS')
  })

  test('il corpo di una voce di elenco si ferma alla voce dopo', () => {
    const corpo = indicizza(ELENCO, 'overview.md').get('R1')?.corpo ?? ''

    expect(corpo).toContain('Verifica')
    expect(corpo).not.toContain('Il riferimento del SUS')
  })

  test('un heading senza àncora esplicita si indicizza con il suo slug', () => {
    const indice = indicizza(HEADING, 'overview.md')

    expect(indice.has('rischi')).toBe(true)
    expect(indice.get('rischi')?.corpo).toContain('Qui non ci sono requisiti')
  })

  test('una sigla qualunque vale, non solo R', () => {
    const md = ['## T4 — La sovrapposizione vera non si riproduce', '', 'Il test prova il ramo che conta.'].join('\n')
    const indice = indicizza(md, 'plan.md')

    expect(indice.get('T4')?.titolo).toBe('La sovrapposizione vera non si riproduce')
    expect(indice.get('T4')?.file).toBe('plan.md')
  })
})

describe('indicizzaTutti', () => {
  test("il primo documento vince sulle chiavi che si ripetono", () => {
    const indice = indicizzaTutti([
      { file: 'overview.md', md: ELENCO },
      { file: 'plan.md', md: '### <a id="r1"></a>**R1** — un altro R1' },
    ])

    expect(indice.get('R1')?.file).toBe('overview.md')
  })
})

describe('riferimenti', () => {
  const indice = indicizzaTutti([
    { file: 'overview.md', md: HEADING },
    { file: 'plan.md', md: '## T4 — La sovrapposizione vera non si riproduce\n\nIl test prova il ramo che conta.' },
  ])

  test('trova le sigle nude e i link alle àncore, in ordine di testo', () => {
    const frase = 'Visto R2, poi R1 e infine [il piano](plan.md#t4--la-sovrapposizione-vera-non-si-riproduce).'
    const esito = riferimenti(frase, indice).map(uno => uno.voce.sigla)

    expect(esito).toEqual(['R2', 'R1', 'T4'])
  })

  test('una sigla che non è nei documenti non diventa una voce', () => {
    expect(riferimenti('il ticket R99 non esiste', indice)).toEqual([])
  })

  test('un link a uno slug lungo ricade sulla sigla che lo apre', () => {
    const esito = riferimenti('vedi [qui](plan.md#t4-qualunque-coda)', indice)

    expect(esito[0]?.voce.sigla).toBe('T4')
  })
})

describe('slug', () => {
  test('come lo calcola GitHub', () => {
    expect(slug('**R1** — La creazione chiede tre campi')).toBe('r1--la-creazione-chiede-tre-campi')
    expect(slug('Perché, davvero?')).toBe('perché-davvero')
  })
})

describe("l'ID del ticket quando la riga di testa non c'è", () => {
  test('lo prende dal titolo', () => {
    expect(testaOverview('# oc:9056 — Il fascicolo\n\n> Stato: approvata').ticket).toBe('oc:9056')
  })

  test('lo prende dal nome della cartella', () => {
    const esito = testaOverview('# Il fascicolo\n', 'docs/features/9056-il-fascicolo')
    expect(esito.ticket).toBe('oc:9056')
  })

  test('la riga di testa vince sugli altri due', () => {
    const md = '# oc:1111 — titolo\n> Ticket: oc:9056\n'
    expect(testaOverview(md, 'docs/features/2222-x').ticket).toBe('oc:9056')
  })

  test('senza ID da nessuna parte non se lo inventa', () => {
    expect(testaOverview('# Un titolo qualunque\n', 'docs/features/senza-numero').ticket).toBe(undefined)
  })
})
