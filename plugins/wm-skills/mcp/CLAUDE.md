# Server MCP per Orchestrator

Espone l'API di Orchestrator come tool tipizzati, così le skill non costruiscono chiamate HTTP
a mano. Il perché delle scelte di fondo sta in
[docs/knowledge/mcp-orchestrator.md](../../../docs/knowledge/mcp-orchestrator.md); qui c'è ciò
che serve per lavorarci dentro.

## Comandi

```bash
go test ./...    # dalla root di mcp/
./build.sh       # ricompila il binario in ../bin/orchestrator-mcp
```

Il binario è **versionato nel repo** e va ricompilato ad ogni release, perché viaggia nel
plugin: chi installa non compila nulla.

`build.sh` produce **solo `darwin/arm64`**. Su Linux o Mac Intel va cambiato `GOOS`/`GOARCH`,
altrimenti il server non parte sulla macchina di chi lo installa.

## Vincoli da conoscere prima di toccare il codice

- **Gli elenchi di valori ammessi (`type`, `status`) si leggono dagli enum PHP di Orchestrator,
  non dalla specifica OpenAPI**: quella generata da Scramble non li espone. Aggiungere un valore
  a mano qui significa farlo divergere dal backend.
- **Gli elenchi finiscono nello schema del tool, non in un controllo interno**: un valore fuori
  elenco diventa inesprimibile per costruzione, rifiutato prima che la chiamata parta. Non
  spostare quella verifica dentro il codice del handler.
- **Ogni tool di scrittura accetta `confirm`**: senza, restituisce l'anteprima della differenza
  e non scrive. È il meccanismo su cui si reggono le conferme delle skill — non aggiungere tool
  di scrittura che scrivano al primo colpo.
- **Comunicazione su stdio, nessuna porta in ascolto**: esclude per costruzione i conflitti con
  i container Docker del team. Non introdurre un trasporto HTTP.
- **Il file delle credenziali è un parametro** (`--auth-file`), non un percorso fisso: serve a
  far convivere produzione e istanza locale con identità diverse.

## Ambiente

Go 1.27.1. L'istanza locale di Orchestrator sta su `http://localhost:8099`, avviata con Docker
dal repo `webmappsrl/orchestrator`: è l'unico ambiente su cui provare le scritture.
