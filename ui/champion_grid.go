// skinhunter/ui/champion_grid.go
package ui

import (
	"log"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewChampionGrid crea la vista usando container.NewGridWrap.
// *** CORRECTION: Reverted to simple GridWrap version ***
func NewChampionGrid(champions []data.ChampionSummary, onChampionSelect func(champ data.ChampionSummary)) fyne.CanvasObject {
	log.Println("Creating Champion Grid UI (GridWrap)...")

	if champions == nil || len(champions) == 0 {
		log.Println("WARN: NewChampionGrid called with nil or empty champions slice.")
		return container.NewCenter(widget.NewLabel("No champions data available to display."))
	}

	log.Printf("Building champion grid UI for %d champions...", len(champions))
	cellWidth := float32(110)
	cellHeight := float32(145)
	cellSize := fyne.NewSize(cellWidth, cellHeight)
	imgTargetSize := fyne.NewSize(80, 80)

	championWidgets := make([]fyne.CanvasObject, 0, len(champions))

	// Build widgets directly from the provided data
	for _, champ := range champions {
		champCopy := champ // Capture range variable

		placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
		placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
		placeholderRect.SetMinSize(imgTargetSize)
		imageContainer := container.NewStack(placeholderRect, container.NewCenter(placeholderIcon))

		// Launch goroutine for image loading
		go func(c data.ChampionSummary, imgCont *fyne.Container) {
			imageURL := data.GetChampionSquarePortraitURL(c)
			if imageURL == data.GetPlaceholderImageURL() {
				return
			}
			uri, err := storage.ParseURI(imageURL)
			if err != nil {
				log.Printf("ERROR: ChampGrid failed to parse URI [%s] for champ %d: %v", imageURL, c.ID, err)
				return
			}
			loadedImage := canvas.NewImageFromURI(uri)
			loadedImage.FillMode = canvas.ImageFillContain
			loadedImage.SetMinSize(imgTargetSize)

			fyne.Do(func() { // Use fyne.Do for UI update
				if imgCont != nil && imgCont.Visible() {
					imgCont.Objects = []fyne.CanvasObject{loadedImage}
					imgCont.Refresh()
				}
			})
		}(champCopy, imageContainer)

		nameLabel := widget.NewLabel(champCopy.Name)
		nameLabel.Alignment = fyne.TextAlignCenter
		nameLabel.Truncation = fyne.TextTruncateEllipsis
		itemContent := container.NewVBox(container.NewCenter(imageContainer), nameLabel)

		// Use TappableCard from utils.go
		tappableCard := NewTappableCard(container.NewPadded(itemContent), func() {
			onChampionSelect(champCopy)
		})
		tappableCard.SetMinSize(cellSize)
		championWidgets = append(championWidgets, tappableCard)
	}

	grid := container.NewGridWrap(cellSize, championWidgets...)
	paddedGrid := container.NewPadded(grid)
	scrollContainer := container.NewScroll(paddedGrid)
	log.Printf("Champion grid UI built.")
	return scrollContainer
}

// --- Remove ChampionGridItem, NewChampionGridItemTemplate ---
// --- Remove TappableContainer, NewTappableContainer ---
// --- Remove approximateGridLayout, NewApproximateGridLayout ---

// --- End of champion_grid.go ---
