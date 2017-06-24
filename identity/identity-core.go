package identity

import (
    "net/http"
    "spotify-bot/identity/manage"
)

// UpsertResponse represents a response from upserting an identity
type UpsertResponse struct {
    IdentityToken string
    IdentityName string
}

// UpsertIdentity creates or updates an identity
func UpsertIdentity(r *http.Request) UpsertResponse {
    clientTokens, exists := r.Header["X-Client-Token"]
    clientToken := clientTokens[0]
    var client *manage.ConnectedClient

    if (!exists) {
        client = manage.Create()
    } else {
        client = manage.GetClient(clientToken)

        if (client == nil) {
            client = manage.Create()
        }
    }

    return UpsertResponse { IdentityToken: client.ClientToken, IdentityName: client.ClientName }
}
