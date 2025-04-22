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

// NewChampionGrid crea la vista de cuadrícula de campeones CENTRADA y responsive.
func NewChampionGrid(onChampionSelect func(champ data.ChampionSummary)) fyne.CanvasObject {
	log.Println("Creating Centered Champion Grid (Async Load)...")
	gridContainer := container.NewMax()
	loadingIndicator := container.NewCenter(container.NewVBox(widget.NewLabel("Loading Champions..."), widget.NewProgressBarInfinite()))
	gridContainer.Add(loadingIndicator)

	go func() {
		champions, err := data.FetchAllChampions()
		var finalContent fyne.CanvasObject

		if err != nil { /* manejo error */
			log.Printf("ERROR: ...")
			finalContent = container.NewCenter(widget.NewLabel("..."))
		} else if len(champions) == 0 { /* manejo vacío */
			log.Println("WARN: ...")
			finalContent = container.NewCenter(widget.NewLabel("..."))
		} else {
			log.Printf("Building centered grid UI for %d champions...", len(champions))
			championWidgets := make([]fyne.CanvasObject, 0, len(champions))

			cellWidth := float32(115)
			cellHeight := float32(135)
			cellSize := fyne.NewSize(cellWidth, cellHeight)
			imgTargetSize := fyne.NewSize(80, 80)

			for _, champ := range champions {
				// ... (creación de imageWidget, nameLabel, itemContent, tappableCard) ...
				var imageWidget fyne.CanvasObject
				imageURL := data.GetChampionSquarePortraitURL(champ)
				uri, parseErr := storage.ParseURI(imageURL)
				if parseErr != nil || imageURL == data.GetPlaceholderImageURL() {
					placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
					placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
					placeholderRect.SetMinSize(imgTargetSize)
					imageWidget = container.NewStack(placeholderRect, container.NewCenter(placeholderIcon))
					imageWidget.Refresh()
				} else {
					img := canvas.NewImageFromURI(uri)
					img.FillMode = canvas.ImageFillContain
					img.SetMinSize(imgTargetSize)
					imageWidget = img
				}
				nameLabel := widget.NewLabel(champ.Name)
				nameLabel.Alignment = fyne.TextAlignCenter
				nameLabel.Truncation = fyne.TextTruncateEllipsis
				itemContent := container.NewVBox(container.NewCenter(imageWidget), nameLabel)
				champCopy := champ
				tappableCard := NewTappableCard(itemContent, func() { log.Printf("..."); onChampionSelect(champCopy) })
				championWidgets = append(championWidgets, tappableCard)
			}

			// *** Usar el CONTENEDOR con LAYOUT PERSONALIZADO ***
			grid := NewCenteredGridWrap(cellSize, championWidgets...)

			// Envolver directamente en Scroll
			scrollContainer := container.NewScroll(grid)
			finalContent = scrollContainer
			log.Printf("Centered champion grid UI built and ready.")
		}

		// Actualizar UI directamente
		if gridContainer != nil {
			gridContainer.Objects = []fyne.CanvasObject{finalContent}
			gridContainer.Refresh()
			log.Printf("Champion grid container updated with final content.")
		} else {
			log.Println("ERROR: gridContainer is nil...")
		}
	}()
	return gridContainer
}
