package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// default log instance
var log = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

type Config struct {
	Host          string
	Port          string
	StaticDir     string
	BarcodeReader string
	Template      string
	Cors          bool
}

type Result struct {
	Success bool   `json:"success"`
	Barcode string `json:"barcode"`
	Result  string `json:"result"` // TODO actual result type
}

func parseOutput(raw []byte) *Result {
	res := &Result {}

	ress := string(raw)

	if ! strings.HasPrefix(ress, "OK") {
		return res
	}

	parts := strings.SplitN(ress, ",", 3)
	res.Success = true
	res.Barcode = parts[1]
	res.Result = parts[2]

	return res
}

func scan(conf *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Limit the size of the memory to 10MB
		r.ParseMultipartForm(10 << 20) // 10 MB

		file, handler, err := r.FormFile("image")
		if err != nil {
			log.Error("can't retrieve file from scan request", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		log.Info("Uploaded file", "name", handler.Filename, "size", handler.Size, "MIME", handler.Header)

		dst, err := os.CreateTemp("", "example.*.jpg")
		if err != nil {
			log.Error("can't create temp file in scan request", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// defer os.Remove(f.Name()) // TODO clean up

		// Copy the uploaded file to the created file on the filesystem
		if _, err := io.Copy(dst, file); err != nil {
			log.Error("can't save temp file in scan request", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := dst.Close(); err != nil {
			log.Error("can't close temp file in scan request", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		out, err := exec.CommandContext(ctx, conf.BarcodeReader, dst.Name()).CombinedOutput()
		if err != nil {
			log.Error("can't read barcode scanner in scan request", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Debug("Barcode reader output", "text", string(out))

		res := parseOutput(out)

		j, err := json.Marshal(res)
		if err != nil {
			log.Error("WTF", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if conf.Cors {
			w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%s:%s", conf.Host, conf.Port))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Content-Length", strconv.Itoa(len(j)))
		written, err := w.Write(j)
		if err != nil {
			log.Error("ResponseWriter says it's sorry", "error", err, "written", written)
		}

	}
}

func homePage(conf *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the embedded HTML template.
		tmpl, err := template.New("home").Parse(conf.Template)
		if err != nil {
			// Return an internal server error if the template parsing fails.
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, fmt.Sprintf("%s:%s/scan", conf.Host, conf.Port))
		if err != nil {
			// Return an internal server error if executing the template fails.
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Processing request", "RemoteAddr", r.RemoteAddr, "Method", r.Method, "URL", r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	config := &Config{
		Host: "http://localhost",
		Port: "54623",
		StaticDir: "static",
		BarcodeReader: "bin/scan",
		Cors: true,
	}

	flag.StringVar(&config.Host,          "host",          config.Host,          "host")
	flag.StringVar(&config.Port,          "port",          config.Port,          "port")
	flag.StringVar(&config.StaticDir,     "staticdir",     config.StaticDir,     "directory containing frontend")
	flag.StringVar(&config.BarcodeReader, "barcodereader", config.BarcodeReader, "path to barcode reader binary")

	flag.Parse()

	tfn := config.StaticDir + "/index.html"
	buf, err := os.ReadFile(tfn)
	if err != nil {
		log.Error("can't open html template file", "name", tfn, "error", err)
		os.Exit(1)
	}
	config.Template = string(buf)

	http.HandleFunc("/", homePage(config))
	http.HandleFunc("/scan", scan(config))

	server := &http.Server{
		Addr:           ":" + config.Port,
		Handler:        logRequest(http.DefaultServeMux),
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Info("Starting HTTP server on port", "port", config.Port)
	// always returns error. ErrServerClosed on graceful close
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// unexpected error. port in use?
		log.Error("can't start HTTP server, ListenAndServe() says:", "error", err)
	}

}
