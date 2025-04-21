// skinhunter/ui/champion_view.go
package ui

import (
	"fmt"
	"log"
	"strings"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewChampionView creates the view displaying champion details and their skins.
func NewChampionView(
	champion data.ChampionSummary,
	onBack func(),
	onSkinSelect func(skin data.Skin, allChromas []data.Chroma),
) fyne.CanvasObject {

	log.Printf("Creating champion view for: %s (ID: %d)", champion.Name, champion.ID)
	loading := widget.NewProgressBarInfinite()
	contentArea := container.NewMax(container.NewCenter(container.NewVBox(widget.NewLabel("Loading Champion Details..."), loading)))

	go func() {
		details, err := data.FetchChampionDetails(champion.ID)
		var viewContent fyne.CanvasObject

		if err != nil {
			log.Printf("Error fetching details for %s (ID: %d): %v", champion.Name, champion.ID, err)
			errorLabel := widget.NewLabel(fmt.Sprintf("Error loading details for %s:\n%v", champion.Name, err))
			errorLabel.Wrapping = fyne.TextWrapWord
			errorLabel.Alignment = fyne.TextAlignCenter
			viewContent = container.NewCenter(errorLabel)
		} else {
			log.Printf("Details fetched successfully for %s", details.Name)

			// --- Left Side: Champion Info (Icon, Name, Title, Bio) ---
			imgSize := float32(80)
			imgAreaSize := fyne.NewSize(imgSize, imgSize)
			var champImageWidget fyne.CanvasObject

			// *** Usa URL de data.Asset y carga con NewImageFromURI ***
			imageUrl := data.Asset(details.SquarePortraitPath)
			imgUri, parseErr := storage.ParseURI(imageUrl)

			if parseErr != nil {
				log.Printf("Error parsing champion portrait URI %s: %v", imageUrl, parseErr)
				placeholder := canvas.NewRectangle(theme.InputBorderColor())
				placeholder.SetMinSize(imgAreaSize)
				champImageWidget = placeholder
			} else {
				imgWidget := canvas.NewImageFromURI(imgUri) // Carga normal
				imgWidget.SetMinSize(imgAreaSize)
				imgWidget.FillMode = canvas.ImageFillContain
				champImageWidget = imgWidget
			}

			champNameLabel := widget.NewLabelWithStyle(details.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			champTitleText := ""
			if details.Title != "" {
				champTitleText = strings.Title(strings.ToLower(details.Title))
			}
			champTitleLabel := widget.NewLabel(champTitleText)
			champTitleLabel.Wrapping = fyne.TextWrapWord
			bioLabel := widget.NewLabel(details.ShortBio)
			bioLabel.Wrapping = fyne.TextWrapWord
			bioScroll := container.NewScroll(bioLabel)
			bioScroll.SetMinSize(fyne.NewSize(250, 100))
			champTextInfo := container.NewVBox(champNameLabel, champTitleLabel)
			champHeader := container.NewHBox(champImageWidget, champTextInfo)
			champInfoBlock := container.NewVBox(champHeader, widget.NewSeparator(), bioScroll)

			// --- Right Side: Skins Grid ---
			skinsTitleLabel := widget.NewLabelWithStyle(fmt.Sprintf("%s Skins", details.Name), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			skinsGrid := NewSkinsGrid(details.Skins, func(skin data.Skin) {
				log.Printf("Skin selected in ChampionView: %s (ID: %d)", skin.Name, skin.ID)
				allChromasForChamp := make([]data.Chroma, 0)
				for _, s := range details.Skins {
					for _, ch := range s.Chromas {
						if ch.OriginSkinID == 0 {
							ch.OriginSkinID = s.ID
						}
						allChromasForChamp = append(allChromasForChamp, ch)
					}
				}
				log.Printf("Passing %d total chromas to dialog for champ %s (Selected Skin: %s)", len(allChromasForChamp), details.Name, skin.Name)
				onSkinSelect(skin, allChromasForChamp)
			})
			var skinsAreaContent fyne.CanvasObject
			if skinsGrid != nil {
				skinsAreaContent = skinsGrid
			} else {
				skinsAreaContent = container.NewCenter(widget.NewLabel("No skins grid available."))
			}
			skinsArea := container.NewBorder(container.NewPadded(skinsTitleLabel), nil, nil, nil, skinsAreaContent)

			// --- Combine Left and Right Panels ---
			split := container.NewHSplit(container.NewPadded(champInfoBlock), container.NewPadded(skinsArea))
			split.Offset = 0.3
			viewContent = split
			log.Printf("Champion view UI built for %s", details.Name)
		}

		// --- Update the Main Content Area ---
		if contentArea != nil {
			contentArea.Objects = []fyne.CanvasObject{viewContent}
			contentArea.Refresh()
			log.Printf("Champion view updated for %s", champion.Name)
		} else {
			log.Println("ERROR: contentArea is nil when trying to update champion view")
		}

	}() // End of goroutine

	return contentArea
}
