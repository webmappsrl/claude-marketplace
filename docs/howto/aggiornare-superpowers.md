# Aggiornare superpowers

`superpowers` non è mantenuto qui: è referenziato dall'upstream
[obra/superpowers](https://github.com/obra/superpowers) in `.claude-plugin/marketplace.json`,
pinnato a `ref: main`.

- Per seguire l'upstream non c'è nulla da fare: `main` si aggiorna da solo ad ogni
  `/plugin marketplace update`.
- Per fissare una versione stabile, sostituisci il `ref` del plugin `superpowers` in
  `.claude-plugin/marketplace.json` con un tag, ad esempio `"ref": "v5.1.0"`.
- Dopo la modifica esegui `claude plugin validate .`.

Un miglioramento al comportamento di una skill `superpowers` va proposto **upstream**, non
duplicato in `wm-skills`.
