package app

type Config struct {
	Address     string
	DocsUseHttp bool

	// The root url of where to find the swagger files
	RootUrl string

	HeaderTitle string
	HeaderLogo  string
}
