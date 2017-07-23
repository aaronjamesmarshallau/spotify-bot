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
	"spotify-bot/identity/manage"
)

// Controller : The controller of spotify.
type controller struct {
	Host *manage.ConnectedClient
	NowPlaying *ThinTrackInfo
	Queue []*ThinTrackInfo
	VoterList map[string]bool
	CsrfID string
	OAuthID string
	CurrentStatus StatusPackage
	CurrentUpvotes int
	CurrentDownvotes int
	UserPaused bool
	AuthToken string
	ChangingSong bool
}

// APIBase is the base URL for the spotify API
const APIBase string = "https://api.spotify.com/v1"

// SearchEndpoint is the endpoint for searching spotify
const SearchEndpoint string = "/search?type=track&q="

// AlbumsEndpoint is the endpoint for getting album info
const AlbumsEndpoint string = "/albums/"

// TracksEndpoint is the endpoint for getting track info
const TracksEndpoint string = "/tracks/"

// AccountsBase is the base URL for accounts
const AccountsBase string = "https://accounts.spotify.com/api"

// TokenEndpoint is the endpoint for getting a token from the Spotify accounts API
const TokenEndpoint string = "/token"

// AppID is the application ID for this bot
const AppID string = "107638fbbd0640c4900de32e810816e0"

// AppSecret is the application secret for this bot
const AppSecret string = "df7a7d40f3b74023a7c429a9f91fad8c"

var instance *controller
var once sync.Once
var quit chan struct{}

