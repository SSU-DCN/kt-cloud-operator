package httpapi

// Meta API for object metadata

type KTMachineStatus struct {
	ID             string           `json:"id"`
	AdminPass      string           `json:"adminPass"`
	Links          []Links          `json:"links"`
	SecurityGroups []SecurityGroups `json:"securityGroups"`
}

type Links struct {
	Rel  string `json:"rel,omitempty"`
	Href string `json:"href,omitempty"`
}

type SecurityGroups struct {
	Name string `json:"name,omitempty"`
}

// var Config cloudapi.Config
// var logger1 *zap.SugaredLogger
