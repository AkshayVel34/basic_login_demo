package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// Global DB reference
var db *sql.DB

// Data model
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	var err error

	// ---------------------------------------------------------
	// FIXED MYSQL CONNECTION STRING
	// ---------------------------------------------------------
	dbUser := "root"
	dbPass := "root@134" // your original password (correctly escaped)
	dbName := "login_demo"

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", dbUser, dbPass, dbName)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to create DB object:", err)
	}

	// Test connection
	if err = db.Ping(); err != nil {
		log.Fatal("MySQL NOT connected:", err)
	}
	log.Println("MySQL connected successfully!")

	//-----------------------------------------------------------
	// ROUTES
	//-----------------------------------------------------------
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/signup", serveSignupPage)
	http.HandleFunc("/home", serveHomePage)

	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/signup", signupHandler)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// -----------------------------------------------------------
// SERVE HTML PAGES
// -----------------------------------------------------------
func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "./static/index.html")
}

func serveSignupPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "./static/signup.html")
}

func serveHomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "./static/home.html")
}

// -----------------------------------------------------------
// LOGIN API
// -----------------------------------------------------------
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var storedPassword string

	err := db.QueryRow("SELECT password FROM users WHERE username=?", creds.Username).
		Scan(&storedPassword)

	if err == sql.ErrNoRows {
		jsonResponse(w, false, "User not found")
		return
	}

	if err != nil {
		jsonResponse(w, false, "DB error")
		return
	}

	// simple plaintext check (you should hash)
	if creds.Password != storedPassword {
		jsonResponse(w, false, "Incorrect password")
		return
	}

	jsonResponse(w, true, "Login successful")
}

// -----------------------------------------------------------
// SIGNUP API
// -----------------------------------------------------------
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)",
		creds.Username, creds.Password)

	if err != nil {
		jsonResponse(w, false, "Username already exists or DB error")
		return
	}

	jsonResponse(w, true, "Signup successful")
}

// -----------------------------------------------------------
// JSON Response Utility
// -----------------------------------------------------------
func jsonResponse(w http.ResponseWriter, success bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}
