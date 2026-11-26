/**
 * Le letture di contorno del pane: le etichette definite parlando, la pulizia
 * del markdown, e il ticket e lo stato scritti in testa a un `overview.md`.
 *
 * Il riconoscimento dei requisiti non sta piu qui: lo fa `documents.ts`,
 * indicizzando ancore e heading invece di inseguire la forma di una riga.
 */

/**
 * Le etichette definite in un messaggio, nella forma `- **NOME** — testo`.
 *
 * NOME in maiuscolo, almeno tre caratteri: cosi una parola qualunque del
 * discorso non diventa una definizione per sbaglio.
 */
export function parseLabels(text: string): Map<string, string> {
  const found = new Map<string, string>()

  for (const line of text.split('\n')) {
    const hit = /^\s*[-*]\s*\*\*([A-Z][A-Z0-9_-]{2,})\*\*\s*[—-]\s*(.+)$/.exec(line)
    if (!hit) continue
    const name = hit[1] ?? ''
    const body = (hit[2] ?? '').trim()
    if (name !== '' && body !== '') found.set(name, body)
  }

  return found
}

/**
 * Il testo senza i pezzi fra backtick.
 *
 * Una sigla fra backtick e la sigla di cui si parla, non il requisito a cui ci
 * si riferisce. Gli spazi restano al posto del codice, cosi le posizioni non
 * si spostano e l'ordine per posizione nel testo resta giusto.
 */
export function withoutCode(text: string): string {
  return text
    .replace(/```[\s\S]*?```/g, block => ' '.repeat(block.length))
    .replace(/`[^`\n]*`/g, piece => ' '.repeat(piece.length))
}

/** Toglie il markdown: nel pane non significa niente ed e solo rumore. */
export function clean(text: string): string {
  return text
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/\s+/g, ' ')
    .trim()
}

/**
 * Il ticket e lo stato scritti in testa a un `overview.md`.
 *
 * Le due righe le detta `wm-plan`: `> Ticket: oc:<ID>` e
 * `> Stato: approvata il <data>` o `> Stato: chiusa il <data> · approvata il <data>`.
 * Senza ticket la prima riga non c'è, e lo stato esiste solo da quando l'overview
 * viene approvata: entrambi i casi tornano `undefined`, non una stringa vuota.
 */
export function overviewHead(
  md: string,
  dir?: string,
): { ticket?: string; status?: string } {
  const status = /^>\s*Stato:\s*(.+?)\s*$/m.exec(md)?.[1]
  const ticket =
    /^>\s*Ticket:\s*(oc:\d+)/m.exec(md)?.[1] ??
    // La riga di testa e la forma prescritta, ma non e l'unico posto dove
    // l'ID e scritto: il titolo lo porta quasi sempre, e il nome della
    // cartella lo porta per contratto (`<ID>-<titolo-in-kebab-case>`).
    // Mostrare `senza ticket` con l'ID scritto due volte li accanto e un
    // dato sbagliato, non un dato mancante.
    /^#\s+(?:oc:)?(\d+)\b/m.exec(md)?.[1]?.replace(/^/, 'oc:') ??
    (dir === undefined ? undefined : /(?:^|\/)(\d+)-/.exec(dir)?.[1]?.replace(/^/, 'oc:'))

  return { ticket, status }
}
