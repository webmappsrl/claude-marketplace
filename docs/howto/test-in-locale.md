# Provare una skill in locale, senza pushare

Per iterare su un `SKILL.md` senza un push ad ogni modifica, aggiungi il marketplace come path
locale. Dentro Claude Code, dalla root del repo:

```
/plugin marketplace add .
/plugin install wm-skills@wm-marketplace
```

Così il plugin punta al filesystem: ogni modifica a un `SKILL.md` è disponibile ricaricando la
sessione, senza commit né push.

In questa modalità l'header di sessione di `wm-plan` mostra `🔧 modalità sviluppo locale` al
posto del check aggiornamenti: è il comportamento atteso, non un errore.

Per tornare alla versione remota:

```
/plugin marketplace remove wm-marketplace
/plugin marketplace add webmappsrl/claude-marketplace
```
