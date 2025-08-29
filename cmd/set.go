package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"godisplay/internal/display"
)

var setCmd = &cobra.Command{
	Use:   "set <display-id> <resolution>",
	Short: "Set display resolution",
	Long: `Set the resolution for a specific display.
    
Resolution can be specified as:
  - Mode number: 12
  - Resolution: 1920x1080
  - With refresh: 1920x1080@60
  - HiDPI mode: 1920x1080@2x (for Retina)
  
Examples:
  godisplay set 1 1920x1080
  godisplay set 1 1920x1080@120
  godisplay set 2 42  # Use mode number directly`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

func init() {
	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	// Parse display ID
	displayID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid display ID '%s': must be a positive number", args[0])
	}

	// Verify display exists
	displays, err := display.GetDisplays()
	if err != nil {
		return fmt.Errorf("failed to get displays: %w", err)
	}

	displayFound := false
	for _, d := range displays {
		if d.ID == uint32(displayID) {
			displayFound = true
			break
		}
	}

	if !displayFound {
		return fmt.Errorf("display %d not found", displayID)
	}

	// Get available modes
	modes, err := display.GetDisplayModes(uint32(displayID))
	if err != nil {
		return fmt.Errorf("failed to get modes for display %d: %w", displayID, err)
	}

	// Parse resolution spec
	modeNumber, err := parseResolutionSpec(args[1], modes)
	if err != nil {
		return fmt.Errorf("invalid resolution specification: %w", err)
	}

	// Safety check
	if viper.GetBool("safe_mode") {
		for _, m := range modes {
			if m.ModeNumber == modeNumber {
				if m.PixelWidth < 800 || m.PixelHeight < 600 {
					return fmt.Errorf("resolution too low (minimum 800x600 in safe mode)")
				}
				break
			}
		}
	}

	// Confirm if verbose
	if viper.GetBool("verbose") {
		fmt.Printf("Setting display %d to mode %d...\n", displayID, modeNumber)
	}

	// Apply the mode
	if err := display.SetDisplayMode(uint32(displayID), modeNumber); err != nil {
		return fmt.Errorf("failed to set display mode: %w", err)
	}

	fmt.Printf("Successfully changed display %d resolution\n", displayID)
	return nil
}

func parseResolutionSpec(spec string, modes []display.Mode) (int, error) {
	// Try direct mode number first
	if modeNum, err := strconv.Atoi(spec); err == nil {
		// Validate mode number exists
		for _, m := range modes {
			if m.ModeNumber == modeNum {
				return modeNum, nil
			}
		}
		return 0, fmt.Errorf("mode number %d not found", modeNum)
	}

	// Parse resolution format
	var width, height int
	var refreshRate float64 = 0
	var wantHiDPI bool

	// Check for @2x suffix for HiDPI
	if strings.HasSuffix(spec, "@2x") {
		wantHiDPI = true
		spec = strings.TrimSuffix(spec, "@2x")
	}

	// Parse WIDTHxHEIGHT[@REFRESH]
	parts := strings.Split(spec, "@")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid resolution format: %s", spec)
	}

	if len(parts) == 2 {
		r, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid refresh rate: %s", parts[1])
		}
		refreshRate = r
	}

	// Parse WIDTHxHEIGHT
	resParts := strings.Split(parts[0], "x")
	if len(resParts) != 2 {
		return 0, fmt.Errorf("invalid resolution format: %s", parts[0])
	}

	w, err := strconv.Atoi(resParts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid width: %s", resParts[0])
	}
	width = w

	h, err := strconv.Atoi(resParts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid height: %s", resParts[1])
	}
	height = h

	// Find matching mode
	var bestMode *display.Mode
	for i := range modes {
		m := &modes[i]

		// Check resolution match
		if m.PixelWidth != width || m.PixelHeight != height {
			continue
		}

		// Check HiDPI requirement
		if wantHiDPI && !m.IsHiDPI {
			continue
		}

		// Check refresh rate if specified
		if refreshRate > 0 {
			if int(m.RefreshRate) != int(refreshRate) {
				continue
			}
			// Exact match found
			return m.ModeNumber, nil
		}

		// Track best match (prefer higher refresh rates)
		if bestMode == nil || m.RefreshRate > bestMode.RefreshRate {
			bestMode = m
		}
	}

	if bestMode == nil {
		return 0, fmt.Errorf("no matching mode found for %s", spec)
	}

	return bestMode.ModeNumber, nil
}
