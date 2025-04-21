// skinhunter/ui/skins_grid.go
package ui

import (
	// "log" // Uncomment if needed for debugging

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewSkinsGrid creates a scrollable grid of skins for a specific champion.
// Takes skins directly, as they are expected to be fetched by the caller (NewChampionView).
func NewSkinsGrid(skins []data.Skin, onSkinSelect func(skin data.Skin)) fyne.CanvasObject {
	// Adjust columns based on SkinItem size (~200px wide + padding)
	// Aim for 3-4 columns typically. Let's use 3.
	grid := container.NewGridWithColumns(3)

	if skins == nil || len(skins) == 0 {
		// log.Println("No skins data provided to NewSkinsGrid")
		return container.NewCenter(widget.NewLabel("No skins found for this champion.")) // Handle nil/empty slice
	}

	foundDisplayableSkins := false
	for _, skin := range skins {
		// SkinItem handles the IsBase check internally now and returns nil if base
		item := SkinItem(skin, onSkinSelect)
		if item != nil { // Only add non-base skins returned by SkinItem
			grid.Add(item)
			foundDisplayableSkins = true
		}
	}

	// If only base skins existed, the grid will be empty
	if !foundDisplayableSkins {
		// log.Printf("Only base skin found for champion (or skins slice was empty/nil initially).")
		return container.NewCenter(widget.NewLabel("No alternate skins available."))
	}

	// Add scroll only if there are items in the grid
	return container.NewScroll(grid)
}
