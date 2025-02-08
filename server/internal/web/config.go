package web

import "github.com/DaanV2/mechanus/server/pkg/config"

var (
	WebConfig        = config.New("web")
	HostFlag         = WebConfig.String("web.host", "", "What host to bind on, such as: \"\", \"localhost\" or \"0.0.0.0\"")
	PortFlag         = WebConfig.Int("web.port", 8080, "The port to server web traffic to")
	StaticFolderFlag = WebConfig.String("web.static.folder", "/web", "The default files to serve")
)
