package server

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/auth/apikeys"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/handlers"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/template"
	"gabe565.com/linx-server/internal/torrent"
	"gabe565.com/linx-server/internal/upload"
	"gabe565.com/linx-server/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func Setup() (*chi.Mux, error) {
	var err error

	if config.Default.ViteURL == "" {
		if err := template.LoadManifest(); err != nil {
			return nil, err
		}
	}

	switch config.Default.SiteURL.Path {
	case "", "/":
		config.Default.SiteURL.Path = "/"
	default:
		config.Default.SiteURL.Path = "/" + strings.Trim(config.Default.SiteURL.Path, "/") + "/"
	}
	config.Default.SelifPath = strings.Trim(config.Default.SelifPath, "/") + "/"

	config.TimeStarted = time.Now()

	var customPages []string
	if config.Default.CustomPagesPath != "" {
		customPages, err = handlers.ListCustomPages(config.Default.CustomPagesPath)
		if err != nil {
			return nil, err
		}

		config.CustomPages = customPages
	}

	// Routing setup
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	if config.Default.SiteURL.Path != "/" {
		r.Use(middleware.StripPrefix(strings.TrimSuffix(config.Default.SiteURL.Path, "/")))
	}
	if config.Default.Header.RealIP {
		r.Use(middleware.RealIP)
	}

	r.Use(middleware.Heartbeat("/ping"))

	if !config.Default.NoLogs {
		r.Use(middleware.Logger)
	}

	r.Use(middleware.Recoverer)
	r.Use(handlers.TurnstileMiddleware)
	r.Use(middleware.GetHead)
	r.Use(NewCSPMiddleware(Options{
		Policy:         GenerateCSP(),
		ReferrerPolicy: config.Default.Header.ReferrerPolicy,
		Frame:          config.Default.Header.XFrameOptions,
	}))
	r.Use(headers.AddHeaders(config.Default.Header.AddHeaders))

	r.Use(RemoveMultipartForm)

	if config.Default.Auth.File != "" {
		r.Use(apikeys.NewAPIKeysMiddleware(apikeys.AuthOptions{
			AuthFile:      config.Default.Auth.File,
			UnauthMethods: []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
			BasicAuth:     config.Default.Auth.Basic,
			SiteName:      config.Default.SiteName,
			SitePath:      config.Default.SiteURL.Path,
		}))
	}

	if len(customPages) != 0 {
		r.Get("/api/custom_page/{name}", handlers.CustomPage(config.Default.CustomPagesPath))
	}

	r.Group(func(r chi.Router) {
		r.Use(
			rateLimit(config.Default.Limit.UploadMaxRequests, config.Default.Limit.UploadInterval.Duration),
			LimitBodySize(int64(config.Default.MaxSize)),
		)

		if config.Default.Auth.AdminPasswordHash != "" {
			r.Use(handlers.AdminAuthMiddleware)
		}

		r.Post("/api/auth", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("Authorized"))
		})
		r.Post("/", upload.POSTHandler)
		r.Put("/", upload.PUTHandler) // Changed from /upload
		r.Put("/{name}", upload.PUTHandler) // Changed from /upload/{name}
		r.Get("/api/ws", upload.WSHandler)
		if config.Default.RemoteUploads {
			r.Get("/upload", upload.Remote)
			r.Get("/upload/{name}", upload.Remote)

			if config.Default.Auth.RemoteFile != "" {
				config.RemoteAuthKeys = apikeys.ReadAuthKeys(config.Default.Auth.RemoteFile)
			}
		}

		r.Get("/api/delete/{name}", handlers.Delete)
		r.Delete("/{name}", handlers.Delete)
	})

	r.Group(func(r chi.Router) {
		r.Use(rateLimit(config.Default.Limit.FileMaxRequests, config.Default.Limit.FileInterval.Duration))

		r.Get("/{name}", handlers.FileAccessHandler)
		r.Post("/{name}", handlers.FileAccessHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(rateLimit(config.Default.Limit.FileMaxRequests, config.Default.Limit.FileInterval.Duration))

		r.Get(path.Join("/", config.Default.SelifPath, "{name}"), handlers.FileServeHandler)

		if !config.Default.NoTorrent {
			r.Get("/torrent/{name}", torrent.FileTorrentHandler)
			r.Get("/{name}/torrent", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/torrent/"+chi.URLParam(r, "name"), http.StatusMovedPermanently)
			})
		}
	})

	r.Get("/api/config", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(template.NewConfig())
		obfuscated := util.Obfuscate(b)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.ServeContent(w, r, "config.json", config.TimeStarted, strings.NewReader(obfuscated))
	})

	// Admin authentication routes
	r.Get("/login", handlers.AdminAuthPage)
	r.Post("/api/login/auth", handlers.AdminAuthAPI)
	r.Get("/api/login/auth", handlers.AdminAuthAPI)



	// ClockHandler for the root path
	r.Get("/", handlers.ClockHandler)

	for _, p := range append(customPages, "API") {
		r.Get("/"+strings.ToLower(p), handlers.AssetHandler(template.WithTitle(p)))
	}

	r.NotFound(handlers.AssetHandler())

	return r, nil
}

func rateLimit(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler {
	limiter := httprate.NewRateLimiter(requestLimit, windowLength,
		httprate.WithKeyByIP(),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			handlers.ErrorMsg(w, r, http.StatusTooManyRequests, "Too many requests")
		}),
	)
	return limiter.Handler
}
