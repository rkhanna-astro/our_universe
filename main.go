package main

import (
	// "errors"
	// "encoding/json"
	"database/sql" // For interacting with the SQL database
	"html/template"
	"log"
	"net/http"
	// "os"
	"fmt" // For printing and formatting output
	"path/filepath"
	"time" // For handling timestamps

	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx driver for database/sql compatibility
)

type Blog struct {
	ID        int
	Title     string
	Content   string
	CreatedAt time.Time
}

var templates *template.Template

var blogs []Blog

func init() {
	// Parse all templates
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {

	dsn := "postgres://postgres:hellorahul@localhost:5432/blogwebsite"

	// Connect to the database
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to the database successfully!")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: Method=%s, Path=%s", r.Method, r.URL.Path)
		homeHandler(db, w, r)
	})
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/add-blog", func(w http.ResponseWriter, r *http.Request) {
		addBlogHandler(db, w, r)
	})
	http.HandleFunc("/save-blog", func(w http.ResponseWriter, r *http.Request) {
		saveBlogHandler(db, w, r)
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Example posts
	posts, err := loadBlogs(db)

	if err != nil {
		http.Error(w, "Unable to load blogs", http.StatusInternalServerError)
		log.Println("Error loading blogs:", err)
		return
	}

	log.Println("Home handler called")
	renderTemplate(w, "index.html", posts)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// Example single post
	post := Blog{Title: "Example Post", Content: "This is an example blog post content."}

	renderTemplate(w, "post.html", post)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addBlogHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse and render the add_blog.html template
	tmpl, err := template.ParseFiles(filepath.Join("templates", "add_blog.html"))
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		log.Println("Error parsing add_blog.html:", err)
		return
	}
	tmpl.Execute(w, nil)
}

func saveBlogHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	title := r.FormValue("title")
	content := r.FormValue("content")

	// Create a new blog
	newBlog := Blog{Title: title, Content: content}

	// Save the blog
	err := saveBlog(db, newBlog)
	if err != nil {
		http.Error(w, "Error saving blog", http.StatusInternalServerError)
		log.Println("Error saving blog:", err)
		return
	}

	// Redirect to the index page or success page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func saveBlog(db *sql.DB, blog Blog) error {
	var id int
	query := `INSERT INTO blogs (title, content) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(query, blog.Title, blog.Content).Scan(&id)
	return err
}

func loadBlogs(db *sql.DB) ([]Blog, error) {
	// Open the blog data file
	log.Println("we reached here")
	query := `SELECT id, title, content, created_at FROM blogs`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []Blog
	for rows.Next() {
		var blog Blog
		if err := rows.Scan(&blog.ID, &blog.Title, &blog.Content, &blog.CreatedAt); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}

	return blogs, nil
}
