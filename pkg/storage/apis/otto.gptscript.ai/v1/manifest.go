package v1

type Manifest struct {
	ID          string
	Name        string
	Description string
	Prompt      string
	Tools       []string
	Agents      []string
}
