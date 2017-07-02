package manage

// PrivateClient represents a client including its secret, which is only known to the server, and the client itself.
type PrivateClient struct {
    IdentitySecret string`json:"identitySecret"`
    IdentityToken string`json:"identityToken"`
    IdentityName string`json:"identityName"`
    VoteHistory []Vote`json:"voteHistory"`
    QueueHistory []QueueEntry`json:"queueHistory"`
}
