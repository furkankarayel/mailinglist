package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mailinglist/mdb"
	"net/http"
)

func setJsonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), target)
}

func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeaders(w)
	data, serverErr := withData()
	if serverErr != nil {
		w.WriteHeader(500)
		serverErrJson, err := json.Marshal(&serverErr)
		if err != nil {
			log.Println(err)
			return
		}
		w.Write(serverErrJson)
		return
	}
	dataJson, err := json.Marshal(&data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	w.Write(dataJson)

}

func returnErr(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (interface{}, error) {
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}

func Serve(db *sql.DB, bind string) {
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/update", UpdateEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))
	http.Handle("/email/getBatch", GetEmailBatch(db))
	log.Printf("JSON API Listening on %s", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatalf("JSON server error: %v", err)
	}
}

func GetEmailBatch(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		queryOptions := mdb.GetEmailBatchQueryParams{}
		fromJson(r.Body, &queryOptions)

		if queryOptions.Count <= 0 || queryOptions.Page <= 0 {
			returnErr(w, errors.New("Page and Count fields are required and must be >0"), 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmailBatch: %v", queryOptions)
			return mdb.GetEmailBatch(db, queryOptions)
		})
	})
}

func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(r.Body, &entry)
		if err := mdb.DeleteEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON DeleteEmail: %v", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(r.Body, &entry)
		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON UpdateEmail: %v", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})

}

func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(r.Body, &entry)

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmail: %v", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(r.Body, &entry)
		if err := mdb.CreateEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON CreateEmail: %v", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}
