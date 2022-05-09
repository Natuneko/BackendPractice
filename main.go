package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Payload struct {
	Flag string `json:"flag"`
	Data string `json:"data"`
}

type Register struct {
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

type AccountDelete struct {
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

func MainHandle(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		path := r.URL.Path[5:]
		switch path {
		case "account":
			// w.Write([]byte("account"))
			switch r.Method {
			case "POST":
				body, _ := ioutil.ReadAll(r.Body)
				var payload Payload
				err := json.Unmarshal(body, &payload)
				if err != nil {
					w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"system error"}`, payload.Flag)))
					return
				}
				fmt.Println(payload.Flag)
				switch payload.Flag {
				case "register":
					data, err := base64.RawStdEncoding.DecodeString(payload.Data)
					// data, err := base64.RawURLEncoding.DecodeString(payload.Data)
					if err != nil {
						fmt.Println(err)
						w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"system error"}`, payload.Flag)))
						return
					}
					var register Register
					err = json.Unmarshal(data, &register)
					if err != nil {
						fmt.Println(err)
						w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"system error"}`, payload.Flag)))
						return
					}

					hash := sha256.Sum256([]byte(register.Password))
					register.Password = hex.EncodeToString(hash[:])

					_, err = db.Exec(`insert into account(mail,password) values (?,?)`, register.Mail, register.Password)
					if err != nil {
						fmt.Println(err)
						w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"system error"}`, payload.Flag)))
						return
					}
					w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"None"}`, payload.Flag)))
					return

				default:
					w.Write([]byte(`{"error":"Not found flag"}`))
					return
				}

			case "DELETE":
				body, _ := ioutil.ReadAll(r.Body)
				var payload Payload
				err := json.Unmarshal(body, &payload)
				if err != nil {
					w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"system error"}`, payload.Flag)))
					return
				}

				switch payload.Flag {
				case "accountdelete":
					data, err := base64.RawStdEncoding.DecodeString(payload.Data)
					// data, err := base64.RawURLEncoding.DecodeString(payload.Data)
					if err != nil {
						fmt.Println(err)
						w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"system error"}`, payload.Flag)))
						return
					}
					var accountdelete AccountDelete
					err = json.Unmarshal(data, &accountdelete)
					if err != nil {
						fmt.Println(err)
						w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"system error"}`, payload.Flag)))
						return
					}

					hash := sha256.Sum256([]byte(accountdelete.Password))
					accountdelete.Password = hex.EncodeToString(hash[:])

					_, err = db.Exec(`delete from account where mail = ? and password = ?`, accountdelete.Mail, accountdelete.Password)
					if err != nil {
						fmt.Println(err)
						w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"system error"}`, payload.Flag)))
						return
					}
					w.Write([]byte(fmt.Sprintf(`{"flag":"%s","error":"None"}`, payload.Flag)))
					return

				default:
					w.Write([]byte(`{"error":"Not found flag"}`))
					return
				}

			default:
				w.Write([]byte(`"error":"Not found"}`))
				return
			}
		default:
			w.Write([]byte(`{"error":"page not found"}`))
		}
	} else {
		w.Write([]byte("page not found"))
	}
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "users.sqlite")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`create table if not exists "account" ("mail" STRING, "password" STRING primary key)`)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", MainHandle)
	go http.ListenAndServe(":448", nil)
	fmt.Println("Start Server")
	ch := make(chan os.Signal)
	<-ch
}
