// skinhunter/ui/champion_grid.go
package ui

import (
	"image/color" // Necesario para color Transparente
	"log"
	"strings"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	// layout no es estrictamente necesario aquí ahora
)

// NewChampionGrid: Prueba con Stack + Rectangle Transparente
func NewChampionGrid(onChampionSelect func(champ data.ChampionSummary)) fyne.CanvasObject {
	const cdragonBase = "https://raw.communitydragon.org/latest"
	log.Println("Creating Champion Grid (Sync Fetch, Testing Stack + Rect)...")

	champions, err := data.FetchAllChampions()
	// ... (manejo de error y lista vacía igual que antes) ...
	if err != nil { /* ... */
		return container.NewCenter(widget.NewLabel("..."))
	}
	if len(champions) == 0 { /* ... */
		return container.NewCenter(widget.NewLabel("..."))
	}

	log.Printf("Building grid with Stack + Rect for %d champions...", len(champions))
	var championWidgets []fyne.CanvasObject
	imgTargetSize := fyne.NewSize(80, 80)
	cellSize := fyne.NewSize(100, 130)

	for _, champ := range champions {
		var imageWidget fyne.CanvasObject

		// ... (lógica para crear imageWidget con NewImageFromURI o placeholder, igual que antes) ...
		rawPath := champ.SquarePortraitPath
		if rawPath == "" { /* ... placeholder ... */
			placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
			placeholderCont := container.NewStack()
			placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
			placeholderRect.SetMinSize(imgTargetSize)
			placeholderCont.Add(placeholderRect)
			placeholderCont.Add(container.NewCenter(placeholderIcon))
			imageWidget = placeholderCont
		} else { /* ... construir URL y cargar imagen ... */
			fixedPath := cdragonBase + strings.Replace(rawPath, "/lol-game-data/assets", "/plugins/rcp-be-lol-game-data/global/default", 1)
			uri, err := storage.ParseURI(fixedPath)
			if err != nil { /* ... placeholder error ... */
				placeholderIcon := widget.NewIcon(theme.ErrorIcon())
				placeholderCont := container.NewStack()
				placeholderRect := canvas.NewRectangle(theme.ErrorColor())
				placeholderRect.SetMinSize(imgTargetSize)
				placeholderCont.Add(placeholderRect)
				placeholderCont.Add(container.NewCenter(placeholderIcon))
				imageWidget = placeholderCont
			} else { /* ... cargar imagen ... */
				img := canvas.NewImageFromURI(uri)
				img.FillMode = canvas.ImageFillContain
				img.SetMinSize(imgTargetSize)
				imageWidget = img
			}
		}

		// --- Crear VBox con imagen y nombre (igual que antes) ---
		nameLabel := widget.NewLabel(champ.Name)
		nameLabel.Alignment = fyne.TextAlignCenter
		nameLabel.Truncation = fyne.TextTruncateEllipsis
		itemContent := container.NewVBox(imageWidget, nameLabel)

		// *** CAMBIO: Usar Rectangle transparente en lugar de Button ***
		overlayRect := canvas.NewRectangle(color.Transparent)
		// Añadir un Tapped handler al rectángulo (requiere un widget custom o manejar eventos de forma diferente)
		// Para simplificar la prueba, de momento no tendrá click.
		// Si esto funciona, podemos hacer un widget TappableRectangle.

		// *** Usar Stack con itemContent y overlayRect ***
		itemWidget := container.NewStack(itemContent, overlayRect) // Rectángulo encima

		championWidgets = append(championWidgets, itemWidget)
	}

	grid := container.NewGridWrap(cellSize, championWidgets...)
	scrollContainer := container.NewScroll(grid)
	log.Printf("Grid built with Stack + Rect. Check if images load.", len(championWidgets))

	return scrollContainer
}
