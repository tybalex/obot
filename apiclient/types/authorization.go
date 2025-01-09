package types

type Authorization struct {
	AuthorizationManifest
	User *User `json:"user,omitempty"`
}

type AuthorizationManifest struct {
	UserID  string `json:"userID,omitempty"`
	AgentID string `json:"agentId,omitempty"`
}

type AuthorizationList List[Authorization]
