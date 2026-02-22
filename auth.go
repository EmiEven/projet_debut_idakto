package main
//middleware + gestion du cookie

import (
	"net/http"
)

// Middelware : vérifie le cookie
/*
prend en paramètre une fonction next
	=> next a la signature func(http.ResponseWriter, *http.Request) => c’est un handler
retourne une fonction qui a la même signature => func(http.ResponseWriter, *http.Request)

=> requireLogin prend un handler, et renvoie un nouveau handler qui ajoute un contrôle avant d’appeler l’original
*/
func requireLogin(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {	
	// cette fonction ⬇ est le “nouveau handler” qui sera réellement utilisé par http.HandleFunc("/secret", ...)
	return func(w http.ResponseWriter, r *http.Request){ // retourne une fonction anonyme (une fonction sans nom)
		/* r.Cookie("session") => cherche un cookie nommé "session" dans la requête
		Si le cookie existe => cookie contient sa valeur, err est nil
		Sinon => err contient une erreur */
		cookie, err := r.Cookie("session")
		// (err != nil) => il n’y a pas de cookie
		//if err != nil || cookie.Value != "ok" {
		if err != nil || cookie.Value == "" {
			/* // => http.Error : envoie une réponse avec un message d’erreur 
			// et un code HTTP 401 Unauthorized
			http.Error(w, "Accès refusé. Connectez-vous sur /login", http.StatusUnauthorized)*/
			// redirection plus propre
			http.Redirect(w, r, "/login", http.StatusFound)
			return // on arrête là, on n’appelle pas le handler suivant
		}

		next(w, r) // on appelle le handler original (ici gestionnaire_secret)
	}
	
}