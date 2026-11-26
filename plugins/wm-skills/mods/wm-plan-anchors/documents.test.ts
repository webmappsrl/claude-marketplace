import { describe, expect, test } from 'claude-code/testing'

import {
  citedPaths,
  citedTickets,
  indexAll,
  indexDocument,
  namedCodes,
  references,
  slug,
} from './documents.ts'
import { overviewHead } from './parse.ts'

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
const LIST = [
  '## Requisiti',
  '',
  '- <a id="r1"></a>**R1** — La creazione chiede tre campi',
  '  - Verifica: una creazione con i soli tre campi risponde 201.',
  '- <a id="r2"></a>**R2** — Il riferimento del SUS resta legato alla domanda',
  '',
  '## Rischi',
].join('\n')

describe('indexDocument', () => {
  test('riconosce i requisiti scritti come heading', () => {
    const index = indexDocument(HEADING, 'overview.md')

    expect(index.has('r1')).toBe(true)
    expect(index.get('R1')?.title).toBe('La creazione chiede tre campi')
    expect(index.get('R1')?.code).toBe('R1')
    expect(index.get('R1')?.level).toBe(3)
  })

  test('riconosce i requisiti scritti come voci di elenco', () => {
    const index = indexDocument(LIST, 'overview.md')

    expect(index.get('R1')?.title).toBe('La creazione chiede tre campi')
    expect(index.get('R1')?.level).toBe(0)
  })

  test('le due forme danno lo stesso indice', () => {
    const a = indexDocument(HEADING, 'overview.md')
    const b = indexDocument(LIST, 'overview.md')

    expect(a.get('R1')?.title).toBe(b.get('R1')?.title)
    expect(a.get('R2')?.title).toBe(b.get('R2')?.title)
  })

  test("il corpo di un heading arriva fino al successivo di pari rango", () => {
    const body = indexDocument(HEADING, 'overview.md').get('R1')?.body ?? ''

    expect(body).toContain('senza dover avere già')
    expect(body).toContain('**Verifica:**')
    expect(body).not.toContain('Il riferimento del SUS')
  })

  test('il corpo di una voce di elenco si ferma alla voce dopo', () => {
    const body = indexDocument(LIST, 'overview.md').get('R1')?.body ?? ''

    expect(body).toContain('Verifica')
    expect(body).not.toContain('Il riferimento del SUS')
  })

  test('un heading senza àncora esplicita si indicizza con il suo slug', () => {
    const index = indexDocument(HEADING, 'overview.md')

    expect(index.has('rischi')).toBe(true)
    expect(index.get('rischi')?.body).toContain('Qui non ci sono requisiti')
  })

  test('una sigla qualunque vale, non solo R', () => {
    const md = ['## T4 — La sovrapposizione vera non si riproduce', '', 'Il test prova il ramo che conta.'].join('\n')
    const index = indexDocument(md, 'plan.md')

    expect(index.get('T4')?.title).toBe('La sovrapposizione vera non si riproduce')
    expect(index.get('T4')?.file).toBe('plan.md')
  })
})

describe('indexAll', () => {
  test("il primo documento vince sulle chiavi che si ripetono", () => {
    const index = indexAll([
      { file: 'overview.md', md: LIST },
      { file: 'plan.md', md: '### <a id="r1"></a>**R1** — un altro R1' },
    ])

    expect(index.get('R1')?.file).toBe('overview.md')
  })
})

