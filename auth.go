package server

import (
	"net/http"
	"time"
	"math"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"golang.org/x/crypto/bcrypt"
	"io"
	"crypto/rand"
	"fmt"
)


type AuthenticatedRequestBody interface{
	GetUserId() int
}

func hmacSha256(message []byte, key []byte)[]byte{
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}


func (s CLinkServer)AuthenticateRequest(w http.ResponseWriter, r *http.Request, body []byte, parsed_body AuthenticatedRequestBody) bool{
	MAX_DRIFT := float64(60 * 60 * 24 * 356)

	date_header := r.Header.Get("CLINK-DATE")

	if date_header == "" {
		s.BadRequestError(w, "date_missing", "CLINK-DATE header is missing")
		return false
	}

	t, err := time.Parse(time.RFC3339,date_header)

	if err != nil {
		s.Log.Warning("Failed to parse date string:" + err.Error())

		s.BadRequestError(w, "date_format", "Failed to parse CLINK-DATE header")
		return false
	}

	drift := math.Abs(float64(time.Now().Unix() - t.Unix()))

	if drift > MAX_DRIFT {
		s.BadRequestError(w, "date_format", "CLINK-DATE header drifts over 60 seconds, check clock on your machine")
		return false
	}

	rows, err := s.DB.Query("SELECT secret FROM users WHERE user_id = ?", parsed_body.GetUserId())

	if ! rows.Next() {
		s.BadRequestError(w, "bad_user_id", "Can't find user-id")
		return false
	}

	var data []byte

	rows.Scan(&data)
	rows.Close()

	expected := hex.EncodeToString(hmacSha256(body, data))
	got := r.Header.Get("CLINK-SIGNATURE")

	s.Log.Infof("Expect %s", expected)

	if !(strings.ToLower(expected) == strings.ToLower(got)) {
		s.BadRequestError(w, "bad signature", "Couldn't verify signature")
		return false
	}




	return true
}

func EncryptPassword(password string) string{
	salt := make([]byte, 8)

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err)
	}

	cost := bcrypt.DefaultCost

	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), cost)

	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("bcrypt$%d$%x$%x", cost, salt, hashed_password)
}