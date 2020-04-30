package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Database Connection String
func InitDB() (*sql.DB, error) {
	dbhost := os.Getenv("DBHOST")
	dbuser := os.Getenv("DBUSER")
	dbpass := os.Getenv("DBPASS")
	dbport := os.Getenv("DBPORT")
	dbname := os.Getenv("DB")
	db, err := sql.Open("mysql", dbuser+":"+dbpass+"@tcp("+dbhost+":"+dbport+")/"+dbname)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// GetEmails function Gets the emails of user and client from database
func GetEmails() ([]string, []string, []string) {
	db, err := InitDB()
	if err != nil {
		log.Printf("Database Connection String Failed {%v}\n", err)
		return nil, nil, nil
	}
	query, err := db.Query(`SELECT Client_user_email,Client_Email,client_user_timezone from emailsender.ClientUserEmail`)
	if err != nil {
		log.Println("Query Failed!")
		return nil, nil, nil
	}
	defer query.Close()
	var UserEmail []string
	var ClientEmail []string
	var UserTimeZone []string
	for query.Next() {
		var user, client, usertimezone string
		err = query.Scan(&user, &client, &usertimezone)
		if err != nil {
			log.Println("Error to Execute Statements")
			return nil, nil, nil
		}
		UserEmail = append(UserEmail, user)
		ClientEmail = append(ClientEmail, client)
		UserTimeZone = append(UserTimeZone, usertimezone)
	}
	return UserEmail, ClientEmail, UserTimeZone
}

// SendEmail Function sends the email.
func sendEmail(body, to string, wg *sync.WaitGroup) {
	defer wg.Done()
	from := os.Getenv("FROM")
	pass := os.Getenv("PASS")
	// to := os.Getenv("TO")
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Google Testing\n" +
		body
	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
	log.Printf("sent, visit to gmail [%s]", to)

}

// Execute function Is the combination of the all functions to run
func Execute(body string, wg *sync.WaitGroup) {
	user, client, usertimezone := GetEmails()
	if user == nil || client == nil || usertimezone == nil {
		return
	}
	LocationTime := ""
	Comparision := ""
	for _, Location := range usertimezone {
		wg.Add(1)
		loc, err := time.LoadLocation(Location)
		if err != nil {
			log.Printf("Error In Location {%v}\n", err)
		}
		current_time := time.Now().In(loc)
		destinationTime := current_time.Format("2006-April-02 15:04:05")
		// fmt.Println(destinationTime)
		LocationTime = destinationTime

		destinationYear := current_time.Year()
		destinationMonth := current_time.Month()
		destinationDay := current_time.Day()
		Time := fmt.Sprintf("%v-%v-%v 18:26:00", destinationYear, destinationMonth, destinationDay)
		Comparision = Time

	}
	if LocationTime == Comparision {
		for _, user := range user {
			wg.Add(1)
			go sendEmail(body, user, wg)
		}

		for _, client := range client {
			wg.Add(1)
			go sendEmail(body, client, wg)
		}
	}
	log.Println("Time Not Matched!")
	log.Println(LocationTime)
	log.Println(Comparision)
	//os.Exit(1)
	wg.Wait()
}

func main() {
	var wg sync.WaitGroup
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				go Execute("Good Morning", &wg)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	wg.Wait()
	fmt.Println("Finished Execution")
}
