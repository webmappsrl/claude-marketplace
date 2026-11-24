import { describe, expect, test } from 'claude-code/testing'

import { identificativi, parseLabels, parseRequirements, pulisci, senzaCodice } from './parse.ts'

const NUOVO = [
  '## Requisiti',
  '',
  '- <a id="r1"></a>**R1** — i POI importati mantengono l\'id Geohub.',
  '- <a id="r2"></a>**R2** — la config dello shard non sovrascrive le chiavi',
  '  gia presenti nel package.',
  '  - Verifica: `tests/Feature/ImportConfigTest.php`',
  '- <a id="r3"></a>**R3** — [UX] la mappa centra sull\'app.',
  '- <a id="r4"></a>**R4** — l\'import non tocca Elastic.',
  '  - Non fatto: [manca il tempo](notes.md#voce)',
  '',
  '## Rischi',
  '',
  '- qui non ci sono requisiti.',
].join('\n')

const VECCHIO = [
  '## Requisiti',
  '',
  '**Parametri e modalita**',
  '',
  '- [ ] `update_story` accetta due parametri nuovi.',
  '- [x] Senza `prepend` si comporta come oggi.',
].join('\n')

describe('il formato dei requisiti', () => {
  test('riconosce gli identificativi nel formato con ancore', () => {
    expect(identificativi(NUOVO)).toEqual(['R1', 'R2', 'R3', 'R4'])
  })

  test('un testo su piu righe rientrate resta un requisito solo', () => {
    const r2 = parseRequirements(NUOVO).get('R2')

    expect(r2?.text).toBe(
      'la config dello shard non sovrascrive le chiavi gia presenti nel package.',
    )
  })

  test('la sigla [UX] si legge e non resta nel testo', () => {
    const r3 = parseRequirements(NUOVO).get('R3')

    expect(r3?.ux, 'R3 porta la sigla [UX]').toBe(true)
    expect(r3?.text).toBe("la mappa centra sull'app.")
    expect(parseRequirements(NUOVO).get('R1')?.ux).toBe(false)
  })

  test('le sottovoci Verifica e Non fatto si distinguono', () => {
    const reqs = parseRequirements(NUOVO)

    expect(reqs.get('R2')?.verifica, 'R2 ha la riga Verifica').toBe(true)
    expect(reqs.get('R1')?.verifica, 'R1 non ce l\'ha').toBe(false)
    expect(reqs.get('R4')?.nonFatto, 'R4 e segnato non fatto').toBe(true)
  })

  test('il vecchio formato a checkbox non viene riconosciuto', () => {
    expect(identificativi(VECCHIO)).toEqual([])
  })

  test('una riga senza ancora non e un requisito', () => {
    expect(identificativi('- **R1** — manca l\'ancora, non vale.')).toEqual([])
  })

  test('su spazzatura non lancia e non riconosce niente', () => {
    expect(identificativi('')).toEqual([])
    expect(identificativi('\u0000 <a id= **R** —')).toEqual([])
    expect(identificativi('- <a id="rX"></a>**RX** — lettera al posto del numero')).toEqual([])
  })
})

describe('le sigle fra backtick', () => {
  test('una sigla fra backtick sparisce, e le posizioni non si spostano', () => {
    const detto = 'quando cito `R1` parlo della sigla, R2 invece lo sto facendo'
    const pulito = senzaCodice(detto)

    expect(pulito.length, 'la lunghezza resta la stessa').toBe(detto.length)
    expect(/\bR1\b/.test(pulito), 'R1 fra backtick e sparito').toBe(false)
    expect(pulito.indexOf('R2'), 'R2 resta dove era').toBe(detto.indexOf('R2'))
  })

  test('anche un blocco di codice intero sparisce', () => {
    const pulito = senzaCodice('prima\n```\n- <a id="r9"></a>**R9** — esempio\n```\ndopo R8')

    expect(/R9/.test(pulito)).toBe(false)
    expect(/R8/.test(pulito)).toBe(true)
  })
})

describe('le etichette definite parlando', () => {
  test('una riga di elenco con nome in maiuscolo diventa una definizione', () => {
    const trovate = parseLabels('- **GATE-COMMIT** — non si committa dentro un lavoro.')

    expect([...trovate.keys()]).toEqual(['GATE-COMMIT'])
    expect(trovate.get('GATE-COMMIT')).toBe('non si committa dentro un lavoro.')
  })

  test('minuscole, nomi corti e righe senza trattino non sono definizioni', () => {
    expect([...parseLabels('- **gate** — minuscolo').keys()]).toEqual([])
    expect([...parseLabels('- **AB** — troppo corto').keys()]).toEqual([])
    expect([...parseLabels('- **NOME** senza trattino').keys()]).toEqual([])
  })
})

test('il markdown esce dal testo mostrato', () => {
  expect(pulisci('il modo in cui `wm-tag` invoca **wm-plan**, vedi [qui](x.md)')).toBe(
    'il modo in cui wm-tag invoca wm-plan, vedi qui',
  )
})
