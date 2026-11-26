/**
 * L'analizzatore dei documenti di una feature.
 *
 * Non riconosce "la riga di un requisito": indicizza tutto quello che in un
 * documento è indirizzabile — ogni àncora `<a id="…"></a>`, ogni heading con
 * il suo slug — e si tiene il blocco di testo che ciascuno apre. Quando
 * wm-plan nomina un identificativo o linka un'àncora, il pane ha già il
 * contenuto da sviluppare.
 *
 * È il motivo per cui questo file esiste al posto del vecchio parser legato al
 * formato: un overview che scrive i requisiti come heading
 *
 *     ### <a id="r1"></a>**R1** — La creazione chiede tre campi
 *
 * e uno che li scrive come voci di elenco
 *
 *     - <a id="r1"></a>**R1** — La creazione chiede tre campi
 *
 * danno lo stesso indice. Nessuna forma di riga da tenere allineata fra più
 * parser: cambia il modo di scrivere l'overview e l'indice continua a leggerlo.
 *
 * Vale per qualunque sigla — `R1` di un requisito, `T4` di un task del piano —
 * perché la sigla non è prescritta: si legge da dov'è scritta.
 */

export type Entry = {
  /** L'àncora con cui ci si arriva: `r1`, `t4-la-sovrapposizione-vera`. */
  id: string
  /** La sigla con cui lo si nomina parlando: `R1`, `T4`. Assente se non c'è. */
  code?: string
  /** Il testo del titolo, senza sigla né markdown. */
  title: string
  /** Il blocco che l'elemento apre, fino al successivo di pari rango. */
  body: string
  /** Il file in cui sta, relativo alla cartella della feature. */
  file: string
  /** 1-6 per un heading, 0 per un'àncora dentro un elenco o in riga. */
  level: number
}

