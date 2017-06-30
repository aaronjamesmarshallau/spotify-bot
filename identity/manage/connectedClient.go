package manage

import (
    "time"
)

// QueueEntry represents when a song has been queued on behalf of the user
type QueueEntry struct {
    TrackID string
    QueueTimestamp time.Time
}

// Vote represents a vote by a connected client
type Vote struct {
    TrackID string
    Upvoted bool
    TimeVoted time.Time
}

// ConnectedClient represents a client connect to spotify-bot
type ConnectedClient struct {
    ClientSecret string
    ClientToken string
    ClientName string
    VoteHistory []Vote
    QueueHistory []QueueEntry
    ConnectionTime time.Time
    LastCommunicated time.Time
}
