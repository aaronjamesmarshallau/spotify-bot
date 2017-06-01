package spotify

type AlbumInfo struct {
	AlbumType string `json:"album_type"`
	Artists []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI string `json:"uri"`
	} `json:"artists"`
	AvailableMarkets []interface{} `json:"available_markets"`
	Copyrights []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"copyrights"`
	ExternalIds struct {
		Upc string `json:"upc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Genres []interface{} `json:"genres"`
	Href string `json:"href"`
	ID string `json:"id"`
	Images []struct {
		Height int `json:"height"`
		URL string `json:"url"`
		Width int `json:"width"`
	} `json:"images"`
	Label string `json:"label"`
	Name string `json:"name"`
	Popularity int `json:"popularity"`
	ReleaseDate string `json:"release_date"`
	ReleaseDatePrecision string `json:"release_date_precision"`
	Tracks struct {
		Href string `json:"href"`
		Items []struct {
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []interface{} `json:"available_markets"`
			DiscNumber int `json:"disc_number"`
			DurationMs int `json:"duration_ms"`
			Explicit bool `json:"explicit"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID string `json:"id"`
			Name string `json:"name"`
			PreviewURL interface{} `json:"preview_url"`
			TrackNumber int `json:"track_number"`
			Type string `json:"type"`
			URI string `json:"uri"`
		} `json:"items"`
		Limit int `json:"limit"`
		Next interface{} `json:"next"`
		Offset int `json:"offset"`
		Previous interface{} `json:"previous"`
		Total int `json:"total"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI string `json:"uri"`
}
