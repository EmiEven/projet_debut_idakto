package main
//librérie dont on a besoin
import (
    "context" // sert à transporter des infos (timeouts, annulation, etc.) dans les appels réseau
    "golang.org/x/oauth2" // gère tout ce qui est OAuth2 (échange de code, génération d’URL, etc.)
    "golang.org/x/oauth2/google" // fournit les endpoints spécifiques de Google (URL d’auth, token, etc.)
    "github.com/coreos/go-oidc" //gère OpenID Connect (provider, vérification d’ID token, claims…).
    "os"
    "github.com/joho/godotenv"
)

var (
    err = godotenv.Load()

    clientID     =  os.Getenv("CLIENT_ID")
    clientSecret =  os.Getenv("CLIENT_SECRET")

    //mettre un client id de idakto

    provider *oidc.Provider // représente le “serveur d’identité” (Google) : ses URLs, sa config, etc
    oauth2Config oauth2.Config // struct qui contient toute la config OAuth2 (clientID, secret, redirect URL, scopes…)
	//=> utiliser pour générer l’URL de login et échanger le code
    IdClusterOauthConfig oauth2.Config
)

// fonction qui initialise tout ce qui concerne OIDC
// elle retourne un error -> soit nil si tout va bien, soit une erreur
func initOIDC() error {
    var err error

    // 1. Récupérer la configuration Google OIDC
	// contacte le serveur OIDC de Google (https://accounts.google.com)
	// récupère sa configuration (endpoints, clés publiques, etc.)
    provider, err = oidc.NewProvider(context.Background(), "https://accounts.google.com")
    if err != nil {
        return err
    }

    // 2. Configurer OAuth2
    oauth2Config = oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURL:  "http://localhost:8080/callback", // (EXACTEMENT la même que dans Google Cloud)
        Endpoint:     google.Endpoint, // contient les URLs d’auth et de token de Google
        Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		/* Scopes :
				- oidc.ScopeOpenID => obligatoire pour OpenID Connect
				- "email" => pour avoir l'email (logique)
				- "profile" => pour avoir d'autres onfis (nom, etc.)
		*/
    }

    // Configurer OAuth2
    IdClusterOauthConfig = oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        //RedirectURL:  "https://auth.pawx.asgard.idenv.fr/demoapp/cbregister", 
        RedirectURL:  "http://localhost:8080/callback",
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://auth.pawx.asgard.idenv.fr/auth/oauth2/auth",// quand je pars du client au serveur d'authentification pour la 1e fois => me redirige vers là où je vais m'authentifier
            TokenURL: "https://auth.pawx.asgard.idenv.fr/token",
        },
        Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
    }

    return nil // si tout s'est bien passé
}


/*func initOIDC() error {
    var err error

    provider, err = oidc.NewProvider(context.Background(), "https://auth.pawx.asgard.idenv.fr/auth")
    if err != nil {
        return err
    }

    // Configurer OAuth2
    IdClusterOauthConfig = oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURL:  "https://auth.pawx.asgard.idenv.fr/demoapp/cbregister", // (EXACTEMENT la même que dans Google Cloud)
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://auth.pawx.asgard.idenv.fr/auth/oauth2",
            TokenURL: "https://auth.pawx.asgard.idenv.fr/token",
        },
        Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
    }

    return nil // si tout s'est bien passé
}*/


/*

il me faut le client_id

*/