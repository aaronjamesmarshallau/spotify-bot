package spotify

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
	Volume float64 `json:"volume"`
	Online bool `json:"online"`
	OpenGraphState struct {
		PrivateSession bool `json:"private_session"`
		PostingDisabled bool `json:"posting_disabled"`
	} `json:"open_graph_state"`
	Running bool `json:"running"`
	CurrentUpvotes int `json:"currentUpvotes"`
	CurrentDownvotes int `json:"currentDownvotes"`
}
