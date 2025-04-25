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
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewChampionView creates the view displaying champion details above their skins grid.
// Uses container.NewBorder, placing the scrollable grid DIRECTLY in the Center.
func NewChampionView(
	champion data.ChampionSummary,
	parentWindow fyne.Window,
	onSkinSelect func(skin data.Skin, allChromas []data.Chroma),
) fyne.CanvasObject {

	log.Printf("Creating champion view V5 (Scroll in Border Center) for: %s (ID: %d)", champion.Name, champion.ID)
	loading := widget.NewProgressBarInfinite()
	viewContainer := container.NewMax()
	viewContainer.Add(container.NewCenter(container.NewVBox(widget.NewLabel("Loading Champion Details..."), loading)))

	go func() {
		details, err := data.FetchChampionDetails(champion.ID)
		var finalContent fyne.CanvasObject

		if err != nil {
			// ... (manejo de error idéntico) ...
			log.Printf("Error fetching details for %s (ID: %d): %v", champion.Name, champion.ID, err)
			errorLabel := widget.NewLabel(fmt.Sprintf("Error loading details for %s:\n%v", champion.Name, err))
			errorLabel.Wrapping = fyne.TextWrapWord
			errorLabel.Alignment = fyne.TextAlignCenter
			finalContent = container.NewCenter(errorLabel)
		} else {
			log.Printf("Details fetched successfully for %s", details.Name)

			// --- Construir Sección Superior (Top) ---
			// Incluye toda la info del campeón Y el título de "Skins"
			imgSize := float32(64)
			imgAreaSize := fyne.NewSize(imgSize, imgSize)
			var champImageWidget fyne.CanvasObject
			// ... (código de carga de imagen idéntico) ...
			imageUrl := data.Asset(details.SquarePortraitPath)
			imgUri, parseErr := storage.ParseURI(imageUrl)
			if parseErr != nil || imageUrl == data.GetPlaceholderImageURL() {
				placeholder := canvas.NewRectangle(theme.InputBorderColor())
				placeholder.SetMinSize(imgAreaSize)
				placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
				champImageWidget = container.NewStack(placeholder, container.NewCenter(placeholderIcon))
			} else {
				imgWidget := canvas.NewImageFromURI(imgUri)
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
			champTextInfo := container.NewVBox(champNameLabel, champTitleLabel)
			champHeader := container.NewHBox(champImageWidget, container.NewPadded(champTextInfo))

			bioExcerpt := details.ShortBio
			const maxBioLen = 180
			if utf8.RuneCountInString(bioExcerpt) > maxBioLen {
				// ... (lógica de truncar bio) ...
				count := 0
				cutoff := 0
				for i := range bioExcerpt {
					count++
					if count >= maxBioLen {
						cutoff = i
						break
					}
				}
				bioExcerpt = bioExcerpt[:cutoff] + "..."
			}
			bioLabel := widget.NewLabel(bioExcerpt)
			bioLabel.Wrapping = fyne.TextWrapWord

			viewMoreButton := widget.NewButton("View more", func() {
				// ... (lógica del diálogo sin cambios) ...
				log.Printf("View more clicked for: %s", details.Name)
				fullBioLabel := widget.NewLabel(details.ShortBio)
				fullBioLabel.Wrapping = fyne.TextWrapWord
				scrollBio := container.NewScroll(fullBioLabel)
				scrollBio.SetMinSize(fyne.NewSize(450, 350))
				dialog.ShowCustom(fmt.Sprintf("%s - Biography", details.Name), "Close", scrollBio, parentWindow)
			})
			bioAndButton := container.NewVBox(bioLabel, viewMoreButton)

			// Título de Skins
			skinsIcon := widget.NewIcon(theme.VisibilityIcon())
			skinsTitleLabel := widget.NewLabelWithStyle(fmt.Sprintf("%s Skins", details.Name), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			skinsTitleHeader := container.NewHBox(skinsIcon, skinsTitleLabel)

			// Agrupar toda la sección superior en un VBox
			topSectionContent := container.NewVBox(
				champHeader,
				bioAndButton,
				widget.NewSeparator(),
				container.NewPadded(skinsTitleHeader), // Título de skins AHORA es parte de la sección superior
			)
			// Añadir padding general a la sección superior
			topSection := container.NewPadded(topSectionContent)

			// --- Construir Sección Central (Center) ---
			// ESTA SECCIÓN ES *SOLAMENTE* LA CUADRÍCULA SCROLLABLE
			skinsGrid := NewSkinsGrid(details.Skins, func(skin data.Skin) {
				// ... (lógica onSkinSelect sin cambios) ...
				log.Printf("Skin selected in ChampionView V5: %s (ID: %d)", skin.Name, skin.ID)
				allChromasForChamp := make([]data.Chroma, 0)
				for _, s := range details.Skins {
					for _, ch := range s.Chromas {
						chromaCopy := ch
						if chromaCopy.OriginSkinID == 0 {
							chromaCopy.OriginSkinID = s.ID
						}
						allChromasForChamp = append(allChromasForChamp, chromaCopy)
					}
				}
				onSkinSelect(skin, allChromasForChamp)
			})

			var centerContent fyne.CanvasObject // Será el grid o un mensaje de error
			if skinsGrid != nil {
				// ASIGNAR EL RESULTADO DE NewSkinsGrid DIRECTAMENTE
				centerContent = skinsGrid
			} else {
				centerContent = container.NewCenter(widget.NewLabel("No skins grid available."))
			}

			// --- Ensamblaje Final con Border Layout ---
			// topSection va arriba.
			// centerContent (el grid scrollable) va al centro y se expandirá.
			finalContent = container.NewBorder(
				topSection,    // Top: Info campeón + Título Skins
				nil,           // Bottom: nil
				nil,           // Left: nil
				nil,           // Right: nil
				centerContent, // Center: ¡El container.Scroll directamente!
			)

			log.Printf("Champion view V5 UI (Scroll in Border Center) built for %s", details.Name)
		}

		// --- Actualizar UI de forma segura ---
		// --- Actualizar el Contenido Principal de Forma Segura ---
		// Ejecutar en el hilo principal de Fyne para evitar problemas de concurrencia
		if viewContainer != nil {
			viewContainer.Objects = []fyne.CanvasObject{finalContent} // Reemplaza el loading
			viewContainer.Refresh()
			log.Printf("Champion view V3 updated directly for %s", champion.Name)
		} else {
			log.Println("ERROR: viewContainer is nil when trying to update champion view V3")
		}
		// --- FIN CORRECCIÓN ---

	}() // Fin de la goroutine

	return viewContainer
}
