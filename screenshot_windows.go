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

ScreenshotData* takeRegionScreenshot(int x, int y, int width, int height) {
    HDC hdcScreen = GetDC(NULL);
    HDC hdcMemDC = CreateCompatibleDC(hdcScreen);
    
    int screenWidth = GetSystemMetrics(SM_CXSCREEN);
    int screenHeight = GetSystemMetrics(SM_CYSCREEN);
    
    // Validate and clamp coordinates
    if (x < 0) x = 0;
    if (y < 0) y = 0;
    if (x + width > screenWidth) width = screenWidth - x;
    if (y + height > screenHeight) height = screenHeight - y;
    if (width <= 0 || height <= 0) {
        ReleaseDC(NULL, hdcScreen);
        DeleteDC(hdcMemDC);
        return NULL;
    }
    
    HBITMAP hbmScreen = CreateCompatibleBitmap(hdcScreen, width, height);
    SelectObject(hdcMemDC, hbmScreen);
    
    BitBlt(hdcMemDC, 0, 0, width, height, hdcScreen, x, y, SRCCOPY);
    
    BITMAPINFOHEADER bi;
    bi.biSize = sizeof(BITMAPINFOHEADER);
    bi.biWidth = width;
    bi.biHeight = -height; // negative for top-down bitmap
    bi.biPlanes = 1;
    bi.biBitCount = 32;
    bi.biCompression = BI_RGB;
    bi.biSizeImage = 0;
    bi.biXPelsPerMeter = 0;
    bi.biYPelsPerMeter = 0;
    bi.biClrUsed = 0;
    bi.biClrImportant = 0;
    
    int dataSize = width * height * 4;
    BYTE* data = (BYTE*)malloc(dataSize);
    
    GetDIBits(hdcScreen, hbmScreen, 0, height, data, (BITMAPINFO*)&bi, DIB_RGB_COLORS);
    
    ScreenshotData* result = (ScreenshotData*)malloc(sizeof(ScreenshotData));
    result->data = data;
    result->width = width;
    result->height = height;
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
    Region *ScreenRegion // nil for full screen
}

type ScreenshotOptions struct {
    Region     *ScreenRegion
    Compress   bool
    MaxWidth   int
    MaxHeight  int
    Quality    int // 1-100, only for JPEG (not used for PNG but kept for future)
}

func TakeScreenshot() (*Screenshot, error) {
    return TakeScreenshotWithOptions(&ScreenshotOptions{})
}

func TakeScreenshotWithOptions(opts *ScreenshotOptions) (*Screenshot, error) {
    var cScreenshot *C.ScreenshotData
    
    if opts.Region != nil {
        cScreenshot = C.takeRegionScreenshot(
            C.int(opts.Region.X),
            C.int(opts.Region.Y), 
            C.int(opts.Region.Width),
            C.int(opts.Region.Height),
        )
    } else {
        cScreenshot = C.takeScreenshot()
    }
    
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
        Region: opts.Region,
    }
    copy(screenshot.Data, data)
    
    return screenshot, nil
}

func TakeRegionScreenshot(x, y, width, height int) (*Screenshot, error) {
    region := &ScreenRegion{
        X:      x,
        Y:      y,
        Width:  width,
        Height: height,
    }
    
    return TakeScreenshotWithOptions(&ScreenshotOptions{
        Region: region,
    })
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

func (s *Screenshot) ToCompressedImage(maxWidth, maxHeight int) image.Image {
    img := s.ToImage()
    
    if maxWidth <= 0 && maxHeight <= 0 {
        return img
    }
    
    bounds := img.Bounds()
    originalWidth := bounds.Dx()
    originalHeight := bounds.Dy()
    
    if maxWidth <= 0 {
        maxWidth = originalWidth
    }
    if maxHeight <= 0 {
        maxHeight = originalHeight
    }
    
    // No compression needed if image is already smaller
    if originalWidth <= maxWidth && originalHeight <= maxHeight {
        return img
    }
    
    // Calculate aspect ratio preserving dimensions
    scaleX := float64(maxWidth) / float64(originalWidth)
    scaleY := float64(maxHeight) / float64(originalHeight)
    scale := scaleX
    if scaleY < scaleX {
        scale = scaleY
    }
    
    newWidth := int(float64(originalWidth) * scale)
    newHeight := int(float64(originalHeight) * scale)
    
    // Create new image with calculated dimensions
    resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
    
    // Simple nearest neighbor scaling
    for y := 0; y < newHeight; y++ {
        for x := 0; x < newWidth; x++ {
            srcX := int(float64(x) / scale)
            srcY := int(float64(y) / scale)
            if srcX < originalWidth && srcY < originalHeight {
                resized.Set(x, y, img.At(srcX, srcY))
            }
        }
    }
    
    return resized
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

func (s *Screenshot) ToPNGBytesWithOptions(opts *ScreenshotOptions) ([]byte, error) {
    var img image.Image
    
    if opts != nil && opts.Compress && (opts.MaxWidth > 0 || opts.MaxHeight > 0) {
        img = s.ToCompressedImage(opts.MaxWidth, opts.MaxHeight)
    } else {
        img = s.ToImage()
    }
    
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