package version

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	Version string = "dev"

	// goreleaser can also pass the specific commit if you want
	Commit string = "HEAD"

	Date string = "?"
)
