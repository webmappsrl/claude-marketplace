# wm-plan: Challenge refactor

## Cosa cambia
La fase Challenge viene spostata dopo la scrittura di `overview.md` ed eseguita tramite un subagente isolato con prompt avversariale. La documentazione viene resa obbligatoria anche in assenza di ticket Orchestrator.

## Perché
- La Challenge eseguita prima dell'overview soffre di confirmation bias: Claude ha già investito contesto nel costruire la soluzione e tende a razionalizzare i rischi invece di esporli.
- Un subagente che legge direttamente il file dal filesystem garantisce isolamento reale del contesto, senza che la conversazione precedente possa contaminare l'analisi.
- Il prompt avversariale esplicito ("trova problemi, non difendere") è necessario perché senza un obiettivo dichiarato il modello produce analisi equilibrate invece che critiche.
- L'assenza di ticket non giustifica l'assenza di documentazione: overview, plan e notes devono esistere sempre.

## Requisiti
- [ ] Fase 3 (Challenge) spostata dopo Fase 4 (overview.md)
- [ ] Challenge eseguita tramite subagente con prompt avversariale
- [ ] Il subagente riceve solo il percorso del file — legge `overview.md` autonomamente dal filesystem
- [ ] Istruzione esplicita di non passare nessun contesto aggiuntivo al subagente
- [ ] Se la Challenge trova buchi, l'overview viene aggiornato prima di procedere al plan
- [ ] Documentazione obbligatoria anche senza ticket (checklist e Fase 4 aggiornate)

## Rischi
- **Subagente che ignora l'istruzione di isolamento** — Claude principale potrebbe comunque includere contesto nel prompt nonostante l'istruzione. Mitigato rendendo la regola esplicita e dichiarando il divieto in modo inequivocabile nella skill.

## Out of scope
- Revisione delle altre fasi del workflow
- Modifica del formato o della struttura di overview.md, plan.md, notes.md

## Moduli toccati
- `plugins/wm-skills/skills/wm-plan/SKILL.md`
