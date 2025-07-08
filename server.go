package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	config         *Config
	configFile     string
	lastScreenshot []byte
	lastUpdate     time.Time
	mu             sync.RWMutex
	stopChan       chan struct{}
	template       *template.Template
}

func NewServer(config *Config, configFile string) *Server {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		// If template file doesn't exist, create a basic embedded template
		tmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head><title>Desktop Surveillance Camera</title></head>
<body>
	<h1>Desktop Surveillance Camera</h1>
	<p>Template file not found. Please create templates/index.html</p>
	<img src="/last" alt="Screenshot">
</body>
</html>`))
	}

	return &Server{
		config:     config,
		configFile: configFile,
		stopChan:   make(chan struct{}),
		template:   tmpl,
	}
}

func (s *Server) updateScreenshot() error {
	return s.updateScreenshotWithOptions(nil)
}

func (s *Server) updateScreenshotWithOptions(opts *ScreenshotOptions) error {
	var screenshot *Screenshot
	var err error

	if opts != nil {
		screenshot, err = TakeScreenshotWithOptions(opts)
	} else {
		// Use config settings for default options
		opts = &ScreenshotOptions{
			Region:    nil,
			Compress:  s.config.Capture.Compression.Enabled,
			MaxWidth:  s.config.Capture.Compression.MaxWidth,
			MaxHeight: s.config.Capture.Compression.MaxHeight,
		}

		// Apply region from config if set
		if s.config.Capture.Region != nil {
			opts.Region = &ScreenRegion{
				X:      s.config.Capture.Region.X,
				Y:      s.config.Capture.Region.Y,
				Width:  s.config.Capture.Region.Width,
				Height: s.config.Capture.Region.Height,
			}
		}

		screenshot, err = TakeScreenshotWithOptions(opts)
	}

	if err != nil {
		return err
	}

	pngData, err := screenshot.ToPNGBytesWithOptions(opts)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.lastScreenshot = pngData
	s.lastUpdate = time.Now()
	s.mu.Unlock()

	return nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Prepare template data
	data := struct {
		Config          *Config
		IntervalSeconds int
	}{
		Config:          s.config,
		IntervalSeconds: int(s.config.Capture.Interval.Seconds()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.template.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

func (s *Server) handleLast(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for custom screenshot options
	opts := s.parseScreenshotOptions(r.URL.Query())

	if s.config.Capture.Mode == "ondemand" {
		err := s.updateScreenshotWithOptions(opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to capture screenshot: %v", err), http.StatusInternalServerError)
			return
		}
	} else if opts != nil {
		// In realtime mode, if custom options are provided, take a fresh screenshot
		err := s.updateScreenshotWithOptions(opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to capture screenshot with custom options: %v", err), http.StatusInternalServerError)
			return
		}
	}

	s.mu.RLock()
	screenshot := make([]byte, len(s.lastScreenshot))
	copy(screenshot, s.lastScreenshot)
	lastUpdate := s.lastUpdate
	s.mu.RUnlock()

	if len(screenshot) == 0 {
		http.Error(w, "No screenshot available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Last-Modified", lastUpdate.UTC().Format(http.TimeFormat))

	w.Write(screenshot)
}

func (s *Server) parseScreenshotOptions(params url.Values) *ScreenshotOptions {
	opts := &ScreenshotOptions{}
	hasCustomOptions := false

	// Parse region parameters
	if x := params.Get("x"); x != "" {
		if xVal, err := strconv.Atoi(x); err == nil {
			if opts.Region == nil {
				opts.Region = &ScreenRegion{}
			}
			opts.Region.X = xVal
			hasCustomOptions = true
		}
	}

	if y := params.Get("y"); y != "" {
		if yVal, err := strconv.Atoi(y); err == nil {
			if opts.Region == nil {
				opts.Region = &ScreenRegion{}
			}
			opts.Region.Y = yVal
			hasCustomOptions = true
		}
	}

	if width := params.Get("width"); width != "" {
		if widthVal, err := strconv.Atoi(width); err == nil && widthVal > 0 {
			if opts.Region == nil {
				opts.Region = &ScreenRegion{}
			}
			opts.Region.Width = widthVal
			hasCustomOptions = true
		}
	}

	if height := params.Get("height"); height != "" {
		if heightVal, err := strconv.Atoi(height); err == nil && heightVal > 0 {
			if opts.Region == nil {
				opts.Region = &ScreenRegion{}
			}
			opts.Region.Height = heightVal
			hasCustomOptions = true
		}
	}

	// Parse compression parameters
	if compress := params.Get("compress"); compress != "" {
		if compressVal, err := strconv.ParseBool(compress); err == nil {
			opts.Compress = compressVal
			hasCustomOptions = true
		}
	}

	if maxWidth := params.Get("max_width"); maxWidth != "" {
		if maxWidthVal, err := strconv.Atoi(maxWidth); err == nil && maxWidthVal > 0 {
			opts.MaxWidth = maxWidthVal
			opts.Compress = true
			hasCustomOptions = true
		}
	}

	if maxHeight := params.Get("max_height"); maxHeight != "" {
		if maxHeightVal, err := strconv.Atoi(maxHeight); err == nil && maxHeightVal > 0 {
			opts.MaxHeight = maxHeightVal
			opts.Compress = true
			hasCustomOptions = true
		}
	}

	if !hasCustomOptions {
		return nil
	}

	return opts
}

func (s *Server) startRealtimeCapture() {
	if s.config.Capture.Mode != "realtime" {
		return
	}

	ticker := time.NewTicker(s.config.Capture.Interval)
	go func() {
		defer ticker.Stop()

		s.updateScreenshot()

		for {
			select {
			case <-ticker.C:
				err := s.updateScreenshot()
				if err != nil {
					fmt.Printf("Failed to capture screenshot: %v\n", err)
				}
			case <-s.stopChan:
				return
			}
		}
	}()
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/last", s.handleLast)
	http.HandleFunc("/config", s.handleConfig)
	http.HandleFunc("/preview", s.handlePreview)
	http.HandleFunc("/screen-info", s.handleScreenInfo)

	s.startRealtimeCapture()

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	fmt.Printf("Starting server on %s\n", addr)
	fmt.Printf("Mode: %s\n", s.config.Capture.Mode)
	if s.config.Capture.Mode == "realtime" {
		fmt.Printf("Capture interval: %v\n", s.config.Capture.Interval)
	}

	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleScreenInfo(w http.ResponseWriter, r *http.Request) {
	// Take a minimal screenshot to get actual screen dimensions
	screenshot, err := TakeScreenshot()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get screen info: %v", err), http.StatusInternalServerError)
		return
	}
	
	screenInfo := map[string]int{
		"width":  screenshot.Width,
		"height": screenshot.Height,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(screenInfo)
}

func (s *Server) Stop() {
	close(s.stopChan)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Return current configuration
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.config)
		return
	}

	if r.Method == "POST" {
		// Update configuration
		var newConfig Config
		err := json.NewDecoder(r.Body).Decode(&newConfig)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Validate the new configuration
		if newConfig.Server.Port < 1 || newConfig.Server.Port > 65535 {
			http.Error(w, "Invalid port number", http.StatusBadRequest)
			return
		}

		if newConfig.Capture.Mode != "ondemand" && newConfig.Capture.Mode != "realtime" {
			http.Error(w, "Invalid capture mode", http.StatusBadRequest)
			return
		}

		// Update in-memory configuration
		s.mu.Lock()
		oldMode := s.config.Capture.Mode
		s.config = &newConfig
		s.mu.Unlock()

		// Save to file if requested
		if r.URL.Query().Get("save") == "true" {
			err = SaveConfig(&newConfig, s.configFile)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// Restart realtime capture if mode changed
		if oldMode != newConfig.Capture.Mode {
			// Note: In a more sophisticated implementation, we might restart the capture goroutine
			// For now, we'll just update the config and let the user restart if needed
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	// Take a screenshot for preview purposes (always full screen, compressed)
	opts := &ScreenshotOptions{
		Region:    nil, // Always full screen for preview
		Compress:  true,
		MaxWidth:  800, // Small preview size
		MaxHeight: 600,
	}

	screenshot, err := TakeScreenshotWithOptions(opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to capture preview: %v", err), http.StatusInternalServerError)
		return
	}

	pngData, err := screenshot.ToPNGBytesWithOptions(opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode preview: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write(pngData)
}
