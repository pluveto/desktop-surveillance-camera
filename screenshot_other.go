//go:build !windows

package main

import (
    "fmt"
    "runtime"
)

type ScreenRegion struct {
    X      int
    Y      int
    Width  int
    Height int
}

type Screenshot struct {
    Width  int
    Height int
    Data   []byte
    Region *ScreenRegion
}

type ScreenshotOptions struct {
    Region     *ScreenRegion
    Compress   bool
    MaxWidth   int
    MaxHeight  int
    Quality    int
}

func TakeScreenshot() (*Screenshot, error) {
    return nil, fmt.Errorf("screenshot functionality is only supported on Windows, current OS: %s", runtime.GOOS)
}

func TakeScreenshotWithOptions(opts *ScreenshotOptions) (*Screenshot, error) {
    return nil, fmt.Errorf("screenshot functionality is only supported on Windows, current OS: %s", runtime.GOOS)
}

func TakeRegionScreenshot(x, y, width, height int) (*Screenshot, error) {
    return nil, fmt.Errorf("screenshot functionality is only supported on Windows, current OS: %s", runtime.GOOS)
}

func (s *Screenshot) SaveToPNG(filename string) error {
    return fmt.Errorf("screenshot functionality is only supported on Windows")
}

func (s *Screenshot) ToPNGBytes() ([]byte, error) {
    return nil, fmt.Errorf("screenshot functionality is only supported on Windows")
}

func (s *Screenshot) ToPNGBytesWithOptions(opts *ScreenshotOptions) ([]byte, error) {
    return nil, fmt.Errorf("screenshot functionality is only supported on Windows")
}

func SaveScreenshotToFile() (string, error) {
    return "", fmt.Errorf("screenshot functionality is only supported on Windows, current OS: %s", runtime.GOOS)
}