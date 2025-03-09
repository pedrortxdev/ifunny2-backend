package config

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	host     = "node.lunarhosting.com.br"
	port     = 3306
	user     = "não vou mostrar isso certo?"
	password = "nem isso ne kkkk"
	dbname   = "isso aqui tambem nao XD"
)

func InitDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		user, password, host, port, dbname)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com banco de dados: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco de dados: %v", err)
	}

	return db, nil
}
