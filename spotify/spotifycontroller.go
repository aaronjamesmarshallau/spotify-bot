package spotify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Status : The current status of Spotify
type Status struct {
	Version int `json:"version"`
	ClientVersion string `json:"client_version"`
	Playing bool `json:"playing"`
	Shuffle bool `json:"shuffle"`
	Repeat bool `json:"repeat"`
	PlayEnabled bool `json:"play_enabled"`
	PrevEnabled bool `json:"prev_enabled"`
	NextEnabled bool `json:"next_enabled"`
	Track struct {
		TrackResource struct {
			Name string `json:"name"`
			URI string `json:"uri"`
			Location struct {
				Og string `json:"og"`
			} `json:"location"`
		} `json:"track_resource"`
		ArtistResource struct {
			Name string `json:"name"`
			URI string `json:"uri"`
			Location struct {
				Og string `json:"og"`
			} `json:"location"`
		} `json:"artist_resource"`
		AlbumResource struct {
			Name string `json:"name"`
			URI string `json:"uri"`
			Location struct {
				Og string `json:"og"`
			} `json:"location"`
		} `json:"album_resource"`
		Length int `json:"length"`
		TrackType string `json:"track_type"`
	} `json:"track"`
	Context struct {
	} `json:"context"`
	PlayingPosition float64 `json:"playing_position"`
	ServerTime int `json:"server_time"`
	Volume int `json:"volume"`
	Online bool `json:"online"`
	OpenGraphState struct {
		PrivateSession bool `json:"private_session"`
		PostingDisabled bool `json:"posting_disabled"`
	} `json:"open_graph_state"`
	Running bool `json:"running"`
	CurrentUpvotes int `json:"currentUpvotes"`
	CurrentDownvotes int `json:"currentDownvotes"`
}

// Response : The formatted response returned when sending a command to spotify
type Response struct {
	Success bool
	Message string
}

// Controller : The controller of spotify.
type controller struct {
	NowPlaying string
	Queue []string
	VoterList map[string]bool
	CsrfID string
	OAuthID string
	CurrentStatus Status
	CurrentUpvotes int
	CurrentDownvotes int
	UserPaused bool
}

var instance *controller
var once sync.Once
var quit chan struct{}

func getSpotifyRequest(endpoint string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://bruhruhurh.spotilocal.com:4371" + endpoint, bytes.NewBuffer([]byte("")))

	if (err != nil) {
		return nil, err
	}

	req.Header.Set("Origin", "https://embed.spotify.com")
	req.Header.Set("Referer", "https://embed.spotify.com/?uri=spotify:track:5Zp4SWOpbuOdnsxLqwgutt")

	return req, nil
}

func getAuthenticatedSpotifyRequest(endpoint string) (*http.Request, error) {
	req, err := getSpotifyRequest(endpoint)

	if (err != nil) {
		return nil, err
	}

	query := req.URL.Query();

	query.Add("oauth", instance.OAuthID)
	query.Add("csrf", instance.CsrfID)

	req.URL.RawQuery = query.Encode()

	return req, nil
}

func getOAuthToken() string {
	url := "https://open.spotify.com/token"
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))

	if (err != nil) {
		return ""
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if (err != nil) {
		return ""
	}

    body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		return ""
	}

	var data map[string]interface{}
   	err = json.Unmarshal(body, &data)

	if (err != nil) {
		return ""
	}

	return data["t"].(string)
}

func getCsrfID() string {
	req, err := getSpotifyRequest("/simplecsrf/token.json");

	if (err != nil) {
		return ""
	}

	query := req.URL.Query()
	unixTime := strconv.FormatInt(time.Now().Unix(), 10)

	query.Add("ref", "")
	query.Add("cors", "")
	query.Add("_", unixTime)
	query.Add("oauth", instance.OAuthID)

	req.URL.RawQuery = query.Encode()

	tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }

	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Do(req)

	if (err != nil) {
		fmt.Println("Error:", err.Error())
		return ""
	}

    body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		return ""
	}

	var data map[string]interface{}
   	err = json.Unmarshal(body, &data)

	if (err != nil) {
		return ""
	}

	return data["token"].(string)
}

func (spotifyController *controller) playNext() {
	if (len(spotifyController.Queue) > 0) {
		nextTrack := spotifyController.Queue[0]
		spotifyController.Queue = spotifyController.Queue[1:]

		spotifyController.Play(nextTrack)
	}
}

// updateNowPlaying : Updates the currently playing track, if needed.
func (spotifyController *controller) updateNowPlaying() {
	status := spotifyController.CurrentStatus
	maxPlaypositionThreshold := status.PlayingPosition + 2

	if (instance.CurrentDownvotes >= 3) {
		spotifyController.playNext();
	}

	if !status.Playing && !instance.UserPaused {
		spotifyController.playNext()
	}

	if float64(status.Track.Length) <= maxPlaypositionThreshold {
		playTimeLeft := float64(status.Track.Length) - status.PlayingPosition

		time.AfterFunc(time.Duration(playTimeLeft) * time.Second, func () {
			spotifyController.playNext()
		})
	}
}

func startStatusLoop() {
	if (quit == nil) {
		ticker := time.NewTicker(2 * time.Second)
		quit = make(chan struct{})
		go func() {
		    for {
		       select {
		        case <- ticker.C:
					fmt.Println("Running update check.")
		            instance.GetStatus()

					instance.updateNowPlaying()
		        case <- quit:
		            ticker.Stop()
		            return
		        }
		    }
		 }()
	 }
}

