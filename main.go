package main

// import des packages
// => fmt : pour afficher du texte ou écrire dans la réponse HTTP
// => net/http : la bibliothèque standard de Go pour créer des serveurs web
import (
    "context" // sinon "undefined: context" même si j'ai déjà importer "context" dans oidc.go, chaque fichier Go importe ce qu'il utilise
    "fmt"
    "github.com/coreos/go-oidc"
    "golang.org/x/crypto/bcrypt" // pour mdp => cryptage
    "golang.org/x/oauth2"        // pour paramètre de session avec google
    "net/http"
	
)


func loginClassic(w http.ResponseWriter, r *http.Request) {

    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }

    email := r.FormValue("email")
    password := r.FormValue("password")

    if email == "" || password == "" {
        http.Error(w, "Champs manquants", http.StatusBadRequest)
        return
    }

    var user User
    result := DB.Where("email = ?", email).First(&user)

    if result.Error != nil {
        http.Error(w, "Utilisateur introuvable", http.StatusUnauthorized)
        return
    }

    // bcrypt => pour mdp haché
    err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        http.Error(w, "Mot de passe incorrect", http.StatusUnauthorized)
        return
    }

    // Création cookie session
    cookie := http.Cookie{
        Name:   "session",
        Value:  user.Email,
        Path:   "/",
        MaxAge: 20, // cookie dure 20s
    }
    http.SetCookie(w, &cookie)

    http.Redirect(w, r, "/secret", http.StatusFound)
}

// définition du handler
// fonction qui sera appelé à chaque fois qu'un utilisateur visite un URL
// => w http.ResponseWriter -> permet d'envoyer une réponse au client(navigateur)
// => r *http.Request -> contient les infos sur la requête (URL, headers, méthode...)
func home(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    // vérifie si déjà connecté
    cookie, err := r.Cookie("session")
    if err == nil && cookie.Value != "" {
        fmt.Fprintf(w, `
            <h1>Bonjour !</h1>
            <a href="/secret"><button>Accéder à la zone secrète</button></a>
        `)
        return
    }

    // page principale login
    fmt.Fprint(w, `
        <h1>Connexion</h1>

        <form method="POST" action="/login-classic">
            <label>Email :</label><br>
            <input type="email" name="email"><br><br>

            <label>Mot de passe :</label><br>
            <input type="password" name="password"><br><br>

            <button type="submit">Se connecter</button>
        </form>

        <br>

        <a href="/inscription">
            <button>S'inscrire</button>
        </a>

        <br><br>

        <h3>Ou</h3>

        <a href="/login">
            <button>Se connecter avec Google</button>
        </a>

        <a href="/login-idCluster">
            <button>Se connecter avec idCluster</button>
        </a>
    `)
}



func handleInscription(w http.ResponseWriter, r *http.Request) {

    // si méthode GET => afficher le formulaire
    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        fmt.Fprint(w, `
            <h1>Inscription</h1>
            <form method="POST" action="/inscription">
                <label>Nom :</label><br>
                <input type="text" name="nom"><br><br>

                <label>Prénom :</label><br>
                <input type="text" name="prenom"><br><br>

                <label>Email :</label><br>
                <input type="email" name="email"><br><br>

                <label>Mot de passe :</label><br>
                <input type="password" name="password"><br><br>

                <label>Confirmer mot de passe :</label><br>
                <input type="password" name="confirm"><br><br>

                <button type="submit">S'inscrire</button>
            </form>
        `)
        return
    }

    // si méthode POST => traiter les données
    if r.Method == http.MethodPost {

        nom := r.FormValue("nom")
        prenom := r.FormValue("prenom")
        email := r.FormValue("email")
        password := r.FormValue("password")
        confirm := r.FormValue("confirm")

        //pour la validation

        if nom == "" || prenom == "" || email == "" || password == "" || confirm == "" {
            http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
            return
        }

        if password != confirm {
            http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
            return
        }

        if len(password) < 4 {
            http.Error(w, "Mot de passe trop court (min 4 caractères)", http.StatusBadRequest)
            return
        }

        var existingUser User
        result := DB.Where("email = ?", email).First(&existingUser)

        if result.Error == nil {
            http.Error(w, "Email déjà utilisé", http.StatusBadRequest)
            return
        }

        // => création de l'utilisateur

        // Hash du mot de passe
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            http.Error(w, "Erreur serveur", http.StatusInternalServerError)
            return
        }

        newUser := User{
            Nom:      nom,
            Prenom:   prenom,
            Email:    email,
            Password: string(hashedPassword),
        }

        DB.Create(&newUser)

        // redirection vers login
        http.Redirect(w, r, "/", http.StatusFound)
    }
}

