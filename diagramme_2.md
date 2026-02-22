# Diagramme de séquence

## 1 - Développement d’un serveur web en Go

Développement d’un serveur web en Go :

avec des routes publiques :
  * `/` -> page de connexion
  * `/inscription` -> création de compte (email, nom, mdp ...)
  * `/login` -> connexion Google
  * `/login-classic` -> connexion email + mdp
  * `/callback` -> retour Google

avec des routes protégées :
  * `/secret` -> on peut y accéder que si connecté (`requireLogin(...)`)

=> logique d’authentification :
  * cookies (nommés session)
  * mdp (hashé avec bcrypt)
  * vérification email + password \
  ou
  * authentification via Google (cookie aussi mais on peut oenlever et laisser ceux de google)

une base SQLite pour stocker les utilisateurs :
  * nom
  * prénom
  * email (unique)
  * mdp hashé


## 2 - Intégration d’OpenID Connect

Intégration d'un système d’authentification basé sur OpenID Connect (OIDC) avec Google.

OpenID Connect : protocole d'identité basé sur OAuth2.

* OAuth 2.0 -> protocole d'autorisation
* OIDC -> protocole d’authentification d’identité

OIDC est une couche d’identité construite au-dessus d’OAuth2.
Elle permet à une application de déléguer la connexion utilisateur à un Identity Provider comme Google.

Dans mon projet, OIDC sert à :
* rediriger l’utilisateur vers Google
* récupérer un ID token signé
* vérifier la signature et la validité
* extraire l’adresse email
* connecter ou créer l’utilisateur

## 3 - Authentification Google (Identity Provider)

Configuration de Google comme Identity Provider OpenID Connect.

```mermaid
sequenceDiagram
  participant Utilisateur
  participant ServeurGo
  participant Google

  Utilisateur ->> ServeurGo : Clique "Se connecter avec Google"
  ServeurGo ->> Google : Redirection OAuth2
  Google ->> Utilisateur : Page de connexion Google
  Utilisateur ->> Google : Saisit ses identifiants
  Google ->> ServeurGo : Redirection avec code d’autorisation
  ServeurGo ->> Google : Échange code contre ID Token
  Google ->> ServeurGo : ID Token (+ Access Token)
  ServeurGo ->> ServeurGo : Vérifie informations + expiration
  ServeurGo ->> Utilisateur : Session créée (cookie)
```


## 4 - Authentification classique (email + mot de passe)

L’utilisateur peut aussi saisir son email + son mdp puis cliquer sur "Se connecter"

Le serveur :
* recherche l’utilisateur par email dans la bdd
* compare le mot de passe avec bcrypt
* crée une session si valide

```mermaid
sequenceDiagram
  participant Utilisateur
  participant ServeurGo
  participant DB

  Utilisateur ->> ServeurGo : POST /login-classic
  ServeurGo ->> DB : Recherche email

  alt Email trouvé
    DB ->> ServeurGo : Utilisateur
    ServeurGo ->> ServeurGo : Vérification bcrypt
    ServeurGo ->> Utilisateur : Cookie session
  else Email inexistant
    DB ->> ServeurGo : Aucun résultat
    ServeurGo ->> Utilisateur : Erreur
  end
```


## 5 - Inscription

L’utilisateur peut cliquer sur "S’inscrire".

Il remplit :
* Nom
* Prénom
* Email
* Mot de passe
* Confirmation mot de passe

Le serveur :
* valide les champs (ex : nb charactères mdp > 6)
* vérifie que l’email n’existe pas déjà
* hash le mot de passe (bcrypt)
* enregistre dans la bdd
* crée une session (cookie nommée session => expiration 20s)

```mermaid
sequenceDiagram
  participant Utilisateur
  participant ServeurGo
  participant DB

  Utilisateur ->> ServeurGo : POST /inscription
  ServeurGo ->> ServeurGo : Validation données
  ServeurGo ->> DB : Vérifier email

  alt Email disponible
    ServeurGo ->> ServeurGo : Hash password
    ServeurGo ->> DB : Insert utilisateur
    ServeurGo ->> Utilisateur : Cookie session
  else Email déjà utilisé
    DB ->> ServeurGo : Utilisateur existant
    ServeurGo ->> Utilisateur : Erreur
  end
```


## 6 - Matching de compte avec l’adresse email

Pour Google :

```mermaid
sequenceDiagram
  participant Google
  participant ServeurGo
  participant DB

  Google ->> ServeurGo : ID Token
  ServeurGo ->> ServeurGo : Extraction email
  ServeurGo ->> DB : Recherche email

  alt Utilisateur existe
    DB ->> ServeurGo : Utilisateur trouvé
  else Utilisateur inexistant
    DB ->> ServeurGo : Aucun résultat
    ServeurGo ->> DB : Création utilisateur
  end
```

Le matching se fait uniquement sur l’email.


## 7 - Stockage des utilisateurs dans SQLite

Utilisation de SQLite via GORM pour stocker :
gorm => orm de go (pour communiquer avec un bdd)
* ID
* Nom
* Prénom
* Email (unique)
* Password (hash bcrypt)


## 8 - Conteneurisation avec Docker

Création d’une image Docker contenant :
* dépendances
* base SQLite
* application exposée sur le port 8080


## 9 - Déploiement dans Kubernetes

Déploiement sur un cluster Kubernetes (ex : Minikube)

Deployment :
* lance le pod
* redémarre si crash
* gère les mises à jour

Service :
* expose l’application
* fournit une IP stable
* redirige le trafic vers le pod

Application accessible via navigateur.


## 10 - Tests automatisés du front web

Utilisation de Playwright pour tester :
  * page login
  * inscription
  * connexion classique
  * connexion Google

