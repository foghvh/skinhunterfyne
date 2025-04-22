// skinhunter/ui/skin_dialog.go
package ui

import (
	"fmt"
	"image/color"
	"log"
	"net/url"

	// "strings" // Ya no es necesario aquí

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

// const cdragonBase = "..." // Ya no es necesario aquí

// ShowSkinDialog displays the detailed skin dialog.
func ShowSkinDialog(skin data.Skin, allChromasForChamp []data.Chroma, parent fyne.Window) {
	// ... (lógica inicial sin cambios) ...
	log.Printf("Showing dialog for skin: %s (ID: %d)", skin.Name, skin.ID)
	selectedChromaID := new(int)
	*selectedChromaID = skin.ID
	filteredChromas := make([]data.Chroma, 0)
	if allChromasForChamp != nil {
		for _, ch := range allChromasForChamp {
			if ch.OriginSkinID == 0 {
				ch.OriginSkinID = data.GetChampionIDFromSkinID(ch.ID)*1000 + (ch.ID % 1000)
			}
			if ch.OriginSkinID == skin.ID {
				filteredChromas = append(filteredChromas, ch)
			}
		}
	}
	log.Printf("Filtered %d chromas for skin ID %d", len(filteredChromas), skin.ID)

	// --- Left Side Elements ---
	imgMinSize := fyne.NewSize(400, 230)
	var leftImage fyne.CanvasObject

	// *** VOLVER a usar URL del paquete data ***
	splashUrl := data.GetSkinSplashURL(skin)
	splashUri, err := storage.ParseURI(splashUrl) // Parsear la URL obtenida
	if err != nil {
		log.Printf("Error parsing splash URI [%s]: %v", splashUrl, err)
		placeholder := canvas.NewRectangle(theme.InputBorderColor())
		placeholder.SetMinSize(imgMinSize)
		leftImage = placeholder
	} else {
		// *** VOLVER a usar canvas.NewImageFromURI ***
		splashImage := canvas.NewImageFromURI(splashUri)
		splashImage.FillMode = canvas.ImageFillContain
		splashImage.SetMinSize(imgMinSize)
		leftImage = splashImage
	}

	// ... (resto layout izquierdo sin cambios) ...
	desc := skin.Description
	if desc == "" {
		desc = "This skin does not have a description."
	}
	descriptionLabel := widget.NewLabel(desc)
	descriptionLabel.Wrapping = fyne.TextWrapWord
	descContainer := container.NewMax(descriptionLabel)
	descriptionScroll := container.NewScroll(descContainer)
	descriptionScroll.SetMinSize(fyne.NewSize(300, 100))
	warningLabel := widget.NewLabel("This skin may not work properly due to game updates.")
	warningIcon := widget.NewIcon(theme.WarningIcon())
	warningBox := container.NewPadded(container.NewHBox(warningIcon, warningLabel))
	leftPanel := container.NewVBox(leftImage, descriptionScroll, warningBox)

	// --- Right Side Elements (sin cambios en lógica principal) ---
	var downloadButton *widget.Button
	downloadButtonText := func() string {
		if selectedChromaID != nil && *selectedChromaID != skin.ID {
			return "Download Chroma"
		}
		return "Download Skin"
	}
	modelViewerURL := parseURL(data.KhadaUrl(skin.ID, 0))
	modelViewerLink := widget.NewHyperlink("View skin on Model viewer", modelViewerURL)
	modelInfoIcon := NewIconButton(theme.InfoIcon(), func() { /* ... */ })
	modelViewerLinkBox := container.NewHBox(modelViewerLink, modelInfoIcon)
	chromaTitle := widget.NewLabel("Chromas")
	chromaInfoIcon := NewIconButton(theme.InfoIcon(), func() { /* ... */ })
	chromaTitleBox := container.NewHBox(chromaTitle, chromaInfoIcon)
	selectDownloadLabel := widget.NewLabel("Select and Download your skin.")
	selectDownloadLabel.TextStyle = fyne.TextStyle{Italic: true}
	selectDownloadLabel.Alignment = fyne.TextAlignCenter
	updateChromaSelection := func(newID int) { /* ... */ }

	// Tab 1: Circles (Usa helper sin cambios visuales)
	circlesGrid := container.NewGridWithColumns(4)
	circlesGrid.Add(createChromaCircleItem("Default", nil, skin.ID, selectedChromaID, updateChromaSelection))
	for _, chroma := range filteredChromas {
		circlesGrid.Add(createChromaCircleItem(chroma.Name, chroma.Colors, chroma.ID, selectedChromaID, updateChromaSelection))
	}
	circlesTabContent := circlesGrid

	// Tab 2: Images (Usa helper que ahora carga imágenes con método anterior)
	imagesGrid := container.NewGridWithColumns(4)
	// Pasar un struct Chroma vacío para el caso "Default" para que la función helper sepa cómo manejarlo
	imagesGrid.Add(createChromaImageItem("Default", data.Chroma{}, skin.ID, selectedChromaID, updateChromaSelection))
	for _, chroma := range filteredChromas {
		// Pasar el struct Chroma completo
		imagesGrid.Add(createChromaImageItem(chroma.Name, chroma, chroma.ID, selectedChromaID, updateChromaSelection))
	}
	imagesTabContent := imagesGrid

	chromaTabs := container.NewAppTabs(
		container.NewTabItem("Circles", circlesTabContent),
		container.NewTabItem("Images", imagesTabContent),
	)

	// ... (resto layout derecho y diálogo sin cambios) ...
	creditsText := "This is going to consume a credit"
	creditsInfoLabel := widget.NewLabel(creditsText)
	creditsBox := container.NewHBox(widget.NewIcon(theme.InfoIcon()), creditsInfoLabel)
	rightPanel := container.NewVBox(modelViewerLinkBox, widget.NewSeparator(), chromaTitleBox, selectDownloadLabel, chromaTabs, layout.NewSpacer(), creditsBox)
	rightPanelContainer := container.NewPadded(rightPanel)
	dialogContentSplit := container.NewHSplit(container.NewPadded(leftPanel), rightPanelContainer)
	dialogContentSplit.Offset = 0.5
	downloadButton = widget.NewButtonWithIcon(downloadButtonText(), theme.DownloadIcon(), func() { /* download logic */
		if selectedChromaID == nil {
			log.Println("Error: selectedChromaID is nil")
			return
		}
		downloadID := *selectedChromaID
		downloadType := "Skin"
		downloadName := skin.Name
		if downloadID != skin.ID {
			downloadType = "Chroma"
			found := false
			for _, ch := range filteredChromas {
				if ch.ID == downloadID {
					downloadName = ch.Name
					found = true
					break
				}
			}
			if !found {
				downloadName = fmt.Sprintf("Chroma ID %d", downloadID)
			}
		}
		log.Printf("DOWNLOAD: Type=%s, Name=%s, ID=%d", downloadType, downloadName, downloadID)
		fyne.CurrentApp().SendNotification(&fyne.Notification{Title: "Download Requested", Content: fmt.Sprintf("Downloading %s: %s (ID: %d)", downloadType, downloadName, downloadID)})
		dialog.ShowInformation("Download", fmt.Sprintf("Placeholder: Download %s: %s (ID: %d)", downloadType, downloadName, downloadID), parent)
	})
	var customDialog dialog.Dialog
	closeButton := widget.NewButton("Close", func() {
		if customDialog != nil {
			customDialog.Hide()
		}
	})
	actionButtons := container.NewBorder(nil, nil, closeButton, downloadButton)
	finalDialogContent := container.NewBorder(nil, container.NewPadded(actionButtons), nil, nil, dialogContentSplit)
	customDialog = dialog.NewCustom(skin.Name, "Dismiss", finalDialogContent, parent)
	customDialog.Resize(fyne.NewSize(850, 550))
	customDialog.Show()
}

// Helper createChromaCircleItem (Sin cambios respecto a la versión anterior)
func createChromaCircleItem(name string, hexColors []string, itemID int, selectedID *int, onSelect func(id int)) fyne.CanvasObject {
	const visualSize float32 = 48
	visualMinSize := fyne.NewSize(visualSize, visualSize)
	var displayElement fyne.CanvasObject
	if name == "Default" {
		placeholder := canvas.NewCircle(theme.InputBorderColor())
		placeholder.Resize(visualMinSize)
		icon := widget.NewIcon(theme.CancelIcon())
		displayElement = container.NewStack(placeholder, container.NewCenter(icon))
	} else if len(hexColors) > 0 {
		clr1, err1 := data.ParseHexColor(hexColors[0])
		if err1 != nil {
			clr1 = color.NRGBA{128, 128, 128, 255}
		}
		if len(hexColors) == 1 || hexColors[1] == "" || hexColors[1] == hexColors[0] {
			circle := canvas.NewCircle(clr1)
			circle.Resize(visualMinSize)
			displayElement = circle
		} else {
			clr2, err2 := data.ParseHexColor(hexColors[1])
			if err2 != nil {
				clr2 = clr1
			}
			gradient := canvas.NewHorizontalGradient(clr1, clr2)
			gradient.SetMinSize(visualMinSize)
			displayElement = gradient
		}
	} else {
		icon := widget.NewIcon(theme.QuestionIcon())
		bg := canvas.NewRectangle(theme.DisabledButtonColor())
		bg.SetMinSize(visualMinSize)
		displayElement = container.NewStack(bg, container.NewCenter(icon))
	}
	selectionIndicator := canvas.NewRectangle(color.Transparent)
	selectionIndicator.StrokeColor = theme.PrimaryColorNamed(theme.ColorBlue)
	selectionIndicator.StrokeWidth = 2
	selectionIndicator.Resize(fyne.NewSize(visualSize+4, visualSize+4))
	selectionIndicator.Hide()
	displayStack := container.NewStack(displayElement, container.NewCenter(selectionIndicator))
	nameLabel := widget.NewLabel(name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	itemContent := container.NewVBox(container.NewCenter(displayStack), nameLabel)
	card := NewTappableCard(container.NewPadded(itemContent), func() { onSelect(itemID) })
	if selectedID != nil && *selectedID == itemID {
		selectionIndicator.Show()
	}
	return card
}

// Helper createChromaImageItem (Modificado para usar URL de data y NewImageFromURI)
// Ahora recibe el struct data.Chroma completo en lugar de la URL
func createChromaImageItem(name string, chroma data.Chroma, itemID int, selectedID *int, onSelect func(id int)) fyne.CanvasObject {
	const imgSize float32 = 64
	imgAreaSize := fyne.NewSize(imgSize, imgSize)
	var visualElement fyne.CanvasObject

	if name == "Default" { // Usar itemID para verificar si es default? O el nombre?
		icon := widget.NewIcon(theme.CancelIcon())
		bg := canvas.NewRectangle(theme.ButtonColor())
		bg.SetMinSize(imgAreaSize)
		visualElement = container.NewStack(bg, container.NewCenter(icon))
		// } else if chroma.ChromaPath == "" { // Verificar si el path del chroma está vacío
		// No es necesario, data.GetChromaImageURL manejará path vacío devolviendo placeholder
	} else {
		// *** Usar data.GetChromaImageURL (que usa data.Asset) ***
		imageURL := data.GetChromaImageURL(chroma)   // Obtener URL
		chromaUri, err := storage.ParseURI(imageURL) // Parsear URL
		if err != nil {
			log.Printf("Error parsing chroma image URI (from data.Asset) [%s]: %v", imageURL, err)
			icon := widget.NewIcon(theme.WarningIcon())
			bg := canvas.NewRectangle(theme.ErrorColor())
			bg.SetMinSize(imgAreaSize)
			visualElement = container.NewStack(bg, container.NewCenter(icon))
		} else {
			// *** Usar canvas.NewImageFromURI ***
			imgWidget := canvas.NewImageFromURI(chromaUri)
			imgWidget.FillMode = canvas.ImageFillContain
			imgWidget.SetMinSize(imgAreaSize)
			visualElement = imgWidget
		}
	}

	selectionIndicator := canvas.NewRectangle(color.Transparent)
	selectionIndicator.StrokeColor = theme.PrimaryColorNamed(theme.ColorBlue)
	selectionIndicator.StrokeWidth = 2
	selectionIndicator.Resize(fyne.NewSize(imgSize+4, imgSize+4))
	selectionIndicator.Hide()
	visualStack := container.NewStack(visualElement, container.NewCenter(selectionIndicator))
	nameLabel := widget.NewLabel(name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	itemContent := container.NewVBox(container.NewCenter(visualStack), nameLabel)
	card := NewTappableCard(container.NewPadded(itemContent), func() { onSelect(itemID) })
	if selectedID != nil && *selectedID == itemID {
		selectionIndicator.Show()
	}
	return card
}

// Helper parseURL (sin cambios)
func parseURL(rawURL string) *url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Error parsing URL '%s': %v", rawURL, err)
		return nil
	}
	return parsed
}
