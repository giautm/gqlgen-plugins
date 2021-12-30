package runtime

// Service is the service object that the
// generated.go file will return for the service
// query
type Service struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Schema  string `json:"schema"`
}
