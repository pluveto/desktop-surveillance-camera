package main

import (
    "fmt"
    "net/http"
    "sync"
    "time"
)

type Server struct {
    config        *Config
    lastScreenshot []byte
    lastUpdate    time.Time
    mu            sync.RWMutex
    stopChan      chan struct{}
}

func NewServer(config *Config) *Server {
    return &Server{
        config:   config,
        stopChan: make(chan struct{}),
    }
}

func (s *Server) updateScreenshot() error {
    screenshot, err := TakeScreenshot()
    if err != nil {
        return err
    }
    
    pngData, err := screenshot.ToPNGBytes()
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
    if s.config.Capture.Mode == "ondemand" {
        err := s.updateScreenshot()
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to capture screenshot: %v", err), http.StatusInternalServerError)
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
    
    return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>桌面监控摄像头</title>
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
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>桌面监控摄像头</h1>
        </div>
        
        <div class="status">
            <strong>运行模式:</strong> %s
            %s
        </div>
        
        <div class="screenshot-container">
            <img id="screenshot" class="screenshot" src="/last" alt="屏幕截图" onload="updateLastUpdate()" onerror="handleImageError()">
        </div>
        
        <div class="controls">
            <button class="btn" onclick="refreshScreenshot()">刷新截图</button>
            %s
        </div>
        
        <div class="last-update" id="lastUpdate"></div>
    </div>

    <script>
        let autoRefreshInterval = null;
        let isRealtime = %t;
        let refreshIntervalSeconds = %d;
        
        function updateLastUpdate() {
            document.getElementById('lastUpdate').textContent = '最后更新: ' + new Date().toLocaleString();
        }
        
        function handleImageError() {
            document.getElementById('lastUpdate').textContent = '截图加载失败';
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
                btn.textContent = '开启自动刷新';
                btn.style.backgroundColor = '#007cba';
            } else {
                autoRefreshInterval = setInterval(refreshScreenshot, refreshIntervalSeconds * 1000);
                btn.textContent = '关闭自动刷新';
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
            <strong>注意:</strong> JavaScript 已禁用。在按需模式下，请手动刷新页面以获取最新截图。
        </div>
    </noscript>
</body>
</html>`,
        mode,
        func() string {
            if mode == "realtime" {
                return fmt.Sprintf("<br><strong>自动刷新间隔:</strong> %d秒", interval)
            }
            return "<br><strong>说明:</strong> 按需模式，点击刷新或刷新页面获取最新截图"
        }(),
        func() string {
            if mode == "realtime" {
                return `<button id="autoRefreshBtn" class="btn" onclick="toggleAutoRefresh()">关闭自动刷新</button>`
            }
            return `<button id="autoRefreshBtn" class="btn" onclick="toggleAutoRefresh()">开启自动刷新</button>`
        }(),
        mode == "realtime",
        interval,
    )
}