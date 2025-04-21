// skinhunter/ui/skins_grid.go
package ui

import (
	// "log"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewSkinsGrid creates a scrollable grid of skins for a specific champion.
func NewSkinsGrid(skins []data.Skin, onSkinSelect func(skin data.Skin)) fyne.CanvasObject {
	// *** Revert to 3 columns to match screenshot ***
	grid := container.NewGridWithColumns(3)

	if skins == nil || len(skins) == 0 {
		return container.NewCenter(widget.NewLabel("No skin data available for this champion."))
	}

	displayableSkinCount := 0
	for _, skin := range skins {
		// Pass the skin struct directly, SkinItem handles IsBase check
		item := SkinItem(skin, onSkinSelect) // SkinItem usa la lógica revertida
		if item != nil {
			grid.Add(item)
			displayableSkinCount++
		}
	}

	if displayableSkinCount == 0 {
		return container.NewCenter(widget.NewLabel("No alternate skins found."))
	}

	return container.NewScroll(grid)
}
