// skinhunter/ui/champion_grid.go
package ui

import (
	"fmt"
	"log"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewChampionGrid creates the champion grid view with lazy image loading
// and padding for centering effect.
func NewChampionGrid(onChampionSelect func(champ data.ChampionSummary)) fyne.CanvasObject {
	log.Println("Creating Champion Grid (Lazy Load, Padded GridWrap)...")
	// Main container that will be updated once data is loaded
	gridContainer := container.NewMax()

	// Initial loading indicator
	loadingIndicator := container.NewCenter(container.NewVBox(widget.NewLabel("Loading Champions..."), widget.NewProgressBarInfinite()))
	gridContainer.Add(loadingIndicator)

	// Fetch champions in background
	go func() {
		champions, err := data.FetchAllChampions()
		var finalContent fyne.CanvasObject

		if err != nil {
			log.Printf("ERROR fetching champions: %v", err)
			errorMsg := fmt.Sprintf("Failed to load champions:\n%v", err)
			finalContent = container.NewCenter(widget.NewLabel(errorMsg))
		} else if len(champions) == 0 {
			log.Println("WARN: No champions found after fetching.")
			finalContent = container.NewCenter(widget.NewLabel("No champions available."))
		} else {
			log.Printf("Building champion grid UI for %d champions...", len(champions))

			// Define cell size for GridWrap layout
			// Adjust these values to control item spacing and number of columns
			cellWidth := float32(110)  // Slightly wider for padding inside TappableCard
			cellHeight := float32(145) // Taller to fit image + label comfortably
			cellSize := fyne.NewSize(cellWidth, cellHeight)
			imgTargetSize := fyne.NewSize(80, 80) // Size for the champion portrait itself

			championWidgets := make([]fyne.CanvasObject, 0, len(champions))

			for _, champ := range champions {
				champCopy := champ // Capture range variable for closure

				// --- Placeholder & Image Container ---
				placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
				placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
				placeholderRect.SetMinSize(imgTargetSize)
				imageContainer := container.NewStack(placeholderRect, container.NewCenter(placeholderIcon))
				imageContainer.Refresh()

				// --- Asynchronous Image Loading ---
				go func(c data.ChampionSummary, imgContainer *fyne.Container) {
					imageURL := data.GetChampionSquarePortraitURL(c)
					if imageURL == data.GetPlaceholderImageURL() {
						// log.Printf("ChampGrid: Skipping load for placeholder URL (Champ %d: %s)", c.ID, c.Name)
						return // No need to load placeholder
					}

					uri, err := storage.ParseURI(imageURL)
					if err != nil {
						log.Printf("ERROR: ChampGrid failed to parse URI [%s] for champ %d: %v", imageURL, c.ID, err)
						// Optionally update placeholder icon to indicate error
						return
					}

					loadedImage := canvas.NewImageFromURI(uri)
					loadedImage.FillMode = canvas.ImageFillContain
					loadedImage.SetMinSize(imgTargetSize)

					// Update the UI safely
					if imgContainer != nil {
						imgContainer.Objects = []fyne.CanvasObject{loadedImage}
						imgContainer.Refresh()
						// log.Printf("ChampGrid: Loaded image for champ %d: %s", c.ID, c.Name)
					} else {
						log.Printf("WARN: ChampGrid imgContainer nil for champ %d", c.ID)
					}
				}(champCopy, imageContainer) // Pass copy and container to goroutine

				// --- Name Label ---
				nameLabel := widget.NewLabel(champCopy.Name)
				nameLabel.Alignment = fyne.TextAlignCenter
				nameLabel.Truncation = fyne.TextTruncateEllipsis // Shorten long names

				// --- Assemble Item Content (Image + Label) ---
				itemContent := container.NewVBox(
					container.NewCenter(imageContainer), // Center the image container
					nameLabel,
				)

				// --- Tappable Card ---
				// Padding added *inside* the tappable card for visual spacing
				tappableCard := NewTappableCard(container.NewPadded(itemContent), func() {
					log.Printf("Champion selected: %s (ID: %d)", champCopy.Name, champCopy.ID)
					onChampionSelect(champCopy)
				})
				// Set the min size hint for the GridWrap layout
				tappableCard.SetMinSize(cellSize)

				championWidgets = append(championWidgets, tappableCard)
			} // End of champion loop

			// *** Use standard GridWrap Layout ***
			grid := container.NewGridWrap(cellSize, championWidgets...)

			// *** Wrap the grid in Padding for centering effect ***
			// Adjust padding values as needed
			paddedGrid := container.NewPadded(grid)

			// *** Wrap the padded grid in a Scroll container ***
			scrollContainer := container.NewScroll(paddedGrid)
			finalContent = scrollContainer // The final UI is the scrollable grid

			log.Printf("Champion grid UI built and ready (using Padded GridWrap).")
		}

		// --- Update the main container's content ---
		// Ensure this runs on the main thread implicitly via Fyne's event loop
		if gridContainer != nil {
			gridContainer.Objects = []fyne.CanvasObject{finalContent} // Replace loading indicator
			gridContainer.Refresh()
			log.Printf("Champion grid container updated with final content.")
		} else {
			log.Println("ERROR: gridContainer is nil when trying to update champion grid.")
		}
	}() // End of background goroutine

	return gridContainer // Return the container that shows loading/final content
}

// --- End of champion_grid.go ---
