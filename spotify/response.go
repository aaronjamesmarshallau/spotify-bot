package spotify

// Response : The formatted response returned when sending a command to spotify
type Response struct {
	Success bool `json:"success"`
	ResponseStatus int `json:"responseStatus"`
	Message string `json:"message"`
}
