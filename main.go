package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"database/sql"
	_ "github.com/lib/pq"
)

var router *chi.Mux
var db *sql.DB

// Post type details/summary model
type PostSummary struct {
	ID int
	Title string
	Content string
	// created_at time.Time 'json:"created_at"'
}

type Posts struct {
	Posts []PostSummary
}

func init() {
	host := os.Getenv("GOTESTDB_HOST")
  port := os.Getenv("GOTESTDB_PORT")
  user := os.Getenv("GOTESTDB_USER")
  password := os.Getenv("GOTESTDB_PASS")
  dbname := os.Getenv("GOTESTDB_NAME")

	router = chi.NewRouter()
	router.Use(middleware.Logger)
  router.Use(middleware.Recoverer)

  psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
          host, port, user, password, dbname)

  localdb, err := sql.Open("postgres", psqlInfo)
  if err != nil {
      panic(err)
  }
	db = localdb

  // err = db.Ping()
  // if err != nil {
  //     panic(err)
  // }

  fmt.Println("Successfully connected to db!")
}

func routers() *chi.Mux {
	router.Get("/posts", AllPosts)
	router.Get("/posts/{id}", DetailPost)
	router.Post("/posts", CreatePost)
	router.Put("/posts/{id}", UpdatePost)
	router.Delete("/posts/{id}", DeletePost)
	router.Get("/posts/insert/{title}/{cont}", InsertPost)

	return router
}

// server starting point
func ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(map[string]string{"message": "Pong"})
	// respondWithJSON(w, http.StatusOK, map[string]string{"message": "Pong"})
}

//--------- API ENDPOINT ------------//

// AllPosts returns all post object data
func AllPosts(w http.ResponseWriter, r *http.Request) {
	postsdict := Posts{}

	rows, err := db.Query("SELECT * FROM post")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		post := PostSummary{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content)
		if err != nil {
			panic(err)
		}
		formattedTitle := strings.TrimRight(post.Title, " ")
		post.Title = formattedTitle
		postsdict.Posts = append(postsdict.Posts, post)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	fmt.Println(postsdict.Posts)

	defer rows.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewDecoder(r.Body)
	// json.NewEncoder(w).Encode(map[string]string{"message": "Need to get all posts"})
	json.NewEncoder(w).Encode(postsdict.Posts)
}

func InsertPost(w http.ResponseWriter, r *http.Request) {
	title := chi.URLParam(r, "title")
	description := chi.URLParam(r, "cont")
	query, err := db.Query("INSERT INTO post(id, title, content) VALUES ($1, $2, $3)",
		rand.Intn(100), title, description)
	if err != nil {
		panic(err)
	}

	defer query.Close()
	// json.NewDecoder(r.Body)
	// json.NewEncoder(w).Encode(map[string]string{"message": "Need to get all posts"})
	json.NewEncoder(w).Encode(query)
}

func DetailPost(w http.ResponseWriter, r *http.Request) {
	// payload := Post{}
	id := chi.URLParam(r, "id")
	fmt.Println(id)
	// row := db.QueryRow
  defer db.Close()
	json.NewEncoder(w).Encode(map[string]string{"message": "DB closed"})
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
/***
CreatePost is an API endpoint that lets you create a new "post" object.
***/
	var post PostSummary
	// json.NewDecoder catches all the POST data from parameter r http request.
	json.NewDecoder(r.Body).Decode(&post)

	// Set dynamic query
	query, err := db.Prepare("INSERT posts SET title=$1, content=$2")
	if err != nil {
		panic(err)
	}

	_, er := query.Exec(post.Title, post.Content)
	if err != nil {
		panic(er)
	}
	defer query.Close()

	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(map[string]string{"message": "Created Post object successfully."})
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var post PostSummary
	id := chi.URLParam(r, "id")
	json.NewDecoder(r.Body).Decode(&post)

	query, err := db.Prepare("UPDATE posts SET title=$1, content=$2 WHERE id=$3")
	if err != nil {
		panic(err)
	}
	_, er := query.Exec(post.Title, post.Content, id)
	if err != nil {
		panic(er)
	}
	defer query.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Updated Post object successfully."})
	// respondwithJSON(w, http.StatusOK, map[string]string{"message": "Updated Post object successfully."})
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	query, err := db.Prepare("DELETE FROM posts WHERE id=$1")
	if err != nil {
		panic(err)
	}
	_, er := query.Exec(id)
	if err != nil {
		panic(er)
	}
	query.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Deleted Post object successfully."})
		// respondWithJSON(w, http.StatusOK, map[string]string{"message": "Deleted Post object successfully."})
}

// Run routes and start server
func main() {
	routers()
	http.ListenAndServe(":8080", router)
}
