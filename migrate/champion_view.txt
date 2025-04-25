// skinhunter/ui/champion_view.go
package ui

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewChampionView creates the view displaying champion details above their skins grid.
// Uses container.NewBorder, placing the scrollable grid (from NewSkinsGrid) DIRECTLY in the Center.
func NewChampionView(
	champion data.ChampionSummary,
	parentWindow fyne.Window,
	onSkinSelect func(skin data.Skin, allChromas []data.Chroma),
) fyne.CanvasObject {

	log.Printf("Creating champion view for: %s (ID: %d)", champion.Name, champion.ID)
	// Use a Max container to hold the loading indicator or the final content
	viewContainer := container.NewMax()
	loading := widget.NewProgressBarInfinite()
	viewContainer.Add(container.NewCenter(container.NewVBox(widget.NewLabel("Loading Champion Details..."), loading)))

	go func() {
		details, err := data.FetchChampionDetails(champion.ID)
		var finalContent fyne.CanvasObject

		if err != nil {
			log.Printf("Error fetching details for %s (ID: %d): %v", champion.Name, champion.ID, err)
			errorLabel := widget.NewLabel(fmt.Sprintf("Error loading details for %s:\n%v", champion.Name, err))
			errorLabel.Wrapping = fyne.TextWrapWord
			errorLabel.Alignment = fyne.TextAlignCenter
			// Center the error message
			finalContent = container.NewCenter(errorLabel)
		} else {
			log.Printf("Details fetched successfully for %s", details.Name)

			// --- Build Top Section (Champion Info + Skins Title) ---
			imgSize := float32(64)
			imgAreaSize := fyne.NewSize(imgSize, imgSize)
			var champImageWidget fyne.CanvasObject
			// Placeholder logic for champion portrait (can also be lazy-loaded if needed)
			imageStack := container.NewStack() // Use stack for potential lazy load
			champImageWidget = imageStack      // Assign stack to the layout first

			placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
			placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
			placeholderRect.SetMinSize(imgAreaSize)
			imageStack.Add(placeholderRect)
			imageStack.Add(container.NewCenter(placeholderIcon))

			// Lazy load champ image (optional, but consistent)
			go func(c data.DetailedChampionData, stack *fyne.Container) {
				imageUrl := data.Asset(c.SquarePortraitPath) // Use Asset func
				if imageUrl == data.GetPlaceholderImageURL() {
					return
				}
				imgUri, parseErr := storage.ParseURI(imageUrl)
				if parseErr != nil {
					log.Printf("WARN: Failed to parse champ portrait URI %s: %v", imageUrl, parseErr)
					return
				}
				imgWidget := canvas.NewImageFromURI(imgUri)
				imgWidget.SetMinSize(imgAreaSize)
				imgWidget.FillMode = canvas.ImageFillContain
				if stack != nil {
					stack.Objects = []fyne.CanvasObject{imgWidget} // Replace placeholder
					stack.Refresh()
				}
			}(*details, imageStack)

			champNameLabel := widget.NewLabelWithStyle(details.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			champTitleText := ""
			if details.Title != "" {
				champTitleText = strings.Title(strings.ToLower(details.Title))
			}
			champTitleLabel := widget.NewLabel(champTitleText)
			champTextInfo := container.NewVBox(champNameLabel, champTitleLabel)
			// Use padding around text info for better spacing
			champHeader := container.NewHBox(champImageWidget, container.NewPadded(champTextInfo))

			// Bio section (truncated)
			bioExcerpt := details.ShortBio
			const maxBioLen = 180 // Keep truncation logic
			if utf8.RuneCountInString(bioExcerpt) > maxBioLen {
				count := 0
				cutoff := 0
				// Correct rune-based truncation
				for i := range bioExcerpt {
					if count >= maxBioLen {
						// Find the end of the previous rune
						cutoff = i
						// Try to break at space before cutoff if possible
						lastSpace := strings.LastIndex(bioExcerpt[:cutoff], " ")
						if lastSpace > maxBioLen-30 { // Only break at space if it's reasonably close
							cutoff = lastSpace
						}
						break
					}
					count++
				}
				if cutoff == 0 { // Handle case where loop finishes without hitting limit
					cutoff = len(bioExcerpt)
				}
				bioExcerpt = bioExcerpt[:cutoff] + "..."
			}
			bioLabel := widget.NewLabel(bioExcerpt)
			bioLabel.Wrapping = fyne.TextWrapWord

			viewMoreButton := widget.NewButton("View more", func() {
				log.Printf("View more clicked for: %s", details.Name)
				fullBioLabel := widget.NewLabel(details.ShortBio)
				fullBioLabel.Wrapping = fyne.TextWrapWord
				scrollBio := container.NewScroll(fullBioLabel)
				// Set a reasonable minimum size for the dialog content
				scrollBio.SetMinSize(fyne.NewSize(450, 350))
				dialog.ShowCustom(fmt.Sprintf("%s - Biography", details.Name), "Close", scrollBio, parentWindow)
			})
			// Align button to the right maybe? Or keep below label. Below is simpler.
			// bioAndButton := container.NewBorder(nil, nil, nil, viewMoreButton, bioLabel) // Alternative layout
			bioAndButton := container.NewVBox(bioLabel, container.NewHBox(layout.NewSpacer(), viewMoreButton)) // Button below, right-aligned

			// Skins Title Header
			skinsIcon := widget.NewIcon(theme.ColorPaletteIcon()) // Changed icon
			skinsTitleLabel := widget.NewLabelWithStyle(fmt.Sprintf("%s Skins", details.Name), fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true})
			skinsTitleHeader := container.NewHBox(skinsIcon, skinsTitleLabel)

			// Group top section elements
			topSectionContent := container.NewVBox(
				champHeader,
				widget.NewSeparator(),
				bioAndButton,
				widget.NewSeparator(),
				container.NewPadded(skinsTitleHeader), // Pad the skins title
			)
			// Add overall padding to the top section
			topSection := container.NewPadded(topSectionContent)

			// --- Build Center Section (Scrollable Skins Grid) ---
			// This now uses the updated NewSkinsGrid which handles lazy loading and padding internally.
			skinsGrid := NewSkinsGrid(details.Skins, func(skin data.Skin) {
				log.Printf("Skin selected in ChampionView: %s (ID: %d)", skin.Name, skin.ID)
				// Collect all chromas *just* for this specific champion
				// This logic seems correct, but ensure Chromas have OriginSkinID set properly in data layer
				allChromasForChamp := make([]data.Chroma, 0)
				for _, s := range details.Skins { // Iterate only through this champion's skins
					for _, ch := range s.Chromas {
						chromaCopy := ch
						// Ensure OriginSkinID is set (should be done in FetchChampionDetails)
						if chromaCopy.OriginSkinID == 0 {
							chromaCopy.OriginSkinID = s.ID
							// log.Printf("WARN: Corrected missing OriginSkinID for chroma %d in champion view", chromaCopy.ID)
						}
						// Only add chromas belonging to *this* champion (redundant check maybe, but safe)
						if data.GetChampionIDFromSkinID(chromaCopy.OriginSkinID) == details.ID {
							allChromasForChamp = append(allChromasForChamp, chromaCopy)
						}
					}
				}
				log.Printf("Passing %d chromas relevant to champion %s to skin dialog", len(allChromasForChamp), details.Name)
				onSkinSelect(skin, allChromasForChamp) // Pass only relevant chromas
			})

			// Check if grid creation returned a valid object
			var centerContent fyne.CanvasObject
			if skinsGrid != nil {
				centerContent = skinsGrid // Assign the scrollable grid directly
			} else {
				// This case should be handled inside NewSkinsGrid now, but keep fallback
				log.Println("WARN: NewSkinsGrid returned nil, showing fallback message.")
				centerContent = container.NewCenter(widget.NewLabel("No skins grid available."))
			}

			// --- Final Assembly using Border Layout ---
			finalContent = container.NewBorder(
				topSection,    // Top: Padded champion info + skins title
				nil,           // Bottom: nil
				nil,           // Left: nil
				nil,           // Right: nil
				centerContent, // Center: The result of NewSkinsGrid (scrollable padded grid)
			)

			log.Printf("Champion view UI built for %s", details.Name)
		}

		// --- Update UI Safely ---
		// Ensure the update happens on the main Fyne thread
		if viewContainer != nil {
			viewContainer.Objects = []fyne.CanvasObject{finalContent} // Replace loading indicator
			viewContainer.Refresh()
			log.Printf("Champion view container updated for %s", champion.Name)
		} else {
			log.Println("ERROR: viewContainer is nil when trying to update champion view")
		}

	}() // End of goroutine

	return viewContainer // Return the container holding loading/final content
}

// --- End of champion_view.go ---