/** L'àncora esplicita in una riga: `<a id="r1"></a>`. */
const ANCHOR = /<a id="([^"]+)"><\/a>/
/** Un heading markdown, con il suo livello. */
const HEADING = /^(#{1,6})\s+(.*)$/
/**
 * La sigla in testa a un titolo: `**R1** — testo`, `R1 — testo`, `T4 — testo`.
 *
 * Una lettera e almeno una cifra, perché una parola qualunque non diventi una
 * sigla per sbaglio, e solo in apertura: `vedi R3` dentro un titolo non lo è.
 */
const LEADING_CODE = /^\*{0,2}([A-Z]+\d+)\*{0,2}\s*[—–-]\s*/

/**
 * Lo slug di un heading, come lo calcola GitHub per l'àncora automatica.
 *
 * Gli spazi NON si collassano: ogni spazio diventa un trattino, uno per uno.
 * È il motivo per cui un titolo come `T4 — La sovrapposizione` dà
 * `t4--la-sovrapposizione`, con due trattini dove stava il trattino lungo, ed
 * è la forma che wm-plan scrive davvero nei link ai documenti.
 */
export function slug(text: string): string {
  return text
    .toLowerCase()
    .replace(/<[^>]+>/g, '')
    .replace(/[`*_\[\]()]/g, '')
    .replace(/[^\p{L}\p{N}\s-]/gu, '')
    .trim()
    .replace(/ /g, '-')
}

/** Toglie il markdown dal titolo: nel pane è solo rumore. */
function cleanText(line: string): string {
  return line
    .replace(/<[^>]+>/g, '')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/\s+/g, ' ')
    .trim()
}

type Marked = {
  line: number
  id: string
  code?: string
  title: string
  level: number
}

/**
 * Indicizza un documento: ogni elemento indirizzabile con il blocco che apre.
 *
 * Il corpo di un heading arriva fino al successivo di pari livello o più alto;
 * quello di una voce di elenco fino al prossimo elemento indicizzato, di
 * qualunque rango. Così un requisito scritto come sezione porta con sé la sua
 * Verifica e le sue fonti, e uno scritto come voce di elenco porta le sue
 * sottovoci, senza che l'indice sappia che cos'è un requisito.
 */
export function indexDocument(md: string, file: string): Map<string, Entry> {
  const lines = md.split('\n')
  const marked: Marked[] = []

  lines.forEach((line, n) => {
    const heading = HEADING.exec(line)
    const anchor = ANCHOR.exec(line)

    if (heading) {
      const level = (heading[1] ?? '').length
      const rest = (heading[2] ?? '').replace(ANCHOR, '').trim()
      const code = LEADING_CODE.exec(rest)?.[1]
      const title = cleanText(rest.replace(LEADING_CODE, ''))
      // L'àncora esplicita vince sullo slug: è quella che non cambia quando il
      // titolo si riformula, ed è quella che wm-plan scrive nei link.
      marked.push({ line: n, id: anchor?.[1] ?? slug(rest), code, title, level })
      return
    }

    if (!anchor) return

    const after = line.slice((anchor.index ?? 0) + anchor[0].length).trim()
    const code = LEADING_CODE.exec(after)?.[1]
    const title = cleanText(after.replace(LEADING_CODE, ''))
    // Livello 0: un'àncora fuori da un heading non apre una sezione, quindi
    // il suo blocco si chiude al primo elemento indicizzato che segue.
    marked.push({ line: n, id: anchor[1] ?? '', code, title, level: 0 })
  })

  const index = new Map<string, Entry>()

  marked.forEach((one, i) => {
    let end = lines.length
    for (let j = i + 1; j < marked.length; j++) {
      const other = marked[j]
      if (other === undefined) continue
      // Un heading chiude il blocco precedente quando è di pari rango o più
      // alto; un elemento senza livello chiude solo gli altri senza livello.
      const closes = one.level === 0 ? true : other.level !== 0 && other.level <= one.level
      if (closes) {
        end = other.line
        break
      }
    }

    const body = lines
      .slice(one.line + 1, end)
      .join('\n')
      .trim()

    const entry: Entry = {
      id: one.id,
      code: one.code,
      title: one.title,
      body,
      file,
      level: one.level,
    }

    if (one.id !== '') index.set(one.id, entry)
    // La sigla è un secondo nome per la stessa voce: `R1` e `r1` portano lì.
    if (one.code !== undefined && !index.has(one.code)) index.set(one.code, entry)
  })

  return index
}

/**
 * Indicizza più documenti in un indice solo.
 *
 * Le chiavi di un file letto dopo non sovrascrivono quelle già presenti:
 * `overview.md` si passa per primo, perché è lì che un identificativo è
 * definito, e `plan.md` o `notes.md` aggiungono i loro senza coprirlo.
 */
export function indexAll(docs: { file: string; md: string }[]): Map<string, Entry> {
  const index = new Map<string, Entry>()

  for (const one of docs) {
    // Il ticket a cui il documento appartiene, letto dal path della cartella.
    // Serve a dare a ogni voce un secondo nome qualificato, `8333:R2`, con cui
    // una sigla si risolve nel ticket giusto quando se ne nominano due.
    const ticket = /docs\/features\/(\d+)-/.exec(one.file)?.[1]

    for (const [key, entry] of indexDocument(one.md, one.file)) {
      if (!index.has(key)) index.set(key, entry)
      if (ticket !== undefined) {
        const qualified = `${ticket}:${key}`
        if (!index.has(qualified)) index.set(qualified, entry)
      }
    }
  }

  return index
}

/**
 * I riferimenti nominati in un testo, con la posizione in cui compaiono.
 *
 * Due forme, entrambe cose che wm-plan scrive davvero: la sigla nuda — `R1`,
 * `T4` — e il link a un'àncora, `[notes.md](notes.md#t4-…)` o `(#r1)`. Il link
 * si risolve prima sull'àncora intera e poi, se non la trova, sulla sigla che
 * la apre: un link a uno slug lungo trova lo stesso la sua voce.
 */
export function references(
  text: string,
  index: Map<string, Entry>,
): { at: number; entry: Entry }[] {
  const found: { at: number; entry: Entry }[] = []
  // I tratti di testo gia consumati da un link: la sigla che sta dentro
  // l'ancora — `#t4--la-sovrapposizione` — e lo stesso riferimento, non un
  // secondo, e contarla due volte riempirebbe il pane di doppioni.
  const taken: { from: number; to: number }[] = []

  for (const hit of text.matchAll(/\(([^)\s]*#([^)\s]+))\)/g)) {
    const anchor = hit[2]
    if (anchor === undefined || hit.index === undefined) continue
    taken.push({ from: hit.index, to: hit.index + hit[0].length })
    const opening = anchor.split('-')[0] ?? ''
    const entry =
      index.get(anchor) ?? index.get(opening) ?? index.get(opening.toUpperCase())
    if (entry) found.push({ at: hit.index, entry })
  }

  // Dove sono nominati i ticket: una sigla che viene dopo `oc:8333` si
  // intende in quel ticket, perche e cosi che se ne parla — «fatto in
  // oc:8333, vedi il suo R2». Senza questo vincerebbe sempre la feature
  // corrente, e il pane mostrerebbe il requisito sbagliato con lo stesso nome.
  const tickets: { at: number; number: string }[] = []
  for (const hit of text.matchAll(/\boc:(\d+)\b/g)) {
    if (hit[1] !== undefined && hit.index !== undefined) {
      tickets.push({ at: hit.index, number: hit[1] })
    }
  }

  // Anche minuscola: `t9` scritto di fretta e la stessa cosa di `T9`, e il
  // pane non deve fare lo schizzinoso. Il rischio di prendere una parola a
  // caso non c'e: una sigla conta solo se l'indice la conosce davvero.
  for (const hit of text.matchAll(/\b([A-Za-z]+\d+)\b/g)) {
    const written = hit[1]
    if (written === undefined || hit.index === undefined) continue
    const at = hit.index
    if (taken.some(one => at >= one.from && at < one.to)) continue
    const code = written.toUpperCase()
    // La qualificazione vale dentro la frase, non per tutto il messaggio:
    // «fatto in oc:8968, vedi il suo R2. Qui siamo su R3» parla di due
    // ticket diversi, e il punto e quello che li separa.
    const before = [...tickets]
      .reverse()
      .find(one => one.at < at && !/[.;\n]/.test(text.slice(one.at, at)))
    const qualified = before === undefined ? undefined : index.get(`${before.number}:${code}`)

    // Se il ticket nominato e fra quelli letti ma quella sigla li dentro non
    // c'e, la si lascia non risolta: ricadere sulla sigla della feature
    // corrente mostrerebbe un requisito diverso con lo stesso nome, cioe un
    // dato sbagliato al posto di un dato mancante.
    const known =
      before !== undefined && [...index.keys()].some(k => k.startsWith(`${before.number}:`))
    if (qualified === undefined && known) continue

    const entry = qualified ?? index.get(code)
    if (entry) found.push({ at, entry })
  }

  return found.sort((a, b) => a.at - b.at)
}

/**
 * I file markdown nominati in un testo.
 *
 * Due forme: il link, `[notes.md](docs/features/x/notes.md#r1)`, e il path
 * nudo che si scrive parlando, `docs/features/x/plan.md`. Serve a sapere
 * quali documenti la sessione sta toccando: sono quelli da indicizzare, non
 * solo i tre della cartella corrente.
 */
export function citedPaths(text: string): string[] {
  const found = new Set<string>()

  for (const hit of text.matchAll(/\]\(([^)\s]+\.md)(?:#[^)\s]*)?\)/g)) {
    if (hit[1] !== undefined) found.add(hit[1])
  }

  for (const hit of text.matchAll(/(?:^|[\s`"'(])([\w./-]+\.md)\b/g)) {
    if (hit[1] !== undefined) found.add(hit[1])
  }

  return [...found]
}

/**
 * Le sigle nominate in un testo, risolte o no.
 *
 * `references` restituisce solo quelle che l'indice conosce: questa dice
 * tutto quello che la sessione ha nominato, così chi disegna il pane può
 * accorgersi di una sigla citata che nei documenti non esiste. Un riferimento
 * che non si trova è un difetto del documento, e va mostrato, non taciuto.
 */
export function namedCodes(text: string): string[] {
  const found = new Set<string>()

  for (const hit of text.matchAll(/\b([A-Z]+\d+)\b/g)) {
    if (hit[1] !== undefined) found.add(hit[1])
  }

  return [...found]
}

/**
 * I ticket nominati in un testo, come numero: `oc:8333` da `8333`.
 *
 * Un ticket non è un path, ma indica un documento quanto un path: «questo è
 * già stato fatto in oc:8333» vuol dire che l'overview di quel ticket va
 * letta. Chi raccoglie i documenti traduce il numero nella sua cartella,
 * `docs/features/<numero>-<titolo>`.
 */
export function citedTickets(text: string): string[] {
  const found = new Set<string>()

  for (const hit of text.matchAll(/\boc:(\d+)\b/g)) {
    if (hit[1] !== undefined) found.add(hit[1])
  }

  return [...found]
}
