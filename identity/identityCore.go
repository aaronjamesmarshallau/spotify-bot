package identity

import (
    "net/http"
    "spotify-bot/identity/manage"
)

// GetAllPublicClients returns all of the publicly visible information about clients
func GetAllPublicClients() []manage.PublicClient {
    return manage.GetAllPublicClients()
}

// UpsertIdentityFromClientToken creates or updates an identity from the provided token
func UpsertIdentityFromClientToken(clientID string, clientSecret string) manage.PrivateClient {
    var client *manage.ConnectedClient

    if (len(clientID) == 0) {
        // if the length of our client id is 0, create a new client
        client = manage.Create()
    } else {
        // if the client exists, get it
        client = manage.GetClient(clientID)

        if (client == nil || client.ClientSecret != clientSecret) {
            // otherwise, if our client is nothing or the secret does not match, create a new client.
            client = manage.Create()
        }
    }

    // Return our less-filtered private info.
    return manage.PrivateClient { IdentitySecret: client.ClientSecret, IdentityToken: client.ClientToken, IdentityName: client.ClientName }
}

// UpsertIdentityFromRequest creates or updates an identity from the provided http.Request
func UpsertIdentityFromRequest(r *http.Request) manage.PrivateClient {
    var clientToken, clientSecret string
    // Try grab info from the request headers
    clientTokens, tokenExists := r.Header["X-Client-Token"]
    clientSecrets, secretExists := r.Header["X-Client-Secret"]

    if (tokenExists && secretExists) {
        // If both exist in the header, set our clientToken and clientSecret
        clientToken = clientTokens[0]
        clientSecret = clientSecrets[0]
    }

    return UpsertIdentityFromClientToken(clientToken, clientSecret)
}

// GetClientFromRequest returns the client it can find from the given request, or nil.
func GetClientFromRequest(r *http.Request) *manage.ConnectedClient {
    var clientToken, clientSecret string
    // Try grab info from the request headers
    clientTokens, tokenExists := r.Header["X-Client-Token"]
    clientSecrets, secretExists := r.Header["X-Client-Secret"]

    if (!tokenExists || !secretExists) {
        return nil
    }

    clientToken = clientTokens[0]
    clientSecret = clientSecrets[0]
    client := manage.GetClient(clientToken)

    if (client == nil || client.ClientSecret != clientSecret) {
        return nil
    }

    return client
}
