package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

func GetDB() *sql.DB {
	fmt.Println("[database] Asked for MySQL connection")
	dbhost := "127.0.0.1:3306"
	if os.Getenv("database") != "" {
		dbhost = os.Getenv("database")
	}
	con, err := sql.Open("mysql", "root:@tcp("+dbhost+")/ginux")
	con.Exec("SET NAMES UTF8")
	if err != nil {
		fmt.Println("[database] Unable to set up connection!")
	}
	con.Ping()
	return con
}
