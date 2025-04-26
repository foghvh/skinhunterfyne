// skinhunter/ui/skins_grid.go
package ui

import (
	"log"
	"sync"
	"time"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	// "image/color" // No longer needed directly here
	// "fyne.io/fyne/v2/canvas" // No longer needed directly here
	// "fyne.io/fyne/v2/storage" // No longer needed directly here
	// "fyne.io/fyne/v2/theme" // No longer needed directly here
	// "fyne.io/fyne/v2/driver/desktop" // No longer needed directly here
	// "fyne.io/fyne/v2/layout" // No longer needed directly here
)

// SkinsGrid is a reusable widget for displaying a grid of skins using GridWrap.
type SkinsGrid struct {
	widget.BaseWidget
	onSkinSelect func(skin data.Skin)

	scroll        *container.Scroll
	gridContainer *fyne.Container // The GridWrap container
	cellSize      fyne.Size

	mu sync.Mutex
}

// NewSkinsGrid creates a new reusable SkinsGrid widget.
func NewSkinsGrid(onSkinSelect func(skin data.Skin)) *SkinsGrid {
	sg := &SkinsGrid{
		onSkinSelect: onSkinSelect,
		cellSize:     fyne.NewSize(210, 200),
	}
	sg.ExtendBaseWidget(sg)
	sg.gridContainer = container.NewGridWrap(sg.cellSize)
	paddedGrid := container.NewPadded(sg.gridContainer)
	sg.scroll = container.NewScroll(paddedGrid)
	return sg
}

// CreateRenderer returns the renderer for the SkinsGrid.
func (sg *SkinsGrid) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(sg.scroll)
}

// UpdateSkins clears the grid and populates it incrementally in the background.
func (sg *SkinsGrid) UpdateSkins(skins []data.Skin) {
	log.Printf("SkinsGrid: Updating with %d skins...", len(skins))
	sg.mu.Lock()
	if sg.gridContainer == nil {
		sg.mu.Unlock()
		return
	}
	sg.gridContainer.Objects = []fyne.CanvasObject{}
	sg.gridContainer.Refresh()
	sg.scroll.Refresh()
	sg.mu.Unlock()

	go func(skinsToProcess []data.Skin) {
		if skinsToProcess == nil {
			fyne.Do(func() { sg.showPlaceholder("No skins data available.") })
			return
		}
		nonBaseSkins := make([]data.Skin, 0, len(skinsToProcess))
		for _, s := range skinsToProcess {
			if !s.IsBase {
				nonBaseSkins = append(nonBaseSkins, s)
			}
		}
		if len(nonBaseSkins) == 0 {
			fyne.Do(func() { sg.showPlaceholder("This champion has no additional skins.") })
			return
		}

		log.Printf("SkinsGrid POPULATE: Starting item creation for %d skins...", len(nonBaseSkins))
		fyne.Do(func() { sg.showPlaceholder("Loading skins...") })

		for i, skin := range nonBaseSkins {
			skinCopy := skin
			// Create SkinItem which returns a TappableCard (defined in utils.go)
			item := SkinItem(skinCopy, func(selectedSkin data.Skin) {
				if sg.onSkinSelect != nil {
					sg.onSkinSelect(selectedSkin)
				}
			})
			if item != nil {
				fyne.Do(func() {
					sg.mu.Lock()
					defer sg.mu.Unlock()
					if sg.gridContainer != nil {
						if len(sg.gridContainer.Objects) == 1 {
							if _, ok := sg.gridContainer.Objects[0].(*fyne.Container); ok {
								sg.gridContainer.Objects = []fyne.CanvasObject{}
							}
						}
						sg.gridContainer.Objects = append(sg.gridContainer.Objects, item)
						sg.gridContainer.Refresh() // Refresh grid on add
						if i > 0 && i%10 == 0 {
							sg.scroll.Refresh()
						} // Refresh scroll periodically
					}
				})
			}
			time.Sleep(20 * time.Millisecond)
		}
		fyne.Do(func() {
			if sg.scroll != nil {
				sg.scroll.Refresh()
			}
			log.Printf("SkinsGrid POPULATE: Population complete.")
		})
	}(skins)
}

// showPlaceholder replaces grid content with a centered message.
func (sg *SkinsGrid) showPlaceholder(text string) {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	if sg.gridContainer != nil {
		placeholder := container.NewCenter(widget.NewLabel(text))
		sg.gridContainer.Objects = []fyne.CanvasObject{placeholder}
		sg.gridContainer.Refresh()
		sg.scroll.Refresh()
	}
}

// --- REMOVE SkinItemWidget struct and related methods ---
// --- REMOVE TappableCard struct and related methods (defined in utils.go) ---

// --- End of skins_grid.go ---
