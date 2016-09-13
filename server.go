package server

import (
	"github.com/op/go-logging"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"crypto/sha256"
	"fmt"
	"encoding/hex"
	"log"
)

type CLinkServer struct {
	DB *sql.DB
	Log *logging.Logger
}

func NewServer(dbLocation string, log *logging.Logger) (*CLinkServer, error){
	db, err := sql.Open("sqlite3", dbLocation)

	if err != nil {
		log.Fatal("Error opening db", err)
		return nil, err
	}

	result := CLinkServer{db, log}

	err = result.noResExecute(`create table if not exists access_logs(
	log_id text primary key not null,
	server_id integer not null,
	time integer not null,
	source_server_id text not null,
	session_id text not null)`)

	if err != nil {
		log.Fatal("Error creating access log table")
		return nil, err
	}

	err = result.noResExecute(`create table if not exists users(
	user_id integer primary key not null,
	email text not null,
	password text not null,
	secret blob not null)`)

	if err != nil {
		log.Fatal("Error creating user table")
		return nil, err
	}

	err = result.noResExecute(`create table if not exists sessions(
	session_id text primary key not null,
	server_id integer not null,
	user_id integer not null,
	connect_time integer not null,
	disconnect_time integer not null)`)

	if err != nil {
		log.Fatal("Error creating connection table")
		return nil, err
	}

	err = result.noResExecute(`create table if not exists servers(
	server_id integer primary key not null,
	domain text not null,
	security_level integer not null)`)

	if err != nil {
		log.Fatal("Error creating connection table")
		return nil, err
	}

	err = result.noResExecute(`create table if not exists own_proxies(
	own_proxy_id integer primary key not null,
	server_id integer not null,
	user_id integer not null,
	acquired_date integer not null)`)

	if err != nil {
		log.Fatal("Error creating connection table")
		return nil, err
	}

	s := sha256.New()
	s.Write([]byte("secret"))
	secret_bytes := s.Sum(nil)

	fmt.Printf("secret is:%s\n", hex.EncodeToString(secret_bytes))

	has_user1, err := result.hasUser(1)

	if err != nil {
		log.Fatal("Error populating user table")
		return nil, err
	}

	if !has_user1 {
		_, err = db.Exec(
			`INSERT INTO users(user_id, email, password, secret) VALUES (?, ?, ?, ?)`,
			1,
			"ilyail3@gmail.com",
			EncryptPassword("sheeps"),
			secret_bytes)

		if err != nil {
			log.Fatal("Error populating user table")
			return nil, err
		}
	}


	return &result, nil
}

func (s CLinkServer) hasUser(user_id int) (bool,error){
	rows, err := s.DB.Query("SELECT COUNT(1) FROM users WHERE user_id = ?", 1)

	if err != nil {
		log.Fatal("Error populating user table")
		return false, err
	}

	defer rows.Close()

	if(!rows.Next()) {
		return false, nil
	} else {
		var num int

		err = rows.Scan(&num)

		if err != nil {
			log.Fatal("Error populating user table")
			return false, err
		}

		return num == 1, nil
	}
}

func (s CLinkServer) noResExecute(query string) error{
	_, err := s.DB.Exec(query)

	return err
}

func (s CLinkServer) Close(){
	s.DB.Close()
}