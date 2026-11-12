#!/usr/bin/env bash
# Controllo meccanico dei documenti di una feature di wm-plan: overview.md, plan.md, notes.md.
#
# Uso:  feature-docs-check.sh <docs/features/<slug>> <prima-del-commit|chiusura>
#
#   prima-del-commit  prima del primo commit del dev: l'overview è ancora quella approvata
#                     (impronta registrata in notes.md)? link e voci delle note sono in regola?
#   chiusura          dopo la chiusura dell'overview: ogni differenza dalla versione approvata,
#                     cioè l'overview del primo commit, ha la sua voce in notes.md, e viceversa;
#                     ogni requisito ha la sua riga Verifica, e il test o simbolo citato esiste.
#
# Output: un rilievo per riga, campi separati da TAB:  <severita>\t<id>\t<messaggio>
#   error  va corretto prima di proseguire
#   hint   da guardare, non blocca (per esempio «oggi» nell'overview chiusa)
# Le differenze fra overview approvata e attuale, quando servono, vanno su stderr.
#
# Exit:   0 nessun error, 1 almeno un error, 2 uso sbagliato
#
# LIMITE, da conoscere: lo script verifica la coerenza fra i documenti, non la verità di ciò che
# dicono. Che «Decide: dev» sia vero, o che il test citato dimostri davvero il requisito, lo
# giudicano il dev e la review.

set -uo pipefail

uso() { echo "uso: $0 <docs/features/<slug>> <prima-del-commit|chiusura>" >&2; exit 2; }
[ $# -eq 2 ] || uso
case "$2" in prima-del-commit|chiusura) ;; *) uso ;; esac
[ -d "$1" ] || { printf 'error\tcartella-assente\t%s non esiste\n' "$1"; exit 1; }
DIR="$(cd "$1" && pwd)"; FASE="$2"
OV="$DIR/overview.md"; NO="$DIR/notes.md"; PL="$DIR/plan.md"
TMP="$(mktemp -d)"; trap 'rm -rf "$TMP"' EXIT

ERRORI=0
emetti() { # severita, id, messaggio
  printf '%s\t%s\t%s\n' "$1" "$2" "$3"
  [ "$1" = error ] && ERRORI=$((ERRORI+1))
  return 0
}

for f in "$OV" "$NO" "$PL"; do
  [ -f "$f" ] || emetti error file-assente "$(basename "$f") non esiste in $DIR"
done
[ "$ERRORI" -eq 0 ] || exit 1

# --- Parser -----------------------------------------------------------------------------------

# Requisiti dell'overview. Forma attesa:
#   - <a id="r3"></a>**R3** — testo, anche su più righe rientrate
#     - Verifica: [File.php](percorso) › `nome del test o del simbolo`
#     - Non fatto: [motivo](notes.md#...)
# Stampa:  T<TAB>r3<TAB>testo normalizzato   |   V<TAB>r3<TAB>riga di verifica   |   N<TAB>r3
REQ_AWK='
function chiudi() {
  if (id != "") {
    gsub(/[[:space:]]+/, " ", testo); sub(/^ /, "", testo); sub(/ $/, "", testo)
    print "T\t" id "\t" testo
  }
  id = ""; testo = ""; sotto = 0
}
/^- <a id="r[0-9]+"><\/a>/ {
  chiudi()
  match($0, /id="r[0-9]+"/); id = substr($0, RSTART + 4, RLENGTH - 5)
  riga = $0
  sub(/^- <a id="r[0-9]+"><\/a>/, "", riga)
  sub(/^\*\*R[0-9]+\*\*[[:space:]]*(—|-)?[[:space:]]*/, "", riga)
  testo = riga
  next
}
id == "" { next }
/^[[:space:]]*$/ || /^#/ || /^- / { chiudi(); next }
/^[[:space:]]+- (\*\*)?Verifica:/ {
  sotto = 1; r = $0; sub(/^[[:space:]]+- (\*\*)?Verifica:(\*\*)?[[:space:]]*/, "", r)
  print "V\t" id "\t" r; next
}
/^[[:space:]]+- (\*\*)?Non fatto:/ { sotto = 1; print "N\t" id; next }
/^[[:space:]]+- / { sotto = 1; next }
sotto == 0 { testo = testo " " $0 }
END { chiudi() }'

