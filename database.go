package main

import (
	"gorm.io/gorm" // ORM => permet de manipuler la base des structs Go au lieu d'écrire du SQL brut
	"gorm.io/driver/sqlite" // driver pour parler à une base SQLite
)

// DB => variable globale qui représente la connexion à la base de données
var DB *gorm.DB // Type : *gorm.DB -> pointeur vers une instance de GORM

type User struct { 
	ID uint `gorm:"primaryKey"` 
	Email string `gorm:"unique"` 
} 


//initDB() ouvre la base users.db et s’assure qu’il y a une table users avec les colonnes id et email
func initDB() error {
	var err error

	/*
	gorm.Open ouvre une connexion à la base
	sqlite.Open("users.db") -> crée (ou ouvre) un fichier users.db dans ton dossier
	&gorm.Config{} -> config par défaut
	Si tout va bien, DB contient la connexion
	*/
	DB, err = gorm.Open(sqlite.Open("users.db"), &gorm.Config{}) 
	if err != nil { 
		return err 
	} 
	
	// Migration : crée la table si elle n'existe pas 
	/*
	“Migration automatique” : GORM regarde la struct User
	et crée/modifie la table users pour qu’elle corresponde au modèle
		Si la table n’existe pas => il la crée
		Si elle existe => il l’adapte (dans certaines limites)
	*/
	return DB.AutoMigrate(&User{})
}