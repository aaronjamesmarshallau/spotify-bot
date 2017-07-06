package main

import (
	"os"
	"encoding/json"
	"fmt"
	"net/http"
	"html/template"
	"spotify-bot/spotify"
	"spotify-bot/identity"
	"spotify-bot/identity/manage"
	"strings"
)


type arguments struct {
	Password string
	Theme string
}

var templates = template.Must(template.ParseFiles("html/index.html"))
var args arguments

func parseArgs(args []string) arguments {
	retVal := arguments { Theme: "default" }

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
		case "--theme":
			retVal.Theme = value
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
			return spotify.Response{ Success: false, ResponseStatus: 3, Message: "Access denied."}
		}

		return fn(client)
	}
}

func makeIdentifiedHandlerWithArgs(fn func (client *manage.ConnectedClient, args map[string][]string) interface{}) func(http.ResponseWriter, *http.Request) interface {} {
	return func(w http.ResponseWriter, r *http.Request) interface {} {
		client := identity.GetClientFromRequest(r)

		if (client == nil) {
			return spotify.Response{ Success: false, ResponseStatus: 3, Message: "Access denied."}
		}

		args := r.URL.Query()

		return fn(client, args)
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

	templates.ExecuteTemplate(w, "index.html", args)
}

func pauseHandler(client *manage.ConnectedClient) interface{} {
	return spotify.GetInstance().Pause(client)
}

func unpauseHandler(client *manage.ConnectedClient) interface{} {
	return spotify.GetInstance().Unpause(client)
}

func queueHandler(client *manage.ConnectedClient, args map[string][]string) interface{} {
	trackID, exists := args["trackId"]

	if (exists) {
		return spotify.GetInstance().Enqueue(client, trackID[0])
	}

	return spotify.GetInstance().Queue
}

func statusHandler(client *manage.ConnectedClient) interface{} {
	return spotify.GetInstance().GetPublicStatus()
}

func upvoteHandler(client *manage.ConnectedClient, args map[string][]string) interface{} {
	return spotify.GetInstance().Upvote(client)
}

func downvoteHandler(client *manage.ConnectedClient, args map[string][]string) interface{} {
	return spotify.GetInstance().Downvote(client)
}

func searchHandler(client *manage.ConnectedClient, args map[string][]string) interface{} {
	return spotify.GetInstance().Search(args["q"][0])
}

func clientHandler(client *manage.ConnectedClient) interface{} {
	return identity.GetAllPublicClients();
}

func identityHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return identity.UpsertIdentityFromRequest(r);
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
	http.HandleFunc("/identify", makeJSONHandler(identityHandler))
	http.HandleFunc("/status", makeJSONHandler(makeIdentifiedHandler(statusHandler)))
	http.HandleFunc("/pause", makeJSONHandler(makeIdentifiedHandler(pauseHandler)))
	http.HandleFunc("/clients", makeJSONHandler(makeIdentifiedHandler(clientHandler)))
	http.HandleFunc("/unpause", makeJSONHandler(makeIdentifiedHandler(unpauseHandler)))
	http.HandleFunc("/search", makeJSONHandler(makeIdentifiedHandlerWithArgs(searchHandler)))
	http.HandleFunc("/queue", makeJSONHandler(makeIdentifiedHandlerWithArgs(queueHandler)))
	http.HandleFunc("/downvote", makeJSONHandler(makeIdentifiedHandlerWithArgs(downvoteHandler)))
	http.HandleFunc("/upvote", makeJSONHandler(makeIdentifiedHandlerWithArgs(upvoteHandler)))
}

func main() {
	args = parseArgs(os.Args[1:]);

	registerHandlers(args.Password)
	spotify.GetInstance()

	http.ListenAndServe(":8080", nil)
}
