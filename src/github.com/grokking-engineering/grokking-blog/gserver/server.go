package gserver

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/grokking-engineering/grokking-blog/handlers"
	"github.com/grokking-engineering/grokking-blog/middlewares"
	"github.com/grokking-engineering/grokking-blog/store"
	"github.com/grokking-engineering/grokking-blog/utils/logs"
)

var l = logs.New("gserver")

type Config struct {
	Server struct {
		ListenAddr    string `json:"LISTEN_ADDR"`
		ContentDir    string `json:"CONTENT_DIR"`
		StaticDir     string `json:"STATIC_DIR"`
		IsDevelopment string `json:"DEVELOPMENT"`
	} `json:"server"`
}

func Start(cfg Config) error {
	s := setup(cfg)
	listenAddr := cfg.Server.ListenAddr

	l.WithFields(logs.M{
		"addr": listenAddr,
	}).Info("Server is listening")
	return http.ListenAndServe(listenAddr, s.Handler)
}

type setupStruct struct {
	Config  Config
	Handler http.Handler
}

func setup(cfg Config) *setupStruct {
	s := &setupStruct{Config: cfg}
	s.setupRoutes()

	return s
}

func commonMiddlewares() func(http.Handler) http.Handler {
	logger := middlewares.NewLogger()
	recovery := middlewares.NewRecovery()

	return func(h http.Handler) http.Handler {
		return recovery(logger(h))
	}
}

func (s *setupStruct) setupRoutes() {
	mainStore := &store.Instance{
		ContentDir: s.Config.Server.ContentDir,
	}
	mainStore.Init()

	isDev := s.Config.Server.IsDevelopment == "1"
	if isDev {
		l.Println("Server is running in DEVELOPMENT MODE")
	}
	mainHandler := &handlers.MainHandler{
		Store: mainStore,
		IsDev: isDev,
	}
	mainHandler.Init()

	router := http.NewServeMux()
	s.Handler = router
	common := commonMiddlewares()

	staticDir := s.Config.Server.StaticDir
	_, err := os.Stat(staticDir)
	if err != nil {
		l.WithError(err).Fatal("Static dir not found")
	}

	router.Handle("/", common(mainHandler))
	router.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(staticDir))))
	router.Handle("/__reload__", common(reloadHandler(mainStore)))
}

func reloadHandler(mainStore *store.Instance) http.Handler {
	lastReload := time.Now()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		now := time.Now()
		if now.Sub(lastReload) > 5*time.Second {
			err := mainStore.ClearCacheAndReload()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Could not reload, keep serving old content. Please check server logs!")
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Reloaded!")
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Must wait 5 seconds before reloading again!")
	})
}
