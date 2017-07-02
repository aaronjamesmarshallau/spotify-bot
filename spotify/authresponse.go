package spotify

// AuthenticationResponse represents a response from Spotify's account API
type AuthenticationResponse struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int `json:"expires_in"`
}
