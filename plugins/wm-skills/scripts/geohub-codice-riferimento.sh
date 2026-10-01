#!/usr/bin/env bash
# Scarica da GitHub il codice che gira davvero per una verifica dell'import Geohub → shard:
# wm-core e wm-package al commit fissato da chi li pubblica, geohub da main. SOLA LETTURA.
#
# Uso:  geohub-codice-riferimento.sh <shard di arrivo> <cartella>
#
#   wm-core     submodule src/app/shared/wm-core nel main di wm-webapp: il frontend è un solo
#               bundle per tutti gli shard
#   wm-package  submodule wm-package nel repo Laravel dello shard: il nome dello shard senza
#               «dev» (maphubdev → maphub), branch develop se lo shard finisce con dev, se no main
#   geohub      main
#
# Output: il codice in <cartella>/codice/, i commit usati in <cartella>/codice.txt.
#
# Sta in un file e non nel SKILL.md perché lì Claude Code sostituisce $1, $2, $3 con gli
# argomenti con cui si lancia la skill, e le funzioni escono rotte.
#
# Exit: 0 ok, 2 uso sbagliato, 4 clone o fetch fallito, 5 repo dello shard assente, senza il
# branch o senza il submodule
#
# Test: GIC_GITHUB_BASE sostituisce https://github.com/webmappsrl con repo locali.

set -uo pipefail

[ $# -eq 2 ] || { echo "Uso: $0 <shard di arrivo> <cartella>" >&2; exit 2; }
SHARD=$1 OUT=$2
BASE=${GIC_GITHUB_BASE:-https://github.com/webmappsrl}

if [[ $SHARD == *dev ]]; then BRANCH=develop; else BRANCH=main; fi
REPO_SHARD=${SHARD%dev}

mkdir -p "$OUT/codice" || exit 2
OUT=$(cd "$OUT" && pwd)
cd "$OUT/codice" || exit 2
TMP=$(mktemp -d "${TMPDIR:-/tmp}/codice-riferimento.XXXXXX")
trap 'rm -rf "$TMP"' EXIT

fallito() { echo "❌ $1: la verifica non ripiega sulle copie locali." >&2; exit 4; }

senza_submodule() {
  echo "❌ $1: chiedi al dev quale repo pubblica lo shard $SHARD." >&2
  exit 5
}

# Commit fissato in <repo> <branch> per il submodule <percorso>. Per il repo dello shard un clone
# che fallisce dopo quello di wm-webapp riuscito vuol dire repo o branch che non esistono.
commit_di() {
  local repo=$1 branch=$2 percorso=$3
  if ! git clone --quiet --depth 1 --branch "$branch" "$BASE/$repo.git" "$TMP/$repo" 2>/dev/null; then
    [ "$repo" = wm-webapp ] && fallito "non riesco a clonare $repo ($branch) da $BASE"
    senza_submodule "$repo non esiste su $BASE o non ha il branch $branch"
  fi
  local riga
  riga=$(git -C "$TMP/$repo" ls-tree HEAD "$percorso")
  [[ $riga == 160000\ commit\ * ]] || senza_submodule "$repo ($branch) non ha il submodule $percorso"
  awk '{print $3}' <<<"$riga"
}

# <repo> al <commit>, nella <cartella>.
scarica_commit() {
  local repo=$1 commit=$2 cartella=$3
  git init --quiet "$cartella" && git -C "$cartella" remote add origin "$BASE/$repo.git" &&
    git -C "$cartella" fetch --quiet --depth 1 origin "$commit" 2>/dev/null &&
    git -C "$cartella" checkout --quiet FETCH_HEAD 2>/dev/null \
    || fallito "non riesco a scaricare $repo al commit $commit"
}

C_CORE=$(commit_di wm-webapp main src/app/shared/wm-core) || exit $?
C_PKG=$(commit_di "$REPO_SHARD" "$BRANCH" wm-package) || exit $?
rm -rf wm-core wm-package geohub
scarica_commit wm-core "$C_CORE" wm-core
scarica_commit wm-package "$C_PKG" wm-package
git clone --quiet --depth 1 --branch main "$BASE/geohub.git" geohub 2>/dev/null \
  || fallito "non riesco a clonare geohub (main) da $BASE"

riga() { printf '%s %s (%s) %s\n' "$1" "$(git -C "$1" rev-parse --short HEAD)" "$2" "$(git -C "$1" log -1 --format=%cs)"; }
{
  riga wm-core "fissato da wm-webapp main"
  riga wm-package "fissato da $REPO_SHARD $BRANCH"
  riga geohub main
} > "$OUT/codice.txt"
cat "$OUT/codice.txt"
