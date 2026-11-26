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

export type Voce = {
  /** L'àncora con cui ci si arriva: `r1`, `t4-la-sovrapposizione-vera`. */
  id: string
  /** La sigla con cui lo si nomina parlando: `R1`, `T4`. Assente se non c'è. */
  sigla?: string
  /** Il testo del titolo, senza sigla né markdown. */
  titolo: string
  /** Il blocco che l'elemento apre, fino al successivo di pari rango. */
  corpo: string
  /** Il file in cui sta, relativo alla cartella della feature. */
  file: string
  /** 1-6 per un heading, 0 per un'àncora dentro un elenco o in riga. */
  livello: number
}

/** L'àncora esplicita in una riga: `<a id="r1"></a>`. */
const ANCORA = /<a id="([^"]+)"><\/a>/
/** Un heading markdown, con il suo livello. */
const HEADING = /^(#{1,6})\s+(.*)$/
/**
 * La sigla in testa a un titolo: `**R1** — testo`, `R1 — testo`, `T4 — testo`.
 *
 * Una lettera e almeno una cifra, perché una parola qualunque non diventi una
 * sigla per sbaglio, e solo in apertura: `vedi R3` dentro un titolo non lo è.
 */
const SIGLA_IN_TESTA = /^\*{0,2}([A-Z]+\d+)\*{0,2}\s*[—–-]\s*/

/**
 * Lo slug di un heading, come lo calcola GitHub per l'àncora automatica.
 *
 * Gli spazi NON si collassano: ogni spazio diventa un trattino, uno per uno.
 * È il motivo per cui un titolo come `T4 — La sovrapposizione` dà
 * `t4--la-sovrapposizione`, con due trattini dove stava il trattino lungo, ed
 * è la forma che wm-plan scrive davvero nei link ai documenti.
 */
export function slug(testo: string): string {
  return testo
    .toLowerCase()
    .replace(/<[^>]+>/g, '')
    .replace(/[`*_\[\]()]/g, '')
    .replace(/[^\p{L}\p{N}\s-]/gu, '')
    .trim()
    .replace(/ /g, '-')
}

/** Toglie il markdown dal titolo: nel pane è solo rumore. */
function testoPulito(riga: string): string {
  return riga
    .replace(/<[^>]+>/g, '')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/\s+/g, ' ')
    .trim()
}

type Segnato = {
  riga: number
  id: string
  sigla?: string
  titolo: string
  livello: number
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
export function indicizza(md: string, file: string): Map<string, Voce> {
  const righe = md.split('\n')
  const segnati: Segnato[] = []

  righe.forEach((riga, n) => {
    const head = HEADING.exec(riga)
    const ancora = ANCORA.exec(riga)

    if (head) {
      const livello = (head[1] ?? '').length
      const resto = (head[2] ?? '').replace(ANCORA, '').trim()
      const sigla = SIGLA_IN_TESTA.exec(resto)?.[1]
      const titolo = testoPulito(resto.replace(SIGLA_IN_TESTA, ''))
      // L'àncora esplicita vince sullo slug: è quella che non cambia quando il
      // titolo si riformula, ed è quella che wm-plan scrive nei link.
      segnati.push({ riga: n, id: ancora?.[1] ?? slug(resto), sigla, titolo, livello })
      return
    }

    if (!ancora) return

    const dopo = riga.slice((ancora.index ?? 0) + ancora[0].length).trim()
    const sigla = SIGLA_IN_TESTA.exec(dopo)?.[1]
    const titolo = testoPulito(dopo.replace(SIGLA_IN_TESTA, ''))
    // Livello 0: un'àncora fuori da un heading non apre una sezione, quindi
    // il suo blocco si chiude al primo elemento indicizzato che segue.
    segnati.push({ riga: n, id: ancora[1] ?? '', sigla, titolo, livello: 0 })
  })

  const indice = new Map<string, Voce>()

  segnati.forEach((uno, i) => {
    let fine = righe.length
    for (let j = i + 1; j < segnati.length; j++) {
      const altro = segnati[j]
      if (altro === undefined) continue
      // Un heading chiude il blocco precedente quando è di pari rango o più
      // alto; un elemento senza livello chiude solo gli altri senza livello.
      const chiude = uno.livello === 0 ? true : altro.livello !== 0 && altro.livello <= uno.livello
      if (chiude) {
        fine = altro.riga
        break
      }
    }

    const corpo = righe
      .slice(uno.riga + 1, fine)
      .join('\n')
      .trim()

    const voce: Voce = {
      id: uno.id,
      sigla: uno.sigla,
      titolo: uno.titolo,
      corpo,
      file,
      livello: uno.livello,
    }

    if (uno.id !== '') indice.set(uno.id, voce)
    // La sigla è un secondo nome per la stessa voce: `R1` e `r1` portano lì.
    if (uno.sigla !== undefined && !indice.has(uno.sigla)) indice.set(uno.sigla, voce)
  })

  return indice
}

/**
 * Indicizza più documenti in un indice solo.
 *
 * Le chiavi di un file letto dopo non sovrascrivono quelle già presenti:
 * `overview.md` si passa per primo, perché è lì che un identificativo è
 * definito, e `plan.md` o `notes.md` aggiungono i loro senza coprirlo.
 */
export function indicizzaTutti(documenti: { file: string; md: string }[]): Map<string, Voce> {
  const indice = new Map<string, Voce>()

  for (const uno of documenti) {
    // Il ticket a cui il documento appartiene, letto dal path della cartella.
    // Serve a dare a ogni voce un secondo nome qualificato, `8333:R2`, con cui
    // una sigla si risolve nel ticket giusto quando se ne nominano due.
    const ticket = /docs\/features\/(\d+)-/.exec(uno.file)?.[1]

    for (const [chiave, voce] of indicizza(uno.md, uno.file)) {
      if (!indice.has(chiave)) indice.set(chiave, voce)
      if (ticket !== undefined) {
        const qualificata = `${ticket}:${chiave}`
        if (!indice.has(qualificata)) indice.set(qualificata, voce)
      }
    }
  }

  return indice
}

/**
 * I riferimenti nominati in un testo, con la posizione in cui compaiono.
 *
 * Due forme, entrambe cose che wm-plan scrive davvero: la sigla nuda — `R1`,
 * `T4` — e il link a un'àncora, `[notes.md](notes.md#t4-…)` o `(#r1)`. Il link
 * si risolve prima sull'àncora intera e poi, se non la trova, sulla sigla che
 * la apre: un link a uno slug lungo trova lo stesso la sua voce.
 */
export function riferimenti(
  testo: string,
  indice: Map<string, Voce>,
): { at: number; voce: Voce }[] {
  const trovati: { at: number; voce: Voce }[] = []
  // I tratti di testo gia consumati da un link: la sigla che sta dentro
  // l'ancora — `#t4--la-sovrapposizione` — e lo stesso riferimento, non un
  // secondo, e contarla due volte riempirebbe il pane di doppioni.
  const presi: { da: number; a: number }[] = []

  for (const hit of testo.matchAll(/\(([^)\s]*#([^)\s]+))\)/g)) {
    const ancora = hit[2]
    if (ancora === undefined || hit.index === undefined) continue
    presi.push({ da: hit.index, a: hit.index + hit[0].length })
    const apertura = ancora.split('-')[0] ?? ''
    const voce =
      indice.get(ancora) ?? indice.get(apertura) ?? indice.get(apertura.toUpperCase())
    if (voce) trovati.push({ at: hit.index, voce })
  }

  // Dove sono nominati i ticket: una sigla che viene dopo `oc:8333` si
  // intende in quel ticket, perche e cosi che se ne parla — «fatto in
  // oc:8333, vedi il suo R2». Senza questo vincerebbe sempre la feature
  // corrente, e il pane mostrerebbe il requisito sbagliato con lo stesso nome.
  const ticketA: { at: number; numero: string }[] = []
  for (const hit of testo.matchAll(/\boc:(\d+)\b/g)) {
    if (hit[1] !== undefined && hit.index !== undefined) {
      ticketA.push({ at: hit.index, numero: hit[1] })
    }
  }

  // Anche minuscola: `t9` scritto di fretta e la stessa cosa di `T9`, e il
  // pane non deve fare lo schizzinoso. Il rischio di prendere una parola a
  // caso non c'e: una sigla conta solo se l'indice la conosce davvero.
  for (const hit of testo.matchAll(/\b([A-Za-z]+\d+)\b/g)) {
    const scritta = hit[1]
    if (scritta === undefined || hit.index === undefined) continue
    const dove = hit.index
    if (presi.some(uno => dove >= uno.da && dove < uno.a)) continue
    const sigla = scritta.toUpperCase()
    // La qualificazione vale dentro la frase, non per tutto il messaggio:
    // «fatto in oc:8968, vedi il suo R2. Qui siamo su R3» parla di due
    // ticket diversi, e il punto e quello che li separa.
    const prima = [...ticketA]
      .reverse()
      .find(
        uno =>
          uno.at < (hit.index ?? 0) &&
          !/[.;\n]/.test(testo.slice(uno.at, hit.index ?? 0)),
      )
    const qualificata = prima === undefined ? undefined : indice.get(`${prima.numero}:${sigla}`)

    // Se il ticket nominato e fra quelli letti ma quella sigla li dentro non
    // c'e, la si lascia non risolta: ricadere sulla sigla della feature
    // corrente mostrerebbe un requisito diverso con lo stesso nome, cioe un
    // dato sbagliato al posto di un dato mancante.
    const letto =
      prima !== undefined && [...indice.keys()].some(k => k.startsWith(`${prima.numero}:`))
    if (qualificata === undefined && letto) continue

    const voce = qualificata ?? indice.get(sigla)
    if (voce) trovati.push({ at: hit.index, voce })
  }

  return trovati.sort((a, b) => a.at - b.at)
}

/**
 * I file markdown nominati in un testo.
 *
 * Due forme: il link, `[notes.md](docs/features/x/notes.md#r1)`, e il path
 * nudo che si scrive parlando, `docs/features/x/plan.md`. Serve a sapere
 * quali documenti la sessione sta toccando: sono quelli da indicizzare, non
 * solo i tre della cartella corrente.
 */
export function pathCitati(testo: string): string[] {
  const trovati = new Set<string>()

  for (const hit of testo.matchAll(/\]\(([^)\s]+\.md)(?:#[^)\s]*)?\)/g)) {
    if (hit[1] !== undefined) trovati.add(hit[1])
  }

  for (const hit of testo.matchAll(/(?:^|[\s`"'(])([\w./-]+\.md)\b/g)) {
    if (hit[1] !== undefined) trovati.add(hit[1])
  }

  return [...trovati]
}

/**
 * Le sigle nominate in un testo, risolte o no.
 *
 * `riferimenti` restituisce solo quelle che l'indice conosce: questa dice
 * tutto quello che la sessione ha nominato, così chi disegna il pane può
 * accorgersi di una sigla citata che nei documenti non esiste. Un riferimento
 * che non si trova è un difetto del documento, e va mostrato, non taciuto.
 */
export function sigleNominate(testo: string): string[] {
  const trovate = new Set<string>()

  for (const hit of testo.matchAll(/\b([A-Z]+\d+)\b/g)) {
    if (hit[1] !== undefined) trovate.add(hit[1])
  }

  return [...trovate]
}

/**
 * I ticket nominati in un testo, come numero: `oc:8333` da `8333`.
 *
 * Un ticket non è un path, ma indica un documento quanto un path: «questo è
 * già stato fatto in oc:8333» vuol dire che l'overview di quel ticket va
 * letta. Chi raccoglie i documenti traduce il numero nella sua cartella,
 * `docs/features/<numero>-<titolo>`.
 */
export function ticketCitati(testo: string): string[] {
  const trovati = new Set<string>()

  for (const hit of testo.matchAll(/\boc:(\d+)\b/g)) {
    if (hit[1] !== undefined) trovati.add(hit[1])
  }

  return [...trovati]
}
