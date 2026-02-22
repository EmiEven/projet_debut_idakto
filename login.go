package main
//pages (login, secret…)

import (
	"fmt" // pour écrire du texte (ici dans la réponse HTTP).
	"net/http" // la bibliothèque standard pour faire un serveur web.
)

// page login (fake login)
// w http.ResponseWriter => w est un objet qui permet d’écrire la réponse envoyée au navigateur.
//r *http.Request => r contient la requête HTTP (URL, méthode, headers, cookies…).
func gestionnaire_login(w http.ResponseWriter, r *http.Request){
	/* cookie := : on crée une variable locale cookie.
	http.Cookie{...} : c’est un littéral de struct 
		=> on construit une valeur de type http.Cookie.*/
	cookie := http.Cookie{
		Name: "session",
		Value: "ok",
		Path: "/", // le cookie est envoyé pour tout le site (/, /secret, etc.)
	}
	/* http.SetCookie => fonction qui ajoute un header Set-Cookie à la réponse.
	w => la réponse HTTP.
	&cookie => on passe l’adresse de la variable cookie (un pointeur).
		=> le navigateur reçoit un cookie session=ok */
	http.SetCookie(w, &cookie)

	fmt.Fprintln(w, "Vous êtes connecté ! Allez sur /secret") // écrit du texte dans la réponse HTTP
}


// page secrète
func gestionnaire_secret(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Voici la page secrète !")
}

