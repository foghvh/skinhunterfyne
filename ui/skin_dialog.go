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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowSkinDialog displays the detailed skin dialog.
func ShowSkinDialog(skin data.Skin, allChromasForChamp []data.Chroma, parent fyne.Window) {
	log.Printf("Showing dialog for skin: %s (ID: %d)", skin.Name, skin.ID)

	// --- State ---
	selectedChromaID := new(int) // Pointer to hold the selected ID
	*selectedChromaID = skin.ID  // Default to base skin ID

	// --- Filter Chromas for this specific skin ---
	filteredChromas := make([]data.Chroma, 0)
	if allChromasForChamp != nil { // Guard against nil slice
		for _, ch := range allChromasForChamp {
			// OriginSkinID should be set during data fetch/processing
			if ch.OriginSkinID == skin.ID {
				filteredChromas = append(filteredChromas, ch)
			}
		}
	}
	log.Printf("Filtered %d chromas for skin ID %d", len(filteredChromas), skin.ID)

	// --- Build Dialog Content ---
	// Define the structure first, then populate

	// Left Side Elements
	splashContainer, splashImage := NewAsyncImage(0, 0) // Let size be determined by container later
	splashImage.FillMode = canvas.ImageFillContain      // Or ImageFillScale if aspect ratio is bad
	SetImageURL(splashImage, splashContainer, data.GetSkinSplashURL(skin))

	desc := skin.Description
	if desc == "" {
		desc = "This skin does not have a description."
	}
	descriptionLabel := widget.NewLabel(desc)
	descriptionLabel.Wrapping = fyne.TextWrapWord
	descriptionScroll := container.NewScroll(descriptionLabel)
	descriptionScroll.SetMinSize(fyne.NewSize(300, 100)) // Give description area some minimum size

	warningLabel := widget.NewLabel("This skin may not work properly due to game updates")
	warningIcon := widget.NewIcon(theme.WarningIcon())
	// Use a padded HBox, simple background isn't standard Fyne widget
	// warningColor := theme.WarningColor() // Removed unused var
	warningBox := container.NewPadded(container.NewHBox(warningIcon, warningLabel))

	leftPanel := container.NewVBox(
		splashContainer, // Image takes available space
		descriptionScroll,
		warningBox,
	)
	leftPanel.Resize(fyne.NewSize(450, 500)) // Approximate size, split will manage final

	// Right Side Elements
	modelViewerLink := widget.NewHyperlink("View on Model viewer", parseURL(data.KhadaUrl(skin.ID, 0))) // Initial URL
	modelInfoIcon := newIconButton(theme.InfoIcon(), func() {
		dialog.ShowInformation("Model Viewer Info", "Preview the in-game appearance of the skin.", parent)
	})
	modelViewerLinkBox := container.NewHBox(modelViewerLink, modelInfoIcon)

	chromaTitle := widget.NewLabel("Chromas")
	chromaInfoIcon := newIconButton(theme.InfoIcon(), func() {
		dialog.ShowInformation("Chroma Info", "Change skin colors.", parent)
	})
	chromaTitleBox := container.NewHBox(chromaTitle, chromaInfoIcon)

	// Chroma Tabs Setup
	var circlesRadio, imagesRadio *widget.RadioGroup // Declare radio groups

	updateChromaSelection := func(newID int) { // Simplified callback
		if *selectedChromaID == newID {
			return // No change
		}
		log.Printf("Chroma selected: ID %d", newID)
		*selectedChromaID = newID

		// Update model viewer link
		newURLStr := data.KhadaUrl(skin.ID, newID)
		modelViewerLink.SetURL(parseURL(newURLStr)) // Update hyperlink URL

		// Update radio group selections visually (find the option string for the newID)
		defaultStr := "Default"
		var optionToSelect string
		if newID == skin.ID {
			optionToSelect = defaultStr
		} else {
			for _, ch := range filteredChromas {
				if ch.ID == newID {
					optionToSelect = ch.Name
					break
				}
			}
		}

		if optionToSelect != "" {
			if circlesRadio != nil {
				circlesRadio.SetSelected(optionToSelect)
			}
			if imagesRadio != nil {
				imagesRadio.SetSelected(optionToSelect)
			}
		} else {
			log.Printf("WARN: Could not find option string for selected chroma ID %d", newID)
			// Fallback to selecting default?
			if circlesRadio != nil {
				circlesRadio.SetSelected(defaultStr)
			}
			if imagesRadio != nil {
				imagesRadio.SetSelected(defaultStr)
			}
			*selectedChromaID = skin.ID                                 // Reset selection state
			modelViewerLink.SetURL(parseURL(data.KhadaUrl(skin.ID, 0))) // Reset URL
		}
	}

	// Tab 1: Circles
	circlesGrid := container.NewGridWithColumns(4)
	circlesOptions := []string{"Default"}
	circlesMap := make(map[string]int)
	circlesMap["Default"] = skin.ID
	circlesGrid.Add(createChromaCircleItem("Default", nil, skin.ID, selectedChromaID, updateChromaSelection))
	for _, chroma := range filteredChromas {
		optionStr := chroma.Name
		circlesOptions = append(circlesOptions, optionStr)
		circlesMap[optionStr] = chroma.ID
		circlesGrid.Add(createChromaCircleItem(chroma.Name, chroma.Colors, chroma.ID, selectedChromaID, updateChromaSelection))
	}
	circlesRadio = widget.NewRadioGroup(circlesOptions, func(selectedOption string) {
		updateChromaSelection(circlesMap[selectedOption])
	})
	circlesRadio.Selected = "Default" // Initial state
	// We primarily use the tappable cards, so hide the radio itself
	circlesRadio.Hide()
	circlesScroll := container.NewScroll(circlesGrid)
	circlesTabContent := container.NewMax(circlesScroll, circlesRadio) // Layer hidden radio and grid
	circlesTabContent.Resize(fyne.NewSize(350, 250))                   // Give tab content size

	// Tab 2: Images
	imagesGrid := container.NewGridWithColumns(4)
	imagesOptions := []string{"Default"}
	imagesMap := make(map[string]int) // Re-use map name, careful with scope if needed
	imagesMap["Default"] = skin.ID
	imagesGrid.Add(createChromaImageItem("Default", "", skin.ID, selectedChromaID, updateChromaSelection))
	for _, chroma := range filteredChromas {
		optionStr := chroma.Name
		imagesOptions = append(imagesOptions, optionStr)
		imagesMap[optionStr] = chroma.ID
		imgURL := data.GetChromaImageURL(chroma)
		imagesGrid.Add(createChromaImageItem(chroma.Name, imgURL, chroma.ID, selectedChromaID, updateChromaSelection))
	}
	imagesRadio = widget.NewRadioGroup(imagesOptions, func(selectedOption string) {
		updateChromaSelection(imagesMap[selectedOption])
	})
	imagesRadio.Selected = "Default" // Initial state
	imagesRadio.Hide()
	imagesScroll := container.NewScroll(imagesGrid)
	imagesTabContent := container.NewMax(imagesScroll, imagesRadio) // Layer hidden radio and grid
	imagesTabContent.Resize(fyne.NewSize(350, 250))                 // Give tab content size

	// Assemble Tabs
	chromaTabs := container.NewAppTabs(
		container.NewTabItem("Circles", circlesTabContent),
		container.NewTabItem("Images", imagesTabContent),
	)

	// Credits Info
	creditsText := "This is going to consume a credit" // TODO: Get actual credits
	creditsInfoLabel := widget.NewLabel(creditsText)
	creditsBox := container.NewHBox(widget.NewIcon(theme.InfoIcon()), creditsInfoLabel)

	rightPanel := container.NewVBox(
		modelViewerLinkBox,
		widget.NewSeparator(),
		chromaTitleBox,
		chromaTabs,         // Let tabs take remaining space
		layout.NewSpacer(), // Push credits down
		creditsBox,
	)

	// --- Combine Panels and Define Dialog ---
	dialogContentSplit := container.NewHSplit(
		container.NewPadded(leftPanel),
		container.NewPadded(rightPanel),
	)
	dialogContentSplit.Offset = 0.55 // Left side slightly larger

	// --- Action Buttons ---
	downloadButtonText := func() string {
		if *selectedChromaID != skin.ID {
			return "Download Chroma"
		}
		return "Download Skin"
	}
	var downloadButton *widget.Button // Declare to allow updating text
	downloadButton = widget.NewButtonWithIcon(downloadButtonText(), theme.DownloadIcon(), func() {
		downloadID := *selectedChromaID
		downloadType := "Skin"
		downloadName := skin.Name
		if downloadID != skin.ID {
			downloadType = "Chroma"
			for _, ch := range filteredChromas {
				if ch.ID == downloadID {
					downloadName = ch.Name
					break
				}
			}
		}
		log.Printf("DOWNLOAD ACTION: Type=%s, Name=%s, ID=%d", downloadType, downloadName, downloadID)
		// TODO: Implement actual download logic via backend calls
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Download Requested",
			Content: fmt.Sprintf("Downloading %s: %s (ID: %d)", downloadType, downloadName, downloadID),
		})
		// dialog.Hide() // Maybe hide dialog after download starts?
	})

	// Add listener to update download button text when chroma changes (a bit tricky)
	// Simplest: Just use a generic "Download" text or update inside the OnSelect callback.
	// Let's modify updateChromaSelection to refresh the button text:
	// Re-declare updateChromaSelection to capture downloadButton
	updateChromaSelection = func(newID int) { // Redefine to capture downloadButton
		if *selectedChromaID == newID {
			return
		}
		log.Printf("Chroma selected: ID %d", newID)
		*selectedChromaID = newID
		newURLStr := data.KhadaUrl(skin.ID, newID)
		modelViewerLink.SetURL(parseURL(newURLStr))
		defaultStr := "Default"
		var optionToSelect string
		if newID == skin.ID {
			optionToSelect = defaultStr
		} else {
			for _, ch := range filteredChromas {
				if ch.ID == newID {
					optionToSelect = ch.Name
					break
				}
			}
		}
		if optionToSelect != "" {
			if circlesRadio != nil {
				circlesRadio.SetSelected(optionToSelect)
			}
			if imagesRadio != nil {
				imagesRadio.SetSelected(optionToSelect)
			}
		} else { // Fallback on bad selection
			log.Printf("WARN: Could not find option string for selected chroma ID %d", newID)
			if circlesRadio != nil {
				circlesRadio.SetSelected(defaultStr)
			}
			if imagesRadio != nil {
				imagesRadio.SetSelected(defaultStr)
			}
			*selectedChromaID = skin.ID                                 // Reset internal state too
			modelViewerLink.SetURL(parseURL(data.KhadaUrl(skin.ID, 0))) // Reset link
		}
		// Update download button text AFTER state change
		downloadButton.SetText(downloadButtonText())
	}
	// End of re-declaration for capture

	// --- Dialog Definition ---
	var customDialog dialog.Dialog // Define variable for the dialog

	closeButton := widget.NewButton("Close", func() {
		customDialog.Hide()
	})
	actionButtons := container.NewHBox(layout.NewSpacer(), closeButton, downloadButton) // Align buttons right
	finalDialogContent := container.NewBorder(nil, actionButtons, nil, nil, dialogContentSplit)

	// Create the custom dialog, passing the already constructed content
	customDialog = dialog.NewCustom(skin.Name, "Dismiss", finalDialogContent, parent) // "Dismiss" is the button label for Escape key
	customDialog.Resize(fyne.NewSize(850, 550))                                       // Target size
	customDialog.Show()
}

