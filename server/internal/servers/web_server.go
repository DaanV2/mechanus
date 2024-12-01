package servers

import (
	"errors"
	"net/http"

	"github.com/DaanV2/mechanus/server/pkg/config"
	xio "github.com/DaanV2/mechanus/server/pkg/extensions/io"
)

var (
	WebConfig  = config.New("server.web")
	FolderFlag = WebConfig.String("server.web.folder", config.StorageFolder("web", "static"), "The folder where static content needs to be served from")
)

type WebServer struct {
}

func NewWebServer() (*WebServer, error) {
	folder := FolderFlag.Value()

	if !xio.DirExists(folder) {
		return nil, errors.New("couldn't find the folder to serve files from: " + folder)
	}

	router := http.NewServeMux()

	router.Handle("/", http.FileServer(http.Dir(folder)))

	return &WebServer{}, nil
}
