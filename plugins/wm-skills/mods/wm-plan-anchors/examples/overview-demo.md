> Ticket: oc:8700

## Requisiti

- <a id="r1"></a>**R1** — I POI importati da Geohub mantengono l'id originale: lo shard non
  rinumera, perché i link condivisi dal cliente puntano a quell'id.
- <a id="r2"></a>**R2** — La config dello shard non sovrascrive le chiavi già presenti nel
  package: l'import aggiunge, non rimpiazza.
  - Verifica: `tests/Feature/ImportConfigTest.php`
- <a id="r3"></a>**R3** — [UX] La mappa si apre centrata sul bounding box dell'app, non sulla
  posizione dell'utente.
- <a id="r4"></a>**R4** — L'import non scrive su Elastic: l'indice resta quello che genera lo
  shard al primo allineamento.
