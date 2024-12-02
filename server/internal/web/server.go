package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/DaanV2/mechanus/server/pkg/config"
	xio "github.com/DaanV2/mechanus/server/pkg/extensions/io"
)

var (
	WebConfig           *config.Config
	FolderFlag          config.Flag[string]
	CampaignsFolderFlag config.Flag[string]
)

func init() {
	WebConfig = config.New("web.server").WithValidate(validateConfig)
	FolderFlag = WebConfig.String("web.server", config.StorageFolder("web", "static"), "The folder where static content needs to be served from")
}

func validateConfig(c *config.Config) error {
	var err error
	folder := FolderFlag.Value()
	if !xio.DirExists(folder) {
		err = errors.Join(err, errors.New("couldn't find the folder to serve files from: "+folder))
	}

	return err
}

type Server struct {
	appCtx context.Context
	router *http.ServeMux
}

func NewWebServer(router *http.ServeMux, appCtx context.Context) (*Server, error) {
	return &Server{
		appCtx,
		router,
	}, nil
}
