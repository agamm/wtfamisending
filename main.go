package main

// db and config
import (
	"database/sql"
	"code.google.com/p/gcfg"
	_ "github.com/lib/pq"
)

// output / logs
import (
	"fmt"
	"github.com/jcelliott/lumber"
	"io"
)

// http / web
import (
	"net/http"
	"net/http/httputil"
	"html/template"
)

// misc
import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Config
type Config struct {
	DB struct {
		UserName string
		Password string
		Name string
	}
	Server struct {
		Port string
	}
}
var config Config


// Globals
var db *sql.DB 
var logger *lumber.MultiLogger
type ResponseHTML struct {
	ResponseString string
}

func main() {

	// Add our loggers
	logger = lumber.NewMultiLogger()
	logConsole := lumber.NewConsoleLogger(lumber.DEBUG)
	logFile, err := lumber.NewRotateLogger("wtf.log", 13337, 9)
	handleErr(nil, err)
	logFile.Level(lumber.WARN)
	logger.AddLoggers(logConsole, logFile)

	// Load config
	err = gcfg.ReadFileInto(&config, "config")
	if err != nil {
	  logger.Error("Error loading configuration!", err)
	}

	// Ineed, we are up and running
	logger.Info("Wop we are up!")

	// DB
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		config.DB.UserName, config.DB.Password, config.DB.Name)

	// Super important not to use := as db is already defined
	db, err = sql.Open("postgres", dbinfo)
	handleErr(nil, err)
	defer db.Close()

	db.SetMaxIdleConns(100)

	// Make sure the database is accessible
	err = db.Ping()
	handleErr(nil, err)

	// HTTP Server
	http.HandleFunc("/wtf/", showRequest)
	http.HandleFunc("/", requestEntry)
	http.ListenAndServe(config.Server.Port, nil)
}

func showRequest(w http.ResponseWriter, r *http.Request) {
	rId := r.URL.Path[len("/wtf/"):]

	var raw_request string
	err := db.QueryRow("SELECT raw_request FROM requests WHERE id=$1", rId).Scan(&raw_request)
	
	// Todo, finout how to check if a specific error was raised.
	//eg. err == "sql: no rows in result set"
	if err != nil {
		logger.Info("%T %q", err, err); // move down after todo (use handleError)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "404 - Wtf are you trying to find...? ")
		return
	}

	query := r.URL.Query()

	if _, ok := query["html"]; ok {
		p := &ResponseHTML{ResponseString: raw_request}
		t, err := template.ParseFiles("response.html")
		handleErr(w, err)
		t.Execute(w, p)
		return
	}

	io.WriteString(w, raw_request)
}

func requestEntry(w http.ResponseWriter, r *http.Request) {
	lastInsertId, dumpString := saveRequest(w, r)

	if lastInsertId == "favicon" {
		w.Header().Set("Cache-Control", "max-age=2592000")
		io.WriteString(w, "Request is too boring, favicon, meh...")
		return
	}

	// Todo, findout why chrome freaks out :{}
	w.Header().Set("Location", "/wtf/" + lastInsertId + "?html=true")
	w.WriteHeader(301)
	w.Write([]byte(dumpString))
}

func saveRequest(w http.ResponseWriter, r *http.Request) (string, string) {
	// Check if favicon and drop insert to db
	if r.URL.Path == "/favicon.ico" {
		return "favicon", ""
	}

	ipAddr := strings.Split(r.RemoteAddr,":")[0] 
	if ipAddr == "" || strings.Contains(ipAddr, "[") {
		ipAddr = "0.0.0.0"
	}

	dump, err := httputil.DumpRequest(r, true)
	handleErr(w, err)

	dumpString := string(dump)

	var lastInsertId string
	hashedRequest := hashRequest(dumpString)
	err = db.QueryRow("INSERT INTO requests (id, raw_request, ip) VALUES($1, $2, $3) returning id;", hashedRequest, dumpString, ipAddr).Scan(&lastInsertId)
	
	// Todo, same as Todo in showRequest
	if err != nil {
		return hashedRequest, dumpString
	}
	handleErr(w, err)

	// Remove the hyphens - I hate them
	lastInsertId = strings.Replace(lastInsertId, "-", "", -1)

	return lastInsertId, dumpString
}

func hashRequest(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	hSum := hash.Sum(nil)
	hStr := hex.EncodeToString(hSum)
	return hStr
}

func handleErr(w http.ResponseWriter, err error) {
	if err != nil {

		logger.Error("(%T) %q", err, err)
		if w != nil {
			http.Error(w, "Some error occured", 500)
			return
		}
	}
}
