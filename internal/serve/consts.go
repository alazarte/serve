package serve

// TODO: the config changes depending on which handler I'm using
type HandlerConfig struct {
	KeyFile  string `json:"key"`
	CertFile string `json:"cert"`
	Hosts    map[string]struct {
		Root        string `json:"root"`
		HandlerType string `json:"handlerType"`
		URL         string `json:"url"`
	} `json:"hosts"`
}

const (
	httpPort        = ":80"
	httpsPort       = ":443"
	RedirectHandler = "redirect"
	PublicHandler   = "public"
	HTMLHandler     = "html"
)
