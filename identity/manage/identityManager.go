package manage

import (
    "math/rand"
    "time"
)

var currentClients map[string]ConnectedClient = make(map[string]ConnectedClient)

func generateRandom(chars string, length int) string {
    myCode := ""

    for i := 0; i < length; i++ {
        myCode += string(chars[rand.Intn(len(chars))])
    }

    return myCode
}

func generateClientName() string {
    clientName := "Anonymous "

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
    }

    return partOne + "-" + partTwo + "-" + partThree + "-" + partFour + "-" + partFive
}

// GetClient returns the client that matches the specified id
func GetClient(id string) *ConnectedClient {
    client, _ := currentClients[id]

    return &client
}

// Create adds a new client and returns the generated representation as a
// ConnectedClient
func Create() *ConnectedClient {
    client := ConnectedClient {}

    client.ClientToken = generateClientIdentifier();
    client.ClientName = generateClientName();
    client.ConnectionTime = time.Now();
    client.LastCommunicated = time.Now();

    currentClients[client.ClientToken] = client

    return &client
}