// Helper to create a chroma item card (Circle version)
func createChromaCircleItem(name string, hexColors []string, itemID int, selectedID *int, onSelect func(id int)) fyne.CanvasObject {

	const visualSize float32 = 48 // Slightly smaller circles for better grid fit
	const itemWidth float32 = 80
	const itemHeight float32 = 110

	placeholderRect := canvas.NewRectangle(theme.ButtonColor()) // Grey background
	placeholderRect.SetMinSize(fyne.NewSize(visualSize, visualSize))
	displayContainer := container.NewStack(placeholderRect) // Base stack

	if name == "Default" {
		icon := widget.NewIcon(theme.CancelIcon())      // Use cancel/ignore icon for default
		displayContainer.Add(container.NewCenter(icon)) // Center icon over placeholder
	} else if len(hexColors) > 0 {
		clr1, err1 := data.ParseHexColor(hexColors[0])
		if err1 != nil {
			clr1 = color.NRGBA{R: 128, G: 128, B: 128, A: 255} // Fallback grey
		}

		if len(hexColors) == 1 || hexColors[1] == "" { // Single color
			circle := canvas.NewCircle(clr1)
			circle.StrokeWidth = 1
			circle.StrokeColor = theme.InputBorderColor()
			displayContainer.Add(container.NewCenter(circle)) // Stack circle
		} else { // Two colors
			clr2, err2 := data.ParseHexColor(hexColors[1])
			if err2 != nil {
				clr2 = clr1
			} // Fallback to first color

			// Simulate split circle with gradient rectangle (best effort without custom render)
			gradient := canvas.NewHorizontalGradient(clr1, clr2) // Horizontal split is simple
			gradient.SetMinSize(fyne.NewSize(visualSize, visualSize))
			displayContainer.Objects = []fyne.CanvasObject{gradient} // Replace placeholder entirely

			// Or keep placeholder and overlay gradient:
			// gradientContainer := container.NewMax(gradient)
			// gradientContainer.Resize(fyne.NewSize(visualSize, visualSize))
			// displayContainer.Add(gradientContainer)
		}
	} else { // No colors provided and not default - Show question mark?
		icon := widget.NewIcon(theme.QuestionIcon())
		displayContainer.Add(container.NewCenter(icon))
	}

	// Ensure final size consistency
	displayContainer.Resize(fyne.NewSize(visualSize, visualSize))

	nameLabel := widget.NewLabel(name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff // Don't wrap chroma names

	// Combine centered display and label
	itemContent := container.NewVBox(
		container.NewCenter(displayContainer), // Center the visual part
		nameLabel,
	)
	itemWrapper := container.NewPadded(itemContent)         // Add padding around
	itemWrapper.Resize(fyne.NewSize(itemWidth, itemHeight)) // Enforce item size

	// Tappable card makes the whole area clickable
	card := newTappableCard(itemWrapper, func() {
		onSelect(itemID) // Trigger selection update
	})

	return card
}

// Helper to create a chroma item card (Image version)
func createChromaImageItem(name, imageURL string, itemID int, selectedID *int, onSelect func(id int)) fyne.CanvasObject {
	const imgSize float32 = 64 // Smaller images for chromas usually
	const itemWidth float32 = 80
	const itemHeight float32 = 110

	var visualElement fyne.CanvasObject

	if name == "Default" {
		icon := widget.NewIcon(theme.CancelIcon())
		iconContainer := container.NewCenter(icon)
		placeholder := canvas.NewRectangle(theme.ButtonColor())
		placeholder.SetMinSize(fyne.NewSize(imgSize, imgSize))
		visualElement = container.NewStack(placeholder, iconContainer)
		visualElement.Resize(fyne.NewSize(imgSize, imgSize)) // Ensure size
	} else {
		imgContainer, imgWidget := NewAsyncImage(imgSize, imgSize)
		SetImageURL(imgWidget, imgContainer, imageURL)
		imgWidget.FillMode = canvas.ImageFillContain // Contain ensures whole image is visible
		visualElement = imgContainer                 // Use the container returned by NewAsyncImage
	}

	nameLabel := widget.NewLabel(name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff

	itemContent := container.NewVBox(
		container.NewCenter(visualElement), // Center the visual part
		nameLabel,
	)
	itemWrapper := container.NewPadded(itemContent)
	itemWrapper.Resize(fyne.NewSize(itemWidth, itemHeight))

	card := newTappableCard(itemWrapper, func() {
		onSelect(itemID)
	})

	return card
}
