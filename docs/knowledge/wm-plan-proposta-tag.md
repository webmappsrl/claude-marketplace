# Proposta dei tag in wm-plan

## Come funziona oggi

`wm-plan` propone i tag Orchestrator da associare al ticket in due momenti distinti.

**Tag di ambiente**, subito dopo `environment-setup`: quello che si sa dal posto in cui si
lavora — il repository da `git remote`, l'area dai flag `stack_type` / `stack_ui`. Due ricerche,
nessuna domanda al dev oltre alla conferma.

**Tag di contenuto**, dopo l'approvazione dell'overview: le caratteristiche che hanno segnato
quel lavoro. Il criterio è la ripetibilità — una caratteristica che tornerà in altri lavori è un
buon tag, una che vale per un ticket solo è il suo titolo. È il momento in cui nascono i tag
nuovi.

La ricerca e il filtro stanno in `plugins/wm-skills/scripts/tag-candidati.sh`, di sola lettura.
Un tag è candidato se ha `description` nulla, non è già sul ticket, e non porta un trimestre nel
nome. Lo script cerca anche le varianti del nome (trattini, spazi, parentesi quadre) prima di
dichiarare che un tag non esiste.

La skill mostra i candidati e scrive solo dopo un sì esplicito del dev, un tag per volta.

## Perché così

- **Il filtro sta in uno script, non nel prompt** (oc:8577): scartare per `description`, per id e
  per trimestre è confronto di stringhe, e uno script lo fa uguale tutte le volte. Stesso criterio
  per cui il lint del `CLAUDE.md` è uscito dal prompt del doctor.
- **Le descrizioni non entrano nel context** (oc:8577): `GET /api/tags` le manda sempre e alcune
  contengono credenziali in chiaro. Misurato: `search=RDO` rende 84.714 caratteri per 7 tag, di
  cui 83.391 di sole descrizioni.
- **Due momenti invece di uno** (oc:8577): all'inizio si sa dove si lavora ma non cosa si sta per
  fare. Tenere tutto all'inizio avrebbe perso i tag di contenuto; spostare tutto alla fine
  avrebbe lasciato senza tag i lavori interrotti a metà.
- **Nessun tag senza un sì esplicito** (oc:8577): due tag su Orchestrator non si possono fondere,
  quindi un doppione si ripara solo a mano, ticket per ticket.
- **Il nome di un tag nuovo esprime una caratteristica ripetibile** (oc:8577), non uno schema
  sintattico: conta cosa qualcuno cercherà fra sei mesi.
- **Se Orchestrator non risponde, la fase si ferma e chiede** (oc:8577): un avviso che scorre via
  mentre il dev legge altro lo lascia convinto che i tag ci siano.

## Come ci siamo arrivati

- **Leggere tutti i tag e confrontarne le descrizioni** (oc:8577, superata): la prima forma in cui
  la feature era stata pensata. Abbandonata alla misura: l'elenco costava circa 21.000 token a
  ticket e portava in context credenziali in chiaro.
- **Il trimestre fra le caratteristiche proposte** (oc:8577, superata): cercarlo pescava per
  sottostringa anche i tag cliente che lo contengono — `[26Q3]FORESTAS` proposto su un lavoro
  interno. Emerso provando lo script sui ticket veri oc:8577 e oc:8545, non dai test. Il trimestre
  lo assegna Orchestrator alla creazione: non c'è niente da proporre.
