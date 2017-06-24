package manage

import (
    "time"
)

// ConnectedClient represents a client connect to spotify-bot
type ConnectedClient struct {
    ClientSecret string
    ClientToken string
    ClientName string
    ConnectionTime time.Time
    LastCommunicated time.Time
}
