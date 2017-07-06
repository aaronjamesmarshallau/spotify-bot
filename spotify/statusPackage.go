package spotify

// StatusPackage is the client-facing status of Spotify
type StatusPackage struct {
	Playing bool `json:"playing"`
	NowPlaying ThinTrackInfo `json:"nowPlaying"`
	PlayPosition float64 `json:"playPosition"`
	CurrentUpvotes int `json:"currentUpvotes"`
	CurrentDownvotes int `json:"currentDownvotes"`
}

// ApplySpotifyStatus applies the Spotify status to the client-facing status
func (clientStatus *StatusPackage) ApplySpotifyStatus(spotifyStatus Status, currentTrackAlbumArt AlbumArtCollection) {
    ctrl := GetInstance()

    clientStatus.Playing = spotifyStatus.Playing
    clientStatus.NowPlaying = *ctrl.NowPlaying
    clientStatus.PlayPosition = spotifyStatus.PlayingPosition
    clientStatus.CurrentUpvotes = ctrl.CurrentUpvotes
    clientStatus.CurrentDownvotes = ctrl.CurrentDownvotes
}
