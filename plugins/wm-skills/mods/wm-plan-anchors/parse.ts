/**
 * Il riconoscimento del formato dei requisiti, isolato dal resto della mod.
 *
 * Sta in un file suo, senza nessun import, per tre motivi: i test lo provano
 * come funzione pura, il controllo in CI lo esegue con node per confrontarlo
 * con l'altro parser del repo (`scripts/feature-docs-check.sh`), e chi cambia
 * il formato trova qui tutto quello che deve cambiare.
 *
 * La forma che riconosce e quella che `skills/wm-plan/SKILL.md` prescrive:
 *   - <a id="r1"></a>**R1** — testo, anche su piu righe rientrate
 *     - Verifica: ...
 */

export type Req = {
  label: string
  text: string
  ux: boolean
  verifica: boolean
  nonFatto: boolean
}

/** La riga che apre un requisito: ancora, identificativo, trattino, testo. */
export const RIGA_REQUISITO = /^-\s*<a id="r(\d+)"><\/a>\s*\*\*(R\d+)\*\*\s*[—-]\s*(.*)$/

/** Legge la sezione Requisiti di un overview nel formato con ancore. */
export function parseRequirements(md: string): Map<string, Req> {
  const found = new Map<string, Req>()
  let current: Req | undefined

  for (const line of md.split('\n')) {
    const head = RIGA_REQUISITO.exec(line)
    if (head) {
      const body = (head[3] ?? '').trim()
      current = {
        label: head[2] ?? '',
        text: body.replace(/^\[UX\]\s*/, ''),
        ux: body.startsWith('[UX]'),
        verifica: false,
        nonFatto: false,
      }
      found.set(current.label, current)
      continue
    }
    if (!current) continue
    if (/^\s+-\s*\*{0,2}(Verifica|Non fatto)/i.test(line)) {
      if (/Verifica/i.test(line)) current.verifica = true
      if (/Non fatto/i.test(line)) current.nonFatto = true
      continue
    }
    if (/^\s+\S/.test(line) && !/^\s*-\s/.test(line)) {
      current.text = `${current.text} ${line.trim()}`
      continue
    }
    if (line.trim() !== '') current = undefined
  }

  return found
}

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
export function senzaCodice(text: string): string {
  return text
    .replace(/```[\s\S]*?```/g, blocco => ' '.repeat(blocco.length))
    .replace(/`[^`\n]*`/g, pezzo => ' '.repeat(pezzo.length))
}

/** Toglie il markdown: nel pane non significa niente ed e solo rumore. */
export function pulisci(text: string): string {
  return text
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/\s+/g, ' ')
    .trim()
}

/**
 * Gli identificativi riconosciuti in un testo, uno per riga.
 *
 * E il punto d'ingresso del controllo in CI: `node parse.ts <file>` stampa
 * cosa il parser riconosce, e lo script confronta quell'elenco con quello
 * che `SKILL.md` prescrive e con quello che riconosce il parser awk.
 */
export function identificativi(md: string): string[] {
  return [...parseRequirements(md).keys()]
}
