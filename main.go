package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

type Movie struct {
	ID       string    `json: "id"`
	Title    string    `json: "title"`
	Director *Director `json: "director"`
}

type Director struct {
	Firstname string `json: "firstname"`
	Lastname  string `json: "lastname"`
}

var movies []Movie

func getMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	for _, item := range movies {
		if item.ID == id {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	for index, item := range movies {
		if item.ID == id {
			movies = append(movies[:index], movies[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(movies)
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var movie Movie
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
	}
	fmt.Fprint(w, "POST request succesful\n")
	movie.ID = strconv.Itoa(rand.Intn(100000000))
	movie.Title = r.FormValue("title")
	movie.Director = &Director{Firstname: r.FormValue("firstname"), Lastname: r.FormValue("lastname")}
	movies = append(movies, movie)
	json.NewEncoder(w).Encode(movies)
}

func updateMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	for index, item := range movies {
		if item.ID == id {
			movies = append(movies[:index], movies[index+1:]...)
			var movie Movie
			_ = json.NewDecoder(r.Body).Decode(&movie)
			movie.ID = id
			movies = append(movies[:index], movie)
			break
		}
	}
	json.NewEncoder(w).Encode(movies)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "method is not supported", http.StatusNotFound)
	}
	fmt.Fprint(w, "hello!")
}

func main() {
	movies = append(movies, Movie{ID: "1", Title: "Movie1", Director: &Director{Firstname: "John", Lastname: "Doe"}})
	movies = append(movies, Movie{ID: "2", Title: "Movie2", Director: &Director{Firstname: "Jane", Lastname: "Smith"}})

	mux := http.NewServeMux()

	// webserver:
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fileServer)
	mux.HandleFunc("/hello", helloHandler)

	// crud api:
	mux.HandleFunc("GET /movies", getMovies)
	mux.HandleFunc("GET /movies/{id}", getMovie)
	mux.HandleFunc("POST /movies", createMovie)
	mux.HandleFunc("PUT /movies/{id}", updateMovie)
	mux.HandleFunc("DELETE /movies/{id}", deleteMovie)

	fmt.Printf("Starting server at port 8080\n")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
