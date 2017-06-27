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
	"spotify-bot/identity/manage"
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

func makeIdentifiedHandler(fn func (client *manage.ConnectedClient) interface{}) func(http.ResponseWriter, *http.Request) interface {} {
	return func(w http.ResponseWriter, r *http.Request) interface {} {
		client := identity.GetClientFromRequest(r)

		if (client == nil) {
			return spotify.Response{ Success: false, Message: "Access denied."}
		}

		return fn(client)
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

func pauseHandler(client *manage.ConnectedClient) interface{} {
	return spotify.GetInstance().Pause(client)
}

func unpauseHandler(client *manage.ConnectedClient) interface{} {
	return spotify.GetInstance().Unpause(client)
}

func queueHandler(w http.ResponseWriter, r *http.Request) interface{} {
	body, err := ioutil.ReadAll(r.Body)

	if (err != nil) {
		fmt.Println("Queue failure: " + err.Error())
	}

	trackInfo := spotify.ThinTrackInfo {}
	err = json.Unmarshal(body, &trackInfo)

	if (len(body) != 0) {
		client := identity.GetClientFromRequest(r);

		return spotify.GetInstance().Enqueue(client, trackInfo)
	}

	return spotify.GetInstance().Queue
}

func statusHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return spotify.GetInstance().GetStatus()
}

func upvoteHandler(client *manage.ConnectedClient) interface{} {
	return spotify.GetInstance().Upvote(client)
}

func downvoteHandler(client *manage.ConnectedClient) interface{} {
	return spotify.GetInstance().Downvote(client)
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

func identityHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return identity.UpsertIdentityFromRequest(r);
}

func clientHandler(client *manage.ConnectedClient) interface{} {
	return identity.GetAllPublicClients();
}

func authHandler(password string) func (w http.ResponseWriter, r *http.Request) interface {} {
	return func(w http.ResponseWriter, r *http.Request) interface {} {
		client := identity.GetClientFromRequest(r)

		if (client == nil) {
			return spotify.Response{Success: false, Message: "Not host."}
		}

		if (len(password) == 0) {
			// No pass mode -- IP must match local IP address
			if (r.RemoteAddr == "127.0.0.1" || strings.HasPrefix(r.RemoteAddr, "[::1]")) {
				// Is local, make host
				return spotify.GetInstance().RegisterHost(client)
			}

			return spotify.Response{Success: false, Message: "Not host."}
		}

		passwordAttempts, exists := r.URL.Query()["pass"]

		if (!exists) {
			return spotify.Response{Success: false, Message: "Not host."}
		}

		passwordAttempt := passwordAttempts[0]

		if (passwordAttempt == password) {
			return spotify.GetInstance().RegisterHost(client)
		}

		return spotify.Response{Success: false, Message: "Not host."}
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
	http.HandleFunc("/search", makeJSONHandler(searchHandler))
	http.HandleFunc("/tracks", makeJSONHandler(tracksHandler))
	http.HandleFunc("/albums", makeJSONHandler(albumsHandler))
	http.HandleFunc("/queue", makeJSONHandler(queueHandler))
	http.HandleFunc("/status", makeJSONHandler(statusHandler))
	http.HandleFunc("/identify", makeJSONHandler(identityHandler))
	http.HandleFunc("/pause", makeJSONHandler(makeIdentifiedHandler(pauseHandler)))
	http.HandleFunc("/unpause", makeJSONHandler(makeIdentifiedHandler(unpauseHandler)))
	http.HandleFunc("/downvote", makeJSONHandler(makeIdentifiedHandler(downvoteHandler)))
	http.HandleFunc("/upvote", makeJSONHandler(makeIdentifiedHandler(upvoteHandler)))
	http.HandleFunc("/clients", makeJSONHandler(makeIdentifiedHandler(clientHandler)))
}

func main() {
	args := parseArgs(os.Args[1:]);

	registerHandlers(args.Password)
	spotify.GetInstance()

	http.ListenAndServe(":8080", nil)
}