// Quand quelqu’un va sur /login, on le renvoie chez Google pour qu’il se connecte
func loginGoogle(w http.ResponseWriter, r *http.Request) {
    // génère l'URL de redirection vers Google
    // state-random = une chaîne pour le paramètre state (normalement un token anti-CSRF)
    // l'URL générée contient : client_id, redirect_uri, scope, response_type=code, state
    url := oauth2Config.AuthCodeURL("state-random", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
    // paramètre à mettre pour forcer la session
    // oauth2.ApprovalForce -> force Google à demander de se reconnecter, même si un cookie Google existe
    // oauth2.AccessTypeOffline → utile pour obtenir un refresh token si tu en as besoin

    // envoie une réponse HTTP 302 au navigateur
    // le navigateur est redirigé vers l’URL de Google
    // Google affiche sa page login
    http.Redirect(w, r, url, http.StatusFound)
}

/*// Quand quelqu’un va sur /login-idCluster, on le renvoie chez idCluster pour qu’il se connecte
func loginIdCluster(w http.ResponseWriter, r *http.Request) {
    // génère l'URL de redirection vers idCluster
    // l'URL générée contient : client_id, redirect_uri, scope, response_type=code, state
    //"https://auth.pawx.asgard.idenv.fr/auth" => url de mon serveur openid
    //url := "https://auth.pawx.asgard.idenv.fr/auth"
    url := "https://auth.pawx.asgard.idenv.fr/demoapp/cbregister"
    // envoie une réponse HTTP 302 au navigateur
    // le navigateur est redirigé vers l’URL de Idcluster
    // Google affiche sa page login
    http.Redirect(w, r, url, http.StatusFound)
}*/

func loginIdCluster(w http.ResponseWriter, r *http.Request) {
    url := IdClusterOauthConfig.AuthCodeURL("state-random", oauth2.SetAuthURLParam("lang", "fr"))
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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
        Name:   "session",
        Value:  claims.Email,
        Path:   "/", // valable pour tout le site
        MaxAge: 20,  // cookie dure 20s
    }
    http.SetCookie(w, &cookie) // ajoute un header Set-Cookie à la réponse

    // 8. Réponse
    //fmt.Fprintf(w, "Connecté avec Google ! Email : %s, ID BD : %d", claims.Email, user.ID)

    // redirection vers login
    http.Redirect(w, r, "/", http.StatusFound)
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

    http.HandleFunc("/login-idCluster", loginIdCluster)

    http.HandleFunc("/login-classic", loginClassic)
    http.HandleFunc("/inscription", handleInscription)

    http.HandleFunc("/callback", callback)

    // routes protégées
    // on ne passe pas directement gestionnaire_secret à HandleFunc
    // on passe requireLogin(gestionnaire_secret) :
    //  requireLogin prend gestionnaire_secret en paramètre (next)
    //  elle renvoie un nouveau handler qui :
    //      vérifie le cookie
    //      puis appelle gestionnaire_secret si tout va bien.
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

/*

faire comme google mais version idCluster
A rajouer :
https://auth.pawx.asgard.idenv.fr/auth
=> c'est la base url de mon serveur openid

requête entière
à construire via la lib Go
comme pour Google auth

https://auth.pawx.asgard.idenv.fr/demoapp/cbregister
=> environnement pawx sur postgres avec openid sur la latest de MFA


Services >> Fournisseurs >> Demo App >> Paramètre OpenID
URIs de retour autorisés > URL => j'ai ajouté http://8080/callback
finalement il faut faire : 
http://localhost:8080

il faut que je mette mon client_id


exemple : là, la location pour google est https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=&prompt=consent&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&response_type=code&scope=openid+email+profile&state=state-random

*/
