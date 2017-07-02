package manage

import "time"

// Vote represents a vote by a connected client
type Vote struct {
    TrackID string
    TrackName string
    Upvoted bool
    TimeVoted time.Time
}
