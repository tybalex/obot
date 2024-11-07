package types

type Assistant struct {
	Metadata
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icons       AgentIcons `json:"icons"`
}

type AssistantList List[Assistant]
