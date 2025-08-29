package display

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework IOKit -framework Foundation
// #include "bridge.h"
import "C"
import (
	"fmt"
	"unsafe"
)

type Display struct {
	ID          uint32
	Width       int
	Height      int
	RefreshRate float64
	ScaleFactor int
	IsBuiltin   bool
	IsOnline    bool
	Name        string
	ModeNumber  int
}

type Mode struct {
	PixelWidth  int
	PixelHeight int
	Width       int
	Height      int
	RefreshRate float64
	IsHiDPI     bool
	IsNative    bool
	ModeNumber  int
}

// GetDisplays returns all active displays
func GetDisplays() ([]Display, error) {
	var count C.int
	cDisplays := C.get_displays(&count)
	if cDisplays == nil {
		return nil, fmt.Errorf("failed to get displays: no displays found or CoreGraphics error")
	}
	if count == 0 {
		C.free_displays(cDisplays, count)
		return nil, fmt.Errorf("no active displays found")
	}
	defer C.free_displays(cDisplays, count)

	// Convert C array to Go slice
	displays := make([]Display, int(count))
	cDisplaySlice := (*[1 << 30]C.DisplayInfo)(unsafe.Pointer(cDisplays))[:count:count]

	for i, cd := range cDisplaySlice {
		displays[i] = Display{
			ID:          uint32(cd.display_id),
			Width:       int(cd.width),
			Height:      int(cd.height),
			RefreshRate: float64(cd.refresh_rate),
			ScaleFactor: int(cd.scale_factor),
			IsBuiltin:   cd.is_builtin != 0,
			IsOnline:    cd.is_online != 0,
			Name:        C.GoString(cd.name),
			ModeNumber:  int(cd.mode_number),
		}
	}

	return displays, nil
}

// GetDisplayModes returns available modes for a specific display
func GetDisplayModes(displayID uint32) ([]Mode, error) {
	var count C.int
	cModes := C.get_display_modes(C.uint32_t(displayID), &count)
	if cModes == nil {
		return nil, fmt.Errorf("failed to get display modes for display %d", displayID)
	}
	defer C.free_display_modes(cModes)

	modes := make([]Mode, int(count))
	cModeSlice := (*[1 << 30]C.DisplayMode)(unsafe.Pointer(cModes))[:count:count]

	for i, cm := range cModeSlice {
		modes[i] = Mode{
			PixelWidth:  int(cm.pixel_width),
			PixelHeight: int(cm.pixel_height),
			Width:       int(cm.width),
			Height:      int(cm.height),
			RefreshRate: float64(cm.refresh_rate),
			IsHiDPI:     cm.is_hidpi != 0,
			IsNative:    cm.is_native != 0,
			ModeNumber:  int(cm.mode_number),
		}
	}

	return modes, nil
}

// SetDisplayMode changes the resolution of a display
func SetDisplayMode(displayID uint32, modeNumber int) error {
	if modeNumber < 0 {
		return fmt.Errorf("invalid mode number: %d", modeNumber)
	}

	result := C.set_display_mode(C.uint32_t(displayID), C.int(modeNumber))
	if result != 0 {
		switch result {
		case -1:
			return fmt.Errorf("mode %d not found for display %d", modeNumber, displayID)
		case 1000:
			return fmt.Errorf("invalid display ID: %d", displayID)
		case 1001:
			return fmt.Errorf("permission denied - may need to grant screen recording permission")
		default:
			return fmt.Errorf("failed to set display mode: CoreGraphics error %d", result)
		}
	}
	return nil
}

// Additional utility functions
func (d Display) IsRetina() bool {
	return d.ScaleFactor > 1
}

func (m Mode) Resolution() string {
	if m.IsHiDPI {
		return fmt.Sprintf("%dx%d HiDPI (%dx%d pixels)",
			m.Width, m.Height, m.PixelWidth, m.PixelHeight)
	}
	return fmt.Sprintf("%dx%d", m.PixelWidth, m.PixelHeight)
}

func (m Mode) AspectRatio() string {
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}

	divisor := gcd(m.PixelWidth, m.PixelHeight)
	return fmt.Sprintf("%d:%d", m.PixelWidth/divisor, m.PixelHeight/divisor)
}
