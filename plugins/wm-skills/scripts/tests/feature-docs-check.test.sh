#!/usr/bin/env bash
# Test di feature-docs-check.sh. Lavora in un repo git temporaneo, mai nel repo vero.
set -uo pipefail
QUI="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$QUI/../feature-docs-check.sh"
FALLITI=0

verifica() { # nome, atteso, ottenuto
  if [ "$2" = "$3" ]; then printf '✔ %s\n' "$1"
  else printf '✘ %s\n   atteso:   %s\n   ottenuto: %s\n' "$1" "$2" "$3"; FALLITI=$((FALLITI+1)); fi
}
ids() { "$SCRIPT" "$F" "$1" 2>/dev/null | awk -F'\t' '$1=="error"{print $2}' | sort -u | paste -sd, -; }

REPO="$(mktemp -d)"; trap 'rm -rf "$REPO"' EXIT
git -C "$REPO" init -q
G() { git -C "$REPO" -c user.name=test -c user.email=test@example.invalid "$@"; }
F="$REPO/docs/features/1-prova"; mkdir -p "$F" "$REPO/tests"
cat > "$REPO/tests/ShowTest.php" <<'EOF'
<?php
it('hides the code until approval', function () {});
EOF

scrivi_overview() { # testo di R2
cat > "$F/overview.md" <<EOF
> Stato: approvata il 2026-11-12

# Prova

## Punto di partenza
Oggi la domanda non si legge.

## Requisiti
- <a id="r1"></a>**R1** — la lettura risponde con lo stato della domanda.
- <a id="r2"></a>**R2** — $1
  come richiesto da [R1](#r1).
EOF
}
scrivi_overview "il numero non compare finché la domanda non è approvata,"
IMPRONTA="$(git -C "$REPO" hash-object -w "$F/overview.md")"
cat > "$F/plan.md" <<'EOF'
> Ticket: oc:1

# Piano
### Task 1
> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#task-1-lettura)
EOF
cat > "$F/notes.md" <<EOF
> Overview approvata il 2026-11-12 · impronta \`$IMPRONTA\`

# Notes

## Modifiche ai requisiti

## Divergenze dal piano, task per task
### Task 1 lettura
Diverso.
EOF

verifica "prima del commit, tutto in regola" "" "$(ids prima-del-commit)"

cp "$F/overview.md" "$REPO/salvata"
echo "Nota sparsa su R1." >> "$F/overview.md"
verifica "identificativo come testo, overview cambiata" "id-senza-link,overview-cambiata" "$(ids prima-del-commit)"
cp "$REPO/salvata" "$F/overview.md"

echo "- [R9](overview.md#r9)" >> "$F/plan.md"
verifica "link a un requisito inesistente" "ancora-assente" "$(ids prima-del-commit)"
sed -i.bak '$d' "$F/plan.md" && rm -f "$F/plan.md.bak"

sed -i.bak 's/^> ⚠️.*$//' "$F/plan.md" && rm -f "$F/plan.md.bak"
verifica "divergenza senza rimando nel piano" "divergenza-senza-rimando" "$(ids prima-del-commit)"
printf '> ⚠️ L%simplementazione ha deviato: [notes.md](notes.md#task-1-lettura)\n' "'" >> "$F/plan.md"

# Primo commit del dev: da qui la versione approvata è quella del commit.
G add -A && G commit -qm "primo commit"

# Chiusura senza aggiornare niente: mancano stato e verifiche.
verifica "chiusura incompleta" "stato-assente,stato-piano-assente,verifica-assente" "$(ids chiusura)"

chiudi() { # testo di R2, riga di verifica di R2
  scrivi_overview "$1"
  sed -i.bak 's/^> Stato: approvata il 2026-11-12$/> Stato: chiusa il 2026-11-20 · approvata il 2026-11-12/' "$F/overview.md"
  python3 - "$F/overview.md" "$2" <<'PY'
import sys
p, v = sys.argv[1], sys.argv[2]
s = open(p).read()
s = s.replace("con lo stato della domanda.\n", "con lo stato della domanda.\n  - Verifica: [ShowTest.php](../../../tests/ShowTest.php) › `hides the code until approval`\n")
s = s.rstrip("\n") + "\n  - Verifica: " + v + "\n"
open(p, "w").write(s)
PY
  rm -f "$F/overview.md.bak"
  grep -q '^> Stato:' "$F/plan.md" || sed -i.bak '1s/^/> Stato: seguito, una divergenza\n/' "$F/plan.md"; rm -f "$F/plan.md.bak"
}
chiudi "il numero non compare finché la domanda non è approvata," "[ShowTest.php](../../../tests/ShowTest.php) › \`hides the code until approval\`"
verifica "chiusura in regola" "" "$(ids chiusura)"

chiudi "il numero non compare finché la domanda non è approvata," "[ShowTest.php](../../../tests/ShowTest.php) › \`a test that does not exist\`"
verifica "test citato che non esiste" "nome-assente" "$(ids chiusura)"

chiudi "il numero compare sempre," "a mano — dev, 2026-11-20: provato su UAT"
verifica "requisito indebolito senza voce" "cambiato-senza-voce" "$(ids chiusura)"

cat >> "$F/notes.md.voce" <<'EOF'
### 2026-11-15 — Il numero si mostra sempre
- **Requisito:** [R2](overview.md#r2)
- **Effetto:** modificato
- **Decide:** dev
- **Testo nuovo:**
  > il numero compare sempre, come richiesto da [R1](overview.md#r1).
EOF
awk -v f="$F/notes.md.voce" '{print} /^## Modifiche ai requisiti/{print ""; while((getline l < f)>0) print l}' "$F/notes.md" > "$F/n" && mv "$F/n" "$F/notes.md" && rm "$F/notes.md.voce"
verifica "requisito modificato con la sua voce" "" "$(ids chiusura)"

chiudi "il numero compare sempre, anche prima," "a mano — dev, 2026-11-20: provato su UAT"
verifica "testo diverso da quello delle note" "testo-diverso" "$(ids chiusura)"

chiudi "il numero compare sempre," "a mano — il dev"
verifica "verifica a mano senza data" "verifica-a-mano-incompleta" "$(ids chiusura)"

sed -i.bak '/<a id="r2">/,$d' "$F/overview.md" && rm -f "$F/overview.md.bak"
verifica "requisito sparito senza voce (la nota punta al vuoto)" "ancora-assente,requisito-sparito" "$(ids chiusura)"

sed -i.bak 's/^- \*\*Effetto:\*\* modificato$/- **Effetto:** tolto/; /^- \*\*Testo nuovo:\*\*$/,/^  > /d' "$F/notes.md" && rm -f "$F/notes.md.bak"
verifica "requisito tolto con la sua voce" "" "$(ids chiusura)"

# --- Casi trovati dalla review della prima versione ------------------------------------------

REPO2="$(mktemp -d)"; trap 'rm -rf "$REPO" "$REPO2"' EXIT
git -C "$REPO2" init -q
G2() { git -C "$REPO2" -c user.name=test -c user.email=test@example.invalid "$@"; }
F="$REPO2/docs/features/2-prova"; mkdir -p "$F" "$REPO2/tests"
printf "<?php\nit('reads', function () {});\n" > "$REPO2/tests/ReadTest.php"
cat > "$F/overview.md" <<'EOF'
> Stato: approvata il 2026-11-12

## Requisiti
- <a id="r1"></a>**R1** — la lettura risponde con:
  - lo stato della domanda;
  - il codice, solo se approvata.
EOF
IMPRONTA="$(git -C "$REPO2" hash-object -w "$F/overview.md")"
cat > "$F/plan.md" <<'EOF'
> Ticket: oc:2

### Task 4 lettura
> ⚠️ Deviato: [notes.md](notes.md#task-41-scrittura)

```php
$this->handlers[$name]($value); // R2 nel codice non è un identificativo
```
EOF
cat > "$F/notes.md" <<EOF
> Overview approvata il 2026-11-12 · impronta \`$IMPRONTA\`

## Modifiche ai requisiti

### 2026-11-13 — La conferma ha un endpoint suo
- **Requisito:** [R2](overview.md#r2)
- **Effetto:** aggiunto
- **Decide:** dev
- **Testo nuovo:**
  > la conferma risponde con lo stato.

## Divergenze dal piano, task per task
### Task 4 lettura
x
### Task 41 scrittura
y
EOF
verifica "requisito aggiunto prima della chiusura, codice nel piano, task con nome simile" \
  "divergenza-senza-rimando" "$(ids prima-del-commit)"
printf '> ⚠️ Deviato: [notes.md](notes.md#task-4-lettura)\n' >> "$F/plan.md"
verifica "rimando aggiunto" "" "$(ids prima-del-commit)"

G2 add -A && G2 commit -qm "primo commit"
PRIMO="$(git -C "$REPO2" rev-parse HEAD)"
sed -i.bak '1s/.*/> Stato: chiusa il 2026-11-20 · approvata il 2026-11-12/; s/  - il codice, solo se approvata./  - il codice, sempre./' "$F/overview.md"
sed -i.bak '1s/^/> Stato: seguito\n/' "$F/plan.md"
cat >> "$F/overview.md" <<'EOF'

  - Verifica: [ReadTest.php](../../../tests/ReadTest.php) › `reads`
- <a id="r2"></a>**R2** — la conferma risponde con lo stato.
  - Verifica: forestas tests/Feature/Sus/ConfirmTest.php › `confirms`
EOF
rm -f "$F"/*.bak
verifica "sottoelenco indebolito senza voce; riga vuota prima di Verifica; Verifica di un altro repo" \
  "cambiato-senza-voce" "$(ids chiusura)"

sed -i.bak 's/  - il codice, sempre./  - il codice, solo se approvata./' "$F/overview.md"; rm -f "$F"/*.bak
cat >> "$F/notes.md.voce" <<'EOF'
### 2026-11-14 — Il codice non si fa
- **Requisito:** [R1](overview.md#r1)
- **Effetto:** non fatto
- **Decide:** dev

### 2026-11-16 — Il codice si fa
- **Requisito:** [R1](overview.md#r1)
- **Effetto:** fatto
- **Decide:** dev
EOF
awk -v f="$F/notes.md.voce" '{print} /^## Modifiche ai requisiti/{print ""; while((getline l < f)>0) print l}' "$F/notes.md" > "$F/n" && mv "$F/n" "$F/notes.md" && rm "$F/notes.md.voce"
verifica "«non fatto» superato da «fatto»" "" "$(ids chiusura)"

# Dopo uno squash il commit che aggiunge overview.md è il merge: il riferimento si passa.
G2 add -A && G2 commit -qm "chiusura"
G2 checkout -q --orphan squash && G2 commit -qm "squash" && SQ="$(git -C "$REPO2" rev-parse HEAD)"
sed -i.bak 's/la conferma risponde con lo stato./la conferma risponde./' "$F/overview.md"; rm -f "$F"/*.bak
verifica "dopo lo squash, con il primo commit come riferimento" "testo-diverso" \
  "$("$SCRIPT" "$F" chiusura "$PRIMO" 2>/dev/null | awk -F'\t' '$1=="error"{print $2}' | sort -u | paste -sd, -)"

# --- Casi trovati dalla review della #38 ------------------------------------------------------

REPO3="$(mktemp -d)"; trap 'rm -rf "$REPO" "$REPO2" "$REPO3"' EXIT
git -C "$REPO3" init -q
G3() { git -C "$REPO3" -c user.name=test -c user.email=test@example.invalid "$@"; }
F="$REPO3/docs/features/3-prova"; mkdir -p "$F" "$REPO3/tests"
printf "<?php\nit('reads the application', function () {});\n" > "$REPO3/tests/ReadTest.php"
cat > "$F/overview.md" <<'EOF'
> Stato: approvata il 2026-11-12

## Requisiti
- <a id="r1"></a>**R1** — la lettura risponde con lo stato.
EOF
IMPRONTA="$(git -C "$REPO3" hash-object -w "$F/overview.md")"
cat > "$F/plan.md" <<'EOF'
> Stato: seguito

1. Passo con codice rientrato:
   ```php
   $this->handlers[$name]($value);
   ```
EOF
nota() { # effetto, requisito
  printf '\n### 2026-11-1%s — voce\n- **Requisito:** [%s](overview.md#%s)\n- **Effetto:** %s\n- **Decide:** dev\n' \
    "$((RANDOM % 9))" "$(echo "$2" | tr r R)" "$2" "$1"
  [ "$1" = aggiunto ] && printf -- '- **Testo nuovo:**\n  > %s\n' "$3"
  return 0
}
{ printf '> Overview approvata il 2026-11-12 · impronta `%s`\n\n## Modifiche ai requisiti\n' "$IMPRONTA"
  nota aggiunto r2 "la conferma risponde."; nota ripristinato r2; } > "$F/notes.md"
verifica "codice rientrato nel piano; «aggiunto» poi «ripristinato»" "" "$(ids prima-del-commit)"
G3 add -A && G3 commit -qm "primo commit"
sed -i.bak '1s/.*/> Stato: chiusa il 2026-11-20/' "$F/overview.md"
cat >> "$F/overview.md" <<'EOF'
  - Verifica: [ReadTest.php](../../../tests/ReadTest.php)
    › `reads the application`
EOF
rm -f "$F"/*.bak
verifica "requisito aggiunto e poi annullato; Verifica su due righe" "" "$(ids chiusura)"

{ printf '> Overview approvata il 2026-11-12 · impronta `%s`\n\n## Modifiche ai requisiti\n' "$IMPRONTA"
  nota tolto r1; nota aggiunto r1 "tutt altro requisito."; } > "$F/notes.md"
sed -i.bak 's/la lettura risponde con lo stato./tutt altro requisito./' "$F/overview.md"; rm -f "$F"/*.bak
verifica "identificativo riusato dopo «tolto»" "id-riusato" "$(ids chiusura)"

echo
[ "$FALLITI" -eq 0 ] && echo "Tutti i test passano." || { echo "$FALLITI test falliti."; exit 1; }
