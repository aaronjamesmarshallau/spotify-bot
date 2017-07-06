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

// GetThinTrackInfo transforms TrackInfo into a ThinTrackInfo object
func GetThinTrackInfo(track TrackInfo) *ThinTrackInfo {
	thinTrack := ThinTrackInfo {}
	albumInfo := track.Album
	albumArt := AlbumArtCollection {}

	for i := 0; i < len(track.Album.Images); i++ {
		if (i == 0) {
			albumArt.LargeArt = albumInfo.Images[i].URL
		}

		if (i == 1) {
			albumArt.MediumArt = albumInfo.Images[i].URL
		}

		if (i == 2) {
			albumArt.SmallArt = albumInfo.Images[i].URL
		}
	}

	thinTrack.TrackID = track.ID
	thinTrack.TrackName = track.Name
	thinTrack.ArtistName = track.Artists[0].Name
	thinTrack.AlbumArt = albumArt
	thinTrack.AlbumName = track.Album.Name
	thinTrack.Duration = track.DurationMs / 1000

	return &thinTrack
}
