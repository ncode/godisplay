
// internal/display/bridge.c
#include <Foundation/Foundation.h>
#include <CoreGraphics/CoreGraphics.h>
#include <IOKit/graphics/IOGraphicsLib.h>
#include <stdlib.h>


typedef struct {
    uint32_t display_id;
    int width;
    int height;
    double refresh_rate;
    int scale_factor;  // 1 for normal, 2 for HiDPI
    int is_builtin;
    int is_online;
    char* name;
    int mode_number;
} DisplayInfo;

typedef struct {
    int pixel_width;
    int pixel_height;
    int width;  // points
    int height; // points
    double refresh_rate;
    int is_hidpi;
    int is_native;
    int mode_number;
} DisplayMode;

// Get all active displays
DisplayInfo* get_displays(int* count) {
    CGDirectDisplayID displays[32];
    uint32_t displayCount;

    if (CGGetActiveDisplayList(32, displays, &displayCount) != kCGErrorSuccess) {
        *count = 0;
        return NULL;
    }

    DisplayInfo* result = (DisplayInfo*)calloc(displayCount, sizeof(DisplayInfo));

    for (int i = 0; i < displayCount; i++) {
        result[i].display_id = displays[i];
        result[i].width = CGDisplayPixelsWide(displays[i]);
        result[i].height = CGDisplayPixelsHigh(displays[i]);

        CGDisplayModeRef mode = CGDisplayCopyDisplayMode(displays[i]);
        if (mode) {
            result[i].refresh_rate = CGDisplayModeGetRefreshRate(mode);
            if (result[i].refresh_rate == 0) {
                result[i].refresh_rate = 60.0; // Default for displays that don't report
            }

            // Check if HiDPI - calculate from pixel vs point dimensions
            result[i].scale_factor = 1;
            size_t width = CGDisplayModeGetWidth(mode);
            size_t pixelWidth = CGDisplayModeGetPixelWidth(mode);
            if (pixelWidth > 0 && width > 0 && pixelWidth != width) {
                result[i].scale_factor = pixelWidth / width;
            }
            result[i].mode_number = CGDisplayModeGetIODisplayModeID(mode);
            CGDisplayModeRelease(mode);
        }

        result[i].is_builtin = CGDisplayIsBuiltin(displays[i]);
        result[i].is_online = CGDisplayIsOnline(displays[i]);

        // Get display name from IOKit
        io_iterator_t it = 0;
        io_service_t service;
        CFDictionaryRef info;

        if (IOServiceGetMatchingServices(kIOMainPortDefault,
                                         IOServiceMatching("IODisplayConnect"),
                                         &it) == KERN_SUCCESS) {
            while ((service = IOIteratorNext(it)) != 0) {
                info = IODisplayCreateInfoDictionary(service, kIODisplayOnlyPreferredName);
                if (info) {
                    NSDictionary* dict = (__bridge NSDictionary*)info;
                    NSDictionary* names = [dict objectForKey:@(kDisplayProductName)];
                    if (names && [names count] > 0) {
                        NSString* name = [names objectForKey:[[names allKeys] objectAtIndex:0]];
                        if (!result[i].name) {  // Only set if not already set
                            result[i].name = strdup([name UTF8String]);
                        }
                    }
                    CFRelease(info);
                }
                IOObjectRelease(service);
                if (result[i].name) {
                    break;  // Found a name for this display, stop searching
                }
            }
        }

        // Always release the iterator, even on failure
        if (it != 0) {
            IOObjectRelease(it);
        }

        if (!result[i].name) {
            result[i].name = strdup("Unknown Display");
        }
    }

    *count = displayCount;
    return result;
}

// Get available modes for a display
DisplayMode* get_display_modes(uint32_t display_id, int* count) {
    CFArrayRef modes = CGDisplayCopyAllDisplayModes(display_id, NULL);
    if (!modes) {
        *count = 0;
        return NULL;
    }

    CFIndex modeCount = CFArrayGetCount(modes);
    DisplayMode* result = (DisplayMode*)calloc(modeCount, sizeof(DisplayMode));

    for (CFIndex i = 0; i < modeCount; i++) {
        CGDisplayModeRef mode = (CGDisplayModeRef)CFArrayGetValueAtIndex(modes, i);

        result[i].pixel_width = CGDisplayModeGetPixelWidth(mode);
        result[i].pixel_height = CGDisplayModeGetPixelHeight(mode);
        result[i].width = CGDisplayModeGetWidth(mode);
        result[i].height = CGDisplayModeGetHeight(mode);
        result[i].refresh_rate = CGDisplayModeGetRefreshRate(mode);
        result[i].mode_number = CGDisplayModeGetIODisplayModeID(mode);

        // Detect HiDPI
        result[i].is_hidpi = (result[i].pixel_width > result[i].width) ? 1 : 0;

        // Check if native
        uint32_t flags = CGDisplayModeGetIOFlags(mode);
        result[i].is_native = (flags & kDisplayModeNativeFlag) ? 1 : 0;
    }

    CFRelease(modes);
    *count = modeCount;
    return result;
}

// Set display resolution
int set_display_mode(uint32_t display_id, int mode_number) {
    CFArrayRef modes = CGDisplayCopyAllDisplayModes(display_id, NULL);
    if (!modes) return -1;

    CGDisplayModeRef targetMode = NULL;
    CFIndex modeCount = CFArrayGetCount(modes);

    for (CFIndex i = 0; i < modeCount; i++) {
        CGDisplayModeRef mode = (CGDisplayModeRef)CFArrayGetValueAtIndex(modes, i);
        if (CGDisplayModeGetIODisplayModeID(mode) == mode_number) {
            targetMode = mode;
            break;
        }
    }

    if (!targetMode) {
        CFRelease(modes);
        return -1;
    }

    CGDisplayConfigRef config;
    CGError err = CGBeginDisplayConfiguration(&config);
    if (err != kCGErrorSuccess) {
        CFRelease(modes);
        return err;
    }

    err = CGConfigureDisplayWithDisplayMode(config, display_id, targetMode, NULL);
    if (err != kCGErrorSuccess) {
        CGCancelDisplayConfiguration(config);
        CFRelease(modes);
        return err;
    }

    err = CGCompleteDisplayConfiguration(config, kCGConfigureForSession);
    CFRelease(modes);

    return err;
}

// Cleanup functions
void free_displays(DisplayInfo* displays, int count) {
    for (int i = 0; i < count; i++) {
        if (displays[i].name) {
            free(displays[i].name);
        }
    }
    free(displays);
}

void free_display_modes(DisplayMode* modes) {
    free(modes);
}
