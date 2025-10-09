package config

const LogPath = "./.brains/.brains.log"

var DefaultConfig = BrainsConfig{
	LoggingEnabled: true,
	AWSRegion:      "us-east-1",
	Model:          "openai.gpt-oss-120b-1:0",
	Personas:       map[string]string{},
	DefaultContext: "**/*",
	DefaultPersona: "",
}
