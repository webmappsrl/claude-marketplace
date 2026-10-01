# Verificare l'import di un'app da Geohub a uno shard

Procedura per chi fa il collaudo di un'app migrata da Geohub a uno shard `wm-package` (Maphub o
Maphub dev), con la skill `wm-geohub-import-check`. Le fasi della skill, una per una, sono nel
[diagramma di flusso](https://webmappsrl.github.io/claude-marketplace/wm-geohub-import-check-diagramma/).

## Prima di cominciare

- Il plugin `wm-skills` installato, e le credenziali di Orchestrator in
  `~/.config/webmapp/orchestrator-auth.json`: servono solo all'ultimo passo, quando il report
  diventa un tag.
- L'accesso a Nova dello shard verificato, con un utente che può lanciare le action dell'app.
- I passi della scaletta di migrazione (oc:8652) fatti fino alla rigenerazione dei file: la skill
  li legge dalla scaletta aggiornata e lo chiede all'inizio, perché senza il report si riempie di
  falsi allarmi. Serve poter aprire l'Artifact della scaletta, condiviso con l'organizzazione.

## Come si lancia

```
/wm-skills:wm-geohub-import-check
```

La skill chiede i due indirizzi del frontend: l'app su Geohub e la stessa app sullo shard, per
esempio `https://28.app.geohub.webmapp.it/` e `https://3.maphubdev.maphub.it/`. Il confronto vale
per l'ambiente di quegli indirizzi: con `3.maphub.it` si verifica la produzione.

## Cosa succede

1. **Confronto.** La skill scarica in sola lettura i file pubblici delle due app e li confronta
   campo per campo. Per un'app di circa 60 tracce e 200 POI serve circa un minuto.
2. **Riepilogo.** Una riga: il confronto è fatto e il report è nel percorso indicato. L'elenco
   completo delle differenze è lì; il dettaglio di un gruppo si chiede alla skill.
3. **Solo ciò che si vede.** Una differenza su un dato che il frontend non legge, o che serve a una
   funzione spenta nell'app su Geohub (per esempio un filtro non attivo nel config) non è un problema: la skill la lascia nel
   report fra quelle «non visibili in questa app» e non la porta nel tag.
4. **Interventi da Nova.** Per ogni differenza la skill cerca, nel codice dell'ambiente verificato
   scaricato da GitHub (il frontend com'è pubblicato, il backend com'è fissato sul branch dello shard: `develop` per Maphub dev, `main` per Maphub), un'action di Nova che la
   risolva, e te la propone con il nome e la pagina da cui lanciarla. **Le action le lanci tu**: la
   skill non scrive mai sullo shard. Per ognuna scegli se farla subito — la skill rifà il confronto
   e, se il problema sparisce, non lo mette nel tag — o lasciarla da fare: in quel caso finisce nel
   tag, fra gli interventi manuali da fare.
5. **Tag e ticket.** Quello che resta è un bug. Se rispondi sì, il report passa a `wm-tag`, che
   crea il tag `[COLLAUDO][<APP>][<ANNO>]<N>` sull'Orchestrator di produzione con una macro area per
   problema e la sezione degli interventi fatti, poi propone i ticket uno per volta. Ogni scrittura
   passa da anteprima e conferma.

## Cosa la skill non verifica

Lo dice anche in fondo al report: coda Horizon e log dell'import, ruolo del proprietario, UGC
(le loro API richiedono autenticazione), la webapp aperta nel browser e la grafica di home e
mappa. Vanno controllati a mano, come da scaletta.

## Dove finiscono i file

In una cartella temporanea del sistema (`/var/folders/…/T/import-check-geohub…` su macOS), fuori
da ogni repo: i file scaricati, il report, le differenze in JSON e la copia del codice da GitHub.
La skill indica il percorso nel riepilogo. Alla fine cancella la copia del codice, che può pesare
centinaia di MB; il resto lo elimina il sistema quando pulisce le cartelle temporanee.

## Se una differenza è attesa

Le differenze che ogni import produce per costruzione — id nuovi, campi che il frontend non
legge, numeri scritti in un altro formato — sono in
`plugins/wm-skills/skills/wm-geohub-import-check/differenze-attese.json`. Se una verifica ne
mostra una nuova che il team giudica innocua, va aggiunta lì, alzando la versione delle regole.
