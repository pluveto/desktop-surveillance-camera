//go:build !windows

package main

import (
    "fmt"
    "runtime"
)

type Screenshot struct {
    Width  int
    Height int
    Data   []byte
}

func TakeScreenshot() (*Screenshot, error) {
    return nil, fmt.Errorf("screenshot functionality is only supported on Windows, current OS: %s", runtime.GOOS)
}

func (s *Screenshot) SaveToPNG(filename string) error {
    return fmt.Errorf("screenshot functionality is only supported on Windows")
}

func (s *Screenshot) ToPNGBytes() ([]byte, error) {
    return nil, fmt.Errorf("screenshot functionality is only supported on Windows")
}

func SaveScreenshotToFile() (string, error) {
    return "", fmt.Errorf("screenshot functionality is only supported on Windows, current OS: %s", runtime.GOOS)
}