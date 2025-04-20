// skinhunter/ui/champion_grid.go
package ui

import (
	"fmt"
	"log"

	"skinhunter/data" // Use your data package

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	// "fyne.io/fyne/v2/layout"
)

// NewChampionGrid creates the scrollable grid of champions.
func NewChampionGrid(onChampionSelect func(champ data.ChampionSummary)) fyne.CanvasObject {
	// Grid container - Adjust columns for desired layout
	// Assuming SkinItem width ~100px (80px img + padding), aim for 8-10 columns.
	grid := container.NewGridWithColumns(9) // Let's try 9

	scrollContainer := container.NewScroll(grid)
	scrollContainer.Hide() // Hide scroll initially

	// Add loading indicator initially
	loading := widget.NewProgressBarInfinite()
	centeredLoading := container.NewCenter(loading)
	// Don't use Stack here, just replace content later
	contentArea := container.NewMax(centeredLoading) // Start with only loading

	// Fetch champions in background
	go func() {
		champions, err := data.FetchAllChampions() // Fetch the list (might init data)

		// --- Prepare UI Update ---
		var newContent fyne.CanvasObject

		if err != nil {
			log.Printf("Error fetching champions for grid: %v", err)
			errorLabel := widget.NewLabel(fmt.Sprintf("Error loading champions:\n%v", err))
			errorLabel.Wrapping = fyne.TextWrapWord
			newContent = container.NewCenter(errorLabel) // Display error
		} else {
			log.Printf("Populating champion grid with %d champions", len(champions))
			grid.Objects = nil // Clear previous content if any
			if len(champions) == 0 {
				grid.Add(widget.NewLabel("No champions found."))
			} else {
				for _, champ := range champions {
					// Capture loop variable correctly
					selectedChamp := champ
					item := ChampionGridItem(selectedChamp, onChampionSelect) // Use the UI element creator
					if item != nil {
						grid.Add(item)
					}
				}
			}
			grid.Refresh()               // Refresh the grid layout
			scrollContainer.Show()       // Show the scroll container
			newContent = scrollContainer // The content is now the scrollable grid
			log.Println("Champion grid populated")
		}

		// --- Apply UI Update ---
		// This runs implicitly on the main thread when widget states change.
		// To be extra safe or handle complex updates, use Events or canvas.Run() if needed.
		contentArea.Objects = []fyne.CanvasObject{newContent} // Replace loading/error with grid/error
		contentArea.Refresh()

	}() // End of goroutine

	return contentArea // Return the container holding loading/error or the final grid
}
