package routes

var (
	TypeRoot   = "root"
	TypePublic = "public"
	TypeProxy  = "proxy"
)

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Handler struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Path    string   `json:"path"`
	Port    string   `json:"port"`
	Headers []Header `json:"headers"`
}

type Config struct {
	Pem      string    `json:"pem"`
	Sk       string    `json:"sk"`
	Debug    string    `json:"debug"`
	Handlers []Handler `json:"handlers"`
}
