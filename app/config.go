package app

type Config struct {
	Address string

	// The root url of where to find the swagger files
	RootUrl string

	HeaderTitle string
	HeaderLogo  string
	Favicon     string
}
