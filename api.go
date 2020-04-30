package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Post struct {
	Client_Id            string `json:"client_id"`
	Client_Email         string `json:"client_email"`
	Client_Timezone      string `json:"client_timezone"`
	Client_User_Email    string `json:"client_user_email"`
	Client_User_TimeZone string `json:"client_user_timezone"`
}

var db *sql.DB
var err error

func main() {
	dbhost := os.Getenv("DBHOST")
	dbuser := os.Getenv("DBUSER")
	dbpass := os.Getenv("DBPASS")
	dbname := os.Getenv("DBNAME")
	dbport := os.Getenv("DBPORT")
	connectingstring := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" + dbport + ")/" + dbname
	db, err = sql.Open("mysql", connectingstring)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	router := mux.NewRouter()
	router.HandleFunc("/posts", getPosts).Methods("GET")
	router.HandleFunc("/posts", createPost).Methods("POST")
	router.HandleFunc("/posts/{id}", getPost).Methods("GET")
	//router.HandleFunc("/posts/{id}", updatePost).Methods("PUT")
	//router.HandleFunc("/posts/{id}", deletePost).Methods("DELETE")
	http.ListenAndServe(":8001", router)
}

//Fetch all Records
func getPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var posts []Post
	result, err := db.Query("SELECT client_id, client_email, client_timezone, client_user_email, client_user_timezone from ClientUserEmail")
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()
	for result.Next() {
		var post Post
		err := result.Scan(&post.Client_Id, &post.Client_Email, &post.Client_Timezone, &post.Client_User_Email, &post.Client_User_TimeZone)
		if err != nil {
			panic(err.Error())
		}
		posts = append(posts, post)
	}
	json.NewEncoder(w).Encode(posts)
}

//create record by POST Method
func createPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stmt, err := db.Prepare("INSERT INTO ClientUserEmail(client_email, client_timezone, client_user_email, client_user_timezone) VALUES(?,?,?,?)")
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	client_email := keyVal["client_email"]
	client_timezone := keyVal["client_timezone"]
	client_user_email := keyVal["client_user_email"]
	client_user_timezone := keyVal["client_user_timezone"]

	_, err = stmt.Exec(client_email, client_timezone, client_user_email, client_user_timezone)
	if err != nil {
		panic(err.Error())
	}
	fmt.Fprintf(w, "New Record is created")
}

//fetch single data by client_id
func getPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result, err := db.Query("select client_id, client_email, client_timezone, client_user_email, client_user_timezone FROM ClientUserEmail WHERE client_id = ?", params["client_id"])
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()
	var post Post
	for result.Next() {
		err := result.Scan(&post.Client_Id, &post.Client_Email, &post.Client_Timezone, &post.Client_User_Email, &post.Client_User_TimeZone)
		if err != nil {
			panic(err.Error())
		}
	}
	json.NewEncoder(w).Encode(post)
}
