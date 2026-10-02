package main

import (
	"fmt"
	"strings"
)

var verificheManuali = []string{
	"Coda Horizon `geohub-import` vuota e log dell'import senza errori (un job fallito può " +
		"risultare «completato» su Horizon).",
	"Proprietario dell'app presente sullo shard con ruolo Editor.",
	"UGC: non confrontati, perché le API degli UGC richiedono autenticazione. Se l'app ne ha, " +
		"vanno controllati a mano, a campione (immagini e autori).",
	"Aprire la webapp dello shard e controllare che carichi e funzioni: una GET da script può " +
		"riuscire mentre il browser fallisce, per esempio per le OPTIONS non gestite (CORS) o per un " +
		"modulo del bundle che non si carica (grafico altimetrico).",
	"Home e mappa a vista, confrontate con l'app su Geohub.",
}

// coppia: «2365 → 343: », o niente per le risorse senza feature (config, icone).
func coppia(e Esempio) string {
	if e.GeohubID == nil && e.ShardID == nil {
		return ""
	}
	g, s := "?", "?"
	if e.GeohubID != nil {
		g = strPy(e.GeohubID)
	}
	if e.ShardID != nil {
		s = strPy(e.ShardID)
	}
	return g + " → " + s + ": "
}

func nome(n any) string {
	if n == nil || n == "" {
		return "(senza nome)"
	}
	return strPy(n)
}

func rigaGruppo(g *Gruppo) string {
	var es []string
	for _, e := range g.Esempi {
		es = append(es, fmt.Sprintf("%sGeohub %s / shard %s", coppia(e), dumpPy(e.Geohub), dumpPy(e.Shard)))
	}
	causa := ""
	if g.PossibileCausa != nil && *g.PossibileCausa != "" {
		causa = " — possibile causa: " + *g.PossibileCausa
	}
	su := ""
	if g.SuGeohub != 0 {
		su = fmt.Sprintf(" su %d", g.SuGeohub)
	}
	return fmt.Sprintf("- %s · %s · %s: %d casi%s — es. %s%s", g.Risorsa, g.Campo, g.Tipo, g.Casi, su,
		strings.Join(es, "; "), causa)
}

// riepilogo: le righe che il comando confronta stampa, una per gruppo.
func riepilogo(d *Diff) string {
	var righe []string
	for _, g := range d.Gruppi {
		righe = append(righe, rigaGruppo(g))
	}
	for _, x := range d.Segnalate {
		righe = append(righe, fmt.Sprintf("- da segnalare, attesa per regola %s · %s · %s: %d casi",
			x.Regola, x.Risorsa, x.Campo, x.Casi))
	}
	righe = append(righe, fmt.Sprintf("Presenti solo su Geohub: %d; solo sullo shard: %d; non scaricate: %d.",
		len(d.SoloGeohub), len(d.SoloShard), len(d.NonScaricate)))
	return strings.Join(righe, "\n")
}

func oNessuna(righe []string) []string {
	if len(righe) == 0 {
		return []string{"Nessuna."}
	}
	return righe
}

func report(d *Diff, m *Manifest) string {
	appID := func(r *Risoluzione) string {
		if r == nil {
			return "?"
		}
		return fmt.Sprint(r.AppID)
	}
	shard, fonte, data := "?", "(non noto)", "(data non nota)"
	if m.Shard != nil {
		shard = m.Shard.Shard
	}
	if m.Geohub != nil {
		fonte = m.Geohub.Fonte
	}
	if m.ScaricatoIl != "" {
		data = m.ScaricatoIl
	}
	r := []string{
		fmt.Sprintf("# Verifica import Geohub %s → %s %s", appID(m.Geohub), shard, appID(m.Shard)),
		"", fmt.Sprintf("Scaricato il %s. Indirizzi degli shard da: %s.", data, fonte), "",
		"## Differenze", "",
	}
	var righe []string
	for _, g := range d.Gruppi {
		righe = append(righe, rigaGruppo(g))
	}
	r = append(r, oNessuna(righe)...)

	r = append(r, "", "## Differenze attese da segnalare", "",
		"Attese per regola, quindi non passano a wm-tag, ma vanno guardate: una di queste può "+
			"comunque essere un errore.", "")
	righe = nil
	for _, x := range d.Segnalate {
		var es []string
		for _, e := range x.Esempi {
			es = append(es, coppia(e)+"Geohub "+dumpPy(e.Geohub))
		}
		righe = append(righe, fmt.Sprintf("- %s · %s · %s: %d casi — es. %s", x.Regola, x.Risorsa, x.Campo,
			x.Casi, strings.Join(es, "; ")))
	}
	r = append(r, oNessuna(righe)...)

	r = append(r, "", "## Presenti da una parte sola", "")
	righe = nil
	for _, x := range d.SoloGeohub {
		righe = append(righe, fmt.Sprintf("- solo su Geohub · %s · %s · %s", x.Risorsa, strPy(x.GeohubID), nome(x.Nome)))
	}
	for _, x := range d.SoloShard {
		gid := "(nessuno)"
		if x.GeohubID != nil {
			gid = strPy(x.GeohubID)
		}
		righe = append(righe, fmt.Sprintf("- solo sullo shard · %s · geohub_id %s · shard %s · %s",
			x.Risorsa, gid, strPy(x.ShardID), nome(x.Nome)))
	}
	r = append(r, oNessuna(righe)...)

	r = append(r, "", "## Chiavi presenti da una parte sola", "",
		"Da stabilire, chiave per chiave, se il frontend la usa. Una chiave che manca sullo shard "+
			"compare anche fra le differenze passate a wm-tag.", "")
	righe = nil
	for _, c := range d.ChiaviDaUnaParte {
		righe = append(righe, fmt.Sprintf("- %s · `%s` · solo su %s · valorizzata in %d feature",
			c.Risorsa, c.Chiave, c.Parte, c.Feature))
	}
	r = append(r, oNessuna(righe)...)

	r = append(r, "", "## Non scaricate", "")
	righe = nil
	for _, x := range d.NonScaricate {
		righe = append(righe, fmt.Sprintf("- %s — %s", x.URL, x.Motivo))
	}
	r = append(r, oNessuna(righe)...)

	r = append(r, "", "## Differenze ignorate per regola", "",
		fmt.Sprintf("Regole di `differenze-attese.json`, versione %s.", strPy(d.RegoleVersione)), "")
	righe = nil
	for _, x := range d.Ignorate {
		righe = append(righe, fmt.Sprintf("- %s: %d casi", x.Regola, x.Casi))
	}
	r = append(r, oNessuna(righe)...)

	r = append(r, "", "## Verifiche da fare a mano", "",
		"Questo report copre solo ciò che il frontend legge via HTTP: non dice che l'import è "+
			"verificato. Restano da fare a mano, come da scaletta di migrazione (fase «Verifica "+
			"post-import»):", "")
	for _, v := range verificheManuali {
		r = append(r, "- [ ] "+v)
	}
	return strings.Join(r, "\n") + "\n"
}
