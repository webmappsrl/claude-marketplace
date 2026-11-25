import type { EngineInterface, Register } from 'claude-code'

import { parseLabels, parseRequirements, pulisci, senzaCodice, testaOverview } from './parse.ts'
import type { Req } from './parse.ts'

const PANE = 'wm-plan-anchors'

type Voce = { label: string; text: string; nota: string }

/**
 * Il branch corrente, come lo vede il motore.
 *
 * `$.vcs.detect` risponde per qualunque sistema di versionamento, non solo git,
 * e può non dare un branch (detached HEAD, o un sistema che non ne ha).
 */
async function branchCorrente($: EngineInterface): Promise<string> {
  try {
    return (await $.vcs.detect({ path: '.' }))?.branch ?? ''
  } catch {
    return ''
  }
}

/** La cartella feature con l'overview piu recente che usa le ancore. */
async function findOverview(
  $: EngineInterface,
): Promise<{ path: string; md: string } | undefined> {
  let best: { path: string; md: string; mtimeMs: number } | undefined

  try {
    for (const entry of await $.fs.list('docs/features')) {
      if (entry.kind !== 'dir') continue
      const path = `docs/features/${entry.name}/overview.md`
      try {
        const stat = await $.fs.stat(path)
        const md = await $.fs.read(path)
        if (typeof md !== 'string') continue
        if (!/<a id="r\d+"><\/a>/.test(md)) continue
        if (!best || stat.mtimeMs > best.mtimeMs) best = { path, md, mtimeMs: stat.mtimeMs }
      } catch {
        continue
      }
    }
  } catch {
    // nessuna docs/features: si passa all'esempio incluso
  }

  if (best) return { path: best.path, md: best.md }

  try {
    const md = await $.fs.read(`${$.plugin.root}/mods/wm-plan-anchors/examples/overview-demo.md`)
    if (typeof md === 'string') return { path: 'esempio incluso nella mod', md }
  } catch {
    return undefined
  }

  return undefined
}

/**
 * I riferimenti citati, dal piu recente.
 *
 * Legge la conversazione invece di aspettare `session.append`: quell'evento un
 * altro plugin puo consumarlo senza passarlo oltre, e la mod resterebbe cieca.
 */
async function citedNow(
  $: EngineInterface,
  known: Map<string, Req>,
  larghezza: number,
  righe: number,
) {
  const rows = await $.session.messages()
  const labels = new Map<string, string>()
  const seen = new Map<string, Voce>()
  const colonne = Math.max(24, larghezza)

  for (const row of rows.slice(-60)) {
    if (row.role !== 'assistant') continue

    const detto = senzaCodice(row.text)

    // Prima le definizioni di questo messaggio, poi i richiami: cosi
    // un'etichetta definita qui e gia risolvibile nello stesso messaggio.
    for (const [name, body] of parseLabels(row.text)) labels.set(name, body)

    // Ogni occorrenza con la sua posizione: l'ordine del pane e quello in cui
    // le cose sono state scritte, non quello in cui sono note.
    const here: { at: number; voce: Voce }[] = []

    for (const hit of detto.matchAll(/\bR(\d+)\b/g)) {
      const req = known.get(`R${hit[1]}`)
      if (!req || hit.index === undefined) continue
      here.push({
        at: hit.index,
        voce: {
          label: req.label,
          text: pulisci(req.text),
          nota: `#${req.label.toLowerCase()}${req.verifica ? ' · Verifica presente' : ' · Verifica assente'}${req.ux ? ' · [UX]' : ''}`,
        },
      })
    }

    for (const [name, body] of labels) {
      for (const hit of detto.matchAll(new RegExp(`\\b${name}\\b`, 'g'))) {
        if (hit.index === undefined) continue
        here.push({
          at: hit.index,
          voce: { label: name, text: pulisci(body), nota: 'etichetta' },
        })
      }
    }

    for (const one of here.sort((a, b) => a.at - b.at)) {
      seen.delete(one.voce.label)
      seen.set(one.voce.label, one.voce)
    }
  }

  // Quante voci entrano davvero: titolo, nota, riga vuota e testo a capo.
  const dal_piu_recente = [...seen.values()].reverse()
  const dentro: Voce[] = []
  let spesa = 0

  for (const voce of dal_piu_recente) {
    const alte = Math.ceil(voce.text.length / colonne) + 3
    if (dentro.length > 0 && spesa + alte > righe) break
    dentro.push(voce)
    spesa += alte
  }

  return dentro
}

