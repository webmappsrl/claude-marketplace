#!/usr/bin/env bash
# Test del comando geohub-import-check (sorgente in mcp/cmd/geohub-import-check, binario in bin/:
# va ricompilato con mcp/build.sh prima di lanciare la suite). Non tocca la rete: environment.ts arriva da file e lo
# scaricamento passa da un server HTTP locale sopra le fixture.
#
# Le fixture in tests/fixtures/geohub-import-check/ sono ritagliate dalla fotografia dei file
# pubblici di Geohub app 28 e Maphub dev app 3 scattata il 30/09/2026, prima delle correzioni di
# oc:8663, oc:8664 e oc:8665. NON vanno rigenerate dopo quelle correzioni: sono la prova che lo
# script trova quei bug.
set -uo pipefail
QUI="$(cd "$(dirname "$0")" && pwd)"
GIC="$QUI/../../bin/geohub-import-check"
FIX="$QUI/fixtures/geohub-import-check"
export GIC_ENV_TS_URL="file://$FIX/environment.ts"
FALLITI=0

verifica() { # nome, atteso, ottenuto
  if [ "$2" = "$3" ]; then
    printf '✔ %s\n' "$1"
  else
    printf '✘ %s\n   atteso:   %s\n   ottenuto: %s\n' "$1" "$2" "$3"
    FALLITI=$((FALLITI+1))
  fi
}

# --- risolvi ---------------------------------------------------------------------------------

