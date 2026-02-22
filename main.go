package main

// import des packages 
// => fmt : pour afficher du texte ou écrire dans la réponse HTTP
// => net/http : la bibliothèque standard de Go pour créer des serveurs web
import (
    "fmt"
    "net/http"
	"context" // sinon "undefined: context" même si j'ai déjà importer "context" dans oidc.go, chaque fichier Go importe ce qu'il utilise
	"github.com/coreos/go-oidc"
)

// définition du handler
// fonction qui sera appelé à chaque fois qu'un utilisateur visite un URL
// => w http.ResponseWriter -> permet d'envoyer une réponse au client(navigateur)
// => r *http.Request -> contient les infos sur la requête (URL, headers, méthode...)
func home(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintln(w, "Serveur Go OK")
    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    cookie, err := r.Cookie("session")

    if err != nil || cookie.Value == "" {
        // Pas connecté
        fmt.Fprint(w, `
            <h1>Bienvenue !</h1>
            <a href="/login"><button>Se connecter avec Google</button></a>
        `)
        return
    }

    // Connecté
    fmt.Fprintf(w, `
        <h1>Bonjour !</h1>
        <a href="/secret"><button>Accéder à la zone secrète</button></a>
    `)
}



// Quand quelqu’un va sur /login, on le renvoie chez Google pour qu’il se connecte
func loginGoogle(w http.ResponseWriter, r *http.Request) {
	// génère l'URL de redirection vers Google
	// state-random = une chaîne pour le paramètre state (normalement un token anti-CSRF)
	// l'URL générée contient : client_id, redirect_uri, scope, response_type=code, state
    url := oauth2Config.AuthCodeURL("state-random")
	// envoie une réponse HTTP 302 au navigateur
	// le navigateur est redirigé vers l’URL de Google
	// Google affiche sa page login
    http.Redirect(w, r, url, http.StatusFound)
}

// /callback reçoit le code, le transforme en token, vérifie l’ID token, 
// lit l’email, crée une session, et confirme la connexion
func callback(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background() // contexte utilisé pour les appels réseau (échange de code, vérification token)

    // 1. Récupérer le code envoyé par Google
	/* r.URL.Query() : récupère les paramètres de la query string.
	.Get("code") : récupère la valeur du paramètre code.
	C’est le code que Google a mis dans l’URL :
    	http://localhost:8080/callback?code=ABCD1234 */
    code := r.URL.Query().Get("code")

    // 2. Échanger le code contre un token
	// envoie une requête à Google pour échanger le code contre des tokens
	// Google répond avec un token qui contient : access_token, id_token, expires_in, etc.
	// Si ça échoue => on renvoie une erreur HTTP 500
    token, err := oauth2Config.Exchange(ctx, code)
    if err != nil {
        http.Error(w, "Impossible d'échanger le code", http.StatusInternalServerError)
        return
    }

    // 3. Extraire l'ID token
	// token.Extra("id_token") => récupère le champ id_token dans la réponse
	// ok => vaut true si la conversion a réussi, false sinon
    rawIDToken, ok := token.Extra("id_token").(string)
    if !ok {
        http.Error(w, "ID token manquant", http.StatusInternalServerError)
        return
    }

    // 4. Vérifier l'ID token
    verifier := provider.Verifier(&oidc.Config{ClientID: clientID}) // il sait quelles clés publiques utiliser, quel issuer, etc
    // verifier.Verify(ctx, rawIDToken) vérifie => la signature du token, qu'il n'est pas expiré, qu'il vient bien de google
	idToken, err := verifier.Verify(ctx, rawIDToken)
    if err != nil {
        http.Error(w, "ID token invalide", http.StatusInternalServerError)
        return
    }

    // 5. Lire les claims (email, nom…)
    var claims struct {
        Email string `json:"email"` // Le tag `` indique à Go : “dans le JSON, le champ s’appelle email”
    }
    idToken.Claims(&claims) // remplit la struct avec les données du token


	// 6. Vérifier si l'utilisateur existe en DB 
	var user User 
	// result est un *gorm.DB
	result := DB.Where("email = ?", claims.Email).First(&user) 
	
	if result.Error != nil { // Pas trouvé => créer un nouvel utilisateur 
		user = User{Email: claims.Email} 
		//ID n’est pas rempli => GORM va le générer automatiquement
		DB.Create(&user) 
	}

    // 7. Créer une session (cookie)
    cookie := http.Cookie{
        Name:  "session",
        Value: claims.Email, 
        Path:  "/", // valable pour tout le site
    }
    http.SetCookie(w, &cookie) // ajoute un header Set-Cookie à la réponse

	
    // 8. Réponse
    fmt.Fprintf(w, "Connecté avec Google ! Email : %s, ID BD : %d", claims.Email, user.ID) 
}




// programme executable => toujours fonction main
func main() {

	/*
	err := initOIDC() :
    	on appelle initOIDC()
    	si tout va bien -> err = nil
    	si problème (Google inaccessible, mauvaise URL, etc.) -> err contient une erreur
	*/
	if err := initOIDC(); err != nil {
		// on arrête le programme immédiatement
		panic(err) // crash volontaire avec un message
		// on fait ça car si OIDC n’est pas bien initialisé le serveur ne peut pas fonctionner correctement
	}

	// idem pour DB
	if err := initDB(); err != nil {
		panic(err)
	}

	/*http.HandleFunc("/", home)
    -> Associe la route / (la racine du site) à la fonction home.
    -> Donc si quelqu’un visite http://localhost:8080/, Go appelle home().*/
	http.HandleFunc("/", home)

	// routes publiques
	//http.HandleFunc("/login", gestionnaire_login)
	http.HandleFunc("/login", loginGoogle)

	http.HandleFunc("/callback", callback)


	// routes protégées
	// on ne passe pas directement gestionnaire_secret à HandleFunc
	// on passe requireLogin(gestionnaire_secret) :
	//	requireLogin prend gestionnaire_secret en paramètre (next)
	//	elle renvoie un nouveau handler qui :
	//		vérifie le cookie
	//		puis appelle gestionnaire_secret si tout va bien.
	http.HandleFunc("/secret", requireLogin(gestionnaire_secret))

	/* http.ListenAndServe(":8080", nil)
    -> Démarre le serveur web. 
    -> ":8080" -> le serveur écoute sur le port 8080.
    -> nil -> Go utilise le multiplexer (router) par défaut.*/ 
	http.ListenAndServe(":8080", nil)

	

}



/*

# Handler => une fonction qui répond à une requête HTTP
	=> fonction qui dit : “Quand quelqu’un visite cette URL, exécute cette fonction.”
	=> en go un handler a toujours cette forme : func(w http.ResponseWriter, r *http.Request)
	=> w -> permet d’envoyer une réponse au navigateur
	=> r -> contient la requête envoyée par le navigateur (URL, cookies, headers…)

# err => variable qui contient une erreur si quelque chose s’est mal passé
	=> si tout va bien : err vaut nill
	=> problème : err contient une erreur

# nil => rien, vide, abscence de valeur
	=> if err == nil => si tout s'est bien passé
	=> if err != nil => s'il y a une erreur

*/

/* CONFIGURER GOOGLE COMME IDENTITY PROVIDER => étape 4

ÉTAPES : 
Aller dans Google Cloud Console
Créer un projet
Activer l'API OAuth (Identity)
Configurer l'écran de consentement OAuth
Créer un OAuth Client ID
Ajouter la redirect URI (http://localhost:8080/callback)
Récupérer les clés :
	client_id = ***
	client_secret = ***

*/