/** Il modello senza il prefisso del fornitore e senza la data. */
function modelloBreve(id: string | undefined): string {
  if (!id || id === '') return '?'
  const senzaData = id.replace(/-\d{8}$/, '')
  const nome = senzaData
    .split('-')
    .find(uno => ['opus', 'sonnet', 'haiku', 'fable'].includes(uno))

  return nome ?? senzaData
}

/**
 * I modelli degli agenti al lavoro, contati.
 *
 * La lista nativa sotto il prompt da nome, durata e token, ma non il modello:
 * questo e l'unico pezzo che manca, e l'unico che vale la pena aggiungere.
 */
async function modelliAlLavoro($: EngineInterface): Promise<string> {
  let vivi
  try {
    vivi = (await $.agent.list()).filter(one => one.status === 'running')
  } catch {
    return ''
  }

  if (vivi.length === 0) return ''

  const conti = new Map<string, number>()
  for (const uno of vivi) {
    const nome = modelloBreve(uno.model)
    conti.set(nome, (conti.get(nome) ?? 0) + 1)
  }

  return [...conti.entries()]
    .map(([nome, quanti]) => (quanti === 1 ? `1 agente su ${nome}` : `${quanti} agenti su ${nome}`))
    .join(', ')
}

/**
 * Il colore della percentuale di context.
 *
 * Chiavi del tema, non colori crudi: seguono il tema scelto dalla persona.
 * Le soglie sono una scelta, non un dato del motore, e si cambiano qui.
 */
