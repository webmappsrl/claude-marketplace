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
    for (const [chiave, voce] of indicizza(uno.md, uno.file)) {
      if (!indice.has(chiave)) indice.set(chiave, voce)
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

  for (const hit of testo.matchAll(/\(([^)\s]*#([^)\s]+))\)/g)) {
    const ancora = hit[2]
    if (ancora === undefined || hit.index === undefined) continue
    const apertura = ancora.split('-')[0] ?? ''
    const voce =
      indice.get(ancora) ?? indice.get(apertura) ?? indice.get(apertura.toUpperCase())
    if (voce) trovati.push({ at: hit.index, voce })
  }

  for (const hit of testo.matchAll(/\b([A-Z]+\d+)\b/g)) {
    const sigla = hit[1]
    if (sigla === undefined || hit.index === undefined) continue
    const voce = indice.get(sigla)
    if (voce) trovati.push({ at: hit.index, voce })
  }

  return trovati.sort((a, b) => a.at - b.at)
}
