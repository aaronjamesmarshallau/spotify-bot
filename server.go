package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"html/template"
	"spotify-bot/spotify"
	"strings"
)

var templates = template.Must(template.ParseFiles("html/index.html"))

// makeJSONHandler : Creates a JSON returning handler from a function that returns a generic interface{}
func makeJSONHandler(fn func(http.ResponseWriter, *http.Request) interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result := fn(w, r)

		var jsonBytes []byte
		var err error

		switch result.(type) {
		default:
			jsonBytes, err = json.MarshalIndent(result, "", "    ")
		case string:
			jsonBytes = []byte(result.(string))
		}

		if err != nil {
			jsonBytes = []byte("{ error: 'Failed to parse headers' }")
		}

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, string(jsonBytes))
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
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
	ipAddress := r.RemoteAddr[0:strings.Index(r.RemoteAddr, ":")]
	return spotify.GetInstance().Pause(ipAddress)
}

func unpauseHandler(w http.ResponseWriter, r *http.Request) interface{} {
	ipAddress := r.RemoteAddr[0:strings.Index(r.RemoteAddr, ":")]
	return spotify.GetInstance().Unpause(ipAddress)
}

func playHandler(w http.ResponseWriter, r *http.Request) interface{} {
	instance := spotify.GetInstance()
	ipAddress := r.RemoteAddr[0:strings.Index(r.RemoteAddr, ":")]

	if (instance.Host != ipAddress) {
		return "{}"
	}

	return instance.Play(r.URL.Query().Get("trackId"))
}

func queueHandler(w http.ResponseWriter, r *http.Request) interface{} {
	param := r.URL.Query().Get("trackId")

	if (len(param) != 0) {
		return spotify.GetInstance().Enqueue(param)
	}

	return spotify.GetInstance().Queue
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

func authHandler(pwd string) func(http.ResponseWriter, *http.Request) interface{} {
	return func (w http.ResponseWriter, r *http.Request) interface{} {
		_, containsAuthId := r.URL.Query()["authId"]

		if (!containsAuthId) {
			if (len(spotify.GetInstance().Host) == 0) {
				return "{ \"auth\": false }"
			} else {
				return "{ \"auth\": true }"
			}
		}

		authId := r.URL.Query().Get("authId")

		if (authId != pwd) {
			return "{ \"error\": \"Invalid Auth ID. Access denied.\" }"
		}

		ipAddress := r.RemoteAddr[0:strings.Index(r.RemoteAddr, ":")]

		return spotify.GetInstance().RegisterHost(ipAddress)
	}
}

func registerHandlers(pwd string) {
	// Setup our default file handlers for html and css content
	cssHandler := http.FileServer(http.Dir("html/css"))
	htmlHandler := http.FileServer(http.Dir("html/images"))
	jsHandler := http.FileServer(http.Dir("html/js"))
	http.Handle("/css/", http.StripPrefix("/css/", cssHandler))
	http.Handle("/images/", http.StripPrefix("/images/", htmlHandler))
	http.Handle("/js/", http.StripPrefix("/js/", jsHandler))

	// Setup our custom handler functions
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/auth", makeJSONHandler(authHandler(pwd)))
	http.HandleFunc("/pause", makeJSONHandler(pauseHandler))
	http.HandleFunc("/unpause", makeJSONHandler(unpauseHandler))
	http.HandleFunc("/play", makeJSONHandler(playHandler))
	http.HandleFunc("/queue", makeJSONHandler(queueHandler))
	http.HandleFunc("/status", makeJSONHandler(statusHandler))
	http.HandleFunc("/downvote", makeJSONHandler(downvoteHandler))
	http.HandleFunc("/upvote", makeJSONHandler(upvoteHandler))
}

func main() {
	var password string
	fmt.Println("Please enter a unique ID:")
	fmt.Scanln(&password)

	registerHandlers(password)
	spotify.GetInstance()

	http.ListenAndServe(":8080", nil)
}
