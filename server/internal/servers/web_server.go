package servers

import (
	"errors"
	"net/http"

	"github.com/DaanV2/mechanus/server/pkg/config"
	xio "github.com/DaanV2/mechanus/server/pkg/extensions/io"
)

type WebServer struct {
}

func NewWebServer(conf config.WebConfig) (*WebServer, error) {
	folder := conf.Folder.Value()

	if !xio.DirExists(folder) {
		return nil, errors.New("couldn't find the folder to serve files from: " + folder)
	}

	router := http.NewServeMux()

	router.Handle("/", http.FileServer(http.Dir(conf.Folder.Value())))


	return &WebServer{}
}
