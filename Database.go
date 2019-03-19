//el programa es para conectar al base de datos y insertar "row" nuevos por Go
//

package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Alicante"
	dbname   = "Login"
)

func database() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	//crear el usuario
	sqlStatement := `
	INSERT INTO user_account (username, user_password)
	VALUES ($1, $2)
	RETURNING id_number`
	personid := 0

	//luego reemplaza "username" y "user_password" con los variables de la entrada del usuario
	err = db.QueryRow(sqlStatement, "llopez", "password1ddd").Scan(&personid)
	//err = db.Ping() -- check to see if code can connect to database
	if err != nil {
		panic(err)
	}

	//fmt.Println("Successfully connected!")

	fmt.Println("New record ID is: ", personid)
}
