package server

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/satori/go.uuid"
	"time"
	"database/sql"
)

type ConnectionRequest struct {
	UserId int `json:"user_id"`
	TargetServerId int `json:"target_server_id"`
	ProxyIds []int `json:"proxy_ids"`
}

func (cr ConnectionRequest) GetUserId() int{
	return cr.UserId
}

func numberInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func connectTransaction(tx *sql.Tx, session_id uuid.UUID, t time.Time, connect_request *ConnectionRequest) error{
	_, err := tx.Exec("INSERT INTO sessions(session_id, server_id, user_id, connect_time, disconnect_time) " +
		"VALUES (?, ?, ?, ?, 0)",
		session_id.String(),
		connect_request.TargetServerId,
		connect_request.UserId,
		t.Unix())

	if err != nil {
		return err
	}

	var source int
	source = 0

	for _, proxy_id := range connect_request.ProxyIds {
		_, err = tx.Exec("INSERT INTO access_logs(log_id, server_id, time, source_server_id, session_id) " +
			"VALUES (?, ?, ?, ?, ?)",
			uuid.NewV4().String(),
			proxy_id,
			t.Unix(),
			source,
			session_id.String())

		if err != nil {
			return err
		}

		source = proxy_id
	}

	_, err = tx.Exec("INSERT INTO access_logs(log_id, server_id, time, source_server_id, session_id) " +
		"VALUES (?, ?, ?, ?, ?)",
		uuid.NewV4().String(),
		connect_request.TargetServerId,
		t.Unix(),
		source,
		session_id.String())

	if err != nil {
		return err
	}

	return nil
}

func (s CLinkServer)Connect(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	var connect_request ConnectionRequest

	err = json.Unmarshal(body, &connect_request)

	if err != nil {
		s.Log.Fatal("Error unmarshalling json", err)
		s.BadRequestError(w, "bad_json", "Error reading json request")
		return
	}

	if !s.AuthenticateRequest(w, r, body, &connect_request){
		return
	}

	if(numberInSlice(connect_request.TargetServerId, connect_request.ProxyIds)){
		s.BadRequestError(w, "target_proxy", "The target cannot be used as a proxy")
		return
	}

	session_id := uuid.NewV4()

	t := time.Now()

	tx, err := s.DB.Begin()

	if err != nil {
		s.Log.Error("Failed to start transaction:" + err.Error())
		s.InternalError(w)
		return
	}

	err = connectTransaction(tx, session_id, t, &connect_request)

	if err != nil {
		s.Log.Error("Failed to write connection:" + err.Error())

		tx.Rollback()

		s.InternalError(w)
	} else {
		tx.Commit()
	}





}