#!/usr/bin/env bash
# Test di geohub-codice-riferimento.sh su repo git locali al posto di GitHub: nessuna rete.
set -uo pipefail
QUI="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$QUI/../geohub-codice-riferimento.sh"
FALLITI=0

verifica() { # nome, atteso, ottenuto
  if [ "$2" = "$3" ]; then
    printf '✔ %s\n' "$1"
  else
    printf '✘ %s\n   atteso:   %s\n   ottenuto: %s\n' "$1" "$2" "$3"
    FALLITI=$((FALLITI+1))
  fi
}

TMP=$(mktemp -d); trap 'rm -rf "$TMP"' EXIT
R="$TMP/remoti"; mkdir -p "$R"
export GIT_AUTHOR_NAME=test GIT_AUTHOR_EMAIL=test@example.com GIT_COMMITTER_NAME=test GIT_COMMITTER_EMAIL=test@example.com

# Un repo con due commit: il fissato è il primo, la testa del branch è il secondo, così il test
# prova che si scarica il commit fissato e non l'ultimo.
due_commit() { # nome → stampa il primo commit
  local d="$TMP/lavoro/$1"
  git init --quiet -b main "$d"
  echo fissato > "$d/versione"; git -C "$d" add versione; git -C "$d" commit --quiet -m uno
  local primo; primo=$(git -C "$d" rev-parse HEAD)
  echo ultimo > "$d/versione"; git -C "$d" commit --quiet -am due
  git clone --quiet --bare "$d" "$R/$1.git"
  echo "$primo"
}
# Un repo che fissa un submodule al commit dato, sul branch dato.
con_submodule() { # nome, branch, percorso, commit
  local d="$TMP/lavoro/$1"
  git init --quiet -b "$2" "$d"
  git -C "$d" update-index --add --cacheinfo "160000,$4,$3"
  git -C "$d" commit --quiet -m submodule
  git clone --quiet --bare "$d" "$R/$1.git"
}

C_CORE=$(due_commit wm-core)
C_PKG=$(due_commit wm-package)
due_commit geohub >/dev/null
con_submodule wm-webapp main src/app/shared/wm-core "$C_CORE"
con_submodule maphub develop wm-package "$C_PKG"
export GIC_GITHUB_BASE="file://$R"

"$SCRIPT" maphubdev "$TMP/out" >/dev/null 2>"$TMP/err"; RC=$?
verifica "shard di sviluppo: exit 0" "0" "$RC"
verifica "wm-core al commit fissato da wm-webapp, non all'ultimo" "fissato" "$(cat "$TMP/out/codice/wm-core/versione")"
verifica "wm-package al commit fissato da maphub develop" "fissato" "$(cat "$TMP/out/codice/wm-package/versione")"
verifica "geohub da main, all'ultimo commit" "ultimo" "$(cat "$TMP/out/codice/geohub/versione")"
verifica "codice.txt con i tre repo e la provenienza" "3 1" \
  "$(wc -l < "$TMP/out/codice.txt" | tr -d ' ') $(grep -c 'wm-package .* (fissato da maphub develop)' "$TMP/out/codice.txt")"
verifica "nessun clone di appoggio lasciato nella cartella" "geohub wm-core wm-package" \
  "$(ls "$TMP/out/codice" | paste -sd' ' -)"

(cd "$TMP" && "$SCRIPT" maphubdev relativo >/dev/null 2>&1)
verifica "cartella relativa: codice.txt accanto a codice/" "si" "$([ -f "$TMP/relativo/codice.txt" ] && echo si || echo no)"

"$SCRIPT" maphub "$TMP/out2" >/dev/null 2>"$TMP/err"; RC=$?
verifica "shard di produzione senza il branch main nel repo: exit 5" "5" "$RC"
verifica "shard di produzione: il messaggio nomina il branch" "1" "$(grep -c 'branch main' "$TMP/err")"

con_submodule altro main README "$C_PKG"
"$SCRIPT" altro "$TMP/out3" >/dev/null 2>"$TMP/err"; RC=$?
verifica "repo dello shard senza submodule wm-package: exit 5" "5" "$RC"

"$SCRIPT" inesistente "$TMP/out4" >/dev/null 2>"$TMP/err"; RC=$?
verifica "repo dello shard inesistente: exit 5" "5" "$RC"

GIC_GITHUB_BASE="file://$TMP/nessuno" "$SCRIPT" maphubdev "$TMP/out5" >/dev/null 2>"$TMP/err"; RC=$?
verifica "GitHub non raggiungibile: exit 4" "4" "$RC"

"$SCRIPT" maphubdev >/dev/null 2>&1; RC=$?
verifica "argomenti mancanti: exit 2" "2" "$RC"

[ "$FALLITI" -eq 0 ] && echo "Tutti i test passati." || { echo "$FALLITI test falliti."; exit 1; }
