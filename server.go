package main

import (
	"os"
	"encoding/json"
	"fmt"
	"net/http"
	"html/template"
	"io/ioutil"
	"spotify-bot/spotify"
	"spotify-bot/identity"
	"strings"
)


type arguments struct {
	Password string
}

var templates = template.Must(template.ParseFiles("html/index.html"))

func parseArgs(args []string) arguments {
	retVal := arguments {}

	for _, element := range args {
		parts := strings.Split(element, ":")
		key := parts[0]
		value := ""

		if (len(parts) > 1) {
			value = parts[1]
		}

		switch (key) {
		case "--pass":
			retVal.Password = value
			break
		}
	}

	return retVal
}

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

func queueHandler(w http.ResponseWriter, r *http.Request) interface{} {
	body, err := ioutil.ReadAll(r.Body)

	if (err != nil) {
		fmt.Println("Queue failure: " + err.Error())
	}

	trackInfo := spotify.ThinTrackInfo {}
	err = json.Unmarshal(body, &trackInfo)

	if (len(body) != 0) {
		return spotify.GetInstance().Enqueue(trackInfo)
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

func albumsHandler(w http.ResponseWriter, r *http.Request) interface {} {
	return spotify.GetInstance().GetAlbumInfo(r.URL.Query().Get("id"))
}

func tracksHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().GetTrackInfo(r.URL.Query().Get("id"))
}

func searchHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().Search(r.URL.Query().Get("q"))
}

func authHandler(pwd string) func(http.ResponseWriter, *http.Request) interface{} {
	return func (w http.ResponseWriter, r *http.Request) interface{} {
		_, containsAuthID := r.URL.Query()["authId"]
		ipAddress := r.RemoteAddr[0:strings.Index(r.RemoteAddr, ":")]

		if (len(pwd) != 0) {
			if (!containsAuthID) {
				if (len(spotify.GetInstance().Host) == 0) {
					return "{ \"auth\": false }"
				}

				return "{ \"auth\": true }"
			}
		} else {
			if (!containsAuthID) {
				if (len(spotify.GetInstance().Host) == 0) {
					return spotify.GetInstance().RegisterHost(ipAddress)
				}

				return "{ \"auth\": true }"
			}
		}

		authID := r.URL.Query().Get("authId")

		if (authID != pwd) {
			return "{ \"error\": \"Invalid Auth ID. Access denied.\" }"
		}

		return spotify.GetInstance().RegisterHost(ipAddress)
	}
}

func identityHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return identity.UpsertIdentity(r);
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
	http.HandleFunc("/search", makeJSONHandler(searchHandler))
	http.HandleFunc("/tracks", makeJSONHandler(tracksHandler))
	http.HandleFunc("/albums", makeJSONHandler(albumsHandler))
	http.HandleFunc("/pause", makeJSONHandler(pauseHandler))
	http.HandleFunc("/unpause", makeJSONHandler(unpauseHandler))
	http.HandleFunc("/queue", makeJSONHandler(queueHandler))
	http.HandleFunc("/status", makeJSONHandler(statusHandler))
	http.HandleFunc("/downvote", makeJSONHandler(downvoteHandler))
	http.HandleFunc("/upvote", makeJSONHandler(upvoteHandler))
	http.HandleFunc("/identify", makeJSONHandler(identityHandler))
}

func main() {
	args := parseArgs(os.Args[1:]);

	registerHandlers(args.Password)
	spotify.GetInstance()

	http.ListenAndServe(":8080", nil)
}