func stopStatusLoop() {
	if (quit != nil) {
		close(quit)
	}
}

// setup : Sets up the Spotify Controller ready to make requests.
func (spotifyController *controller) setup() {
	attempt := 1
	for ((len(spotifyController.OAuthID) == 0 || len(spotifyController.CsrfID) == 0) && attempt <= 3) {
		msg := ""
		if attempt == 1 {
			msg = "Initializing Spotify Controller..."
		} else {
			msg = "Initialization failed. Retrying (attempt " + strconv.Itoa(attempt) + ")"
		}

		fmt.Println(msg)
		spotifyController.OAuthID = getOAuthToken()
		spotifyController.CsrfID = getCsrfID()

		attempt++;
	}

	if attempt > 3 {
		panic("Unable to authenticated after three attempts.")
	}

	fmt.Println("Controller initialized successfully:")
	fmt.Println("\tOAuth:", spotifyController.OAuthID)
	fmt.Println("\tCsrf:", spotifyController.CsrfID)

	startStatusLoop()
}

// GetInstance : Gets the instance of the spotify controller
func GetInstance() *controller {
	once.Do(func () {
		instance = &controller { CurrentDownvotes: 0, CurrentUpvotes: 0, VoterList: make(map[string]bool) }
		instance.setup()
	})

	return instance
}

func getJSON(endpoint string) ([]byte, error) {
	req, err := getAuthenticatedSpotifyRequest(endpoint)

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if (err != nil) {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		return nil, err
	}

	return body, nil
}

// Play : Plays the given track immediately.
func (spotifyController *controller) Play(trackID string) Response {
	body, err := getJSON("/remote/play.json?uri=spotify:track:" + trackID)

	if (err != nil) {
		return Response { Success: false, Message: "An error has occurred: " + err.Error() }
	}

	err = json.Unmarshal(body, &instance.CurrentStatus)

	if (err != nil) {
		return Response { Success: false, Message: "Unable to parse response." }
	}

	spotifyController.NowPlaying = trackID
	spotifyController.CurrentUpvotes = 0
	spotifyController.CurrentDownvotes = 0
	spotifyController.VoterList = make(map[string]bool)

	return Response { Success: true, Message: "Request made. Response: " + string(body) }
}

// Pause : Pauses the currently playing track
func (spotifyController *controller) Pause() Response {
	body, err := getJSON("/remote/pause.json?pause=true")

	if (err != nil) {
		return Response { Success: false, Message: "An error has occurred: " + err.Error() }
	}

	err = json.Unmarshal(body, &instance.CurrentStatus)

	if (err != nil) {
		return Response { Success: false, Message: "Unable to parse response." }
	}

	instance.UserPaused = true

	return Response { Success: true, Message: "Request made. Response: " + string(body) }
}

// Unpause : Unpauses the currently playing track
func (spotifyController *controller) Unpause() Response {
	body, err := getJSON("/remote/pause.json?pause=false")

	if (err != nil) {
		return Response { Success: false, Message: "An error has occurred: " + err.Error() }
	}

	err = json.Unmarshal(body, &instance.CurrentStatus)

	if (err != nil) {
		return Response { Success: false, Message: "Unable to parse response." }
	}

	instance.UserPaused = false

	return Response { Success: true, Message: "Request made. Response: " + string(body) }
}

func (spotifyController *controller) Enqueue(trackID string) Response {
	spotifyController.Queue = append(spotifyController.Queue, trackID)

	return Response { Success: true, Message: "Track queued." }
}

func (spotifyController *controller) Upvote(ip string) Response {
	if (!spotifyController.VoterList[ip]) {
		spotifyController.CurrentUpvotes++
		spotifyController.VoterList[ip] = true

		return Response { Success: true, Message: "Current downvotes: " + strconv.Itoa(spotifyController.CurrentUpvotes) }
	}

	return Response { Success: false, Message: "Already voted." }
}

func (spotifyController *controller) Downvote(ip string) Response {
	if (!spotifyController.VoterList[ip]) {
		spotifyController.CurrentDownvotes++
		spotifyController.VoterList[ip] = true

		return Response { Success: true, Message: "Current downvotes: " + strconv.Itoa(spotifyController.CurrentDownvotes) }
	}

	return Response { Success: false, Message: "Already voted." }
}

// GetStatus : Gets the current status of the spotify player.
func (spotifyController *controller) GetStatus() Status {
	body, err := getJSON("/remote/status.json")

	if (err != nil) {
		instance.CurrentStatus.CurrentUpvotes = instance.CurrentUpvotes
		instance.CurrentStatus.CurrentDownvotes = instance.CurrentDownvotes
		return instance.CurrentStatus
	}

	err = json.Unmarshal(body, &instance.CurrentStatus)

	if (err != nil) {
		instance.CurrentStatus.CurrentUpvotes = instance.CurrentUpvotes
		instance.CurrentStatus.CurrentDownvotes = instance.CurrentDownvotes
		return instance.CurrentStatus
	}

	instance.CurrentStatus.CurrentUpvotes = instance.CurrentUpvotes
	instance.CurrentStatus.CurrentDownvotes = instance.CurrentDownvotes

	return instance.CurrentStatus
}
