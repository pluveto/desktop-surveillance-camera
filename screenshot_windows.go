//go:build windows

package main

/*
#cgo LDFLAGS: -lgdi32 -luser32 -lshcore
#include <windows.h>
#include <wingdi.h>

typedef struct {
    BYTE* data;
    int width;
    int height;
    int size;
} ScreenshotData;

ScreenshotData* takeScreenshot() {
    // Set DPI awareness to get actual screen resolution
    SetProcessDPIAware();
    
    HDC hdcScreen = GetDC(NULL);
    HDC hdcMemDC = CreateCompatibleDC(hdcScreen);
    
    // Get actual screen dimensions (not DPI scaled)
    int screenWidth = GetSystemMetrics(SM_CXVIRTUALSCREEN);
    int screenHeight = GetSystemMetrics(SM_CYVIRTUALSCREEN);
    int screenX = GetSystemMetrics(SM_XVIRTUALSCREEN);
    int screenY = GetSystemMetrics(SM_YVIRTUALSCREEN);
    
    // If virtual screen metrics return 0, fall back to primary screen
    if (screenWidth == 0 || screenHeight == 0) {
        screenWidth = GetSystemMetrics(SM_CXSCREEN);
        screenHeight = GetSystemMetrics(SM_CYSCREEN);
        screenX = 0;
        screenY = 0;
    }
    
    HBITMAP hbmScreen = CreateCompatibleBitmap(hdcScreen, screenWidth, screenHeight);
    SelectObject(hdcMemDC, hbmScreen);
    
    BitBlt(hdcMemDC, 0, 0, screenWidth, screenHeight, hdcScreen, screenX, screenY, SRCCOPY);
    
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
    // Set DPI awareness to get actual screen resolution
    SetProcessDPIAware();
    
    HDC hdcScreen = GetDC(NULL);
    HDC hdcMemDC = CreateCompatibleDC(hdcScreen);
    
    int screenWidth = GetSystemMetrics(SM_CXVIRTUALSCREEN);
    int screenHeight = GetSystemMetrics(SM_CYVIRTUALSCREEN);
    int screenX = GetSystemMetrics(SM_XVIRTUALSCREEN);
    int screenY = GetSystemMetrics(SM_YVIRTUALSCREEN);
    
    // If virtual screen metrics return 0, fall back to primary screen
    if (screenWidth == 0 || screenHeight == 0) {
        screenWidth = GetSystemMetrics(SM_CXSCREEN);
        screenHeight = GetSystemMetrics(SM_CYSCREEN);
        screenX = 0;
        screenY = 0;
    }
    
    // Adjust coordinates for virtual screen offset
    x += screenX;
    y += screenY;
    
    // Validate and clamp coordinates
    if (x < screenX) x = screenX;
    if (y < screenY) y = screenY;
    if (x + width > screenX + screenWidth) width = screenX + screenWidth - x;
    if (y + height > screenY + screenHeight) height = screenY + screenHeight - y;
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

// Set text to clipboard
int setClipboardText(const char* text) {
    if (!OpenClipboard(NULL)) {
        return 0;
    }
    
    EmptyClipboard();
    
    int len = strlen(text);
    HGLOBAL hMem = GlobalAlloc(GMEM_MOVEABLE, len + 1);
    if (!hMem) {
        CloseClipboard();
        return 0;
    }
    
    char* pMem = (char*)GlobalLock(hMem);
    strcpy(pMem, text);
    GlobalUnlock(hMem);
    
    SetClipboardData(CF_TEXT, hMem);
    CloseClipboard();
    
    return 1;
}

// Simulate keyboard input
void simulateKeyPress(WORD vkCode) {
    INPUT input;
    input.type = INPUT_KEYBOARD;
    input.ki.wVk = vkCode;
    input.ki.wScan = 0;
    input.ki.dwFlags = 0;
    input.ki.time = 0;
    input.ki.dwExtraInfo = 0;
    
    // Key down
    SendInput(1, &input, sizeof(INPUT));
    
    // Key up
    input.ki.dwFlags = KEYEVENTF_KEYUP;
    SendInput(1, &input, sizeof(INPUT));
}

// Simulate Ctrl+V followed by Enter
void simulatePasteAndEnter() {
    INPUT inputs[4];
    
    // Ctrl down
    inputs[0].type = INPUT_KEYBOARD;
    inputs[0].ki.wVk = VK_CONTROL;
    inputs[0].ki.wScan = 0;
    inputs[0].ki.dwFlags = 0;
    inputs[0].ki.time = 0;
    inputs[0].ki.dwExtraInfo = 0;
    
    // V down
    inputs[1].type = INPUT_KEYBOARD;
    inputs[1].ki.wVk = 'V';
    inputs[1].ki.wScan = 0;
    inputs[1].ki.dwFlags = 0;
    inputs[1].ki.time = 0;
    inputs[1].ki.dwExtraInfo = 0;
    
    // V up
    inputs[2].type = INPUT_KEYBOARD;
    inputs[2].ki.wVk = 'V';
    inputs[2].ki.wScan = 0;
    inputs[2].ki.dwFlags = KEYEVENTF_KEYUP;
    inputs[2].ki.time = 0;
    inputs[2].ki.dwExtraInfo = 0;
    
    // Ctrl up
    inputs[3].type = INPUT_KEYBOARD;
    inputs[3].ki.wVk = VK_CONTROL;
    inputs[3].ki.wScan = 0;
    inputs[3].ki.dwFlags = KEYEVENTF_KEYUP;
    inputs[3].ki.time = 0;
    inputs[3].ki.dwExtraInfo = 0;
    
    // Send Ctrl+V
    SendInput(4, inputs, sizeof(INPUT));
    
    // Small delay
    Sleep(50);
    
    // Send Enter
    simulateKeyPress(VK_RETURN);
}

// Simulate mouse click at specified coordinates
void simulateMouseClick(int x, int y) {
    // Get screen dimensions for coordinate validation
    int screenWidth = GetSystemMetrics(SM_CXVIRTUALSCREEN);
    int screenHeight = GetSystemMetrics(SM_CYVIRTUALSCREEN);
    int screenX = GetSystemMetrics(SM_XVIRTUALSCREEN);
    int screenY = GetSystemMetrics(SM_YVIRTUALSCREEN);
    
    if (screenWidth == 0 || screenHeight == 0) {
        screenWidth = GetSystemMetrics(SM_CXSCREEN);
        screenHeight = GetSystemMetrics(SM_CYSCREEN);
        screenX = 0;
        screenY = 0;
    }
    
    // Adjust coordinates for virtual screen
    x += screenX;
    y += screenY;
    
    // Validate coordinates
    if (x < screenX || x >= screenX + screenWidth || y < screenY || y >= screenY + screenHeight) {
        return; // Invalid coordinates
    }
    
    // Set cursor position
    SetCursorPos(x, y);
    
    // Small delay to ensure cursor position is set
    Sleep(10);
    
    INPUT inputs[2];
    
    // Mouse down
    inputs[0].type = INPUT_MOUSE;
    inputs[0].mi.dx = 0;
    inputs[0].mi.dy = 0;
    inputs[0].mi.mouseData = 0;
    inputs[0].mi.dwFlags = MOUSEEVENTF_LEFTDOWN;
    inputs[0].mi.time = 0;
    inputs[0].mi.dwExtraInfo = 0;
    
    // Mouse up
    inputs[1].type = INPUT_MOUSE;
    inputs[1].mi.dx = 0;
    inputs[1].mi.dy = 0;
    inputs[1].mi.mouseData = 0;
    inputs[1].mi.dwFlags = MOUSEEVENTF_LEFTUP;
    inputs[1].mi.time = 0;
    inputs[1].mi.dwExtraInfo = 0;
    
    SendInput(2, inputs, sizeof(INPUT));
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
    
    // Bilinear interpolation for better quality
    for y := 0; y < newHeight; y++ {
        for x := 0; x < newWidth; x++ {
            srcX := float64(x) / scale
            srcY := float64(y) / scale
            
            // Get the four surrounding pixels
            x1 := int(srcX)
            y1 := int(srcY)
            x2 := x1 + 1
            y2 := y1 + 1
            
            // Clamp coordinates
            if x1 < 0 { x1 = 0 }
            if y1 < 0 { y1 = 0 }
            if x2 >= originalWidth { x2 = originalWidth - 1 }
            if y2 >= originalHeight { y2 = originalHeight - 1 }
            
            // Calculate interpolation weights
            dx := srcX - float64(x1)
            dy := srcY - float64(y1)
            
            // Get the four pixels
            p11 := img.At(x1, y1)
            p12 := img.At(x1, y2)
            p21 := img.At(x2, y1)
            p22 := img.At(x2, y2)
            
            // Convert to RGBA
            r11, g11, b11, a11 := p11.RGBA()
            r12, g12, b12, a12 := p12.RGBA()
            r21, g21, b21, a21 := p21.RGBA()
            r22, g22, b22, a22 := p22.RGBA()
            
            // Perform bilinear interpolation
            r := (1-dx)*(1-dy)*float64(r11) + dx*(1-dy)*float64(r21) + (1-dx)*dy*float64(r12) + dx*dy*float64(r22)
            g := (1-dx)*(1-dy)*float64(g11) + dx*(1-dy)*float64(g21) + (1-dx)*dy*float64(g12) + dx*dy*float64(g22)
            b := (1-dx)*(1-dy)*float64(b11) + dx*(1-dy)*float64(b21) + (1-dx)*dy*float64(b12) + dx*dy*float64(b22)
            a := (1-dx)*(1-dy)*float64(a11) + dx*(1-dy)*float64(a21) + (1-dx)*dy*float64(a12) + dx*dy*float64(a22)
            
            // Convert back to 8-bit and set pixel
            resized.Set(x, y, color.RGBA{
                uint8(r / 257), // Convert from 16-bit to 8-bit
                uint8(g / 257),
                uint8(b / 257),
                uint8(a / 257),
            })
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

// SetClipboardText sets text to Windows clipboard
func SetClipboardText(text string) error {
    cText := C.CString(text)
    defer C.free(unsafe.Pointer(cText))
    
    result := C.setClipboardText(cText)
    if result == 0 {
        return fmt.Errorf("failed to set clipboard text")
    }
    
    return nil
}

// SimulatePasteAndEnter simulates Ctrl+V followed by Enter key press
func SimulatePasteAndEnter() {
    C.simulatePasteAndEnter()
}

// SimulateMouseClick simulates a mouse click at the specified screen coordinates
func SimulateMouseClick(x, y int) {
    C.simulateMouseClick(C.int(x), C.int(y))
}

// SendTextToClipboardAndPaste sends text to clipboard and simulates paste+enter
func SendTextToClipboardAndPaste(text string) error {
    err := SetClipboardText(text)
    if err != nil {
        return err
    }
    
    // Small delay to ensure clipboard is set
    time.Sleep(10 * time.Millisecond)
    
    SimulatePasteAndEnter()
    return nil
}