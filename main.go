//Tugas Pemweb 2

package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"

    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"
)

// User represents the user model
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
type Message struct {
    Message string `json:"Message"`
}



var db *sql.DB

// BasicAuth middleware for basic authentication
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        username, password, ok := r.BasicAuth()

        if !ok || !validateCredentials(username, password) {
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Validate username and password (this could be replaced by a database check)
func validateCredentials(username, password string) bool {
    return username == "admin" && password == "pemweb2"
}

// Get all users (READ)
func getUsers(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name FROM users")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        users = append(users, user)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

// Create a new user (CREATE)
func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result, err := db.Exec("INSERT INTO users (name) VALUES (?)", user.Name)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    id, err := result.LastInsertId()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    user.ID = int(id)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// Update an existing user (UPDATE)
func updateUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err = db.Exec("UPDATE users SET name = ? WHERE id = ?", user.Name, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    user.ID, _ = strconv.Atoi(id)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// Delete a user (DELETE)
func deleteUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    _, err := db.Exec("DELETE FROM users WHERE id = ?", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    var message Message

    message = Message{Message : "Berhasil Di Hapus"}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(message)
}

// Main function to set up the routes and start the server
func main() {
    // Database connection string modified for XAMPP
    dsn := "root:@tcp(127.0.0.1:3306)/tugas_pemweb" // Adjusted to connect to XAMPP's MySQL
    var err error
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }
    defer db.Close()

    // Setting up the router
    router := mux.NewRouter()

    // Routes
    router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }).Methods("GET")

    // Protect routes with basicAuth middleware
    router.HandleFunc("/users", basicAuth(getUsers)).Methods("GET")
    router.HandleFunc("/users", basicAuth(createUser)).Methods("POST")
    router.HandleFunc("/users/{id}", basicAuth(updateUser)).Methods("PUT")
    router.HandleFunc("/users/{id}", basicAuth(deleteUser)).Methods("DELETE")

    // Start server
    fmt.Println("Server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
