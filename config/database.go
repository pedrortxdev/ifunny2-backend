package config

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	host     = "node.lunarhosting.com.br"
	port     = 3306
	user     = "u166_KylX8MoFxd"
	password = "5QV03xUdLvP@38RSkca!pDVR"
	dbname   = "s166_go-dev"
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
