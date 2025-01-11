package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var secrets = map[string]string{
	"token":   "supersecrettoken",
	"admin":   "admin123",
	"hidden":  "hidden_value",
	"message": "This is a secret message.",
}

func main() {
	// Set up logging to the terminal
	log.SetOutput(os.Stdout) // Log to the terminal (stdout)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/api/secrets", secretsHandler)

	fmt.Println("Hackable server is running on :8080...")
	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil) //change this to your local IP in this format: 000.000.0.000:8080
	if err != nil {
		log.Printf("Error starting server: %v", err)
		os.Exit(1)
	}
}

// Log request and actions for each handler
func logRequest(r *http.Request, action string) {
	log.Printf("Request: %s %s from %s - Action: %s", r.Method, r.URL.Path, r.RemoteAddr, action)
}

// Home page handler
func homeHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "Rendering home page")
	fmt.Fprintf(w, `<html>
		<head><title>Hackable Go Server</title></head>
		<body>
			<h1>Welcome to the Advanced Hackable Go Server</h1>
			<p>Endpoints to explore:</p>
			<ul>
				<li>/upload - Upload files</li>
				<li>/read?file=filename - Read files or execute commands</li>
				<li>/admin - Secret admin panel</li>
				<li>/echo?msg=yourmessage - Echo your message</li>
				<li>/api/secrets - Access secret values</li>
			</ul>
		</body>
	</html>`)
}

// File upload handler
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "Handling file upload")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		log.Printf("Error: Invalid method for /upload, expected POST")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to parse file: "+err.Error(), http.StatusBadRequest)
		log.Printf("Error: Failed to parse file: %v", err)
		return
	}
	defer file.Close()

	filename := header.Filename
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error: Failed to read file: %v", err)
		return
	}

	// Save file in the uploads directory
	filePath := "./uploads/" + filename
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error: Failed to save file: %v", err)
		return
	}

	log.Printf("File uploaded successfully: %s", filePath)
	fmt.Fprintf(w, "File uploaded successfully: %s", filePath)
}

// File read handler with command execution
func readHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "Handling file read or command execution")

	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "No file specified", http.StatusBadRequest)
		log.Printf("Error: No file specified in the request")
		return
	}

	// Hidden feature: Execute commands if file starts with "cmd:"
	if strings.HasPrefix(file, "cmd:") {
		command := strings.TrimPrefix(file, "cmd:")
		// Fix: Use cmd /C for executing Windows commands, allowing multiple commands
		output, err := exec.Command("cmd", "/C", command).CombinedOutput()
		if err != nil {
			http.Error(w, "Command execution failed: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Error: Command execution failed: %v", err)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write(output)
		log.Printf("Executed command: %s", command)
		return
	}

	// Directory traversal vulnerability
	filePath := "./uploads/" + file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error: Failed to read file: %v", err)
		return
	}

	log.Printf("File read successfully: %s", filePath)
	w.Write(data)
}

// Admin panel handler with command execution
func adminHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.URL.Query().Get("auth")

	// Weak security logic
	if auth == secrets["admin"] {
		command := r.URL.Query().Get("exec")
		if command != "" {
			output, err := exec.Command("cmd", "/C", command).CombinedOutput()
			if err != nil {
				http.Error(w, "Command execution failed: "+err.Error(), http.StatusInternalServerError)
				log.Printf("Error: Command execution failed: %v", err)
				return
			}
			fmt.Fprintf(w, `<h1>Command Execution Result:</h1><pre>%s</pre>`, output)
			log.Printf("Executed command by admin: %s", command)
			return
		}

		log.Printf("Admin accessed the panel")
		fmt.Fprintf(w, `<h1>Welcome, Admin!</h1>
		<p>Here are some secret server files:</p>
		<ul>
			<li><a href="/read?file=secret.txt">secret.txt</a></li>
			<li><a href="/read?file=passwords.txt">passwords.txt</a></li>
		</ul>
		<p>Use /admin?auth=admin123&exec=command to run commands.</p>`)
	} else {
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		log.Printf("Unauthorized access attempt to admin panel from %s", r.RemoteAddr)
	}
}

// Echo handler with potential injection vulnerability
func echoHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "Handling echo message")

	msg := r.URL.Query().Get("msg")
	if msg == "" {
		http.Error(w, "No message provided", http.StatusBadRequest)
		log.Printf("Error: No message provided in echo request")
		return
	}

	// Potential HTML injection
	fmt.Fprintf(w, `<html><body><h1>Your Message:</h1><p>%s</p></body></html>`, msg)
	log.Printf("Echo message sent: %s", msg)
}

// API secrets handler
func secretsHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if !strings.Contains(token, secrets["token"]) {
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		log.Printf("Unauthorized access attempt to secrets API from %s", r.RemoteAddr)
		return
	}

	// Return secrets as JSON
	jsonData, _ := json.Marshal(secrets)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
	log.Printf("Secrets API accessed by %s", r.RemoteAddr)
}
