package spotify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"strings"
	"time"
	"encoding/base64"
)

// Controller : The controller of spotify.
type controller struct {
	Host string
	NowPlaying string
	Queue []string
	VoterList map[string]bool
	CsrfID string
	OAuthID string
	CurrentStatus Status
	CurrentUpvotes int
	CurrentDownvotes int
	UserPaused bool
	AuthToken string
}

const API_BASE string = "https://api.spotify.com/v1"
const SEARCH_ENDPOINT string = "/search?type=track&q="
const ALBUMS_ENDPOINT string = "/albums/"
const TRACKS_ENDPOINT string = "/tracks/"

const ACCOUNTS_BASE string = "https://accounts.spotify.com/api"
const TOKEN_ENDPOINT string = "/token"
const APP_ID string = "107638fbbd0640c4900de32e810816e0"
const APP_SECRET string = "df7a7d40f3b74023a7c429a9f91fad8c"

var instance *controller
var once sync.Once
var quit chan struct{}

func (ctrl *controller) getSpotifyAPIRequest(endpoint string, params string) (*http.Request, error) {
	var URL *url.URL
	urlString := API_BASE + endpoint + params
	URL, err := url.Parse(strings.Replace(urlString, " ", "%20", -1))

	if (err != nil) {
		panic("fuck")
	}

	fmt.Println(URL.String())

	req, err := http.NewRequest("GET", URL.String(), bytes.NewBuffer([]byte("")))

	if (err != nil) {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer " + ctrl.AuthToken)

	return req, nil
}

func getAuthenticationSpotifyRequest(appId, appSecret string) (*http.Request, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	req, err := http.NewRequest("POST", ACCOUNTS_BASE + TOKEN_ENDPOINT, bytes.NewBufferString(data.Encode()))

	if (err != nil) {
		return nil, err
	}

	authCode := base64.StdEncoding.EncodeToString([]byte(appId + ":" + appSecret));

	req.Header.Set("Authorization", "Basic " + authCode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

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

func (spotifyController *controller) getEqualizedVotes() int {
	return spotifyController.CurrentUpvotes - spotifyController.CurrentDownvotes;
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

	if (instance.getEqualizedVotes() >= 3) {
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
		fmt.Println("Starting status loop.")
		quit = make(chan struct{})
		go func() {
			for range time.Tick(time.Duration(2) * time.Second) {
				select {
				case <- quit:
					return;
				default:
					fmt.Println("Running update check.")
					instance.GetStatus()

					instance.updateNowPlaying()

					break;
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
	for ((len(spotifyController.OAuthID) == 0 || len(spotifyController.CsrfID) == 0 || len(spotifyController.AuthToken) == 0) && attempt <= 3) {
		msg := ""
		if attempt == 1 {
			msg = "Initializing Spotify Controller..."
		} else {
			msg = "Initialization failed. Retrying (attempt " + strconv.Itoa(attempt) + ")"
		}

		fmt.Println(msg)
		spotifyController.OAuthID = getOAuthToken()
		spotifyController.CsrfID = getCsrfID()
		authResponse, err := spotifyController.Authenticate()

		if (err == nil) {
			spotifyController.AuthToken = authResponse.AccessToken
			fmt.Println("What" + spotifyController.AuthToken)
		} else {
			fmt.Println("Error: " + err.Error())
		}

		attempt++;
	}

	if attempt > 3 {
		panic("Unable to authenticate after three attempts.")
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

func (spotifyController *controller) Authenticate() (AuthenticationResponse, error) {
	req, err := getAuthenticationSpotifyRequest(APP_ID, APP_SECRET)

	if (err != nil) {
		panic(err)
	}

	client := &http.Client {}
	resp, err := client.Do(req)

	if (err != nil) {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		panic(err)
	}

	outResponse := AuthenticationResponse {}
	err = json.Unmarshal(body, &outResponse)

	return outResponse, err
}

func (spotifyController *controller) GetTrackInfo(trackId string) TrackInfo {
	req, err := spotifyController.getSpotifyAPIRequest(TRACKS_ENDPOINT, trackId)

	if (err != nil) {
		panic(err)
	}

	client := &http.Client {}
	resp, err := client.Do(req)

	if (err != nil) {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		panic(err)
	}

	outResponse := TrackInfo {}
	err = json.Unmarshal(body, &outResponse)

	if (err != nil) {
		panic(err)
	}

	return outResponse
}

func (spotifyController *controller) GetAlbumInfo(albumId string) AlbumInfo {
	req, err := spotifyController.getSpotifyAPIRequest(ALBUMS_ENDPOINT, albumId)

	if (err != nil) {
		panic(err)
	}

	client := &http.Client {}
	resp, err := client.Do(req)

	if (err != nil) {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		panic(err)
	}

	outResponse := AlbumInfo {}
	err = json.Unmarshal(body, &outResponse)

	if (err != nil) {
		panic(err)
	}

	return outResponse
}

func (spotifyController *controller) Search(query string) SearchResults {
	req, err := spotifyController.getSpotifyAPIRequest(SEARCH_ENDPOINT, query)

	if (err != nil) {
		panic(err)
	}

	client := &http.Client {}
	resp, err := client.Do(req)

	if (err != nil) {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		panic(err)
	}

	outResponse := SearchResults {}
	err = json.Unmarshal(body, &outResponse)

	if (err != nil) {
		fmt.Println("Error: " + err.Error() + "; Body: " + string(body[:]))
	}

	return outResponse
}

func (spotifyController *controller) RegisterHost(ipaddress string) Response {
	if (len(spotifyController.Host) == 0) {
		spotifyController.Host = ipaddress
		return Response { Message: "Host registered.", Success: true }
	}

	return Response { Message: "Host already registered.", Success: false}
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
func (spotifyController *controller) Pause(requestingIp string) Response {
	if (spotifyController.Host != requestingIp) {
		return Response { Success: false, Message: "You are not the registered host - you cannot directly control playback. " + spotifyController.Host + " vs " + requestingIp }
	}

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
func (spotifyController *controller) Unpause(requestingIp string) Response {
	if (spotifyController.Host != requestingIp) {
		return Response { Success: false, Message: "You are not the registered host - you cannot directly control playback. " + spotifyController.Host + " vs " + requestingIp }
	}

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
