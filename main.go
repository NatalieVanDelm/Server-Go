package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

type ApiConfigData struct {
	TMDBAccessToken string `json: "TMDBAccessToken"`
}

type Movie struct {
	ID       string `json: "id"`
	Title    string `json: "title"`
	Overview string `json: "overview"`
}

type Response struct {
	Page       int    `json: "page"`
	Results    *Movie `json: "results"`
	TotalPages int    `json: "total_pages"`
}

var movies []Movie

func loadApiConfig(filename string) (ApiConfigData, error) {
	bytes, err := ioutil.ReadFile(filename)

	if err != nil {
		return ApiConfigData{}, err
	}

	var c ApiConfigData

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return ApiConfigData{}, err
	}
	return c, nil
}

func queryTitle(w http.ResponseWriter, r *http.Request) {
	apiConfigData, _ := loadApiConfig(".apiConfig")

	title := r.PathValue("title")

	url := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?query=%s&include_adult=false&language=en-US&page=1", title)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	authorization := fmt.Sprintf("Bearer %s", apiConfigData.TMDBAccessToken)
	req.Header.Add("Authorization", authorization)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))

	var response Response
	json.NewDecoder(res.Body).Decode(&response)
	var movie Movie

	json.NewEncoder(w).Encode(movie)
}

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
	movie.Overview = r.FormValue("overview")
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the HTML template
	tmpl, err := template.ParseFiles("./static/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template with the list of items as data
	err = tmpl.Execute(w, movies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	movies = append(movies, Movie{ID: "1", Title: "Movie1", Overview: "goodgood movie"})

	mux := http.NewServeMux()

	// webserver:
	fileServer := http.FileServer(http.Dir("./"))
	mux.Handle("/static/", fileServer)
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/tmdb/{title}", queryTitle)

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
