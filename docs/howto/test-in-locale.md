# Provare una skill in locale, senza pushare

Per iterare su un `SKILL.md` senza un push ad ogni modifica, aggiungi il marketplace come path
locale. Dentro Claude Code, dalla root del repo:

```
/plugin marketplace add ./
/plugin install wm-skills@wm-marketplace
```

Con `.` al posto di `./` il comando fallisce.

L'installazione **copia** il plugin in `~/.claude/plugins/cache/wm-marketplace/wm-skills/<versione>/`,
ed è quella copia che le sessioni usano: una modifica nel repo non arriva da sola. Per provarla,
senza commit né push, reinstalla e riavvia la sessione:

```
/plugin uninstall wm-skills@wm-marketplace
/plugin install wm-skills@wm-marketplace
```

`/plugin update` non basta finché `version` non cambia: la copia risulta già aggiornata. Per
controllare che la copia sia quella nuova, cerca in `~/.claude/plugins/cache/wm-marketplace/wm-skills/*/`
il file appena modificato.

In questa modalità l'header di sessione di `wm-plan` mostra `🔧 modalità sviluppo locale` al
posto del check aggiornamenti: è il comportamento atteso, non un errore.

Per tornare alla versione remota:

```
/plugin marketplace remove wm-marketplace
/plugin marketplace add webmappsrl/claude-marketplace
```
