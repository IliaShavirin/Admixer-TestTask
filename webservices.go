package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	_ "github.com/denisenkom/go-mssqldb"
)

type Request struct {
	RequestID  int    `json:"request_id,omitempty"`
	UrlPackage []int  `json:"url_package"`
	IP         string `json:"ip"`
}

type Response struct {
	Price float64 `json:"price"`
}

var db *sql.DB

var server = "admixer.database.windows.net"
var port = 1433
var user = "Test"
var password = "Admixer#221"
var database = "Admixer-testTask"

func main() {
	// Build connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;encrypt=true;TrustServerCertificate=false",
		server, user, password, port, database)
	var err error

	// Create connection pool
	db, err = sql.Open("sqlserver", connString)

	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	createTable(db)
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc("/url", urlHandler)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func createTable(db *sql.DB) {
	if !tableExists(db, "Admixer") {
		_, err := db.Exec("CREATE TABLE Admixer (id DECIMAL, url VARCHAR(100))")
		if err != nil {
			fmt.Println("Error creating table:", err)
			panic(err.Error())
		}
	}

	row := db.QueryRow("SELECT COUNT(*) FROM Admixer")
	var count int
	if err := row.Scan(&count); err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		_, err := db.Exec(`INSERT INTO Admixer (id, url) VALUES
            (1, 'http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=1'),
            (2, 'http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=2'),
            (3, 'http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=3'),
            (4, 'http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=4')`)

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Table created successfully")
}

func getURL(id int64) (string, error) {
	var url string
	connString := fmt.Sprintf("SELECT url FROM Admixer WHERE id = %d", id)
	err := db.QueryRow(connString).Scan(&url)
	if err != nil {
		return "", err
	}
	fmt.Println("url:", url)
	return url, nil
}

func urlHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		postRequest(w, r)
	case "GET":
		getRequest(w, r)
	default:
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func isValidIP(ip string) bool {
	if ip == "" {
		return false
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || parsedIP.To4() == nil {
		return false
	}
	return true
}

func tableExists(db *sql.DB, tableName string) bool {

	connString := fmt.Sprintf("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME =  '%s'", tableName)
	rows, err := db.Query(connString)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
	}

	return count > 0
}

func handleRequest(w http.ResponseWriter, r *http.Request, urlPackages []int, ip string) {
	if !isValidIP(ip) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var maxPrice float64

	for _, urlID := range urlPackages {
		url, err := getURL(int64(urlID))
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var jsonResponse map[string]interface{}
		err = json.Unmarshal(body, &jsonResponse)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		price, ok := jsonResponse["price"].(float64)

		if !ok {
			price = 0
		}

		if price > maxPrice {
			maxPrice = price
		}
	}

	// Create response
	response := Response{
		Price: maxPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getRequest(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")

	fmt.Println("requestID:", requestID)

	urlPackagesStr := r.URL.Query()["url_package"]

	urlPackages := make([]int, len(urlPackagesStr))

	for i, s := range urlPackagesStr {
		n, err := strconv.Atoi(s)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		urlPackages[i] = n
	}

	ip := r.URL.Query().Get("ip")

	handleRequest(w, r, urlPackages, ip)
}

func postRequest(w http.ResponseWriter, r *http.Request) {
	var request Request

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	requestID := request.RequestID

	fmt.Println("requestID:", requestID)

	urlPackages := request.UrlPackage
	ip := request.IP

	handleRequest(w, r, urlPackages, ip)
}