OUT=$("$GIC" risolvi https://28.app.geohub.webmapp.it/ | jq -c '[.shard,.app_id,.layout,.awsApi]')
verifica "Geohub: shard, id, layout vecchio, awsApi" \
  '["geohub",28,"vecchio","https://wmfe.s3.eu-central-1.amazonaws.com/geohub"]' "$OUT"

OUT=$("$GIC" risolvi https://3.maphubdev.maphub.it/ | jq -c '[.shard,.app_id,.layout,.awsApi]')
verifica "shard maphubdev: shard, id, layout nuovo, awsApi" \
  '["maphubdev",3,"nuovo","https://dev.maphub.it/wmfe/maphubdev"]' "$OUT"

OUT=$("$GIC" risolvi https://fiemaps.it/ | jq -c '[.shard,.app_id]')
verifica "dominio cliente risolto da redirects" '["geohub",29]' "$OUT"

ERR=$("$GIC" risolvi https://3.inesistente.maphub.it/ 2>&1 >/dev/null); RC=$?
verifica "shard sconosciuto: exit 2" "2" "$RC"
verifica "shard sconosciuto: messaggio col nome" "1" "$(grep -c inesistente <<<"$ERR")"

ERR=$(GIC_ENV_TS_URL="file://$FIX/environment-rotto.ts" "$GIC" risolvi https://3.maphubdev.maphub.it/ 2>&1 >/dev/null); RC=$?
verifica "environment.ts non leggibile: exit 2" "2" "$RC"
verifica "environment.ts non leggibile: messaggio" "1" "$(grep -c 'environment.ts' <<<"$ERR")"

OUT=$(GIC_ENV_TS_URL="http://127.0.0.1:9/environment.ts" "$GIC" risolvi https://3.maphubdev.maphub.it/ 2>/dev/null | jq -r '.fonte')
verifica "GitHub irraggiungibile: copia locale dichiarata" "1" "$(grep -c 'copia locale del 2026-09-30' <<<"$OUT")"

# --- scarica ---------------------------------------------------------------------------------
# Un server HTTP locale serve una copia delle fixture con la stessa forma degli indirizzi veri;
# un environment.ts di prova fa puntare geohub e maphubdev a quel server.

TMP=$(mktemp -d); trap 'kill $SRV 2>/dev/null; rm -rf "$TMP"' EXIT
PORTA=$(python3 -c 'import socket;s=socket.socket();s.bind(("127.0.0.1",0));print(s.getsockname()[1])')
W="$TMP/www"; mkdir -p "$W"/g/{conf,icons,pois,tracks} "$W"/g-el "$W"/g-origin/api/taxonomy/where \
  "$W"/s/3 "$W"/s/tracks "$W"/s-el "$W"/s-origin/api/v2/app
cp "$FIX/geohub/config.json" "$W/g/conf/28.json"; cp "$FIX/geohub/icons.json" "$W/g/icons/28.json"
cp "$FIX/geohub/pois.geojson" "$W/g/pois/28.geojson"; cp "$FIX"/geohub/tracks/*.json "$W/g/tracks/"
cp "$FIX/geohub/elastic.json" "$W/g-el/index.html"; cp "$FIX"/geohub/where/*.json "$W/g-origin/api/taxonomy/where/"
for f in "$W"/g-origin/api/taxonomy/where/*.json; do mv "$f" "${f%.json}"; done
cp "$FIX/shard/config.json" "$FIX/shard/icons.json" "$FIX/shard/pois.geojson" "$W/s/3/"
cp "$FIX"/shard/tracks/*.json "$W/s/tracks/"; cp "$FIX/shard/app-all.json" "$W/s-origin/api/v2/app/all"
# Elastic dello shard elenca anche la traccia 777, che non esiste: deve risultare non scaricata.
jq '.hits += [{"id":777,"name":"Traccia inesistente"}]' "$FIX/shard/elastic.json" > "$W/s-el/index.html"
B="http://127.0.0.1:$PORTA"
sed -e "/^  geohub: {/,/^  },/{s#origin: '[^']*'#origin: '$B/g-origin'#;s#elasticApi: '[^']*'#elasticApi: '$B/g-el'#;s#awsApi: '[^']*'#awsApi: '$B/g'#;}" \
    -e "/^  maphubdev: {/,/^  },/{s#origin: '[^']*'#origin: '$B/s-origin'#;s#elasticApi: '[^']*'#elasticApi: '$B/s-el'#;s#awsApi: '[^']*'#awsApi: '$B/s'#;}" \
    "$FIX/environment.ts" > "$TMP/environment-locale.ts"
(cd "$W" && exec python3 -m http.server "$PORTA" --bind 127.0.0.1 >/dev/null 2>&1) & SRV=$!
for _ in $(seq 50); do curl -s "$B/" >/dev/null && break; sleep 0.1; done
scarica() { GIC_ENV_TS_URL="file://$TMP/environment-locale.ts" "$GIC" scarica --no-immagini \
  --geohub https://28.app.geohub.webmapp.it/ --shard https://3.maphubdev.maphub.it/ --out "$1"; }

scarica "$TMP/out" >/dev/null 2>"$TMP/err"; RC=$?
verifica "scarica: exit 0 anche con una traccia mancante" "0" "$RC"
verifica "scarica: tre tracce per parte" "3 3" \
  "$(ls "$TMP/out/geohub/tracks" | wc -l | tr -d ' ') $(ls "$TMP/out/shard/tracks" | wc -l | tr -d ' ')"
verifica "scarica: pois.geojson dalle due parti" "3 6" \
  "$(jq '.features|length' "$TMP/out/geohub/pois.geojson") $(jq '.features|length' "$TMP/out/shard/pois.geojson")"
verifica "scarica: where Geohub delle tracce" "Toscana" "$(jq -r '.name.it' "$TMP/out/geohub/where/9.json")"
verifica "scarica: traccia 777 non scaricata con 404" "1" \
  "$(jq '[.non_scaricate[]|select((.url|test("tracks/777.json")) and (.motivo|test("404")))]|length' "$TMP/out/manifest.json")"

# I campioni di immagini si scelgono sulle feature abbinate; il tipo senza immagini non c'è.
OUT=$("$GIC" campioni "$FIX" | jq -c '[.[]|[.tipo,(.geohub_id|tostring)]]')
verifica "campioni: un originale per tipo presente" \
  '[["traccia_evidenza","31696"],["poi_evidenza","2565"],["poi_galleria","2565"]]' "$OUT"

jq '(.[]|select(.id==3)|.properties.geohub_id)=99' "$FIX/shard/app-all.json" > "$W/s-origin/api/v2/app/all"
scarica "$TMP/out2" >/dev/null 2>"$TMP/err"; RC=$?
verifica "geohub_id diverso: exit 3" "3" "$RC"
verifica "geohub_id diverso: nessuna traccia scaricata" "no" "$([ -d "$TMP/out2/shard/tracks" ] && echo si || echo no)"
cp "$FIX/shard/app-all.json" "$W/s-origin/api/v2/app/all"

# --- confronta -------------------------------------------------------------------------------

REGOLE="$QUI/../../skills/wm-geohub-import-check/differenze-attese.json"
confronta() { "$GIC" confronta "$1" --regole "$REGOLE" >/dev/null 2>"$TMP/err"; }
C="$TMP/c"; cp -R "$FIX" "$C"; confronta "$C"; RC=$?; D="$C/diff.json"
verifica "confronta: exit 0" "0" "$RC"
gruppo() { jq -c "[.gruppi[]|select($1)]|length" "$D"; }

verifica "oc:8665 name assente su tutti i POI del pois.geojson" "3" \
  "$(jq '[.gruppi[]|select(.risorsa=="pois.geojson" and .campo=="name" and .tipo=="assente")|.casi]|add' "$D")"
verifica "oc:8664 related_url \"[]\" come testo su tutte e 3 le tracce" "1" \
  "$(gruppo '.risorsa=="tracks" and .campo=="related_url" and .tipo=="tipo_diverso" and .casi==3 and all(.esempi[];.shard=="[]")')"
verifica "oc:8664 related_url come testo su tutti e 3 i POI abbinati (\"false\" e oggetto)" "1" \
  "$(gruppo '.risorsa=="pois.geojson" and .campo=="related_url" and .tipo=="tipo_diverso" and .casi==3')"
verifica "oc:8663 immagine in evidenza dei POI, esempio 2466 → 268" "1" \
  "$(gruppo '.risorsa=="pois.geojson" and .campo=="feature_image" and .tipo=="conteggio_immagini" and any(.esempi[];.geohub_id==2466 and .shard_id==268)')"
verifica "oc:8663 galleria dei POI: 6 su Geohub, meno sullo shard (2565 → 350)" "1" \
  "$(gruppo '.risorsa=="pois.geojson" and .campo=="image_gallery" and .tipo=="conteggio_immagini"')"
verifica "metriche a 0 in Elastic con possibile causa" "1" \
  "$(gruppo '.risorsa=="elastic" and .campo=="distance" and (.possibile_causa|test("Reindicizza Scout")) and any(.esempi[];.geohub_id==31696 and .shard==0)')"
verifica "POI solo sullo shard: 106258, 106260 e quello senza geohub_id" '[99999,106258,106260]' \
  "$(jq -c '[.solo_shard[]|select(.risorsa=="pois.geojson")|(.geohub_id // .shard_id)]|sort' "$D")"
verifica "nome in Elastic con una sola lingua: regola su 2 tracce" "2" \
  "$(jq '[.ignorate[]|select(.regola=="elastic_nome_una_lingua")|.casi]|add' "$D")"
verifica "nome in Elastic con due lingue: resta una differenza (31723)" "1" \
  "$(gruppo '.risorsa=="elastic" and .campo=="name" and .casi==1 and .esempi[0].geohub_id==31723')"
verifica "searchable una volta sola fra le chiavi presenti solo su Geohub" "1" \
  "$(jq '[.chiavi_da_una_parte[]|select(.risorsa=="pois.geojson" and .chiave=="searchable" and .parte=="geohub")]|length' "$D")"
verifica "nessun gruppo per chiavi attese o URL delle immagini" "0" \
  "$(gruppo '.campo|test("geohub_synced_at|created_at|updated_at|feature_image\\.url|^id$")')"
verifica "where OSMFeatures in più su POI e tracce: nessuna differenza" "0" \
  "$(gruppo '.campo=="taxonomy_where" and .risorsa!="elastic"')"
verifica "regole: versione 4" "4" "$(jq '.regole_versione' "$D")"

verifica "tassonomie del config abbinate per identifier, non per posizione" "0" \
  "$(gruppo '.risorsa=="config.json" and (.campo|test("identifier$"))')"
verifica "POI collegati uguali: nessuna differenza sulle tracce" "0" "$(gruppo '.campo|test("^related_pois")')"

# Un POI collegato che manca sullo shard è una differenza della traccia.
C5="$TMP/c5"; cp -R "$FIX" "$C5"
jq '.properties.related_pois|=.[1:]' "$FIX/shard/tracks/39.json" > "$C5/shard/tracks/39.json"
confronta "$C5"; D="$C5/diff.json"
verifica "POI collegato mancante sulla traccia 31723 → 39" "1" \
  "$(gruppo '.risorsa=="tracks" and .campo=="related_pois" and any(.esempi[];.geohub_id==31723)')"
D="$C/diff.json"

# Una where di Geohub che manca sullo shard, quando lo shard ha già le where calcolate dalla
# geometria (OSMFeatures), è una differenza attesa da segnalare: non va fra i gruppi.
C2="$TMP/c2"; cp -R "$FIX" "$C2"
jq '(.features[]|select(.properties.geohub_id==2466)|.properties)|=(.taxonomy_where|=with_entries(select(.value.it!="Andora")))|(.features[]|select(.properties.geohub_id==2466)|.properties.taxonomyWheres)|=map(select(.!="Andora"))' \
  "$FIX/shard/pois.geojson" > "$C2/shard/pois.geojson"
confronta "$C2"; D="$C2/diff.json"
verifica "where mancante con where OSMFeatures presenti: non è un gruppo" "0" \
  "$(gruppo '.campo=="taxonomy_where" and .risorsa=="pois.geojson"')"
verifica "where mancante con where OSMFeatures presenti: segnalata con l'esempio" "1" \
  "$(jq '[.segnalate[]|select(.regola=="where_geometria" and .risorsa=="pois.geojson" and any(.esempi[];.geohub=="Andora" and .geohub_id==2466))]|length' "$D")"
verifica "report: sezione delle differenze attese da segnalare con Andora" "1" \
  "$(sed -n '/^## Differenze attese da segnalare/,/^## /p' "$C2/report.md" | grep -c 'Andora')"

# Senza nessuna where OSMFeatures sullo shard la where mancante resta una differenza, con la
# sincronizzazione come possibile causa.
C9="$TMP/c9"; cp -R "$FIX" "$C9"
jq '(.features[]|select(.properties.geohub_id==2466)|.properties)|=(.taxonomy_where={}|.taxonomyWheres=[])' \
  "$FIX/shard/pois.geojson" > "$C9/shard/pois.geojson"
confronta "$C9"; D="$C9/diff.json"
verifica "where mancanti senza OSMFeatures: gruppo con possibile causa" "1" \
  "$(gruppo '.campo=="taxonomy_where" and .risorsa=="pois.geojson" and (.possibile_causa|test("Sincronizza")) and any(.esempi[];.geohub_id==2466)')"

# Un valore JSON scritto come testo, "false" dove Geohub non ha nulla: stesso difetto di oc:8664
# (trovato dal vivo su 90 POI di app 28 → 3).
C6="$TMP/c6"; cp -R "$FIX" "$C6"
jq '(.features[]|select(.properties.geohub_id==2466)|.properties.related_url)="false"' \
  "$FIX/shard/pois.geojson" > "$C6/shard/pois.geojson"
confronta "$C6"; D="$C6/diff.json"
verifica "related_url \"false\" come testo: tipo diverso" "1" \
  "$(gruppo '.risorsa=="pois.geojson" and .campo=="related_url" and .tipo=="tipo_diverso" and any(.esempi[];.shard=="false")')"

# Nomi delle where uguali a meno di trattini e maiuscole: «Massa Carrara» su Geohub e «Massa-Carrara»
# in OSMFeatures (trovato dal vivo sulla traccia 31688 → 66).
C7="$TMP/c7"; cp -R "$FIX" "$C7"
echo '{"id":66,"name":{"it":"Massa Carrara"}}' > "$C7/geohub/where/66.json"
jq '(.properties.taxonomy_where[]|select(.it=="Lucca"))|=(.it="Massa-Carrara")' "$FIX/shard/tracks/39.json" > "$C7/shard/tracks/39.json"
jq '(.hits[]|select(.id==39)|.taxonomyWheres)|=map(if .=="Lucca" then "Massa-Carrara" else . end)' "$FIX/shard/elastic.json" > "$C7/shard/elastic.json"
confronta "$C7"; D="$C7/diff.json"
verifica "where: trattino e spazio sono lo stesso nome (traccia 31723 → 39)" "0" "$(gruppo '.campo=="taxonomy_where" and .risorsa=="tracks"')"

# In Elastic la possibile causa «sincronizzazione» non si dà se il file della traccia ha le where
# OSMFeatures: il documento Elastic non le distingue.
C8="$TMP/c8"; cp -R "$FIX" "$C8"
jq '(.hits[]|select(.id==39)|.taxonomyWheres)|=map(select(.!="Toscana"))' "$FIX/shard/elastic.json" > "$C8/shard/elastic.json"
confronta "$C8"; D="$C8/diff.json"
verifica "Elastic: where mancanti con where OSMFeatures nel file della traccia segnalate (7 + Toscana)" "0 1" \
  "$(gruppo '.risorsa=="elastic" and .campo=="taxonomy_where"') $(jq '[.segnalate[]|select(.risorsa=="elastic" and .regola=="where_geometria" and .casi==8)]|length' "$D")"

# Rumore (regole versione 3): formato diverso a contenuto uguale e chiavi che il frontend non legge.
D="$C/diff.json"
verifica "rumore: numero scritto come testo (ele \"21\" e 21)" "0" "$(gruppo '.campo=="ele"')"
verifica "rumore: chiavi non lette dal frontend" "0" \
  "$(gruppo '.campo|test("^(gpx_url|kml_url|geojson_url|author_email|user_id|user_can_download|dem_data|manual_data)$")')"
verifica "rumore: campi interni di Elastic" "0" "$(gruppo '.risorsa=="elastic" and (.campo|test("^(doc|name_keyword|searchable)$"))')"
verifica "rumore: slope in più sullo shard" "0" "$(gruppo '.campo=="slope"')"
verifica "rumore: chiavi vuote su tutte le feature non elencate" "0" \
  "$(jq '[.chiavi_da_una_parte[]|select(.feature==0)]|length' "$D")"
verifica "rumore: config con numeri \"16\" e \"16.00\"" "0" "$(gruppo '.campo|test("poiIcon")')"
verifica "non rumore: searchable dei POI resta (serve alla ricerca locale)" "1" \
  "$(gruppo '.risorsa=="pois.geojson" and .campo=="searchable"')"
C10="$TMP/c10"; cp -R "$FIX" "$C10"
echo '{"txn-a":"<svg a='"'"'1'"'"'  b=\"2\"><g/></svg>"}' > "$C10/geohub/icons.json"
printf '{"txn-a":"    <svg a=\\"1\\" b=\\"2\\">\\n        <g></g>\\n    </svg>\\n"}' > "$C10/shard/icons.json"
confronta "$C10"; D="$C10/diff.json"
verifica "rumore: SVG uguale a meno di spazi, a capo, apici e tag vuoti chiusi" "0" "$(gruppo '.risorsa=="icons.json"')"
echo '{"txn-a":"<svg a=\"1\"><g/></svg>"}' > "$C10/shard/icons.json"
confronta "$C10"; D="$C10/diff.json"
verifica "SVG davvero diverso resta una differenza" "1" "$(gruppo '.risorsa=="icons.json"')"
D="$C/diff.json"

# Un POI senza tipo su Geohub riceve sullo shard il tipo generico «Punto di interesse»
# (wm-package EcPoi.php:276-280): atteso, da segnalare, non una differenza.
C12="$TMP/c12"; cp -R "$FIX" "$C12"
jq '(.features[]|select(.properties.id==2511)|.properties)|=(.taxonomy.poi_type=null|.taxonomy.poi_types=[]|.taxonomyIdentifiers|=map(select(startswith("poi_type_")|not)))' \
  "$FIX/geohub/pois.geojson" > "$C12/geohub/pois.geojson"
jq '(.features[]|select(.properties.geohub_id==2511)|.properties)|=(.taxonomy.poi_type={"identifier":"poi","name":{"it":"Punto di interesse"}}|.taxonomy.poi_types=[{"identifier":"poi","name":{"it":"Punto di interesse"}}]|.taxonomyIdentifiers=["poi_type_poi"])' \
  "$FIX/shard/pois.geojson" > "$C12/shard/pois.geojson"
confronta "$C12"; D="$C12/diff.json"
verifica "tipo generico sullo shard per un POI senza tipo: segnalato" "1" \
  "$(jq '[.segnalate[]|select(.regola=="poi_type_generico" and any(.esempi[];.geohub_id==2511))]|length' "$D")"
verifica "tipo generico: nessuna differenza sui tipi del POI 2511" "0" \
  "$(jq '[.gruppi[]|.esempi[]|select(.geohub_id==2511 and ((.shard|tostring|test("poi_type_poi|Punto di interesse"))))]|length' "$D")"

# Per un campo assente il report dice su quante feature Geohub ha il valore: «3 casi su 3» dice che
# l'import lo perde sempre, anche se le feature sono poche.
D="$C/diff.json"
verifica "assente: numero di feature che su Geohub hanno il valore (name dei POI, 3 su 3)" "3 3" \
  "$(jq -r '.gruppi[]|select(.risorsa=="pois.geojson" and .campo=="name" and .tipo=="assente")|"\(.casi) \(.su_geohub)"' "$D")"
verifica "report: «casi su N» nella riga del gruppo" "1" \
  "$(grep -c 'pois.geojson · name · assente: 3 casi su 3' "$C/report.md")"

# name null sullo shard vale come assente.
C3="$TMP/c3"; cp -R "$FIX" "$C3"
jq '(.properties.name)=null' "$FIX/shard/tracks/64.json" > "$C3/shard/tracks/64.json"
confronta "$C3"; D="$C3/diff.json"
verifica "name null sullo shard: assente" "1" \
  "$(gruppo '.risorsa=="tracks" and .campo=="name" and .tipo=="assente" and any(.esempi[];.shard_id==64)')"

# Elastic dello shard vuoto: possibile causa, nessuna eccezione.
C4="$TMP/c4"; cp -R "$FIX" "$C4"; echo '{"hits":[]}' > "$C4/shard/elastic.json"
confronta "$C4"; RC=$?; D="$C4/diff.json"
verifica "Elastic vuoto: exit 0" "0" "$RC"
verifica "Elastic vuoto: Reindicizza Scout" "1" "$(gruppo '.codice=="elastic_vuoto" and (.possibile_causa|test("Reindicizza Scout"))')"

# Una traccia non scaricata non è né diversa né uguale: sta solo fra le non scaricate.
confronta "$TMP/out"; D="$TMP/out/diff.json"
verifica "traccia non scaricata solo fra le non scaricate" "1 0" \
  "$(jq '[.non_scaricate[]|select(.url|test("777"))]|length' "$D") $(jq '[.. | objects | select(.shard_id? == 777)]|length' "$D")"

# --- report --------------------------------------------------------------------------------

R="$TMP/out/report.md"
verifica "report: titolo con le due app" "# Verifica import Geohub 28 → maphubdev 3" "$(head -1 "$R")"
verifica "report: sezioni nell'ordine" \
  "## Differenze|## Differenze attese da segnalare|## Presenti da una parte sola|## Chiavi presenti da una parte sola|## Non scaricate|## Differenze ignorate per regola|## Verifiche da fare a mano" \
  "$(grep '^## ' "$R" | paste -sd'|' -)"
verifica "report: versione delle regole" "1" "$(grep -c 'versione 4' "$R")"
verifica "report: verifiche manuali" "4" \
  "$(sed -n '/^## Verifiche da fare a mano/,$p' "$R" | grep -cE 'geohub-import|Editor|UGC|CORS')"
verifica "report: rimando alla scaletta" "1" "$(sed -n '/^## Verifiche da fare a mano/,$p' "$R" | grep -c 'scaletta')"
verifica "report: traccia 777 fra le non scaricate" "1" "$(sed -n '/^## Non scaricate/,/^## /p' "$R" | grep -c 'tracks/777.json')"
"$GIC" confronta "$C" --regole "$REGOLE" >/dev/null 2>&1
verifica "report: name dei POI anche fra le differenze" "1" \
  "$(sed -n '/^## Differenze$/,/^## /p' "$C/report.md" | grep -c 'pois.geojson · name · assente')"
verifica "report: nessun None al posto di nomi e id" "0" "$(grep -c 'None' "$C/report.md")"
C11="$TMP/c11"; cp -R "$FIX" "$C11"
jq '.MAP.controls={"x":"<svg width=\"40\">'"$(printf 'a%.0s' $(seq 400))"'</svg>"}' "$FIX/geohub/config.json" > "$C11/geohub/config.json"
confronta "$C11"
verifica "report: SVG lunghi abbreviati" "0" "$(awk 'length>600' "$C11/report.md" | grep -c '<svg')"
verifica "report senza manifest: Non scaricate = Nessuna" "Nessuna." \
  "$(sed -n '/^## Non scaricate/,/^## /p' "$C/report.md" | sed -n 3p)"
OUT=$("$GIC" confronta "$C" --regole "$REGOLE" | grep -c 'pois.geojson · name · assente')
verifica "riepilogo su stdout con il gruppo dei nomi" "1" "$OUT"

[ "$FALLITI" -eq 0 ] && echo "Tutti i test passati." || { echo "$FALLITI test falliti."; exit 1; }
