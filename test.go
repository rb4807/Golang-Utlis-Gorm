package main

import (
	"html/template"
	"log"
	"net/http"
)

// Define your data structure to pass to the template
type PageData struct {
	Title string
	Items []string
	User  User
}

type User struct {
	Name  string
	Email string
}

func text() {
	// Define server routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)

	// Serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Sample data to pass to the template
	data := PageData{
		Title: "Welcome to Go Templates",
		Items: []string{"Item 1", "Item 2", "Item 3"},
		User: User{
			Name:  "John Doe",
			Email: "john@example.com",
		},
	}

	// Create a new template set
	tmpl := template.New("layout")

	// Parse the template files
	tmpl, err := tmpl.ParseFiles("templates/layout.html", "templates/home.html")
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the template with data
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "About Us",
		User: User{
			Name:  "John Doe",
			Email: "john@example.com",
		},
	}

	// Create a new template set
	tmpl := template.New("layout")

	// Parse the template files
	tmpl, err := tmpl.ParseFiles("templates/layout.html", "templates/about.html")
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
