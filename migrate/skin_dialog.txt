// skinhunter/ui/skin_dialog.go
package ui

import (
	"fmt"
	"image/color"
	"log"
	"runtime/debug"
	"sync"

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

type chromaItemUI struct {
	widget    fyne.CanvasObject
	nameLabel *widget.Label
}

func ShowSkinDialog(skin data.Skin, allChromasForChamp []data.Chroma, parent fyne.Window) {
	log.Printf("Showing dialog for skin: %s (ID: %d)", skin.Name, skin.ID)
	if allChromasForChamp == nil {
		allChromasForChamp = []data.Chroma{}
	}
	selectedChromaID := new(int)
	*selectedChromaID = skin.ID
	filteredChromas := make([]data.Chroma, 0)
	for _, ch := range allChromasForChamp {
		if ch.OriginSkinID == 0 {
			ch.OriginSkinID = data.DeriveOriginSkinID(ch.ID)
		}
		if ch.OriginSkinID == skin.ID {
			filteredChromas = append(filteredChromas, ch)
		}
	}
	log.Printf("Found %d chromas associated with skin ID %d ('%s')", len(filteredChromas), skin.ID, skin.Name)

	var downloadButton *widget.Button
	var circlesGrid, imagesGrid *fyne.Container
	var chromaCircleItems = make(map[int]*chromaItemUI)
	var chromaImageItems = make(map[int]*chromaItemUI)
	var uiMutex sync.Mutex

	// --- Izquierda: Imagen y Descripción ---
	imgMinSize := fyne.NewSize(400, 230)
	imageStack := container.NewStack()
	leftImage := imageStack
	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgMinSize)
	imageStack.Add(placeholderRect)
	imageStack.Add(container.NewCenter(placeholderIcon))
	go func(s data.Skin, stack *fyne.Container) { /* ... Carga Imagen Splash ... */
		splashUrl := data.GetSkinSplashURL(s)
		if splashUrl == data.GetPlaceholderImageURL() {
			return
		}
		splashUri, err := storage.ParseURI(splashUrl)
		if err != nil {
			return
		}
		splashImage := canvas.NewImageFromURI(splashUri)
		splashImage.FillMode = canvas.ImageFillContain
		splashImage.SetMinSize(imgMinSize)
		fyne.Do(func() {
			if stack != nil && stack.Visible() {
				stack.Objects = []fyne.CanvasObject{splashImage}
				stack.Refresh()
			}
		})
	}(skin, imageStack)

	desc := skin.Description
	if desc == "" {
		desc = "This skin does not have a description."
	}
	descriptionLabel := widget.NewLabel(desc)
	descriptionLabel.Wrapping = fyne.TextWrapWord
	descScroll := container.NewScroll(descriptionLabel)
	descScroll.SetMinSize(fyne.NewSize(350, 350)) // Ancho ajustado, altura mínima

	// *** Layout Panel Izquierdo (VBox simple) ***
	leftPanel := container.NewVBox(
		leftImage,
		descScroll, // Descripción directamente debajo
		// Sin warningBox
	)
	// Añadir padding general al panel izquierdo
	paddedLeftPanel := container.NewPadded(leftPanel)
	// -----------------------------------------

	// --- Derecha: Chromas y Acciones ---
	updateSelectionUI := func(newID int) { /* ... sin cambios ... */
		*selectedChromaID = newID
		if downloadButton != nil {
			btnTxt := "Download Skin"
			if *selectedChromaID != skin.ID {
				btnTxt = "Download Chroma"
			}
			downloadButton.SetText(btnTxt)
		}
		if circlesGrid != nil {
			circlesGrid.Refresh()
		}
		if imagesGrid != nil {
			imagesGrid.Refresh()
		}
	}
	modelViewerURL := parseURL(data.KhadaUrl(skin.ID, *selectedChromaID))
	modelViewerLink := widget.NewHyperlink("View on Model Viewer", modelViewerURL)
	modelViewerBox := container.NewHBox(modelViewerLink, layout.NewSpacer(), NewIconButton(theme.InfoIcon(), func() { dialog.ShowInformation("Model Viewer", "...", parent) }))
	chromaTitle := widget.NewLabelWithStyle("Chromas", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	chromaTitleBox := container.NewHBox(chromaTitle, layout.NewSpacer(), NewIconButton(theme.InfoIcon(), func() { dialog.ShowInformation("Chromas", "...", parent) }))
	// *** Texto simplificado ***
	selectDownloadLabel := widget.NewLabel("Select variation:")
	selectDownloadLabel.TextStyle = fyne.TextStyle{Italic: true}

	circlesGrid = container.NewGridWithColumns(4)
	imagesGrid = container.NewGridWithColumns(4)
	defaultCircleUI := createChromaCircleItem("Default", nil, skin.ID, selectedChromaID, updateSelectionUI)
	defaultImageUI := createChromaImageItem("Default", data.Chroma{ID: skin.ID, OriginSkinID: skin.ID}, skin.ID, selectedChromaID, updateSelectionUI)
	circlesGrid.Add(defaultCircleUI.widget)
	imagesGrid.Add(defaultImageUI.widget)
	for _, chroma := range filteredChromas {
		chromaCopy := chroma
		circleUI := createChromaCircleItem("Loading...", chromaCopy.Colors, chromaCopy.ID, selectedChromaID, updateSelectionUI)
		imageUI := createChromaImageItem("Loading...", chromaCopy, chromaCopy.ID, selectedChromaID, updateSelectionUI)
		circlesGrid.Add(circleUI.widget)
		imagesGrid.Add(imageUI.widget)
		uiMutex.Lock()
		chromaCircleItems[chromaCopy.ID] = &circleUI
		chromaImageItems[chromaCopy.ID] = &imageUI
		uiMutex.Unlock()
	}

	// *** Altura Scroll Chromas (Ajustada para ~2 filas) ***
	imgItemHeight := float32(64 + 40 + 20) // Altura item imagen (imagen+label+padding)
	chromaScrollHeight := imgItemHeight*2 + theme.Padding()*3
	circlesTabContent := container.NewScroll(circlesGrid)
	circlesTabContent.SetMinSize(fyne.NewSize(350, chromaScrollHeight))
	imagesTabContent := container.NewScroll(imagesGrid)
	imagesTabContent.SetMinSize(fyne.NewSize(350, chromaScrollHeight))

	chromaTabs := container.NewAppTabs(container.NewTabItemWithIcon("Colors", theme.ColorPaletteIcon(), circlesTabContent), container.NewTabItemWithIcon("Previews", theme.DocumentIcon(), imagesTabContent))
	chromaTabs.SetTabLocation(container.TabLocationTop)

	creditsText := "Note: Downloading may require credits (Not Implemented)."
	creditsInfoLabel := widget.NewLabel(creditsText)
	creditsBox := container.NewHBox(widget.NewIcon(theme.InfoIcon()), creditsInfoLabel)

	// *** Layout Panel Derecho (VBox) ***
	rightPanel := container.NewVBox(
		container.NewPadded(modelViewerBox),
		widget.NewSeparator(),
		container.NewPadded(chromaTitleBox),
		container.NewPadded(selectDownloadLabel), // Con padding
		chromaTabs,                               // Tabs con scroll
		layout.NewSpacer(),                       // Empuja créditos abajo
		container.NewPadded(creditsBox),
	)
	// Añadir padding general al panel derecho
	paddedRightPanel := container.NewPadded(rightPanel)
	// -----------------------------------------

	// --- Goroutine para fetch nombres (Sin cambios aquí) ---
	go func(champID int, skinID int) { /* ... como antes ... */
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered fetching chroma names: %v\n%s", r, string(debug.Stack()))
			}
		}()
		log.Printf("Fetching rich chroma data for champion %d from Supabase...", champID)
		richData, err := data.FetchChampionJsonFromSupabase(champID)
		if err != nil {
			log.Printf("WARN: Failed to fetch rich chroma data: %v", err)
			return
		}
		chromaNames := make(map[int]string)
		if skinsData, ok := richData["skins"].([]interface{}); ok { /* ... parseo JSON ... */
			for _, skinInterface := range skinsData {
				if skinMap, ok := skinInterface.(map[string]interface{}); ok {
					currentSkinIDFloat, okID := skinMap["id"].(float64)
					if !okID {
						continue
					}
					currentSkinID := int(currentSkinIDFloat)
					if currentSkinID == skinID {
						if chromasData, ok := skinMap["chromas"].([]interface{}); ok {
							for _, chromaInterface := range chromasData {
								if chromaMap, ok := chromaInterface.(map[string]interface{}); ok {
									chromaIDFloat, okCID := chromaMap["id"].(float64)
									chromaNameStr, okCName := chromaMap["name"].(string)
									if okCID && okCName {
										chromaNames[int(chromaIDFloat)] = chromaNameStr
									}
								}
							}
						}
						break
					}
				}
			}
		}
		log.Printf("Found %d chroma names from Supabase.", len(chromaNames))
		fyne.Do(func() {
			uiMutex.Lock()
			defer uiMutex.Unlock()
			for id, name := range chromaNames {
				if itemUI, ok := chromaCircleItems[id]; ok && itemUI.nameLabel != nil {
					itemUI.nameLabel.SetText(name)
				}
				if itemUI, ok := chromaImageItems[id]; ok && itemUI.nameLabel != nil {
					itemUI.nameLabel.SetText(name)
				}
			}
			if circlesGrid != nil {
				circlesGrid.Refresh()
			}
			if imagesGrid != nil {
				imagesGrid.Refresh()
			}
		})
	}(data.GetChampionIDFromSkinID(skin.ID), skin.ID)

	// --- Layout principal del diálogo (HBox) ---
	// Poner panel izquierdo y derecho uno al lado del otro
	dialogMainContent := container.NewHSplit(
		paddedLeftPanel,
		paddedRightPanel,
	)
	dialogMainContent.Offset = 0.5 // Ajusta el punto de división si es necesario (0.0 a 1.0)

	// --- Botones de Acción (Abajo a la derecha) ---
	downloadButton = widget.NewButtonWithIcon("Download Skin", theme.DownloadIcon(), func() { /* ... descarga ... */
		if selectedChromaID == nil {
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
		log.Printf("DOWNLOAD ACTION: Type=%s, Name=%s, ID=%d (Origin Skin: %d)", downloadType, downloadName, downloadID, skin.ID)
		currentApp := fyne.CurrentApp()
		if currentApp != nil {
			currentApp.SendNotification(&fyne.Notification{Title: "Download Initiated", Content: fmt.Sprintf("Starting download for %s: %s (ID: %d)", downloadType, downloadName, downloadID)})
		}
		dialog.ShowInformation("Download", fmt.Sprintf("Placeholder: Download initiated for %s: %s (ID: %d)", downloadType, downloadName, downloadID), parent)
	})
	closeButton := widget.NewButton("Close", func() {})
	// Usa Border para poner los botones abajo a la derecha
	actionButtons := container.NewBorder(
		nil,                // top
		nil,                // bottom (los botones estarán aquí)
		layout.NewSpacer(), // left spacer
		container.NewHBox(downloadButton, closeButton), // right (botones juntos)
		nil, // center (vacío)
	)

	// --- Diálogo Final ---
	finalDialogContent := container.NewBorder(
		nil,                                // Top: Título ya está en el diálogo
		container.NewPadded(actionButtons), // Bottom: Botones con padding
		nil,                                // Left
		nil,                                // Right
		dialogMainContent,                  // Center: El HSplit
	)

	customDialog := dialog.NewCustom(skin.Name, "", finalDialogContent, parent)
	closeButton.OnTapped = customDialog.Hide
	customDialog.Resize(fyne.NewSize(850, 550))
	customDialog.Show()
}

