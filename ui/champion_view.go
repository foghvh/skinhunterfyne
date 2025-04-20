// skinhunter/ui/champion_view.go
package ui

import (
	"fmt"
	"log"
	"strings" // For capitalizing title

	"skinhunter/data"

	"fyne.io/fyne/v2"
	// Removed canvas import
	"fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/layout"
	// Removed theme import
	"fyne.io/fyne/v2/widget"
)

// NewChampionView creates the view displaying champion details and their skins.
func NewChampionView(
	champion data.ChampionSummary, // Pass summary, fetch details inside
	onBack func(),
	onSkinSelect func(skin data.Skin, allChromas []data.Chroma), // Pass chromas too
) fyne.CanvasObject {

	log.Printf("Creating champion view for: %s (ID: %d)", champion.Name, champion.ID)
	loading := widget.NewProgressBarInfinite()
	// Main container that will hold the final split layout or loading/error
	contentArea := container.NewMax(container.NewCenter(loading))

	go func() {
		// --- Fetch Detailed Data ---
		details, err := data.FetchChampionDetails(champion.ID) // Fetch details by ID
		var viewContent fyne.CanvasObject                      // Content to display

		if err != nil {
			log.Printf("Error fetching details for %s (ID: %d): %v", champion.Name, champion.ID, err)
			errorLabel := widget.NewLabel(fmt.Sprintf("Error loading details for %s", champion.Name))
			errorLabel.Wrapping = fyne.TextWrapWord
			viewContent = container.NewCenter(errorLabel) // Display error
		} else {
			log.Printf("Details fetched successfully for %s", details.Name)
			// --- Build the Champion View UI ---

			// Left Side: Champion Info (Icon, Name, Title, Bio)
			imgSize := float32(80)
			imgContainer, imgWidget := NewAsyncImage(imgSize, imgSize)
			// Use details.SquarePortraitPath which should be correctly resolved by Asset
			SetImageURL(imgWidget, imgContainer, data.Asset(details.SquarePortraitPath))

			champNameLabel := widget.NewLabelWithStyle(details.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			champTitleText := ""
			if details.Title != "" {
				champTitleText = strings.Title(strings.ToLower(details.Title)) // Capitalize
			}
			champTitleLabel := widget.NewLabel(champTitleText)
			champTitleLabel.Wrapping = fyne.TextWrapWord

			// Basic bio
			bioLabel := widget.NewLabel(details.ShortBio)
			bioLabel.Wrapping = fyne.TextWrapWord
			bioScroll := container.NewScroll(bioLabel)
			bioScroll.SetMinSize(fyne.NewSize(300, 100)) // Limit description height

			champInfoBlock := container.NewVBox(
				container.NewHBox(imgContainer, container.NewVBox(champNameLabel, champTitleLabel)),
				widget.NewSeparator(),
				bioScroll,
				// TODO: Add Role/Damage/Difficulty icons here if needed
			)

			// Right Side: Skins Grid
			skinsTitleLabel := widget.NewLabelWithStyle(fmt.Sprintf("%s Skins", details.Name), fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: false}) // Style like screenshot
			// The skin grid needs the detailed skins from `details.Skins`
			skinsGrid := NewSkinsGrid(details.Skins, func(skin data.Skin) { // Pass fetched skins
				log.Printf("Skin selected in ChampionView: %s (ID: %d)", skin.Name, skin.ID)
				// Find all chromas for *this specific champion* from the detailed data
				// Pass *all* chromas found within the champion's details.Skins array.
				var allChromasForChamp []data.Chroma
				// The skin passed already *should* have its chromas if fetched correctly by GetSkinDetails/FetchChampionDetails
				// The dialog will filter these again, but let's pass all chromas found under this champ.
				for _, s := range details.Skins {
					for _, ch := range s.Chromas {
						// Ensure OriginSkinID is set if it wasn't during fetch
						if ch.OriginSkinID == 0 {
							ch.OriginSkinID = s.ID
						}
						allChromasForChamp = append(allChromasForChamp, ch)
					}
				}

				log.Printf("Passing %d total chromas to dialog for champ %s", len(allChromasForChamp), details.Name)
				onSkinSelect(skin, allChromasForChamp) // Pass the selected skin and all chromas for this champ
			})

			skinsArea := container.NewBorder(container.NewPadded(skinsTitleLabel), nil, nil, nil, skinsGrid)

			// Combine Left and Right panels
			split := container.NewHSplit(
				container.NewPadded(champInfoBlock),
				container.NewPadded(skinsArea),
			)
			split.Offset = 0.3 // Adjust ratio (Left side smaller)

			viewContent = split // Display the split view
			log.Printf("Champion view UI built for %s", details.Name)
		}

		// --- Update the Main Content Area ---
		contentArea.Objects = []fyne.CanvasObject{viewContent}
		contentArea.Refresh()
		log.Printf("Champion view updated for %s", champion.Name)

	}() // End of goroutine

	return contentArea // Return the container that holds loading/error or the final content
}
