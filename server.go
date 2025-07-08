package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	config         *Config
	lastScreenshot []byte
	lastUpdate     time.Time
	mu             sync.RWMutex
	stopChan       chan struct{}
}

func NewServer(config *Config) *Server {
	return &Server{
		config:   config,
		stopChan: make(chan struct{}),
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(s.getIndexHTML()))
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

	s.startRealtimeCapture()

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	fmt.Printf("Starting server on %s\n", addr)
	fmt.Printf("Mode: %s\n", s.config.Capture.Mode)
	if s.config.Capture.Mode == "realtime" {
		fmt.Printf("Capture interval: %v\n", s.config.Capture.Interval)
	}

	return http.ListenAndServe(addr, nil)
}

func (s *Server) Stop() {
	close(s.stopChan)
}

func (s *Server) getIndexHTML() string {
	mode := s.config.Capture.Mode
	interval := int(s.config.Capture.Interval.Seconds())

	// Get region info
	regionInfo := ""
	if s.config.Capture.Region != nil {
		regionInfo = fmt.Sprintf("<br><strong>Region:</strong> %dx%d at (%d,%d)",
			s.config.Capture.Region.Width, s.config.Capture.Region.Height,
			s.config.Capture.Region.X, s.config.Capture.Region.Y)
	}

	// Get compression info
	compressionInfo := ""
	if s.config.Capture.Compression.Enabled {
		compressionInfo = fmt.Sprintf("<br><strong>Compression:</strong> Max %dx%d",
			s.config.Capture.Compression.MaxWidth, s.config.Capture.Compression.MaxHeight)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Desktop Surveillance Camera</title>
    <style>
        body {
            margin: 0;
            padding: 20px;
            font-family: Arial, sans-serif;
            background-color: #f0f0f0;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 20px;
        }
        .status {
            background-color: #e7f3ff;
            border: 1px solid #b3d9ff;
            border-radius: 4px;
            padding: 10px;
            margin-bottom: 20px;
        }
        .screenshot-container {
            text-align: center;
            margin-bottom: 20px;
        }
        .screenshot {
            max-width: 100%%;
            max-height: 80vh;
            border: 2px solid #ddd;
            border-radius: 4px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .controls {
            text-align: center;
            margin-top: 20px;
        }
        .btn {
            background-color: #007cba;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 4px;
            cursor: pointer;
            margin: 0 5px;
            font-size: 16px;
        }
        .btn:hover {
            background-color: #005a87;
        }
        .btn:disabled {
            background-color: #ccc;
            cursor: not-allowed;
        }
        .last-update {
            margin-top: 10px;
            color: #666;
            font-size: 14px;
        }
        .api-info {
            background-color: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 15px;
            margin-top: 20px;
            font-family: monospace;
            font-size: 12px;
        }
        .api-info h4 {
            margin-top: 0;
            font-family: Arial, sans-serif;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Desktop Surveillance Camera</h1>
        </div>
        
        <div class="status">
            <strong>Mode:</strong> %s
            %s
            %s
            %s
        </div>
        
        <div class="screenshot-container">
            <img id="screenshot" class="screenshot" src="/last" alt="Screenshot" onload="updateLastUpdate()" onerror="handleImageError()">
        </div>
        
        <div class="controls">
            <button class="btn" onclick="refreshScreenshot()">Refresh Screenshot</button>
            %s
        </div>
        
        <div class="last-update" id="lastUpdate"></div>
        
        <div class="api-info">
            <h4>API Usage Examples</h4>
            <div><strong>Full screenshot:</strong> /last</div>
            <div><strong>Region screenshot:</strong> /last?x=100&y=100&width=800&height=600</div>
            <div><strong>Compressed screenshot:</strong> /last?compress=true&max_width=800&max_height=600</div>
            <div><strong>Combined:</strong> /last?x=0&y=0&width=1920&height=1080&compress=true&max_width=640&max_height=480</div>
        </div>
    </div>

    <script>
        let autoRefreshInterval = null;
        let isRealtime = %t;
        let refreshIntervalSeconds = %d;
        
        function updateLastUpdate() {
            document.getElementById('lastUpdate').textContent = 'Last updated: ' + new Date().toLocaleString();
        }
        
        function handleImageError() {
            document.getElementById('lastUpdate').textContent = 'Screenshot failed to load';
        }
        
        function refreshScreenshot() {
            const img = document.getElementById('screenshot');
            img.src = '/last?' + new Date().getTime();
        }
        
        function toggleAutoRefresh() {
            const btn = document.getElementById('autoRefreshBtn');
            if (autoRefreshInterval) {
                clearInterval(autoRefreshInterval);
                autoRefreshInterval = null;
                btn.textContent = 'Enable Auto Refresh';
                btn.style.backgroundColor = '#007cba';
            } else {
                autoRefreshInterval = setInterval(refreshScreenshot, refreshIntervalSeconds * 1000);
                btn.textContent = 'Disable Auto Refresh';
                btn.style.backgroundColor = '#d32f2f';
            }
        }
        
        if (isRealtime) {
            autoRefreshInterval = setInterval(refreshScreenshot, refreshIntervalSeconds * 1000);
        }
        
        updateLastUpdate();
    </script>
    
    <noscript>
        <div style="background-color: #fff3cd; border: 1px solid #ffeaa7; border-radius: 4px; padding: 10px; margin: 20px 0;">
            <strong>Note:</strong> JavaScript is disabled. In on-demand mode, please manually refresh the page to get the latest screenshot.
        </div>
    </noscript>
</body>
</html>`,
		mode,
		func() string {
			if mode == "realtime" {
				return fmt.Sprintf("<br><strong>Auto refresh interval:</strong> %d seconds", interval)
			}
			return "<br><strong>Description:</strong> On-demand mode, click refresh or reload page for latest screenshot"
		}(),
		regionInfo,
		compressionInfo,
		func() string {
			if mode == "realtime" {
				return `<button id="autoRefreshBtn" class="btn" onclick="toggleAutoRefresh()">Disable Auto Refresh</button>`
			}
			return `<button id="autoRefreshBtn" class="btn" onclick="toggleAutoRefresh()">Enable Auto Refresh</button>`
		}(),
		mode == "realtime",
		interval,
	)
}
