package spotify

// AlbumArtCollection represents the collection of possible album art
type AlbumArtCollection struct {
	SmallArt string `json:"smallArt"`
	MediumArt string `json:"mediumArt"`
	LargeArt string `json:"largeArt"`
}

// ThinTrackInfo represents the minimal version of a track from Spotify
type ThinTrackInfo struct {
	TrackID string `json:"trackId"`
	TrackName string `json:"trackName"`
	ArtistName string `json:"artistName"`
	AlbumArt AlbumArtCollection `json:"albumArt"`
	AlbumName string `json:"albumName"`
	Duration float64 `json:"duration"`
}
