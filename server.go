package main

import (
	"net/http"
	"os"
	"path/filepath"

	httputils "github.com/3bl3gamer/go-http-utils"
	"github.com/ansel1/merry"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
)

func StartHTTPServer(address string, env Env, imgSearcher *ImageSearcher, tags map[string]string) error {
	ex, err := os.Executable()
	if err != nil {
		return merry.Wrap(err)
	}
	baseDir := filepath.Dir(ex)

	var bundleFPath, stylesFPath string

	// Config
	wrapper := &httputils.Wrapper{
		ShowErrorDetails: env.IsDev(),
		TemplateHandler: &httputils.TemplateHandler{
			CacheParsed: env.IsProd(),
			BasePath:    baseDir + "/www/templates",
			ParamsFunc: func(r *http.Request, ctx *httputils.MainCtx, params httputils.TemplateCtx) error {
				params["BundleFPath"] = bundleFPath
				params["StylesFPath"] = stylesFPath
				return nil
			},
			LogBuild: func(path string) { log.Info().Str("path", path).Msg("building template") },
		},
		LogError: func(err error, r *http.Request) {
			log.Error().Stack().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("")
		},
	}

	router := httprouter.New()
	route := func(method, path string, chain ...interface{}) {
		router.Handle(method, path, wrapper.WrapChain(chain...))
	}

	handleImages := func(wr http.ResponseWriter, r *http.Request, ps httprouter.Params) (interface{}, error) {
		return map[string]interface{}{"images": imgSearcher.images, "tags": tags}, nil
	}

	// Routes
	route("GET", "/", HandleIndex)
	route("GET", "/images", handleImages)

	route("GET", "/api/explode", func(wr http.ResponseWriter, r *http.Request, ps httprouter.Params) (interface{}, error) {
		return nil, merry.New("test API error")
	})
	route("GET", "/explode", func(wr http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		return merry.New("test error")
	})
	for _, dir := range imgSearcher.imagesDirs {
		router.ServeFiles("/img/"+filepath.Base(dir.Suffix)+"/*filepath", http.Dir(dir.Path))
	}
	router.ServeFiles("/thumb/*filepath", http.Dir(imgSearcher.thumbDir))

	if env.IsDev() {
		devServerAddress, err := httputils.RunBundleDevServerNear(address, baseDir+"/www", "--configHost", "--configPort")
		if err != nil {
			log.Fatal().Err(err)
		}
		bundleFPath = "http://" + devServerAddress + "/bundle.js"
		stylesFPath = "http://" + devServerAddress + "/bundle.css"
	} else {
		distPath := baseDir + "/www/dist"
		bundleFPath, stylesFPath, err = httputils.LastJSAndCSSFNames(distPath, "bundle.", "bundle.")
		if err != nil {
			log.Fatal().Err(err)
		}
		bundleFPath = "/dist/" + bundleFPath
		stylesFPath = "/dist/" + stylesFPath
		router.ServeFiles("/dist/*filepath", http.Dir(distPath))
	}
	log.Info().Str("fpath", bundleFPath).Msg("bundle")
	log.Info().Str("fpath", stylesFPath).Msg("styles")

	// Server
	log.Info().Msg("starting server on " + address)
	return merry.Wrap(http.ListenAndServe(address, router))
}

func HandleIndex(wr http.ResponseWriter, r *http.Request, ps httprouter.Params) (httputils.TemplateCtx, error) {
	return map[string]interface{}{"FPath": "index.html", "Block": "index.html"}, nil
}