// Helper createChromaCircleItem - Ajusta tamaño círculo y placeholder
func createChromaCircleItem(name string, hexColors []string, itemID int, selectedID *int, onSelect func(id int)) chromaItemUI {
	const visualSize float32 = 48
	visualMinSize := fyne.NewSize(visualSize, visualSize)
	var displayElement fyne.CanvasObject
	if name == "Default" {
		baseIcon := widget.NewIcon(theme.RadioButtonCheckedIcon())
		bgCircle := canvas.NewCircle(theme.InputBackgroundColor())
		bgCircle.Resize(visualMinSize)
		displayElement = container.NewStack(bgCircle, container.NewCenter(baseIcon))
	} else if len(hexColors) > 0 {
		clr1, err1 := data.ParseHexColor(hexColors[0])
		if err1 != nil {
			clr1 = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
		}
		if len(hexColors) >= 2 && hexColors[1] != "" && hexColors[1] != hexColors[0] {
			clr2, err2 := data.ParseHexColor(hexColors[1])
			if err2 != nil {
				clr2 = clr1
			}
			gradient := canvas.NewHorizontalGradient(clr1, clr2)
			gradient.SetMinSize(visualMinSize)
			displayElement = gradient // Gradiente ya tiene tamaño
		} else {
			// *** CORRECTION: Forzar tamaño de círculo sólido ***
			circle := canvas.NewCircle(clr1)
			bgRect := canvas.NewRectangle(color.Transparent)
			bgRect.SetMinSize(visualMinSize)
			displayElement = container.NewStack(bgRect, circle) // Poner círculo sobre rectángulo con tamaño
		}
	} else {
		icon := widget.NewIcon(theme.QuestionIcon())
		bg := canvas.NewRectangle(theme.DisabledButtonColor())
		bg.SetMinSize(visualMinSize)
		displayElement = container.NewStack(bg, container.NewCenter(icon))
	}

	selectionIndicator := canvas.NewRectangle(color.Transparent)
	selectionIndicator.StrokeColor = theme.PrimaryColor()
	selectionIndicator.StrokeWidth = 2
	selectionIndicator.Resize(fyne.NewSize(visualSize+4, visualSize+4))
	indicatorWrapper := NewSelectionIndicatorWrapper(displayElement, selectionIndicator, func() bool { return selectedID != nil && *selectedID == itemID })

	nameLabel := widget.NewLabel(name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	itemContent := container.NewVBox(container.NewCenter(indicatorWrapper), nameLabel)
	card := NewTappableCard(container.NewPadded(itemContent), func() {
		if selectedID != nil && *selectedID != itemID {
			onSelect(itemID)
		}
	})

	// *** Altura consistente ***
	imgItemHeight := float32(64 + 40 + 20)
	circleItemWidth := visualSize + 40
	card.SetMinSize(fyne.NewSize(circleItemWidth, imgItemHeight))

	return chromaItemUI{widget: card, nameLabel: nameLabel}
}

// Helper createChromaImageItem - Ajusta placeholder
func createChromaImageItem(name string, chroma data.Chroma, itemID int, selectedID *int, onSelect func(id int)) chromaItemUI {
	const imgSize float32 = 64
	imgAreaSize := fyne.NewSize(imgSize, imgSize)
	imageStack := container.NewStack()
	visualElement := imageStack
	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgAreaSize)
	imageStack.Add(placeholderRect)
	imageStack.Add(container.NewCenter(placeholderIcon))
	if name != "Default" {
		go func(ch data.Chroma, stack *fyne.Container) { /* ... carga imagen async ... */
			imageURL := data.GetChromaImageURL(ch)
			if imageURL == data.GetPlaceholderImageURL() {
				return
			}
			chromaUri, err := storage.ParseURI(imageURL)
			if err != nil {
				return
			}
			imgWidget := canvas.NewImageFromURI(chromaUri)
			imgWidget.FillMode = canvas.ImageFillContain
			imgWidget.SetMinSize(imgAreaSize)
			fyne.Do(func() {
				if stack != nil && stack.Visible() {
					stack.Objects = []fyne.CanvasObject{imgWidget}
					stack.Refresh()
				}
			})
		}(chroma, imageStack)
	} else {
		defaultIcon := widget.NewIcon(theme.RadioButtonCheckedIcon())
		imageStack.Objects = []fyne.CanvasObject{placeholderRect, container.NewCenter(defaultIcon)}
	}
	selectionIndicator := canvas.NewRectangle(color.Transparent)
	selectionIndicator.StrokeColor = theme.PrimaryColor()
	selectionIndicator.StrokeWidth = 2
	selectionIndicator.Resize(fyne.NewSize(imgSize+4, imgSize+4))
	indicatorWrapper := NewSelectionIndicatorWrapper(visualElement, selectionIndicator, func() bool { return selectedID != nil && *selectedID == itemID })

	nameLabel := widget.NewLabel(name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	itemContent := container.NewVBox(container.NewCenter(indicatorWrapper), nameLabel)
	card := NewTappableCard(container.NewPadded(itemContent), func() {
		if selectedID != nil && *selectedID != itemID {
			onSelect(itemID)
		}
	})

	// *** Altura consistente ***
	imgItemHeight := float32(64 + 40 + 20)
	imgItemWidth := imgSize + 40
	card.SetMinSize(fyne.NewSize(imgItemWidth, imgItemHeight))

	return chromaItemUI{widget: card, nameLabel: nameLabel}
}

// --- SelectionIndicatorWrapper (sin cambios) ---
type SelectionIndicatorWrapper struct {
	widget.BaseWidget
	content    fyne.CanvasObject
	indicator  fyne.CanvasObject
	isSelected func() bool
}

func NewSelectionIndicatorWrapper(content, indicator fyne.CanvasObject, isSelected func() bool) *SelectionIndicatorWrapper {
	w := &SelectionIndicatorWrapper{content: content, indicator: indicator, isSelected: isSelected}
	w.ExtendBaseWidget(w)
	w.indicator.Hide()
	return w
}
func (w *SelectionIndicatorWrapper) CreateRenderer() fyne.WidgetRenderer {
	if w.isSelected != nil && w.isSelected() {
		w.indicator.Show()
	} else {
		w.indicator.Hide()
	}
	c := container.NewStack(w.content, container.NewCenter(w.indicator))
	return widget.NewSimpleRenderer(c)
}
func (w *SelectionIndicatorWrapper) Refresh() {
	if w.isSelected != nil && w.isSelected() {
		w.indicator.Show()
	} else {
		w.indicator.Hide()
	}
	w.BaseWidget.Refresh()
}
func (w *SelectionIndicatorWrapper) MinSize() fyne.Size { return w.content.MinSize() }

// --- End of skin_dialog.go ---