# Voci della sezione «## Modifiche ai requisiti» di notes.md. Forma attesa:
#   ### 2026-11-12 — titolo
#   - **Requisito:** [R3](overview.md#r3)
#   - **Effetto:** aggiunto | modificato | tolto | non fatto
#   - **Decide:** dev
#   - **Testo nuovo:**
#     > testo del requisito, senza identificativo; i link agli altri requisiti come
#     > [R1](overview.md#r1), che nel confronto vale come [R1](#r1)
# Stampa:  E<TAB>titolo<TAB>r3<TAB>effetto<TAB>decide<TAB>testo nuovo normalizzato
NOTE_AWK='
function chiudi() {
  if (titolo != "") {
    gsub(/[[:space:]]+/, " ", nuovo); sub(/^ /, "", nuovo); sub(/ $/, "", nuovo)
    print "E\t" titolo "\t" req "\t" eff "\t" dec "\t" nuovo
  }
  titolo = ""; req = ""; eff = ""; dec = ""; nuovo = ""; innuovo = 0
}
/^## / { chiudi(); dentro = ($0 ~ /^## Modifiche ai requisiti/); next }
!dentro { next }
/^### / { chiudi(); titolo = substr($0, 5); next }
titolo == "" { next }
/^- \*\*Requisito:\*\*/ { innuovo = 0; if (match($0, /#r[0-9]+\)/)) req = substr($0, RSTART + 1, RLENGTH - 2); next }
/^- \*\*Effetto:\*\*/ {
  innuovo = 0; e = $0; sub(/^- \*\*Effetto:\*\*[[:space:]]*/, "", e); sub(/[[:space:]]+$/, "", e)
  eff = tolower(e); next
}
/^- \*\*Decide:\*\*/ { innuovo = 0; d = $0; sub(/^- \*\*Decide:\*\*[[:space:]]*/, "", d); dec = d; next }
/^- \*\*Testo nuovo:\*\*/ { innuovo = 1; next }
/^- / { innuovo = 0; next }
innuovo && /^[[:space:]]*>/ { t = $0; sub(/^[[:space:]]*>[[:space:]]?/, "", t); gsub(/\]\(overview\.md#/, "](#", t); nuovo = nuovo " " t; next }
END { chiudi() }'

slug() { # titolo Markdown -> àncora GitHub
  perl -CSD -ne 'chomp; $_ = lc; s/[^\p{L}\p{N} _-]//g; s/ /-/g; print "$_\n"'
}

# --- Controlli comuni alle due fasi -----------------------------------------------------------

awk "$REQ_AWK" "$OV" > "$TMP/ora"
awk "$NOTE_AWK" "$NO" > "$TMP/voci"
# Requisiti tolti: la loro àncora non esiste più, e il link della voce «tolto» resta rotto per forza.
awk -F'\t' '$4 == "tolto" { print $3 }' "$TMP/voci" > "$TMP/tolti"

# Identificativi definiti due volte.
grep -oE 'id="r[0-9]+"' "$OV" | sort | uniq -d | while read -r d; do echo "$d"; done > "$TMP/doppi"
while read -r d; do emetti error id-doppio "overview.md: $d definito più di una volta"; done < "$TMP/doppi"

# Link: il file esiste, l'àncora esiste.
for f in "$OV" "$PL" "$NO"; do
  while IFS= read -r m; do
    n="${m%%:*}"; t="${m#*:}"; t="${t#](}"; t="${t%)}"
    case "$t" in http://*|https://*|mailto:*) continue ;; esac
    p="${t%%#*}"; frag=""; [ "$p" != "$t" ] && frag="${t#*#}"
    if [ -z "$p" ]; then tgt="$f"; else tgt="$(dirname "$f")/$p"; fi
    if [ ! -e "$tgt" ]; then
      emetti error link-rotto "$(basename "$f"):$n: $p non esiste"; continue
    fi
    [ -n "$frag" ] || continue
    case "$frag" in
      r[0-9]*)
        grep -qF "id=\"$frag\"" "$tgt" || grep -qxF "$frag" "$TMP/tolti" || emetti error ancora-assente "$(basename "$f"):$n: $p#$frag — nessun requisito con id=\"$frag\"" ;;
      L[0-9]*)
        emetti error link-a-riga "$(basename "$f"):$n: $t — i numeri di riga cambiano, cita il nome del test o del simbolo" ;;
      *)
        case "$tgt" in
          *.md) grep '^#' "$tgt" | sed 's/^#* *//' | slug | grep -qxF "$frag" \
                  || emetti error ancora-assente "$(basename "$f"):$n: $p#$frag — nessun titolo con questa àncora" ;;
        esac ;;
    esac
  done < <(grep -noE '\]\([^) ]+\)' "$f")
done

# Identificativi scritti come testo invece che come link.
for f in "$OV" "$PL" "$NO"; do
  sed -E 's/\[R[0-9]+\]\([^)]*\)//g; s/<a id="r[0-9]+"><\/a>\*\*R[0-9]+\*\*//g; s/`[^`]*`//g' "$f" \
    | grep -nE '(^|[^A-Za-z0-9_/#-])R[0-9]+([^A-Za-z0-9_]|$)' \
    | while IFS= read -r m; do printf '%s\n' "$m"; done > "$TMP/testo"
  while IFS= read -r m; do
    emetti error id-senza-link "$(basename "$f"):${m%%:*}: un identificativo va scritto come link, per esempio [R3](overview.md#r3)"
  done < "$TMP/testo"
done

# Voci delle note: campi obbligatori.
while IFS="$(printf '\t')" read -r _ titolo req eff dec nuovo; do
  [ -n "$req" ] || emetti error voce-senza-requisito "notes.md «${titolo}»: manca il link al requisito"
  case "$eff" in
    aggiunto|modificato|tolto|"non fatto") ;;
    *) emetti error effetto-sconosciuto "notes.md «${titolo}»: Effetto «${eff}» — ammessi: aggiunto, modificato, tolto, non fatto" ;;
  esac
  [ -n "$dec" ] || emetti error decisore-assente "notes.md «${titolo}»: manca Decide"
  case "$eff" in aggiunto|modificato)
    [ -n "$nuovo" ] || emetti error testo-nuovo-assente "notes.md «${titolo}»: un requisito $eff va riscritto per intero in «Testo nuovo»" ;;
  esac
