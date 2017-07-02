package manage

import (
    "math/rand"
    "time"
)

var currentClients map[string]*ConnectedClient = make(map[string]*ConnectedClient)
var potentialClientNames = []string {
    "Platypus",
    "Echidna",
    "Wombat",
    "Kangaroo",
    "Shark",
    "Snake",
    "Dolphin",
    "Crocodile",
    "Possum",
    "Emu",
    "Wallaby",
    "Koala",
    "Dingo",
    "Quokka",
    "Kookaburra"}

func generateRandom(chars string, length int) string {
    myCode := ""

    for i := 0; i < length; i++ {
        myCode += string(chars[rand.Intn(len(chars))])
    }

    return myCode
}

func generateClientName() string {
    clientName := "Anonymous "

    clientName += potentialClientNames[rand.Intn(len(potentialClientNames))]

    return clientName
}

func generateClientIdentifier() string {
    ID := ""
    var partOne, partTwo, partThree, partFour, partFive string

    for _, exists := currentClients[ID]; len(ID) == 0 || exists; {
        partOne = generateRandom("abcdef0123456789", 8)
        partTwo = generateRandom("abcdef0123456789", 4)
        partThree = generateRandom("abcdef0123456789", 4)
        partFour = generateRandom("abcdef0123456789", 4)
        partFive = generateRandom("abcdef0123456789", 12)

        ID = partOne + "-" + partTwo + "-" + partThree + "-" + partFour + "-" + partFive
    }

    return ID
}

func generateClientSecret() string {
    secret := ""
    uniqueGenerated := false

    for !uniqueGenerated {
        secret = generateRandom("abcdefghijklmnopqrstuvwxyz1234567890$-_#=", 64)

        uniqueGenerated = true

        for _, client := range currentClients {
            if (client.ClientSecret == secret) {
                uniqueGenerated = false
            }
        }
    }

    return secret
}

// GetAllPublicClients returns all publicly visible information about all connected clients.
func GetAllPublicClients() []PublicClient {
    clients := make([]PublicClient, 0);

    for _, client := range currentClients {
        thisClient := PublicClient { IdentityToken: client.ClientToken, IdentityName: client.ClientName, VoteHistory: client.VoteHistory, QueueHistory: client.QueueHistory }

        clients = append(clients, thisClient)
    }

    return clients;
}

// GetClient returns the client that matches the specified id
func GetClient(id string) *ConnectedClient {
    client, exists := currentClients[id]

    if (exists) {
        return client
    }

    return nil
}

// Create adds a new client and returns the generated representation as a
// ConnectedClient
func Create() *ConnectedClient {
    client := ConnectedClient {}

    client.ClientSecret = generateClientSecret();
    client.ClientToken = generateClientIdentifier();
    client.ClientName = generateClientName();
    client.VoteHistory = make([]Vote, 0);
    client.QueueHistory = make([]QueueEntry, 0);
    client.ConnectionTime = time.Now();
    client.LastCommunicated = time.Now();

    currentClients[client.ClientToken] = &client

    return &client
}
