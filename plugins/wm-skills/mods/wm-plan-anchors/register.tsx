import type { EngineInterface, Register } from 'claude-code'

import {
  citedPaths,
  citedTickets,
  indexAll,
  namedCodes,
  references,
} from './documents.ts'
import type { Entry as DocEntry } from './documents.ts'
import { clean, overviewHead, parseLabels, withoutCode } from './parse.ts'

const PANE = 'wm-plan-anchors'

type Item = { label: string; text: string; note: string }

/** I documenti della feature su cui il pane sta lavorando. */
type Feature = { path: string; dir: string; md: string }

/** Un agente al lavoro: cosa fa, e l'id con cui se ne apre la trascrizione. */
type Agent = { id?: string; text: string }

/**
 * Il branch corrente, come lo vede il motore.
 *
 * `$.vcs.detect` risponde per qualunque sistema di versionamento, non solo git,
 * e può non dare un branch (detached HEAD, o un sistema che non ne ha).
 */
async function currentBranch($: EngineInterface): Promise<string> {
  try {
    return (await $.vcs.detect({ path: '.' }))?.branch ?? ''
  } catch {
    return ''
  }
}

/**
 * I documenti già letti, con la mtime che avevano.
 *
 * Il pane si ridisegna a ogni fine turno e, finché un agente lavora, ogni due
 * secondi: senza questa cache rileggerebbe e riparserebbe ogni volta decine
 * di KB che non sono cambiati. Si tiene la mtime e si rilegge solo il file
 * che si è mosso, così l'indice non è mai stantio e il lavoro si fa una volta.
 */
const readCache = new Map<string, { mtimeMs: number; md: string }>()

/**
 * L'ultimo indice costruito, con la firma dei documenti da cui viene.
 *
 * La firma è l'elenco dei path con la loro mtime: finché non cambia, l'indice
 * è ancora buono e non si riparsa niente. Sta in memoria e non su disco
 * perché la mod vive quanto la sessione — un file temporaneo costerebbe una
 * scrittura in più e la stat andrebbe fatta lo stesso.
 */
let heldIndex: { signature: string; index: Map<string, DocEntry> } | undefined

/** Legge un markdown, dalla cache se la mtime non è cambiata. */
async function readIfChanged(
  $: EngineInterface,
  path: string,
): Promise<{ file: string; md: string } | undefined> {
  try {
    const stat = await $.fs.stat(path)
    const before = readCache.get(path)
    if (before && before.mtimeMs === stat.mtimeMs) return { file: path, md: before.md }

    const md = await $.fs.read(path)
    if (typeof md !== 'string') return undefined
    readCache.set(path, { mtimeMs: stat.mtimeMs, md })

    return { file: path, md }
  } catch {
    return undefined
  }
}

/**
 * I documenti che la sessione sta toccando.
 *
 * Non è più "i tre file della cartella della feature": è ogni `.md` di quella
 * cartella, più ogni `.md` che la conversazione nomina — un link a un altro
 * ticket, una pagina di `docs/knowledge/`, un README. Se la sessione elenca
 * un riferimento, il documento che lo definisce deve essere fra questi, o il
 * pane non ha modo di svilupparlo.
 *
 * L'ordine è quello che decide chi vince su un id ripetuto: prima l'overview
 * della feature corrente, poi il resto della sua cartella, poi i citati.
 */
