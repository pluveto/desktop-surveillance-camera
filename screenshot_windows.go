//go:build windows

package main

/*
#include <windows.h>
#include <wingdi.h>

typedef struct {
    BYTE* data;
    int width;
    int height;
    int size;
} ScreenshotData;

ScreenshotData* takeScreenshot() {
    HDC hdcScreen = GetDC(NULL);
    HDC hdcMemDC = CreateCompatibleDC(hdcScreen);
    
    int screenWidth = GetSystemMetrics(SM_CXSCREEN);
    int screenHeight = GetSystemMetrics(SM_CYSCREEN);
    
    HBITMAP hbmScreen = CreateCompatibleBitmap(hdcScreen, screenWidth, screenHeight);
    SelectObject(hdcMemDC, hbmScreen);
    
    BitBlt(hdcMemDC, 0, 0, screenWidth, screenHeight, hdcScreen, 0, 0, SRCCOPY);
    
    BITMAPINFOHEADER bi;
    bi.biSize = sizeof(BITMAPINFOHEADER);
    bi.biWidth = screenWidth;
    bi.biHeight = -screenHeight; // negative for top-down bitmap
    bi.biPlanes = 1;
    bi.biBitCount = 32;
    bi.biCompression = BI_RGB;
    bi.biSizeImage = 0;
    bi.biXPelsPerMeter = 0;
    bi.biYPelsPerMeter = 0;
    bi.biClrUsed = 0;
    bi.biClrImportant = 0;
    
    int dataSize = screenWidth * screenHeight * 4;
    BYTE* data = (BYTE*)malloc(dataSize);
    
    GetDIBits(hdcScreen, hbmScreen, 0, screenHeight, data, (BITMAPINFO*)&bi, DIB_RGB_COLORS);
    
    ScreenshotData* result = (ScreenshotData*)malloc(sizeof(ScreenshotData));
    result->data = data;
    result->width = screenWidth;
    result->height = screenHeight;
    result->size = dataSize;
    
    DeleteObject(hbmScreen);
    DeleteDC(hdcMemDC);
    ReleaseDC(NULL, hdcScreen);
    
    return result;
}

void freeScreenshot(ScreenshotData* screenshot) {
    if (screenshot) {
        if (screenshot->data) {
            free(screenshot->data);
        }
        free(screenshot);
    }
}
*/
import "C"
import (
    "bytes"
    "fmt"
    "image"
    "image/color"
    "image/png"
    "os"
    "path/filepath"
    "time"
    "unsafe"
)

type Screenshot struct {
    Width  int
    Height int
    Data   []byte
}

func TakeScreenshot() (*Screenshot, error) {
    cScreenshot := C.takeScreenshot()
    if cScreenshot == nil {
        return nil, fmt.Errorf("failed to take screenshot")
    }
    defer C.freeScreenshot(cScreenshot)
    
    width := int(cScreenshot.width)
    height := int(cScreenshot.height)
    size := int(cScreenshot.size)
    
    data := C.GoBytes(unsafe.Pointer(cScreenshot.data), C.int(size))
    
    screenshot := &Screenshot{
        Width:  width,
        Height: height,
        Data:   make([]byte, len(data)),
    }
    copy(screenshot.Data, data)
    
    return screenshot, nil
}

func (s *Screenshot) ToImage() *image.RGBA {
    img := image.NewRGBA(image.Rect(0, 0, s.Width, s.Height))
    
    for y := 0; y < s.Height; y++ {
        for x := 0; x < s.Width; x++ {
            offset := (y*s.Width + x) * 4
            if offset+3 < len(s.Data) {
                b := s.Data[offset]
                g := s.Data[offset+1]
                r := s.Data[offset+2]
                a := s.Data[offset+3]
                img.Set(x, y, color.RGBA{r, g, b, a})
            }
        }
    }
    
    return img
}

func (s *Screenshot) SaveToPNG(filename string) error {
    img := s.ToImage()
    
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    return png.Encode(file, img)
}

func (s *Screenshot) ToPNGBytes() ([]byte, error) {
    img := s.ToImage()
    
    var buf bytes.Buffer
    err := png.Encode(&buf, img)
    if err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func SaveScreenshotToFile() (string, error) {
    screenshot, err := TakeScreenshot()
    if err != nil {
        return "", err
    }
    
    timestamp := time.Now().Format("20060102_150405")
    filename := fmt.Sprintf("screenshot_%s.png", timestamp)
    
    err = screenshot.SaveToPNG(filename)
    if err != nil {
        return "", err
    }
    
    absPath, err := filepath.Abs(filename)
    if err != nil {
        return filename, nil
    }
    
    return absPath, nil
}