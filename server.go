package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"html/template"
	"spotify-server/spotify"
)

var templates = template.Must(template.ParseFiles("html/index.html"))

// makeJSONHandler : Creates a JSON returning handler from a function that returns a generic interface{}
func makeJSONHandler(fn func(http.ResponseWriter, *http.Request) interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result := fn(w, r)
		jsonBytes, err := json.MarshalIndent(result, "", "    ")

		if err != nil {
			jsonBytes = []byte("{ error: 'Failed to parse headers' }")
		}

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, string(jsonBytes))
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Error: 404 Not Found")
}

// rootHandler : The handler for the root directory
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if (r.URL.Path != "/") {
		errorHandler(w, r)
		return
	}

	templates.ExecuteTemplate(w, "index.html", nil)
}

func pauseHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().Pause()
}

func unpauseHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().Unpause()
}

func playHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().Play(r.URL.Query().Get("trackId"))
}

func queueHandler(w http.ResponseWriter, r *http.Request) interface{} {
	param := r.URL.Query().Get("trackId")

	if (len(param) != 0) {
		return spotify.GetInstance().Enqueue(param)
	}

	result, err := json.Marshal(spotify.GetInstance().Queue)

	if (err != nil) {
		return "{ error: \"An error occurred while serializing the play queue.\" }"
	}

	return result
}

func statusHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().GetStatus()
}

func upvoteHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().Upvote(r.RemoteAddr)
}

func downvoteHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().Downvote(r.RemoteAddr)
}

func registerHandlers() {
	// Setup our default file handlers for html and css content
	cssHandler := http.FileServer(http.Dir("html/css"))
	htmlHandler := http.FileServer(http.Dir("html/images"))
	jsHandler := http.FileServer(http.Dir("html/js"))
	http.Handle("/css/", http.StripPrefix("/css/", cssHandler))
	http.Handle("/images/", http.StripPrefix("/images/", htmlHandler))
	http.Handle("/js/", http.StripPrefix("/js/", jsHandler))

	// Setup our custom handler functions
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/pause", makeJSONHandler(pauseHandler))
	http.HandleFunc("/unpause", makeJSONHandler(unpauseHandler))
	http.HandleFunc("/play", makeJSONHandler(playHandler))
	http.HandleFunc("/queue", makeJSONHandler(queueHandler))
	http.HandleFunc("/status", makeJSONHandler(statusHandler))
	http.HandleFunc("/downvote", makeJSONHandler(downvoteHandler))
	http.HandleFunc("/upvote", makeJSONHandler(upvoteHandler))
}

func main() {
	registerHandlers()
	spotify.GetInstance()

	http.ListenAndServe(":8080", nil)
}
