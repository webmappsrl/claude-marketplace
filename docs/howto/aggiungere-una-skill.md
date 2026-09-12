# Aggiungere una skill a wm-skills

I passi marcati **(dev)** non li esegue Claude: la prima regola del `CLAUDE.md` vieta
all'agente commit e push.

1. Crea la cartella `plugins/wm-skills/skills/<nome-skill>/`. Il nome è in kebab-case con
   prefisso obbligatorio `wm-`.
2. Crea `SKILL.md` con il frontmatter YAML:

   ```markdown
   ---
   name: wm-nome-skill
   description: Una frase che spiega quando Claude deve invocare questa skill.
   ---

   Corpo della skill in Markdown…
   ```

   `name` deve coincidere col nome della cartella. Solo `name` e `description` sono
   obbligatori: evita campi extra.
3. Scrivi `description` come una frase con soggetto — "Use when…" oppure "Usa quando…" — non
   come un titolo: è il testo su cui Claude decide se invocare la skill, ed è la **fonte
   autoritativa** su quando si attiva. Non ripeterlo nel `CLAUDE.md`.
4. Scrivi il corpo in tono imperativo e diretto, rivolto a Claude come agente che esegue
   istruzioni. Sezioni `##` per separare fasi o categorie, elenchi puntati per passi atomici.
   La lunghezza va da poche decine di righe, per una checklist, a qualche centinaio per un
   workflow articolato.
5. Verifica il coupling: se la skill nuova condivide un contratto con una esistente, aggiorna
   entrambe nello stesso lavoro e aggiungi la riga alla tabella in `CLAUDE.md` →
   `## Regole del repo`.
6. `claude plugin validate .` prima di dichiarare il lavoro pronto.
7. **(dev)** Commit e push su `main`.
