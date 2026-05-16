package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mailingService/mdb"
	"net/http"
)

func setJsonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func fromJson[T any](body io.Reader, target T) (T, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
	return target, nil
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

func returnErrorJson(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (interface{}, error) {
		errorMessage := struct {
			Error string
		}{
			Error: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}

func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}

		_, err := fromJson(req.Body, &entry)
		if err != nil {
			return
		}

		if err := mdb.CreateEmail(db, entry.EmailAddress); err != nil {
			returnErrorJson(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("Create email successfully with address : %v\n", entry.EmailAddress)
			return mdb.GetEmail(db, entry.EmailAddress)
		})
	})
}

func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		entry := mdb.EmailEntry{}

		_, err := fromJson(req.Body, &entry)
		if err != nil {
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("Get email successfully with address : %v\n", entry.EmailAddress)
			return mdb.GetEmail(db, entry.EmailAddress)
		})
	})
}

func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			return
		}
		entry := mdb.EmailEntry{}

		_, err := fromJson(req.Body, &entry)
		if err != nil {
			return
		}

		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnErrorJson(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("Update email successfully with address : %v\n", entry.EmailAddress)
			return mdb.GetEmail(db, entry.EmailAddress)
		})
	})
}

func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}

		_, err := fromJson(req.Body, &entry)
		if err != nil {
			return
		}

		if err := mdb.DeleteEmail(db, entry.EmailAddress); err != nil {
			returnErrorJson(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("Delete email successfully with address : %v\n", entry.EmailAddress)
			return mdb.GetEmail(db, entry.EmailAddress)
		})
	})
}

func GetEmailPaginated(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}

		queryOptions := mdb.GetEmailPaginatedQueryParams{}
		_, err := fromJson(req.Body, &queryOptions)
		if err != nil {
			return
		}

		if queryOptions.Count <= 0 || queryOptions.Page <= 0 {
			returnErrorJson(
				w, errors.New("the Page and Count fields are required and must be greater than zero"),
				http.StatusBadRequest)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("Get paginated email successfully with options : %v\n", queryOptions)
			return mdb.GetEmailPaginated(db, queryOptions)
		})
	})
}

func Serve(db *sql.DB, bind string) {
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/paginated", GetEmailPaginated(db))
	http.Handle("/email/update", UpdateEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))

	log.Printf("Json API Server Listening on %s\n", bind)

	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatalf("failed to start http server with error %v", err)
	}
}
