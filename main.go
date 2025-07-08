package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "runtime"
    "syscall"
)

const (
    defaultConfigFile = "config.json"
    version          = "1.0.0"
)

func main() {
    var (
        configFile = flag.String("config", defaultConfigFile, "配置文件路径")
        showHelp   = flag.Bool("help", false, "显示帮助信息")
        showVersion = flag.Bool("version", false, "显示版本信息")
        testMode   = flag.Bool("test", false, "测试截图功能")
    )
    flag.Parse()

    if *showHelp {
        printHelp()
        return
    }

    if *showVersion {
        fmt.Printf("桌面监控摄像头 v%s\n", version)
        fmt.Printf("构建平台: %s/%s\n", runtime.GOOS, runtime.GOARCH)
        return
    }

    if *testMode {
        testScreenshot()
        return
    }

    if runtime.GOOS != "windows" {
        log.Printf("警告: 当前运行在 %s 平台，截图功能仅在 Windows 平台可用", runtime.GOOS)
    }

    config, err := LoadConfig(*configFile)
    if err != nil {
        log.Fatalf("加载配置文件失败: %v", err)
    }

    validateConfig(config)

    server := NewServer(config, *configFile)

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        fmt.Println("\n正在关闭服务器...")
        server.Stop()
        os.Exit(0)
    }()

    err = server.Start()
    if err != nil {
        log.Fatalf("启动服务器失败: %v", err)
    }
}

func printHelp() {
    fmt.Printf(`桌面监控摄像头 v%s

使用方法:
  %s [选项]

选项:
  -config string
        配置文件路径 (默认: %s)
  -test
        测试截图功能
  -version
        显示版本信息
  -help
        显示此帮助信息

配置文件格式 (JSON):
{
  "server": {
    "host": "0.0.0.0",
    "port": 9981
  },
  "capture": {
    "mode": "ondemand",    // "ondemand" 或 "realtime"
    "interval": "5s"       // 仅在 realtime 模式下使用
  }
}

运行模式:
  ondemand  - 按需模式：只有在访问 /last 端点时才截图
  realtime  - 实时模式：按配置的间隔自动截图

API 端点:
  /         - 网页界面
  /last     - 获取最新截图 (PNG 格式)

示例:
  %s                           # 使用默认配置启动
  %s -config my.json           # 使用指定配置文件启动
  %s -test                     # 测试截图功能
`, version, os.Args[0], defaultConfigFile, os.Args[0], os.Args[0], os.Args[0])
}

func testScreenshot() {
    fmt.Println("正在测试截图功能...")
    
    if runtime.GOOS != "windows" {
        fmt.Printf("错误: 截图功能仅在 Windows 平台可用，当前平台: %s\n", runtime.GOOS)
        return
    }
    
    filename, err := SaveScreenshotToFile()
    if err != nil {
        fmt.Printf("截图失败: %v\n", err)
        return
    }
    
    fmt.Printf("截图成功保存至: %s\n", filename)
}

func validateConfig(config *Config) {
    if config.Server.Port < 1 || config.Server.Port > 65535 {
        log.Fatalf("无效的端口号: %d", config.Server.Port)
    }
    
    if config.Capture.Mode != "ondemand" && config.Capture.Mode != "realtime" {
        log.Fatalf("无效的捕获模式: %s，只支持 'ondemand' 或 'realtime'", config.Capture.Mode)
    }
    
    if config.Capture.Mode == "realtime" && config.Capture.Interval <= 0 {
        log.Fatalf("实时模式下截图间隔必须大于 0")
    }
    
    if config.Capture.Interval.Seconds() < 1 {
        log.Printf("警告: 截图间隔过短 (%v)，可能会影响性能", config.Capture.Interval)
    }
}