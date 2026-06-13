package split

import "os"

// FromEnv builds a Provider from the plugin's environment. Secrets and endpoint
// come from the plugin process, never from the Rollops target spec (Rollops
// passes only the flag name, environment, and percentage).
//
//	SPLIT_API_URL  base URL (default https://api.split.io)
//	SPLIT_TOKEN    Admin API key (required)
func FromEnv() Provider {
	base := os.Getenv("SPLIT_API_URL")
	if base == "" {
		base = "https://api.split.io"
	}
	return Provider{
		BaseURL: base,
		Token:   os.Getenv("SPLIT_TOKEN"),
	}
}
