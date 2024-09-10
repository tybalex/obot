package v1

type ToolManifest struct {
	Name     string
	Params   map[string]string
	Body     string
	Metadata map[string]string
}
