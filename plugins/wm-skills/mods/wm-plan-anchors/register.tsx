import type { EngineInterface, Register } from 'claude-code'

import { indicizzaTutti, riferimenti } from './documento.ts'
import type { Voce as VoceDoc } from './documento.ts'
import { parseLabels, pulisci, senzaCodice, testaOverview } from './parse.ts'

const PANE = 'wm-plan-anchors'

type Voce = { label: string; text: string; nota: string }

/** I documenti della feature su cui il pane sta lavorando. */
type Feature = {
  path: string
  dir: string
  md: string
  documenti: { file: string; md: string }[]
}

/** Un agente al lavoro: cosa fa, e l'id con cui se ne apre la trascrizione. */
type Agente = { id?: string; testo: string }

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

/**
 * I documenti della feature più recente sotto `docs/features/`.
 *
 * Non si filtra più per formato della riga dei requisiti: prima si cercava
 * `<a id="r\d+"></a>` e un overview che scrive i requisiti come heading
 * veniva scartato in silenzio, pane vuoto e nessuna spiegazione. Si prende la
 * cartella toccata per ultima e si leggono i tre documenti: l'indice ci trova
 * dentro quello che c'è.
 */
async function findOverview($: EngineInterface): Promise<Feature | undefined> {
  let best: { dir: string; mtimeMs: number } | undefined

  try {
    for (const entry of await $.fs.list('docs/features')) {
      if (entry.kind !== 'dir') continue
      const dir = `docs/features/${entry.name}`
      try {
        const stat = await $.fs.stat(`${dir}/overview.md`)
        if (!best || stat.mtimeMs > best.mtimeMs) best = { dir, mtimeMs: stat.mtimeMs }
      } catch {
        continue
      }
    }
  } catch {
    // nessuna docs/features: si passa all'esempio incluso
  }

  if (best) {
    const documenti: { file: string; md: string }[] = []
    // L'ordine conta: un identificativo è definito nell'overview, e plan.md
    // o notes.md lo richiamano senza doverlo ridefinire.
    for (const file of ['overview.md', 'plan.md', 'notes.md']) {
      try {
        const md = await $.fs.read(`${best.dir}/${file}`)
        if (typeof md === 'string') documenti.push({ file, md })
      } catch {
        continue
      }
    }
    const overview = documenti.find(uno => uno.file === 'overview.md')
    if (overview) {
      return { path: `${best.dir}/overview.md`, dir: best.dir, md: overview.md, documenti }
    }
  }

  try {
    const md = await $.fs.read(`${$.plugin.root}/mods/wm-plan-anchors/examples/overview-demo.md`)
    if (typeof md === 'string') {
      return {
        path: 'esempio incluso nella mod',
        dir: '',
        md,
        documenti: [{ file: 'overview.md', md }],
      }
    }
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
  indice: Map<string, VoceDoc>,
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

    // I riferimenti li trova l'indice: una sigla nuda come `R1` o `T4`, e i
    // link alle ancore. Il pane non sa piu che forma abbia la riga che li
    // definisce, e non deve saperlo.
    for (const { at, voce } of riferimenti(detto, indice)) {
      const dove = voce.file === 'overview.md' ? '' : ` · ${voce.file}`
      here.push({
        at,
        voce: {
          label: voce.sigla ?? voce.titolo,
          text: pulisci(voce.titolo === '' ? voce.corpo : `${voce.titolo}. ${voce.corpo}`),
          nota: `#${voce.id}${dove}`,
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
 * Gli agenti al lavoro: cosa sta facendo ciascuno, e con quale modello.
 *
 * La lista nativa sotto il prompt dà nome, durata e token; `description` è la
 * frase con cui l'agente è stato lanciato — cioè il compito che sta eseguendo
 * in questo momento — e il modello non si vede da nessuna parte. Sono i due
 * pezzi che mancano, ed è l'unica ragione per cui questa riga esiste.
 */
async function agentiAlLavoro($: EngineInterface): Promise<Agente[]> {
  let vivi
  try {
    vivi = (await $.agent.list()).filter(one => one.status === 'running')
  } catch {
    return []
  }

  // Oltre tre, il pane diventa la lista degli agenti invece dei riferimenti:
  // i primi tre e un conteggio di quelli rimasti.
  const mostrati: Agente[] = vivi.slice(0, 3).map(uno => ({
    id: uno.id,
    testo: `▸ ${uno.description === '' ? uno.type : uno.description} · ${modelloBreve(uno.model)}`,
  }))
  if (vivi.length > 3) mostrati.push({ testo: `▸ e altri ${vivi.length - 3}` })

  return mostrati
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
    const { Box, Button, Text } = $.ui.resolve(e)
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

    const agenti = await agentiAlLavoro($)

    // L'overview si legge prima della testa: il ticket e lo stato del lavoro
    // stanno in quelle due righe, e vanno mostrati anche quando non c'è
    // nessun riferimento citato.
    const overview = await findOverview($)
    const testaDoc = overview === undefined ? {} : testaOverview(overview.md, overview.dir)
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
        {agenti.map(agente =>
          agente.id === undefined ? (
            <Text dimColor wrap="truncate-end">{agente.testo}</Text>
          ) : (
            // Premendo la riga, la trascrizione dell'agente prende la vista
            // principale: è il modo più corto per vedere cosa sta davvero
            // facendo, invece di aspettare il report.
            <Button key={agente.id} onPress={async () => {
              const aperto = await $.ui.view({ agentId: agente.id })
              if ('deny' in aperto && aperto.deny !== undefined) $.ui.toast(aperto.deny)
            }}>
              <Text dimColor wrap="truncate-end">{agente.testo}</Text>
            </Button>
          ),
        )}
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
            Nessun overview sotto docs/features/, e l'esempio incluso nella mod non si legge.
          </Text>
        </Box>
      )
    }

    const indice = indicizzaTutti(overview.documenti)

    if (indice.size === 0) {
      return (
        <Box flexDirection="column">
          {testa}
          <Text dimColor wrap="wrap">
            Documenti trovati ma niente da indicizzare: nessun heading e nessuna ancora.
          </Text>
          <Text dimColor wrap="truncate-start">{overview.path}</Text>
        </Box>
      )
    }

    const list = await citedNow(
      $,
      indice,
      larghezza,
      Math.max(6, (e.viewport?.rows ?? 24) - 3 - agenti.length),
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
