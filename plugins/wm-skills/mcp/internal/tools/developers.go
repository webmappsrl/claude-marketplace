package tools

// developerIDs sono gli utenti Orchestrator che sono sviluppatori del team, per id.
// Quando il creatore di un ticket è uno di loro, la risposta scritta in
// customer_request la riceve lui e non un cliente: l'anteprima di update_story
// non avvisa dell'email e la skill può applicare senza una conferma a parte.
//
// L'API di Orchestrator non espone gli utenti né i loro ruoli (/api/users non
// esiste), quindi l'elenco è scritto qui. Per aggiungere un dev: id dal
// pannello di Orchestrator, una riga con id e nome.
var developerIDs = map[int]string{
	104: "Rubens Garofalo",
	116: "Giuseppe Bonfanti",
	154: "Alessandro Peci",
	172: "Carla Cupani",
}

// developerCreator restituisce il nome del creatore del ticket se è un dev del
// team, altrimenti stringa vuota. creator_id arriva dal JSON come numero.
func developerCreator(story map[string]any) string {
	id, ok := story["creator_id"].(float64)
	if !ok {
		return ""
	}
	return developerIDs[int(id)]
}
