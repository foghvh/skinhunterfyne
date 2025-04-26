// skinhunter/ui/skin_dialog.go
package ui

import (
	"fmt"
	"image/color"
	"log"

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
			ch.OriginSkinID = data.GetChampionIDFromSkinID(ch.ID)*1000 + (ch.ID % 1000)
		}
		if ch.OriginSkinID == skin.ID {
			filteredChromas = append(filteredChromas, ch)
		}
	}
	log.Printf("Found %d chromas associated with skin ID %d ('%s')", len(filteredChromas), skin.ID, skin.Name)

	imgMinSize := fyne.NewSize(400, 230)
	imageStack := container.NewStack()
	leftImage := imageStack
	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgMinSize)
	imageStack.Add(placeholderRect)
	imageStack.Add(container.NewCenter(placeholderIcon))
	go func(s data.Skin, stack *fyne.Container) { // Load splash
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
				log.Printf("Dialog loaded splash for skin %d", s.ID)
			}
		})
	}(skin, imageStack)

	desc := skin.Description
	if desc == "" {
		desc = "This skin does not have a description."
	}
	descriptionLabel := widget.NewLabel(desc)
	descriptionLabel.Wrapping = fyne.TextWrapWord
	descContainer := container.NewMax(descriptionLabel)
	descriptionScroll := container.NewScroll(descContainer)
	descriptionScroll.SetMinSize(fyne.NewSize(300, 100))
	warningLabel := widget.NewLabel("Note: Skin functionality depends on game compatibility.")
	warningIcon := widget.NewIcon(theme.WarningIcon())
	warningBox := container.NewPadded(container.NewHBox(warningIcon, warningLabel))
	leftPanel := container.NewVBox(leftImage, descriptionScroll, warningBox)

	var downloadButton *widget.Button
	var circlesGrid, imagesGrid *fyne.Container
	updateSelectionUI := func(newID int) {
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
	modelViewerURL := parseURL(data.KhadaUrl(skin.ID, 0))
	modelViewerLink := widget.NewHyperlink("View on Model Viewer", modelViewerURL)
	modelInfoIcon := NewIconButton(theme.InfoIcon(), func() { dialog.ShowInformation("Model Viewer", "...", parent) })
	modelViewerLinkBox := container.NewHBox(modelViewerLink, layout.NewSpacer(), modelInfoIcon)
	chromaTitle := widget.NewLabel("Chromas")
	chromaInfoIcon := NewIconButton(theme.InfoIcon(), func() { dialog.ShowInformation("Chromas", "...", parent) })
	chromaTitleBox := container.NewHBox(chromaTitle, layout.NewSpacer(), chromaInfoIcon)
	selectDownloadLabel := widget.NewLabel("Select a variation below:")
	selectDownloadLabel.TextStyle = fyne.TextStyle{Italic: true}
	selectDownloadLabel.Alignment = fyne.TextAlignCenter

	circlesGrid = container.NewGridWithColumns(4)
	circlesGrid.Add(createChromaCircleItem("Default", nil, skin.ID, selectedChromaID, updateSelectionUI))
	for _, chroma := range filteredChromas {
		circlesGrid.Add(createChromaCircleItem(chroma.Name, chroma.Colors, chroma.ID, selectedChromaID, updateSelectionUI))
	}
	circlesTabContent := container.NewScroll(circlesGrid)
	imagesGrid = container.NewGridWithColumns(4)
	imagesGrid.Add(createChromaImageItem("Default", data.Chroma{OriginSkinID: skin.ID, ID: skin.ID}, skin.ID, selectedChromaID, updateSelectionUI))
	for _, chroma := range filteredChromas {
		imagesGrid.Add(createChromaImageItem(chroma.Name, chroma, chroma.ID, selectedChromaID, updateSelectionUI))
	}
	imagesTabContent := container.NewScroll(imagesGrid)
	chromaTabs := container.NewAppTabs(container.NewTabItemWithIcon("Circles", theme.ColorPaletteIcon(), circlesTabContent), container.NewTabItemWithIcon("Images", theme.DocumentIcon(), imagesTabContent))
	chromaTabs.SetTabLocation(container.TabLocationTop)

	creditsText := "Note: Downloading may require credits (Not Implemented)."
	creditsInfoLabel := widget.NewLabel(creditsText)
	creditsBox := container.NewHBox(widget.NewIcon(theme.InfoIcon()), creditsInfoLabel)
	rightPanel := container.NewVBox(modelViewerLinkBox, widget.NewSeparator(), chromaTitleBox, selectDownloadLabel, chromaTabs, layout.NewSpacer(), creditsBox)
	rightPanelContainer := container.NewPadded(rightPanel)
	dialogContentSplit := container.NewHSplit(container.NewPadded(leftPanel), rightPanelContainer)
	dialogContentSplit.Offset = 0.5

	downloadButton = widget.NewButtonWithIcon("Download Skin", theme.DownloadIcon(), func() { /* ... download logic ... */
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

// Helper createChromaCircleItem
func createChromaCircleItem(name string, hexColors []string, itemID int, selectedID *int, onSelect func(id int)) fyne.CanvasObject { /* ... as before ... */
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
			displayElement = gradient
		} else {
			circle := canvas.NewCircle(clr1)
			circle.Resize(visualMinSize)
			displayElement = circle
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
	card.SetMinSize(fyne.NewSize(visualSize+20, visualSize+40))
	return card
}

// Helper createChromaImageItem
func createChromaImageItem(name string, chroma data.Chroma, itemID int, selectedID *int, onSelect func(id int)) fyne.CanvasObject { /* ... as before ... uses fyne.Do ... */
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
		go func(ch data.Chroma, stack *fyne.Container) {
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
	card.SetMinSize(fyne.NewSize(imgSize+20, imgSize+40))
	return card
}

// SelectionIndicatorWrapper definition remains the same
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