describe('references', () => {
  const index = indexAll([
    { file: 'overview.md', md: HEADING },
    { file: 'plan.md', md: '## T4 — La sovrapposizione vera non si riproduce\n\nIl test prova il ramo che conta.' },
  ])

  test('trova le sigle nude e i link alle àncore, in ordine di testo', () => {
    const sentence = 'Visto R2, poi R1 e infine [il piano](plan.md#t4--la-sovrapposizione-vera-non-si-riproduce).'
    const got = references(sentence, index).map(one => one.entry.code)

    expect(got).toEqual(['R2', 'R1', 'T4'])
  })

  test('una sigla che non è nei documenti non diventa una voce', () => {
    expect(references('il ticket R99 non esiste', index)).toEqual([])
  })

  test('un link a uno slug lungo ricade sulla sigla che lo apre', () => {
    const got = references('vedi [qui](plan.md#t4-qualunque-coda)', index)

    expect(got[0]?.entry.code).toBe('T4')
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
    expect(overviewHead('# oc:9056 — Il fascicolo\n\n> Stato: approvata').ticket).toBe('oc:9056')
  })

  test('lo prende dal nome della cartella', () => {
    const got = overviewHead('# Il fascicolo\n', 'docs/features/9056-il-fascicolo')
    expect(got.ticket).toBe('oc:9056')
  })

  test('la riga di testa vince sugli altri due', () => {
    const md = '# oc:1111 — titolo\n> Ticket: oc:9056\n'
    expect(overviewHead(md, 'docs/features/2222-x').ticket).toBe('oc:9056')
  })

  test('senza ID da nessuna parte non se lo inventa', () => {
    expect(overviewHead('# Un titolo qualunque\n', 'docs/features/senza-numero').ticket).toBe(undefined)
  })
})

describe('citedPaths', () => {
  test('prende i link e i path nudi, senza doppioni', () => {
    const text = 'Vedi [le note](docs/features/x/notes.md#r1) e anche docs/features/x/plan.md, poi notes.md.'
    expect(citedPaths(text).sort()).toEqual([
      'docs/features/x/notes.md',
      'docs/features/x/plan.md',
      'notes.md',
    ])
  })

  test('un testo senza markdown non produce path', () => {
    expect(citedPaths('nessun documento qui')).toEqual([])
  })
})

describe('namedCodes', () => {
  test('dice tutto quello che è stato nominato, risolto o no', () => {
    expect(namedCodes('T1, T3 e R99 sono citati').sort()).toEqual(['R99', 'T1', 'T3'])
  })
})

describe('citedTickets', () => {
  test('prende il numero dei ticket nominati, senza doppioni', () => {
    expect(citedTickets('fatto in oc:8333, vedi anche oc:9056 e di nuovo oc:8333').sort()).toEqual([
      '8333',
      '9056',
    ])
  })

  test('un numero che non è un ticket non conta', () => {
    expect(citedTickets('la riga 8333 del file')).toEqual([])
  })
})

describe('una sigla si qualifica con il ticket nominato nella stessa frase', () => {
  const index = indexAll([
    { file: 'docs/features/9056-corrente/overview.md', md: '### <a id="r2"></a>**R2** — quello di 9056' },
    { file: 'docs/features/8968-altro/overview.md', md: '### <a id="r2"></a>**R2** — quello di 8968' },
  ])

  test('il ticket nominato prima della sigla vince', () => {
    const got = references('fatto in oc:8968, vedi il suo R2', index)
    expect(got[0]?.entry.title).toBe('quello di 8968')
  })

  test('dopo il punto la qualificazione non vale più', () => {
    const got = references('fatto in oc:8968. Qui siamo su R2', index)
    expect(got[0]?.entry.title).toBe('quello di 9056')
  })

  test('senza ticket nominato vince il primo documento', () => {
    expect(references('siamo su R2', index)[0]?.entry.title).toBe('quello di 9056')
  })

  test('una sigla che nel ticket nominato non esiste resta non risolta', () => {
    expect(references('in oc:8968 il suo R7', index)).toEqual([])
  })
})

describe('una sigla scritta in minuscolo vale come in maiuscolo', () => {
  const index = indexDocument('## T9 — Il canale si chiude\n\nTesto del task.', 'plan.md')

  test('t9 trova T9', () => {
    expect(references('ricordami t9', index)[0]?.entry.code).toBe('T9')
  })

  test('una parola che somiglia a una sigla ma non è nell indice non conta', () => {
    expect(references('la versione v2 del file', index)).toEqual([])
  })
})