func (ctrl *controller) getSpotifyAPIRequest(endpoint string, params string) (*http.Request, error) {
	var URL *url.URL
	urlString := APIBase + endpoint + params
	URL, err := url.Parse(strings.Replace(urlString, " ", "%20", -1))

	if (err != nil) {
		return nil, err
	}

	req, err := http.NewRequest("GET", URL.String(), bytes.NewBuffer([]byte("")))

	if (err != nil) {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer " + ctrl.AuthToken)

	return req, nil
}

func getAuthenticationSpotifyRequest(appID, appSecret string) (*http.Request, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	req, err := http.NewRequest("POST", AccountsBase + TokenEndpoint, bytes.NewBufferString(data.Encode()))

	if (err != nil) {
		return nil, err
	}

	authCode := base64.StdEncoding.EncodeToString([]byte(appID + ":" + appSecret));

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

func (ctrl *controller) getEqualizedVotes() int {
	return ctrl.CurrentUpvotes - ctrl.CurrentDownvotes;
}

func (ctrl *controller) playNext() {
	if (len(ctrl.Queue) > 0) {
		ctrl.ChangingSong = true;
		nextTrack := ctrl.Queue[0]
		ctrl.Queue = ctrl.Queue[1:]

		ctrl.playImmediately(nextTrack)
		ctrl.ChangingSong = false;
	}
}

// updateNowPlaying updates the currently playing track, if needed.
func (ctrl *controller) updateNowPlaying() {
	if ctrl.ChangingSong {
		return;
	}

	status := ctrl.CurrentStatus
	maxPlaypositionThreshold := status.PlayPosition + 2

	if (instance.getEqualizedVotes() <= -3) {
		ctrl.playNext();
		return;
	}

	if !status.Playing && !instance.UserPaused {
		ctrl.playNext()
		return;
	}

	if float64(status.NowPlaying.Duration) <= maxPlaypositionThreshold {
		playTimeLeft := float64(status.NowPlaying.Duration) - status.PlayPosition

		time.AfterFunc(time.Duration(playTimeLeft) * time.Second, func () {
			ctrl.playNext()
		})
	}
}

func startStatusLoop() {
	if (quit == nil) {
		fmt.Println("Starting status loop.")
		quit = make(chan struct{})
		go func() {
			for range time.Tick(time.Duration(1) * time.Second) {
				select {
				case <- quit:
					return;
				default:
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

// setup sets up the Spotify Controller ready to make requests.
func (ctrl *controller) setup() {
	attempt := 1
	for ((len(ctrl.OAuthID) == 0 || len(ctrl.CsrfID) == 0 || len(ctrl.AuthToken) == 0) && attempt <= 3) {
		msg := ""
		if attempt == 1 {
			msg = "Initializing Spotify Controller..."
		} else {
			msg = "Initialization failed. Retrying (attempt " + strconv.Itoa(attempt) + ")"
		}

		fmt.Println(msg)
		ctrl.OAuthID = getOAuthToken()
		ctrl.CsrfID = getCsrfID()
		err := ctrl.Authenticate()

		if (err != nil) {
			fmt.Println("Error: " + err.Error())
		}

		attempt++;
	}

	if attempt > 3 {
		panic("Unable to authenticate after three attempts.")
	}

	fmt.Println("Controller initialized successfully:")
	fmt.Println("\tOAuth:", ctrl.OAuthID)
	fmt.Println("\tCsrf:", ctrl.CsrfID)

	startStatusLoop()
}

// GetInstance gets the instance of the Spotify controller
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

func (ctrl *controller) Authenticate() error {
	req, err := getAuthenticationSpotifyRequest(AppID, AppSecret)
	outResponse := AuthenticationResponse {}

	if (err != nil) {
		return err
	}

	client := &http.Client {}
	resp, err := client.Do(req)

	if (err != nil) {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		return err
	}

	err = json.Unmarshal(body, &outResponse)

	if (err != nil) {
		return err
	}

	ctrl.AuthToken = outResponse.AccessToken

	return nil
}

func (ctrl *controller) GetTrackInfo(trackID string) TrackInfo {
	req, err := ctrl.getSpotifyAPIRequest(TracksEndpoint, trackID)
		outResponse := TrackInfo {}

	if (err != nil) {
		return outResponse
	}

	client := &http.Client {}
	resp, err := client.Do(req)

	if (err != nil) {
		ctrl.Authenticate()
		return ctrl.GetTrackInfo(trackID)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		return outResponse
	}

	err = json.Unmarshal(body, &outResponse)

	if (err != nil) {
		return outResponse
	}

	return outResponse
}

func (ctrl *controller) GetAlbumInfo(albumID string) AlbumInfo {
	req, err := ctrl.getSpotifyAPIRequest(AlbumsEndpoint, albumID)
	outResponse := AlbumInfo {}

	if (err != nil) {
		return outResponse
	}

	client := &http.Client {}
	resp, err := client.Do(req)

	if (err != nil) {
		ctrl.Authenticate()
		return ctrl.GetAlbumInfo(albumID)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		return outResponse
	}

	err = json.Unmarshal(body, &outResponse)

	if (err != nil) {
		return outResponse
	}

	return outResponse
}

func (ctrl *controller) Search(query string) SearchResults {
	req, err := ctrl.getSpotifyAPIRequest(SearchEndpoint, query)
		outResponse := SearchResults {}

	if (err != nil) {
		return outResponse
	}

	client := &http.Client {}
	resp, err := client.Do(req)

	if (err != nil) {
		ctrl.Authenticate()
		return ctrl.Search(query)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		return outResponse
	}

	err = json.Unmarshal(body, &outResponse)

	if (err != nil) {
		fmt.Println("Error: " + err.Error() + "; Body: " + string(body[:]))
	}

	return outResponse
}

func (ctrl *controller) playImmediately(track *ThinTrackInfo) Response {
	body, err := getJSON("/remote/play.json?uri=spotify:track:" + track.TrackID)

	if (err != nil) {
		return Response { Success: false, ResponseStatus: 8, Message: "An error has occurred: " + err.Error() }
	}

	err = json.Unmarshal(body, &instance.CurrentStatus)

	if (err != nil) {
		return Response { Success: false, ResponseStatus: 8, Message: "Unable to parse response." }
	}

	ctrl.NowPlaying = track
	ctrl.CurrentUpvotes = 0
	ctrl.CurrentDownvotes = 0
	ctrl.VoterList = make(map[string]bool)

	fmt.Println("Playing next song: " + track.ArtistName + " - " + track.TrackName)

	return Response { Success: true, ResponseStatus: 1, Message: "Request made. Response: " + string(body) }
}

// Play : Plays the given track immediately.
func (ctrl *controller) Play(client *manage.ConnectedClient, track *ThinTrackInfo) Response {
	if (client == nil) {
		return Response { Success: false, ResponseStatus: 8, Message: "Access denied." }
	}

	return ctrl.playImmediately(track)
}

// Pause : Pauses the currently playing track
func (ctrl *controller) Pause(client *manage.ConnectedClient) Response {
	if (client == nil || ctrl.Host == nil) {
		return Response { Success: false, ResponseStatus: 8, Message: "The host has not been set." }
	}

	if (ctrl.Host.ClientToken != client.ClientToken) {
		return Response { Success: false, ResponseStatus: 8, Message: "You are not the registered host - you cannot directly control playback. " + ctrl.Host.ClientToken + " vs " + client.ClientToken }
	}

	body, err := getJSON("/remote/pause.json?pause=true")

	if (err != nil) {
		return Response { Success: false, ResponseStatus: 8, Message: "An error has occurred: " + err.Error() }
	}

	err = json.Unmarshal(body, &instance.CurrentStatus)

	if (err != nil) {
		return Response { Success: false, ResponseStatus: 8, Message: "Unable to parse response." }
	}

	instance.UserPaused = true

	return Response { Success: true, ResponseStatus: 1, Message: "Request made. Response: " + string(body) }
}

// Unpause : Unpauses the currently playing track
func (ctrl *controller) Unpause(client *manage.ConnectedClient) Response {
	if (client == nil || ctrl.Host.ClientToken != client.ClientToken) {
		return Response { Success: false, ResponseStatus: 8, Message: "You are not the registered host - you cannot directly control playback. " + ctrl.Host.ClientToken + " vs " + client.ClientToken }
	}

	body, err := getJSON("/remote/pause.json?pause=false")

	if (err != nil) {
		return Response { Success: false, ResponseStatus: 8, Message: "An error has occurred: " + err.Error() }
	}

	err = json.Unmarshal(body, &instance.CurrentStatus)

	if (err != nil) {
		return Response { Success: false, ResponseStatus: 8, Message: "Unable to parse response." }
	}

	instance.UserPaused = false

	return Response { Success: true, ResponseStatus: 1, Message: "Request made. Response: " + string(body) }
}

// Enqueue queues up the provided song
func (ctrl *controller) Enqueue(client *manage.ConnectedClient, trackID string) Response {
	lowerBound := time.Now().Add(time.Duration(-3) * time.Minute)
	songsAfter := 0

	if (client == nil) {
		return Response { Success: false, ResponseStatus: 3, Message: "Cannot queue track -- your client is not registered." }
	}

	for i := 0; i < len(client.QueueHistory); i++ {
		queueEntry := client.QueueHistory[i];

		if (queueEntry.QueueTimestamp.After(lowerBound)) {
			songsAfter++
		}
	}

	canQueue := songsAfter < 3

	if (canQueue) {
		track := ctrl.GetTrackInfo(trackID)
		thinTrack := GetThinTrackInfo(track)

		fmt.Println("Client '" + client.ClientName + "' queued song '" + thinTrack.TrackName + "' (" + thinTrack.TrackID + ") \nTracks queued in last 3 minutes: " + strconv.Itoa(songsAfter));

		ctrl.Queue = append(ctrl.Queue, thinTrack)
		client.QueueHistory = append(client.QueueHistory, manage.QueueEntry {TrackID: thinTrack.TrackID, TrackName: thinTrack.TrackName, QueueTimestamp: time.Now() })

		return Response { Success: true, ResponseStatus: 1, Message: "Track queued." }
	}

	return Response { Success: false, ResponseStatus: 8, Message: "Cannot queue track -- you have reached the queue limit of 3 songs in 3 minutes" }
}

func (ctrl *controller) GetRemainingTime() float64 {
	return ctrl.CurrentStatus.NowPlaying.Duration - ctrl.CurrentStatus.PlayPosition
}

func (ctrl *controller) Upvote(client *manage.ConnectedClient) Response {
	lowerBound := client.ConnectionTime.Add(time.Duration(ctrl.GetRemainingTime()) * time.Second)
	connectedLongEnough := time.Now().After(lowerBound)
	alreadyVoted := ctrl.VoterList[client.ClientToken]

	if (connectedLongEnough && !alreadyVoted) {
		ctrl.CurrentUpvotes++
		ctrl.VoterList[client.ClientToken] = true
		client.VoteHistory = append(client.VoteHistory, manage.Vote { TrackID: ctrl.NowPlaying.TrackID, TrackName: ctrl.NowPlaying.TrackName, Upvoted: true, TimeVoted: time.Now() });

		return Response { Success: true, ResponseStatus: 1, Message: "Current downvotes: " + strconv.Itoa(ctrl.CurrentUpvotes) }
	}

	if (!connectedLongEnough) {
		return Response { Success: false, ResponseStatus: 8, Message: "Unable to vote during the first song." }
	}

	return Response { Success: false, ResponseStatus: 3, Message: "Unable to vote." }
}

func (ctrl *controller) Downvote(client *manage.ConnectedClient) Response {
	lowerBound := client.ConnectionTime.Add(time.Duration(ctrl.GetRemainingTime()) * time.Second)
	connectedLongEnough := time.Now().After(lowerBound)
	alreadyVoted := ctrl.VoterList[client.ClientToken]

	if (connectedLongEnough && !alreadyVoted) {
		ctrl.CurrentDownvotes++
		ctrl.VoterList[client.ClientToken] = true
		client.VoteHistory = append(client.VoteHistory, manage.Vote { TrackID: ctrl.NowPlaying.TrackID, TrackName: ctrl.NowPlaying.TrackName, Upvoted: true, TimeVoted: time.Now() });

		return Response { Success: true, ResponseStatus: 1, Message: "Current downvotes: " + strconv.Itoa(ctrl.CurrentDownvotes) }
	}

	if (!connectedLongEnough) {
		return Response { Success: false, ResponseStatus: 8, Message: "Unable to vote during the first song." }
	}

	return Response { Success: false, ResponseStatus: 3, Message: "Unable to vote." }
}

// GetStatus gets the current status of the Spotify player
func (ctrl *controller) GetStatus() {
	body, err := getJSON("/remote/status.json")
	spotifyStatus := Status {}
	currentTrackAlbumArt := AlbumArtCollection {}

	if (err != nil) {
		instance.CurrentStatus.CurrentUpvotes = instance.CurrentUpvotes
		instance.CurrentStatus.CurrentDownvotes = instance.CurrentDownvotes
		return
	}

	err = json.Unmarshal(body, &spotifyStatus)

	if (err != nil) {
		instance.CurrentStatus.CurrentUpvotes = instance.CurrentUpvotes
		instance.CurrentStatus.CurrentDownvotes = instance.CurrentDownvotes
		return
	}

	newTrackID := strings.Replace(spotifyStatus.Track.TrackResource.URI, "spotify:track:", "", -1)

	if (instance.NowPlaying == nil) {
		instance.NowPlaying = &ThinTrackInfo {}
	}

	if instance.NowPlaying.TrackID != newTrackID {
		trackInfo := ctrl.GetTrackInfo(newTrackID)

		fmt.Println("Don't match (new song): " + instance.CurrentStatus.NowPlaying.TrackID + " vs " + newTrackID)

		ctrl.NowPlaying = GetThinTrackInfo(trackInfo)
	}

	instance.CurrentStatus.ApplySpotifyStatus(spotifyStatus, currentTrackAlbumArt)
}

func (ctrl *controller) GetPublicStatus() StatusPackage {
	return instance.CurrentStatus
}

// RegisterHost registers the provided client as the host
func (ctrl *controller) RegisterHost(client *manage.ConnectedClient) Response {
	ctrl.Host = client;
	return Response {Success: true, ResponseStatus: 1, Message: "You have been successfully registered as the host."}
}