done < "$TMP/voci"

# Divergenze dal piano: ogni task annotato nelle note ha il suo rimando ⚠️ nel piano.
awk '/^## /{d=($0 ~ /^## Divergenze dal piano/); next} d && /^### /{print substr($0,5)}' "$NO" | slug \
  | while IFS= read -r s; do grep -qF "notes.md#$s" "$PL" || echo "$s"; done > "$TMP/divergenze"
while IFS= read -r s; do
  emetti error divergenza-senza-rimando "plan.md: nessuna riga ⚠️ rimanda a notes.md#$s"
done < "$TMP/divergenze"

# --- Prima del primo commit -------------------------------------------------------------------

if [ "$FASE" = prima-del-commit ]; then
  grep -q '^> Stato: approvata il' "$OV" || emetti error stato-assente "overview.md: manca la riga «> Stato: approvata il …»"
  IMPRONTA="$(grep -oE 'impronta `[0-9a-f]{7,40}`' "$NO" | head -1 | grep -oE '[0-9a-f]{7,40}')"
  if [ -z "$IMPRONTA" ]; then
    emetti error impronta-assente "notes.md: manca l'impronta dell'overview approvata"
  else
    ATTUALE="$(git -C "$DIR" hash-object "$OV")"
    case "$ATTUALE" in
      "$IMPRONTA"*) ;;
      *)
        emetti error overview-cambiata "overview.md è cambiata dopo l'approvazione: ogni modifica va in notes.md, «Modifiche ai requisiti»"
        if git -C "$DIR" cat-file -e "$IMPRONTA" 2>/dev/null; then
          git -C "$DIR" cat-file -p "$IMPRONTA" > "$TMP/approvata"
          diff -u "$TMP/approvata" "$OV" >&2
        fi ;;
    esac
  fi
fi

# --- Chiusura ---------------------------------------------------------------------------------

