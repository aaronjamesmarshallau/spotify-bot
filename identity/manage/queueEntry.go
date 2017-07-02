package manage

import "time"

// QueueEntry represents when a song has been queued on behalf of the user
type QueueEntry struct {
    TrackID string
    TrackName string
    QueueTimestamp time.Time
}
