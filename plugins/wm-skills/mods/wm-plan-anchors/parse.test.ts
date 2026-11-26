import { describe, expect, test } from 'claude-code/testing'

import { parseLabels, pulisci, senzaCodice, testaOverview } from './parse.ts'



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

describe('la testa di un overview', () => {
  test('legge ticket e stato', () => {
    const md = [
      '> Ticket: oc:8700',
      '> Stato: chiusa il 2026-11-25 · approvata il 2026-11-20',
      '',
      '# Titolo',
    ].join('\n')

    expect(testaOverview(md).ticket).toBe('oc:8700')
    expect(testaOverview(md).stato).toBe('chiusa il 2026-11-25 · approvata il 2026-11-20')
  })

  test('un overview senza ticket non inventa un id', () => {
    const md = '> Stato: approvata il 2026-11-20\n\n# Titolo'

    expect(testaOverview(md).ticket).toBeUndefined()
    expect(testaOverview(md).stato).toBe('approvata il 2026-11-20')
  })

  test('prima dell\'approvazione lo stato non c\'è ancora', () => {
    expect(testaOverview('> Ticket: oc:8700\n\n# Titolo').stato).toBeUndefined()
  })

  test('una riga di testo che nomina un ticket non è la riga di testa', () => {
    expect(testaOverview('Il lavoro nasce dal Ticket: oc:1 di ieri.').ticket).toBeUndefined()
  })
})