if [ "$FASE" = chiusura ]; then
  grep -q '^> Stato: chiusa il' "$OV" || emetti error stato-assente "overview.md: manca la riga «> Stato: chiusa il …»"
  grep -q '^> Stato:' "$PL" || emetti error stato-piano-assente "plan.md: manca la riga «> Stato: …» che dice se il piano è stato seguito"

  BASE="$(git -C "$DIR" log --diff-filter=A --format=%H -1 -- overview.md 2>/dev/null)"
  REL="$(git -C "$DIR" ls-files --full-name overview.md 2>/dev/null)"
  if [ -z "$BASE" ] || [ -z "$REL" ]; then
    emetti error baseline-assente "overview.md non è in nessun commit: la versione approvata è quella del primo commit del dev"
  else
    git -C "$DIR" show "$BASE:$REL" | awk "$REQ_AWK" > "$TMP/base"
    awk -F'\t' '
      FILENAME == ARGV[1] { if ($1 == "T") { b[$2] = $3; ids[$2] = 1 }; next }
      FILENAME == ARGV[2] {
        if ($1 == "T") { o[$2] = $3; ids[$2] = 1 } else if ($1 == "V") v[$2] = 1; else if ($1 == "N") nf[$2] = 1
        next
      }
      FILENAME == ARGV[3] {
        r = $3; if (r == "") next
        ha[r, $4] = 1
        if ($4 == "aggiunto" || $4 == "modificato") nuovo[r] = $6
        next
      }
      function err(id, msg) { print "error\t" id "\t" msg }
      END {
        for (id in ids) {
          R = toupper(substr(id, 1, 1)) substr(id, 2)
          inb = (id in b); ino = (id in o)
          if (inb && !ino && !ha[id, "tolto"]) err("requisito-sparito", R ": era nella versione approvata e non c\047è più, e nessuna voce «tolto» lo spiega")
          if (ino && ha[id, "tolto"]) err("tolto-ma-presente", R ": una voce lo dà per tolto, ma è ancora nell\047overview")
          if (!inb && ino && !ha[id, "aggiunto"]) err("aggiunto-senza-voce", R ": non era nella versione approvata e nessuna voce «aggiunto» lo spiega")
          if (inb && ino && b[id] != o[id] && !ha[id, "modificato"]) err("cambiato-senza-voce", R ": il testo è diverso dalla versione approvata e nessuna voce «modificato» lo spiega")
          if (inb && ino && b[id] == o[id] && ha[id, "modificato"]) err("voce-senza-modifica", R ": una voce lo dà per modificato, ma il testo è quello approvato")
          if (ino && (id in nuovo) && nuovo[id] != o[id]) err("testo-diverso", R ": il testo nell\047overview non è quello scritto in «Testo nuovo» nelle note")
          if (ha[id, "non fatto"] && !nf[id]) err("non-fatto-senza-segno", R ": una voce lo dà per non fatto, ma nell\047overview manca la riga «Non fatto»")
          if (nf[id] && !ha[id, "non fatto"]) err("segno-senza-voce", R ": è segnato «Non fatto» ma nessuna voce lo spiega")
          if (ino && !nf[id] && !v[id]) err("verifica-assente", R ": manca la riga «Verifica»")
        }
      }' "$TMP/base" "$TMP/ora" "$TMP/voci" > "$TMP/confronto"
    while IFS="$(printf '\t')" read -r _ id msg; do emetti error "$id" "$msg"; done < "$TMP/confronto"
    if [ -s "$TMP/confronto" ] && ! git -C "$DIR" diff --quiet "$BASE" -- overview.md 2>/dev/null; then
      git -C "$DIR" diff "$BASE" -- overview.md >&2
    fi
  fi

  # La verifica cita qualcosa che esiste davvero.
  while IFS="$(printf '\t')" read -r _ id riga; do
    R="R${id#r}"
    case "$riga" in
      *"a mano"*)
        echo "$riga" | grep -qE '[0-9]{4}-[0-9]{2}-[0-9]{2}' \
          || emetti error verifica-a-mano-incompleta "$R: una verifica a mano dice chi l'ha fatta, quando (AAAA-MM-GG) e come"
        continue ;;
    esac
    link="$(echo "$riga" | grep -oE '\]\([^) ]+\)' | head -1)"; link="${link#](}"; link="${link%)}"
    nome="$(echo "$riga" | sed -nE 's/.*› *`([^`]+)`.*/\1/p')"
    if [ -z "$link" ] || [ -z "$nome" ]; then
      emetti error verifica-incompleta "$R: la verifica è «[file](percorso) › \`nome del test o del simbolo\`», oppure «a mano — chi, AAAA-MM-GG: come»"
      continue
    fi
    tgt="$DIR/${link%%#*}"
    [ -f "$tgt" ] && { grep -qF -- "$nome" "$tgt" || emetti error nome-assente "$R: «${nome}» non compare in ${link%%#*}"; }
  done < <(grep '^V' "$TMP/ora")

  # «oggi» nell'overview chiusa, fuori dal «Punto di partenza»: da guardare, non blocca.
  awk '/^## /{p=($0 ~ /^## Punto di partenza/)} !p && tolower($0) ~ /(^|[^a-zà-ù])oggi([^a-zà-ù]|$)/ {print NR}' "$OV" \
    | while read -r n; do printf 'hint\toggi\toverview.md:%s: «oggi» in un overview chiusa — descrive ancora lo stato giusto?\n' "$n"; done
fi

[ "$ERRORI" -eq 0 ] || exit 1
exit 0
