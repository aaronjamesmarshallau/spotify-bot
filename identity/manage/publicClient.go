package manage

// PublicClient represents a response from upserting an identity
type PublicClient struct {
    IdentityToken string`json:"identityToken"`
    IdentityName string`json:"identityName"`
    VoteHistory []Vote`json:"voteHistory"`
    QueueHistory []QueueEntry`json:"queueHistory"`
}
