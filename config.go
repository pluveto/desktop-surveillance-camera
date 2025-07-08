package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
)

type Config struct {
    Server   ServerConfig   `json:"server"`
    Capture  CaptureConfig  `json:"capture"`
}

type ServerConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

type CaptureConfig struct {
    Mode        string        `json:"mode"`        // "realtime" or "ondemand"
    Interval    time.Duration `json:"interval"`    // for realtime mode
    Region      *RegionConfig `json:"region"`      // optional screen region
    Compression CompressionConfig `json:"compression"` // image compression settings
}

type RegionConfig struct {
    X      int `json:"x"`
    Y      int `json:"y"`
    Width  int `json:"width"`
    Height int `json:"height"`
}

type CompressionConfig struct {
    Enabled   bool `json:"enabled"`
    MaxWidth  int  `json:"max_width"`
    MaxHeight int  `json:"max_height"`
}

func (c *Config) MarshalJSON() ([]byte, error) {
    type Alias Config
    return json.Marshal(&struct {
        *Alias
        Capture struct {
            Mode        string             `json:"mode"`
            Interval    string             `json:"interval"`
            Region      *RegionConfig      `json:"region"`
            Compression CompressionConfig  `json:"compression"`
        } `json:"capture"`
    }{
        Alias: (*Alias)(c),
        Capture: struct {
            Mode        string             `json:"mode"`
            Interval    string             `json:"interval"`
            Region      *RegionConfig      `json:"region"`
            Compression CompressionConfig  `json:"compression"`
        }{
            Mode:        c.Capture.Mode,
            Interval:    c.Capture.Interval.String(),
            Region:      c.Capture.Region,
            Compression: c.Capture.Compression,
        },
    })
}

func (c *Config) UnmarshalJSON(data []byte) error {
    type Alias Config
    aux := &struct {
        *Alias
        Capture struct {
            Mode        string             `json:"mode"`
            Interval    string             `json:"interval"`
            Region      *RegionConfig      `json:"region"`
            Compression CompressionConfig  `json:"compression"`
        } `json:"capture"`
    }{
        Alias: (*Alias)(c),
    }
    
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    
    c.Capture.Mode = aux.Capture.Mode
    c.Capture.Region = aux.Capture.Region
    c.Capture.Compression = aux.Capture.Compression
    
    if aux.Capture.Interval != "" {
        interval, err := time.ParseDuration(aux.Capture.Interval)
        if err != nil {
            return fmt.Errorf("invalid interval format: %v", err)
        }
        c.Capture.Interval = interval
    }
    
    return nil
}

func DefaultConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Host: "0.0.0.0",
            Port: 9981,
        },
        Capture: CaptureConfig{
            Mode:     "ondemand",
            Interval: 5 * time.Second,
            Region:   nil, // full screen
            Compression: CompressionConfig{
                Enabled:   false,
                MaxWidth:  1920,
                MaxHeight: 1080,
            },
        },
    }
}

func LoadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        if os.IsNotExist(err) {
            config := DefaultConfig()
            err := SaveConfig(config, filename)
            if err != nil {
                return nil, fmt.Errorf("failed to create default config: %v", err)
            }
            fmt.Printf("Created default config file: %s\n", filename)
            return config, nil
        }
        return nil, err
    }
    
    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return nil, fmt.Errorf("failed to parse config: %v", err)
    }
    
    return &config, nil
}

func SaveConfig(config *Config, filename string) error {
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, data, 0644)
}