async function collectDocuments(
  $: EngineInterface,
  dir: string,
  cited: string[],
  tickets: string[],
): Promise<{ documents: { file: string; md: string }[]; signature: string }> {
  const taken: { file: string; md: string }[] = []
  const signatures: string[] = []
  const seen = new Set<string>()

  const add = async (path: string) => {
    if (seen.has(path)) return
    seen.add(path)
    const one = await readIfChanged($, path)
    if (!one) return
    taken.push({ file: path.startsWith(`${dir}/`) ? path.slice(dir.length + 1) : path, md: one.md })
    signatures.push(`${path}@${readCache.get(path)?.mtimeMs ?? 0}`)
  }

  if (dir !== '') {
    // L'ordine non e quello del filesystem: a parita di sigla vince chi viene
    // letto prima, e `T1` deve risolvere sempre allo stesso documento da una
    // sessione all'altra. Overview, piano, note, poi il resto in ordine.
    const first = ['overview.md', 'plan.md', 'notes.md']
    for (const name of first) await add(`${dir}/${name}`)
    try {
      const rest = (await $.fs.list(dir))
        .filter(entry => entry.kind !== 'dir' && entry.name.endsWith('.md'))
        .map(entry => entry.name)
        .filter(name => !first.includes(name))
        .sort()
      for (const name of rest) await add(`${dir}/${name}`)
    } catch {
      // la cartella non si legge: restano i citati
    }
  }

  // Un ticket nominato porta la sua cartella: il numero si traduce cercando
  // `docs/features/<numero>-…`, come fa la skill, invece di ri-derivare lo
  // slug dal titolo — che nel frattempo puo essere cambiato.
  if (tickets.length > 0) {
    try {
      const dirs = await $.fs.list('docs/features')
      for (const number of [...tickets].sort()) {
        const found = dirs.find(
          entry => entry.kind === 'dir' && entry.name.startsWith(`${number}-`),
        )
        if (!found) continue
        for (const name of ['overview.md', 'plan.md', 'notes.md']) {
          await add(`docs/features/${found.name}/${name}`)
        }
      }
    } catch {
      // nessuna docs/features leggibile: restano i path citati
    }
  }

  for (const path of [...cited].sort()) {
    if (!path.endsWith('.md')) continue
    // Un path relativo scritto parlando si intende dentro la cartella della
    // feature: `notes.md` vuol dire quello di questa feature, non uno in radice.
    if (!path.includes('/') && dir !== '') await add(`${dir}/${path}`)
    await add(path)
  }

  return { documents: taken, signature: signatures.join('|') }
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
    const one = await readIfChanged($, `${best.dir}/overview.md`)
    if (one) return { path: `${best.dir}/overview.md`, dir: best.dir, md: one.md }
  }

  try {
    const md = await $.fs.read(`${$.plugin.root}/mods/wm-plan-anchors/examples/overview-demo.md`)
    if (typeof md === 'string') return { path: 'esempio incluso nella mod', dir: '', md }
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
function citedNow(
  rows: { role: string; text: string }[],
  index: Map<string, DocEntry>,
  width: number,
  rowBudget: number,
) {
  // I documenti letti usano le sigle? Basta che una voce ne abbia una.
  const usesCodes = [...index.values()].some(entry => entry.code !== undefined)
  const labels = new Map<string, string>()
  const seen = new Map<string, Item>()
  const columns = Math.max(24, width)

  for (const row of rows.slice(-60)) {
    const said = withoutCode(row.text)

    // Prima le definizioni di questo messaggio, poi i richiami: cosi
    // un'etichetta definita qui e gia risolvibile nello stesso messaggio.
    for (const [name, body] of parseLabels(row.text)) labels.set(name, body)

    // Ogni occorrenza con la sua posizione: l'ordine del pane e quello in cui
    // le cose sono state scritte, non quello in cui sono note.
    const here: { at: number; item: Item }[] = []

    // I riferimenti li trova l'indice: una sigla nuda come `R1` o `T4`, e i
    // link alle ancore. Il pane non sa piu che forma abbia la riga che li
    // definisce, e non deve saperlo.
    // Una sigla nominata che nei documenti non c'e: si mostra lo stesso,
    // perche un riferimento che non si trova e un difetto del documento, e
    // tacerlo lo lascia addosso al dev da scoprire al merge.
    //
    // Ma solo se questi documenti le sigle le usano: in un repo i cui
    // overview non ne hanno, ogni sigla nominata parlando risulterebbe
    // mancante e il pane sarebbe un elenco di falsi allarmi.
    for (const code of usesCodes ? namedCodes(said) : []) {
      if (index.has(code)) continue
      const at = said.indexOf(code)
      if (at < 0) continue
      here.push({
        at,
        item: { label: code, text: 'citato ma non trovato nei documenti letti', note: 'manca' },
      })
    }

    for (const { at, entry } of references(said, index)) {
      const where = entry.file === 'overview.md' ? '' : ` · ${entry.file}`
      here.push({
        at,
        item: {
          label: entry.code ?? entry.title,
          text: clean(entry.title === '' ? entry.body : `${entry.title}. ${entry.body}`),
          note: `#${entry.id}${where}`,
        },
      })
    }

    for (const [name, body] of labels) {
      for (const hit of said.matchAll(new RegExp(`\\b${name}\\b`, 'g'))) {
        if (hit.index === undefined) continue
        here.push({
          at: hit.index,
          item: { label: name, text: clean(body), note: 'etichetta' },
        })
      }
    }

    for (const one of here.sort((a, b) => a.at - b.at)) {
      seen.delete(one.item.label)
      seen.set(one.item.label, one.item)
    }
  }

  // Quante voci entrano davvero: titolo, nota, riga vuota e testo a capo.
  const newestFirst = [...seen.values()].reverse()
  const shown: Item[] = []
  let spent = 0

  for (const item of newestFirst) {
    const tall = Math.ceil(item.text.length / columns) + 3
    if (shown.length > 0 && spent + tall > rowBudget) break
    shown.push(item)
    spent += tall
  }

  return shown
}

/** Il modello senza il prefisso del fornitore e senza la data. */
function shortModel(id: string | undefined): string {
  if (!id || id === '') return '?'
  const withoutDate = id.replace(/-\d{8}$/, '')
  const name = withoutDate
    .split('-')
    .find(one => ['opus', 'sonnet', 'haiku', 'fable'].includes(one))

  return name ?? withoutDate
}

/**
 * Gli agenti al lavoro: cosa sta facendo ciascuno, e con quale modello.
 *
 * La lista nativa sotto il prompt dà nome, durata e token; `description` è la
 * frase con cui l'agente è stato lanciato — cioè il compito che sta eseguendo
 * in questo momento — e il modello non si vede da nessuna parte. Sono i due
 * pezzi che mancano, ed è l'unica ragione per cui questa riga esiste.
 */
async function runningAgents($: EngineInterface): Promise<Agent[]> {
  let alive
  try {
    alive = (await $.agent.list()).filter(one => one.status === 'running')
  } catch {
    return []
  }

  // Oltre tre, il pane diventa la lista degli agenti invece dei riferimenti:
  // i primi tre e un conteggio di quelli rimasti.
  const shown: Agent[] = alive.slice(0, 3).map(one => ({
    id: one.id,
    text: `▸ ${one.description === '' ? one.type : one.description} · ${shortModel(one.model)}`,
  }))
  if (alive.length > 3) shown.push({ text: `▸ e altri ${alive.length - 3}` })

  return shown
}

/**
 * Il colore della percentuale di context.
 *
 * Chiavi del tema, non colori crudi: seguono il tema scelto dalla persona.
 * Le soglie sono una scelta, non un dato del motore, e si cambiano qui.
 */
function contextColor(percent: number): 'success' | 'warning' | 'error' {
  if (percent >= 75) return 'error'
  if (percent >= 50) return 'warning'

  return 'success'
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    // Un pane che nessuno ha chiesto si siede solo da 144 colonne di terminale.
    // Sotto quella larghezza resta non disegnato: senza questo avviso la mod
    // sarebbe muta, e il dev non saprebbe che c'è un pane da vedere.
    const opened = await $.ui.open({ id: PANE, title: 'Riferimenti' })
    if (!opened.isPlaced) {
      const why = 'reason' in opened ? opened.reason : 'motivo non dato'
      $.ui.toast(`Pane dei riferimenti non disegnato: ${why}`)
    }

    // La testa del pane non si ridisegna da sola mentre un agente lavora:
    // senza questo, il conteggio resterebbe fermo proprio quando cambia.
    $.clock.every(2000, async () => {
      try {
        const alive = (await $.agent.list()).filter(one => one.status === 'running')
        if (alive.length > 0) void $.ui.invalidate('ui.render')
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
    const width = e.props?.bodyColumns ?? e.viewport?.columns ?? 40

    let percent: number | undefined
    let trouble = ''
    try {
      percent = (await $.session.usage()).context.percent
      if (percent === undefined) trouble = 'percentuale non data'
    } catch (error) {
      trouble = `usage() rifiutata: ${String(error)}`
    }

    // Il modello del loop principale, come lo mostra /model. Quello degli
    // agenti è un'altra cosa e sta nella riga sotto: qui si legge con quale
    // modello sta rispondendo la sessione, che altrove non si vede mai.
    let model = ''
    try {
      model = await $.session.model()
    } catch {
      model = '?'
    }

    const agents = await runningAgents($)

    // L'overview si legge prima della testa: il ticket e lo stato del lavoro
    // stanno in quelle due righe, e vanno mostrati anche quando non c'è
    // nessun riferimento citato.
    const overview = await findOverview($)
    const head = overview === undefined ? {} : overviewHead(overview.md, overview.dir)
    const branch = await currentBranch($)

    // Sempre in testa, anche quando il dato manca: una riga muta non si
    // distingue da un hook che non gira.
    const header = (
      <Box flexDirection="column" marginBottom={1}>
        <Box>
          <Text dimColor>{model === '' ? '? ' : `${model}  `}</Text>
          <Text dimColor>context </Text>
          {percent === undefined ? (
            <Text dimColor>{`— ${trouble}`}</Text>
          ) : (
            <Text color={contextColor(percent)} bold>{`${percent}%`}</Text>
          )}
          <Text dimColor>
            {percent !== undefined && percent >= 75
              ? '  conviene un compact'
              : percent !== undefined && percent >= 50
                ? '  ci si avvicina'
                : ''}
          </Text>
        </Box>
        {agents.map(agent =>
          agent.id === undefined ? (
            <Text dimColor wrap="truncate-end">{agent.text}</Text>
          ) : (
            // Premendo la riga, la trascrizione dell'agente prende la vista
            // principale: è il modo più corto per vedere cosa sta davvero
            // facendo, invece di aspettare il report.
            <Button key={agent.id} onPress={async () => {
              const opened = await $.ui.view({ agentId: agent.id })
              if ('deny' in opened && opened.deny !== undefined) $.ui.toast(opened.deny)
            }}>
              <Text dimColor wrap="truncate-end">{agent.text}</Text>
            </Button>
          ),
        )}
        <Box>
          <Text bold>{head.ticket ?? (overview === undefined ? 'nessun overview' : 'senza ticket')}</Text>
          <Text dimColor>{branch === '' ? '' : `  ${branch}`}</Text>
        </Box>
        {head.status === undefined ? undefined : <Text dimColor>{head.status}</Text>}
        <Text dimColor>{'─'.repeat(Math.max(8, width))}</Text>
      </Box>
    )

    if (!overview) {
      return (
        <Box flexDirection="column">
          {header}
          <Text dimColor wrap="wrap">
            Nessun overview sotto docs/features/, e l'esempio incluso nella mod non si legge.
          </Text>
        </Box>
      )
    }

    // La conversazione si legge una volta sola: da qui escono sia i documenti
    // che la sessione nomina, sia i riferimenti da sviluppare.
    const rows = await $.session.messages()
    // Anche quello che scrive il dev, non solo quello che scrivo io: «com'era
    // T5?» e il caso piu naturale di tutti, e con il solo ruolo assistant il
    // pane restava muto proprio quando gli si chiedeva qualcosa.
    const said = rows.slice(-60).filter(one => one.role === 'assistant' || one.role === 'user')
    const cited = [...new Set(said.flatMap(one => citedPaths(one.text)))]
    const tickets = [...new Set(said.flatMap(one => citedTickets(one.text)))]

    const collected = await collectDocuments($, overview.dir, cited, tickets)

    // L'indice si ricostruisce solo se un documento si è mosso o se la
    // sessione ne ha nominato uno nuovo: la firma dice entrambe le cose.
    const index =
      heldIndex?.signature === collected.signature
        ? heldIndex.index
        : indexAll(collected.documents)
    heldIndex = { signature: collected.signature, index }

    if (index.size === 0) {
      return (
        <Box flexDirection="column">
          {header}
          <Text dimColor wrap="wrap">
            Documenti trovati ma niente da indicizzare: nessun heading e nessuna ancora.
          </Text>
          <Text dimColor wrap="truncate-start">{overview.path}</Text>
        </Box>
      )
    }

    const list = citedNow(
      said,
      index,
      width,
      Math.max(6, (e.viewport?.rows ?? 24) - 3 - agents.length),
    )

    // In quiete il pane mostra solo la testa: dire a ogni turno quanti
    // requisiti sono stati letti e da quale file non serve a nessuno, e il
    // file è comunque scritto in fondo appena una voce compare. I due casi
    // che vanno detti - nessun overview, formato non riconosciuto - restano
    // sopra: il silenzio qui vuol dire "letto, nessun riferimento citato".
    if (list.length === 0) {
      return <Box flexDirection="column">{header}</Box>
    }

    return (
      <Box flexDirection="column">
        {header}
        {list.map((one, position) => (
          <Box flexDirection="column" marginBottom={1}>
            <Text bold>
              {one.label}
              {position === 0 ? '  ← citato ora' : ''}
            </Text>
            <Text wrap="wrap">{one.text}</Text>
            <Text dimColor>{one.note}</Text>
          </Box>
        ))}
        <Text dimColor wrap="truncate-start">{overview.path}</Text>
      </Box>
    )
  })
}
