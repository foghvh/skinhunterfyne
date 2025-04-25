// skinhunter/ui/skins_grid.go
package ui

import (
	"log"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	// "fyne.io/fyne/v2/layout" // Not needed for GridWrap
	"fyne.io/fyne/v2/widget"
)

// NewSkinsGrid creates a grid of skins with lazy image loading,
// responsive columns using GridWrap, and padding for centering.
func NewSkinsGrid(skins []data.Skin, onSkinSelect func(skin data.Skin)) fyne.CanvasObject {

	if skins == nil {
		log.Println("WARN: NewSkinsGrid called with nil skins slice.")
		// Return a centered message immediately
		return container.NewCenter(widget.NewLabel("No skins data available."))
	}

	// Filter out base skins
	nonBaseSkins := make([]data.Skin, 0, len(skins))
	for _, s := range skins {
		if !s.IsBase {
			nonBaseSkins = append(nonBaseSkins, s)
		}
	}

	if len(nonBaseSkins) == 0 {
		log.Println("INFO: Champion has no non-base skins.")
		// Return a centered message immediately
		return container.NewCenter(widget.NewLabel("This champion has no additional skins."))
	}

	log.Printf("Building skins grid UI for %d non-base skins...", len(nonBaseSkins))

	// Define cell size for GridWrap layout
	// Corresponds to the size set in SkinItem's TappableCard
	cellWidth := float32(210)
	cellHeight := float32(200)
	cellSize := fyne.NewSize(cellWidth, cellHeight)

	// Create skin item widgets (using the modified SkinItem function)
	skinItems := make([]fyne.CanvasObject, 0, len(nonBaseSkins))
	for _, skin := range nonBaseSkins {
		skinCopy := skin // Capture range variable
		item := SkinItem(skinCopy, func(selectedSkin data.Skin) {
			// The onSelect passed to SkinItem already logs, just call the outer handler
			onSkinSelect(selectedSkin)
		})
		if item != nil {
			skinItems = append(skinItems, item)
		} else {
			log.Printf("WARN: SkinItem returned nil for skin %s (ID: %d)", skinCopy.Name, skinCopy.ID)
		}
	}

	if len(skinItems) == 0 {
		// Should not happen if nonBaseSkins was > 0, but safety check
		log.Println("ERROR: No valid skin items generated for skins grid.")
		return container.NewCenter(widget.NewLabel("Error displaying skins."))
	}

	log.Printf("Created %d skin item widgets for grid.", len(skinItems))

	// *** Use standard GridWrap Layout ***
	// The cellSize should match the MinSize set on the TappableCard in SkinItem
	grid := container.NewGridWrap(cellSize, skinItems...)

	// *** Wrap the grid in Padding for centering effect ***
	paddedGrid := container.NewPadded(grid)

	// *** Wrap the padded grid in a Scroll container ***
	// This scroll container is the final object returned
	// It allows scrolling if the grid content exceeds available vertical space
	scrollableGrid := container.NewScroll(paddedGrid)

	log.Printf("Skins grid UI built (using Padded GridWrap).")
	return scrollableGrid
}

// --- End of skins_grid.go ---
