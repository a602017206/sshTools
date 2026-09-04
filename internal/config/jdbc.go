package config

type JDBCManifest struct {
	Version int          `json:"version"`
	Drivers []JDBCDriver `json:"drivers"`
}

type JDBCDriver struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Category           string              `json:"category,omitempty"`
	RecommendedVersion string              `json:"recommendedVersion"`
	Profiles           []JDBCDriverProfile `json:"profiles"`
}

type JDBCDriverProfile struct {
	ID             string     `json:"id"`
	Version        string     `json:"version"`
	DriverClass    string     `json:"driverClass"`
	URLTemplate    string     `json:"urlTemplate"`
	DefaultPort    int        `json:"defaultPort"`
	JRERequirement string     `json:"jre"`
	Jars           []JDBCJar  `json:"jars"`
	Properties     []JDBCProp `json:"properties,omitempty"`
	Source         string     `json:"source,omitempty"`
	Installed      bool       `json:"installed,omitempty"`
	InstallPath    string     `json:"installPath,omitempty"`
}

type JDBCJar struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	URL    string `json:"url,omitempty"`
}

type JDBCProp struct {
	Name         string `json:"name"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Required     bool   `json:"required,omitempty"`
}