function colorePerContext(percento: number): 'success' | 'warning' | 'error' {
  if (percento >= 75) return 'error'
  if (percento >= 50) return 'warning'

  return 'success'
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    // Un pane che nessuno ha chiesto si siede solo da 144 colonne di terminale.
    // Sotto quella larghezza resta non disegnato: senza questo avviso la mod
    // sarebbe muta, e il dev non saprebbe che c'è un pane da vedere.
    const aperto = await $.ui.open({ id: PANE, title: 'Riferimenti' })
    if (!aperto.isPlaced) {
      const motivo = 'reason' in aperto ? aperto.reason : 'motivo non dato'
      $.ui.toast(`Pane dei riferimenti non disegnato: ${motivo}`)
    }

    // La testa del pane non si ridisegna da sola mentre un agente lavora:
    // senza questo, il conteggio resterebbe fermo proprio quando cambia.
    $.clock.every(2000, async () => {
      try {
        const vivi = (await $.agent.list()).filter(one => one.status === 'running')
        if (vivi.length > 0) void $.ui.invalidate('ui.render')
      } catch {
        // niente da aggiornare
      }
    })

    return next(e)
  })

  // Il pane si riapre da solo: chiuso per sbaglio, torna al turno dopo.
  on('turn.complete', async ($, e, next) => {
    const done = await next(e)
    void $.ui.open({ id: PANE, title: 'Riferimenti' })
    void $.ui.invalidate('ui.render')

    return done
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Box, Text } = $.ui.resolve(e)
    const larghezza = e.props?.bodyColumns ?? e.viewport?.columns ?? 40

    let percento: number | undefined
    let guaio = ''
    try {
      percento = (await $.session.usage()).context.percent
      if (percento === undefined) guaio = 'percentuale non data'
    } catch (errore) {
      guaio = `usage() rifiutata: ${String(errore)}`
    }

    // Il modello del loop principale, come lo mostra /model. Quello degli
    // agenti è un'altra cosa e sta nella riga sotto: qui si legge con quale
    // modello sta rispondendo la sessione, che altrove non si vede mai.
    let modello = ''
    try {
      modello = await $.session.model()
    } catch {
      modello = '?'
    }

    const agenti = await modelliAlLavoro($)

    // L'overview si legge prima della testa: il ticket e lo stato del lavoro
    // stanno in quelle due righe, e vanno mostrati anche quando non c'è
    // nessun riferimento citato.
    const overview = await findOverview($)
    const testaDoc = overview === undefined ? {} : testaOverview(overview.md)
    const branch = await branchCorrente($)

    // Sempre in testa, anche quando il dato manca: una riga muta non si
    // distingue da un hook che non gira.
    const testa = (
      <Box flexDirection="column" marginBottom={1}>
        <Box>
          <Text dimColor>{modello === '' ? '? ' : `${modello}  `}</Text>
          <Text dimColor>context </Text>
          {percento === undefined ? (
            <Text dimColor>{`— ${guaio}`}</Text>
          ) : (
            <Text color={colorePerContext(percento)} bold>{`${percento}%`}</Text>
          )}
          <Text dimColor>
            {percento !== undefined && percento >= 75
              ? '  conviene un compact'
              : percento !== undefined && percento >= 50
                ? '  ci si avvicina'
                : ''}
          </Text>
        </Box>
        {agenti === '' ? undefined : <Text dimColor>{agenti}</Text>}
        <Box>
          <Text bold>{testaDoc.ticket ?? (overview === undefined ? 'nessun overview' : 'senza ticket')}</Text>
          <Text dimColor>{branch === '' ? '' : `  ${branch}`}</Text>
        </Box>
        {testaDoc.stato === undefined ? undefined : <Text dimColor>{testaDoc.stato}</Text>}
        <Text dimColor>{'─'.repeat(Math.max(8, larghezza))}</Text>
      </Box>
    )

    if (!overview) {
      return (
        <Box flexDirection="column">
          {testa}
          <Text dimColor wrap="wrap">
            Nessun overview con ancore sotto docs/features/, e l'esempio incluso nella mod non si
            legge.
          </Text>
        </Box>
      )
    }

    const known = parseRequirements(overview.md)

    if (known.size === 0) {
      return (
        <Box flexDirection="column">
          {testa}
          <Text dimColor wrap="wrap">
            Overview trovato ma nessun requisito riconosciuto: il formato della riga non e quello
            con l'ancora.
          </Text>
          <Text dimColor wrap="truncate-start">{overview.path}</Text>
        </Box>
      )
    }

    const list = await citedNow(
      $,
      known,
      larghezza,
      Math.max(6, (e.viewport?.rows ?? 24) - 2 - (agenti === '' ? 2 : 3)),
    )

    // In quiete il pane mostra solo la testa: dire a ogni turno quanti
    // requisiti sono stati letti e da quale file non serve a nessuno, e il
    // file è comunque scritto in fondo appena una voce compare. I due casi
    // che vanno detti - nessun overview, formato non riconosciuto - restano
    // sopra: il silenzio qui vuol dire "letto, nessun riferimento citato".
    if (list.length === 0) {
      return <Box flexDirection="column">{testa}</Box>
    }

    return (
      <Box flexDirection="column">
        {testa}
        {list.map((one, index) => (
          <Box flexDirection="column" marginBottom={1}>
            <Text bold>
              {one.label}
              {index === 0 ? '  ← citato ora' : ''}
            </Text>
            <Text wrap="wrap">{one.text}</Text>
            <Text dimColor>{one.nota}</Text>
          </Box>
        ))}
        <Text dimColor wrap="truncate-start">{overview.path}</Text>
      </Box>
    )
  })
}
