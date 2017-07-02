package spotify

// ThinTrackInfo represents the minimal version of a track from Spotify
type ThinTrackInfo struct {
	TrackID string `json:"trackId"`
	TrackName string `json:"trackName"`
	ArtistName string `json:"artistName"`
	AlbumArt string `json:"albumArt"`
	AlbumName string `json:"albumName"`
}
