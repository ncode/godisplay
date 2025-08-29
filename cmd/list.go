package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"godisplay/internal/display"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	showAll       bool
	displayID     uint32
	groupByAspect bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List displays and available resolutions",
	Long:  `List all connected displays and their available resolution modes.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&showAll, "all", "a", false,
		"show all modes including duplicates")
	listCmd.Flags().Uint32VarP(&displayID, "display", "d", 0,
		"show modes for specific display ID")
	listCmd.Flags().BoolVarP(&groupByAspect, "group", "g", false,
		"group resolutions by aspect ratio")
}

func runList(cmd *cobra.Command, args []string) error {
	displays, err := display.GetDisplays()
	if err != nil {
		return fmt.Errorf("failed to get displays: %w", err)
	}

	if viper.GetBool("json") {
		return outputJSON(displays)
	}

	// Text output
	for _, d := range displays {
		if displayID != 0 && d.ID != displayID {
			continue
		}

		fmt.Printf("\nDisplay %d:\n", d.ID)
		fmt.Printf("  Status: %s\n", getDisplayStatus(d))
		fmt.Printf("  Current: %dx%d @ %.0fHz\n", d.Width, d.Height, d.RefreshRate)
		if d.IsRetina() {
			fmt.Printf("  Type: Retina (HiDPI %dx scale)\n", d.ScaleFactor)
		}

		modes, err := display.GetDisplayModes(d.ID)
		if err != nil {
			fmt.Printf("  Error getting modes: %v\n", err)
			continue
		}

		// Sort and deduplicate modes
		modes = processModes(modes)

		if groupByAspect {
			printModesGrouped(modes, d.ModeNumber)
		} else {
			printModes(modes, d.ModeNumber)
		}
	}

	return nil
}

func getDisplayStatus(d display.Display) string {
	status := []string{}
	if d.IsOnline {
		status = append(status, "Online")
	} else {
		status = append(status, "Offline")
	}
	if d.IsBuiltin {
		status = append(status, "Built-in")
	} else {
		status = append(status, "External")
	}
	return strings.Join(status, ", ")
}

func processModes(modes []display.Mode) []display.Mode {
	// Remove duplicates unless --all flag is set
	if !showAll {
		seen := make(map[string]bool)
		filtered := []display.Mode{}

		for _, m := range modes {
			key := fmt.Sprintf("%dx%d@%.0f", m.PixelWidth, m.PixelHeight, m.RefreshRate)
			if !seen[key] {
				seen[key] = true
				filtered = append(filtered, m)
			}
		}
		modes = filtered
	}

	// Sort by resolution (descending) then refresh rate
	sort.Slice(modes, func(i, j int) bool {
		if modes[i].PixelWidth != modes[j].PixelWidth {
			return modes[i].PixelWidth > modes[j].PixelWidth
		}
		if modes[i].PixelHeight != modes[j].PixelHeight {
			return modes[i].PixelHeight > modes[j].PixelHeight
		}
		return modes[i].RefreshRate > modes[j].RefreshRate
	})

	return modes
}

func printModes(modes []display.Mode, currentModeNumber int) {
	fmt.Println("\n  Available modes:")
	for _, m := range modes {
		prefix := "   "
		if m.ModeNumber == currentModeNumber {
			prefix = " * " // Current resolution
		} else if m.IsNative {
			prefix = " ✓ " // Native resolution
		}

		hidpi := ""
		if m.IsHiDPI {
			hidpi = " ⚡" // Lightning bolt for HiDPI, like RDM
		}

		fmt.Printf("%s [%d] %dx%d @ %.0fHz%s (%s)\n", prefix, m.ModeNumber, m.PixelWidth, m.PixelHeight, m.RefreshRate, hidpi, m.AspectRatio())
	}
}

func printModesGrouped(modes []display.Mode, currentModeNumber int) {
	grouped := make(map[string][]display.Mode)

	for _, m := range modes {
		ratio := m.AspectRatio()
		grouped[ratio] = append(grouped[ratio], m)
	}

	fmt.Println("\n  Available modes by aspect ratio:")
	for ratio, groupModes := range grouped {
		fmt.Printf("\n  %s:\n", ratio)
		for _, m := range groupModes {
			prefix := "   "
			if m.ModeNumber == currentModeNumber {
				prefix = " * "
			} else if m.IsNative {
				prefix = " ✓ "
			}

			hidpi := ""
			if m.IsHiDPI {
				hidpi = " ⚡"
			}

			fmt.Printf("%s [%d] %dx%d @ %.0fHz%s\n", prefix, m.ModeNumber, m.PixelWidth, m.PixelHeight, m.RefreshRate, hidpi)
		}
	}
}

func outputJSON(displays []display.Display) error {
	type JSONOutput struct {
		Displays []display.Display         `json:"displays"`
		Modes    map[uint32][]display.Mode `json:"modes,omitempty"`
	}

	output := JSONOutput{
		Displays: displays,
		Modes:    make(map[uint32][]display.Mode),
	}

	// If specific display requested, include modes
	if displayID != 0 {
		modes, err := display.GetDisplayModes(displayID)
		if err == nil {
			output.Modes[displayID] = processModes(modes)
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
