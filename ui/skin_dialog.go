// skinhunter/ui/skin_dialog.go
package ui

import (
	"fmt"
	"image/color"
	"log"

	// Keep for Khada link parsing
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

// ShowSkinDialog displays the detailed skin dialog.
// Uses data package functions for URLs, ensuring consistency.
func ShowSkinDialog(skin data.Skin, allChromasForChamp []data.Chroma, parent fyne.Window) {
	log.Printf("Showing dialog for skin: %s (ID: %d)", skin.Name, skin.ID)
	if allChromasForChamp == nil {
		log.Println("WARN: ShowSkinDialog received nil allChromasForChamp slice.")
		allChromasForChamp = []data.Chroma{} // Avoid nil pointer dereference
	}

	selectedChromaID := new(int) // Pointer to track selection
	*selectedChromaID = skin.ID  // Default selection is the base skin

	// Filter chromas specifically belonging to the *selected skin*
	// The allChromasForChamp passed in should ideally already be filtered by champion,
	// but this ensures we only show chromas for *this* skin.
	filteredChromas := make([]data.Chroma, 0)
	for _, ch := range allChromasForChamp {
		// Ensure OriginSkinID is set (should be done in data layer)
		if ch.OriginSkinID == 0 {
			// Attempt to derive if missing (less reliable)
			derivedOriginID := data.GetChampionIDFromSkinID(ch.ID)*1000 + (ch.ID % 1000)
			log.Printf("WARN: Chroma %d (%s) missing OriginSkinID in dialog. Derived as %d.", ch.ID, ch.Name, derivedOriginID)
			ch.OriginSkinID = derivedOriginID // Use derived ID cautiously
		}
		// Check if the chroma belongs to the currently viewed skin
		if ch.OriginSkinID == skin.ID {
			filteredChromas = append(filteredChromas, ch)
		}
	}
	log.Printf("Found %d chromas associated with skin ID %d ('%s')", len(filteredChromas), skin.ID, skin.Name)

	// --- Left Side Elements ---
	imgMinSize := fyne.NewSize(400, 230) // Keep desired size
	var leftImage fyne.CanvasObject
	imageStack := container.NewStack() // Use stack for lazy loading splash
	leftImage = imageStack

	// Initial placeholder
	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgMinSize)
	imageStack.Add(placeholderRect)
	imageStack.Add(container.NewCenter(placeholderIcon))
	imageStack.Refresh()

	// Lazy load splash image
	go func(s data.Skin, stack *fyne.Container) {
		splashUrl := data.GetSkinSplashURL(s) // Use data package function
		if splashUrl == data.GetPlaceholderImageURL() {
			return // Don't load placeholder
		}
		splashUri, err := storage.ParseURI(splashUrl)
		if err != nil {
			log.Printf("Error parsing splash URI [%s] in dialog: %v", splashUrl, err)
			// Optionally update placeholder to show error
			return
		}

		splashImage := canvas.NewImageFromURI(splashUri)
		splashImage.FillMode = canvas.ImageFillContain // Maintain aspect ratio
		splashImage.SetMinSize(imgMinSize)

		if stack != nil {
			stack.Objects = []fyne.CanvasObject{splashImage} // Replace placeholder
			stack.Refresh()
			log.Printf("Dialog loaded splash for skin %d", s.ID)
		}
	}(skin, imageStack)

	// Description
	desc := skin.Description
	if desc == "" {
		desc = "This skin does not have a description." // Default text
	}
	descriptionLabel := widget.NewLabel(desc)
	descriptionLabel.Wrapping = fyne.TextWrapWord
	// Scrollable description area
	descContainer := container.NewMax(descriptionLabel) // Max ensures label uses available width
	descriptionScroll := container.NewScroll(descContainer)
	descriptionScroll.SetMinSize(fyne.NewSize(300, 100)) // Min size for scroll area

	// Warning message (static)
	warningLabel := widget.NewLabel("Note: Skin functionality depends on game compatibility.") // Adjusted text
	warningIcon := widget.NewIcon(theme.WarningIcon())
	warningBox := container.NewPadded(container.NewHBox(warningIcon, warningLabel))

	// Assemble left panel
	leftPanel := container.NewVBox(leftImage, descriptionScroll, warningBox)

	// --- Right Side Elements ---
	var downloadButton *widget.Button // Declare button variable

	// Function to update download button text and Model Viewer link based on selection
	updateSelectionUI := func(newID int) {
		*selectedChromaID = newID
		// Update download button text
		if downloadButton != nil {
			buttonText := "Download Skin"
			if *selectedChromaID != skin.ID {
				buttonText = "Download Chroma"
			}
			downloadButton.SetText(buttonText)
		}
		// Update Model Viewer link (implementation needs widget reference or recreation)
		// For simplicity, we might just keep the base skin link or update it here if complex state management is added.
		// Khada link logic remains the same in data package.
	}

	// Model Viewer Link (initially points to base skin)
	modelViewerURL := parseURL(data.KhadaUrl(skin.ID, 0)) // Initial URL
	modelViewerLink := widget.NewHyperlink("View on Model Viewer", modelViewerURL)
	// TODO: Update modelViewerLink URL when chroma is selected if needed
	modelInfoIcon := NewIconButton(theme.InfoIcon(), func() {
		dialog.ShowInformation("Model Viewer", "Opens the skin/chroma on modelviewer.lol in your browser.", parent)
	})
	modelViewerLinkBox := container.NewHBox(modelViewerLink, layout.NewSpacer(), modelInfoIcon) // Spacer pushes icon right

	// Chroma Selection Title
	chromaTitle := widget.NewLabel("Chromas")
	chromaInfoIcon := NewIconButton(theme.InfoIcon(), func() {
		dialog.ShowInformation("Chromas", "Select a chroma variation. The 'Default' option represents the base skin.", parent)
	})
	chromaTitleBox := container.NewHBox(chromaTitle, layout.NewSpacer(), chromaInfoIcon) // Spacer pushes icon right

	// Helper text
	selectDownloadLabel := widget.NewLabel("Select a variation below:")
	selectDownloadLabel.TextStyle = fyne.TextStyle{Italic: true}
	selectDownloadLabel.Alignment = fyne.TextAlignCenter

	// --- Chroma Tabs ---
	// Tab 1: Circles (Visual representation using colors)
	circlesGrid := container.NewGridWithColumns(4) // Responsive grid for circles
	circlesGrid.Add(createChromaCircleItem("Default", nil, skin.ID, selectedChromaID, updateSelectionUI))
	for _, chroma := range filteredChromas {
		chromaCopy := chroma // Capture range variable
		circlesGrid.Add(createChromaCircleItem(chromaCopy.Name, chromaCopy.Colors, chromaCopy.ID, selectedChromaID, updateSelectionUI))
	}
	circlesTabContent := container.NewScroll(circlesGrid) // Make circles scrollable if many

	// Tab 2: Images (Visual representation using chroma images)
	imagesGrid := container.NewGridWithColumns(4) // Responsive grid for images
	// Pass empty Chroma for default case
	imagesGrid.Add(createChromaImageItem("Default", data.Chroma{OriginSkinID: skin.ID, ID: skin.ID}, skin.ID, selectedChromaID, updateSelectionUI))
	for _, chroma := range filteredChromas {
		chromaCopy := chroma // Capture range variable
		imagesGrid.Add(createChromaImageItem(chromaCopy.Name, chromaCopy, chromaCopy.ID, selectedChromaID, updateSelectionUI))
	}
	imagesTabContent := container.NewScroll(imagesGrid) // Make images scrollable if many

	chromaTabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Circles", theme.ColorPaletteIcon(), circlesTabContent),
		container.NewTabItemWithIcon("Images", theme.DocumentIcon(), imagesTabContent),
	)
	chromaTabs.SetTabLocation(container.TabLocationTop) // Tabs at top are standard

	// --- Credits Info --- (Static for now)
	creditsText := "Note: Downloading may require credits (Not Implemented)."
	creditsInfoLabel := widget.NewLabel(creditsText)
	creditsBox := container.NewHBox(widget.NewIcon(theme.InfoIcon()), creditsInfoLabel)

	// --- Assemble Right Panel ---
	rightPanel := container.NewVBox(
		modelViewerLinkBox,
		widget.NewSeparator(),
		chromaTitleBox,
		selectDownloadLabel,
		chromaTabs,         // Tabs take up flexible space
		layout.NewSpacer(), // Pushes credits box down
		creditsBox,
	)
	rightPanelContainer := container.NewPadded(rightPanel) // Pad the whole right side

	// --- Dialog Content Split ---
	dialogContentSplit := container.NewHSplit(container.NewPadded(leftPanel), rightPanelContainer)
	dialogContentSplit.Offset = 0.5 // Start with equal split

	// --- Action Buttons ---
	downloadButton = widget.NewButtonWithIcon("Download Skin", theme.DownloadIcon(), func() {
		if selectedChromaID == nil {
			log.Println("Error: selectedChromaID is nil in download action")
			dialog.ShowError(fmt.Errorf("internal error: selection tracking failed"), parent)
			return
		}
		downloadID := *selectedChromaID // Get the selected ID
		downloadType := "Skin"
		downloadName := skin.Name // Default to base skin name

		if downloadID != skin.ID {
			downloadType = "Chroma"
			// Find the selected chroma's name
			found := false
			for _, ch := range filteredChromas {
				if ch.ID == downloadID {
					downloadName = ch.Name // Use specific chroma name
					found = true
					break
				}
			}
			if !found {
				// Fallback if name somehow not found
				downloadName = fmt.Sprintf("Chroma ID %d", downloadID)
				log.Printf("WARN: Could not find name for selected chroma ID %d", downloadID)
			}
		}

		log.Printf("DOWNLOAD ACTION: Type=%s, Name=%s, ID=%d (Origin Skin: %d)", downloadType, downloadName, downloadID, skin.ID)
		// Placeholder for actual download logic
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Download Initiated",
			Content: fmt.Sprintf("Starting download for %s: %s (ID: %d)", downloadType, downloadName, downloadID),
		})
		// Show confirmation/info dialog
		dialog.ShowInformation("Download", fmt.Sprintf("Placeholder: Download initiated for %s: %s (ID: %d)", downloadType, downloadName, downloadID), parent)
		// Here you would trigger the actual download process with downloadID
	})

	// Need a variable for the dialog itself to close it
	var customDialog dialog.Dialog

	closeButton := widget.NewButton("Close", func() {
		if customDialog != nil {
			customDialog.Hide()
		} else {
			log.Println("Error: customDialog is nil on close attempt")
		}
	})

	// Arrange buttons using Border layout for standard alignment
	actionButtons := container.NewBorder(nil, nil, closeButton, downloadButton) // Close left, Download right

	// --- Final Dialog Assembly ---
	finalDialogContent := container.NewBorder(
		nil,                                // Top: nil
		container.NewPadded(actionButtons), // Bottom: Padded action buttons
		nil,                                // Left: nil
		nil,                                // Right: nil
		dialogContentSplit,                 // Center: The main HSplit content
	)

	// Create and show the custom dialog
	customDialog = dialog.NewCustom(skin.Name, "Dismiss", finalDialogContent, parent)
	customDialog.Resize(fyne.NewSize(850, 550)) // Keep specified size
	customDialog.Show()
}

// Helper createChromaCircleItem: Creates a tappable circle for chroma selection.
// Uses selectedID pointer for state tracking and onSelect callback.
func createChromaCircleItem(name string, hexColors []string, itemID int, selectedID *int, onSelect func(id int)) fyne.CanvasObject {
	const visualSize float32 = 48
	visualMinSize := fyne.NewSize(visualSize, visualSize)
	var displayElement fyne.CanvasObject

	// Create the visual part (circle, gradient, icon)
	if name == "Default" || itemID == *selectedID { // Show icon for default or initially selected
		// Base skin representation (e.g., simple icon)
		baseIcon := widget.NewIcon(theme.CheckButtonCheckedIcon()) // Or theme.RadioButtonCheckedIcon()
		bgCircle := canvas.NewCircle(theme.InputBackgroundColor()) // Background circle
		bgCircle.Resize(visualMinSize)
		displayElement = container.NewStack(bgCircle, container.NewCenter(baseIcon))
	} else if len(hexColors) > 0 {
		// Attempt to parse the first color
		clr1, err1 := data.ParseHexColor(hexColors[0])
		if err1 != nil {
			log.Printf("WARN: Error parsing hex color '%s' for chroma '%s': %v. Using gray.", hexColors[0], name, err1)
			clr1 = color.NRGBA{R: 128, G: 128, B: 128, A: 255} // Fallback gray
		}

		// Check for second color to create gradient
		if len(hexColors) >= 2 && hexColors[1] != "" && hexColors[1] != hexColors[0] {
			clr2, err2 := data.ParseHexColor(hexColors[1])
			if err2 != nil {
				log.Printf("WARN: Error parsing second hex color '%s' for chroma '%s': %v. Using first color.", hexColors[1], name, err2)
				clr2 = clr1 // Fallback to first color
			}
			// Use gradient if two distinct colors are valid
			gradient := canvas.NewHorizontalGradient(clr1, clr2)
			gradient.SetMinSize(visualMinSize) // Ensure gradient fills the area
			// Apply a circular mask (more complex, maybe skip for simplicity or use rectangle)
			// For simplicity, let's use a rectangle gradient for now.
			// displayElement = container.NewMask(gradient, canvas.NewCircle(color.White)) // Example mask
			displayElement = gradient
		} else {
			// Solid color circle if only one color or second is same/invalid
			circle := canvas.NewCircle(clr1)
			circle.Resize(visualMinSize)
			displayElement = circle
		}
	} else {
		// Fallback if no colors provided (shouldn't happen ideally)
		icon := widget.NewIcon(theme.QuestionIcon())
		bg := canvas.NewRectangle(theme.DisabledButtonColor())
		bg.SetMinSize(visualMinSize)
		displayElement = container.NewStack(bg, container.NewCenter(icon))
	}

	// Selection Indicator (rectangle border)
	selectionIndicator := canvas.NewRectangle(color.Transparent)
	selectionIndicator.StrokeColor = theme.PrimaryColor()
	selectionIndicator.StrokeWidth = 2
	selectionIndicator.Resize(fyne.NewSize(visualSize+4, visualSize+4)) // Slightly larger than visual element
	selectionIndicator.Hide()                                           // Hidden by default

	// Stack the visual element and the indicator
	displayStack := container.NewStack(displayElement, container.NewCenter(selectionIndicator))

	// Label for the chroma name
	nameLabel := widget.NewLabel(name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.Truncation = fyne.TextTruncateEllipsis // Truncate long names
	nameLabel.Wrapping = fyne.TextWrapOff

	// Combine stack and label vertically
	itemContent := container.NewVBox(container.NewCenter(displayStack), nameLabel)

	// Create tappable card, update selection state on tap
	card := NewTappableCard(container.NewPadded(itemContent), func() {
		// Update internal state via pointer
		*selectedID = itemID
		// Trigger external callback (e.g., to update download button)
		onSelect(itemID)
		// Refresh UI elements that depend on selection (this needs better state management)
		// selectionIndicator.Show() // Show indicator for this one
		// Need to hide indicators on others - requires reference to all items or a refresh of the parent grid
		log.Printf("Chroma circle selected: %s (ID: %d)", name, itemID)
	})

	// Initial check if this item should be selected
	if selectedID != nil && *selectedID == itemID {
		selectionIndicator.Show()
	}

	// Set min size for the card itself to guide grid layout
	card.SetMinSize(fyne.NewSize(visualSize+20, visualSize+40)) // Estimate size needed for circle + label + padding

	return card
}

// Helper createChromaImageItem: Creates a tappable image tile for chroma selection.
// Uses lazy loading for chroma images.
func createChromaImageItem(name string, chroma data.Chroma, itemID int, selectedID *int, onSelect func(id int)) fyne.CanvasObject {
	const imgSize float32 = 64 // Size for the image tile
	imgAreaSize := fyne.NewSize(imgSize, imgSize)
	var visualElement fyne.CanvasObject
	imageStack := container.NewStack() // Use stack for placeholder/image
	visualElement = imageStack         // Assign stack to layout

	// --- Placeholder ---
	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgAreaSize)
	imageStack.Add(placeholderRect)
	imageStack.Add(container.NewCenter(placeholderIcon))
	imageStack.Refresh()

	// --- Asynchronous Image Loading ---
	// Don't try to load image if it's the "Default" placeholder item
	if name != "Default" {
		go func(ch data.Chroma, stack *fyne.Container) {
			imageURL := data.GetChromaImageURL(ch) // Use data package function
			if imageURL == data.GetPlaceholderImageURL() {
				return // Don't load placeholder URL
			}
			chromaUri, err := storage.ParseURI(imageURL)
			if err != nil {
				log.Printf("Error parsing chroma image URI [%s] for chroma %d: %v", imageURL, ch.ID, err)
				// Optionally update placeholder to show error
				return
			}

			imgWidget := canvas.NewImageFromURI(chromaUri)
			imgWidget.FillMode = canvas.ImageFillContain
			imgWidget.SetMinSize(imgAreaSize)

			if stack != nil {
				stack.Objects = []fyne.CanvasObject{imgWidget} // Replace placeholder
				stack.Refresh()
				// log.Printf("Dialog loaded chroma image for %d", ch.ID)
			}
		}(chroma, imageStack)
	} else {
		// Special handling for "Default" item if needed (e.g., different icon)
		defaultIcon := widget.NewIcon(theme.CheckButtonCheckedIcon()) // Or other indicator
		imageStack.Objects = []fyne.CanvasObject{placeholderRect, container.NewCenter(defaultIcon)}
		imageStack.Refresh()
	}

	// --- Selection Indicator ---
	selectionIndicator := canvas.NewRectangle(color.Transparent)
	selectionIndicator.StrokeColor = theme.PrimaryColor()
	selectionIndicator.StrokeWidth = 2
	selectionIndicator.Resize(fyne.NewSize(imgSize+4, imgSize+4))
	selectionIndicator.Hide()

	// Stack visual element and indicator
	visualStack := container.NewStack(visualElement, container.NewCenter(selectionIndicator))

	// --- Label ---
	nameLabel := widget.NewLabel(name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff

	// --- Assemble Content ---
	itemContent := container.NewVBox(container.NewCenter(visualStack), nameLabel)

	// --- Tappable Card ---
	card := NewTappableCard(container.NewPadded(itemContent), func() {
		*selectedID = itemID
		onSelect(itemID)
		// Need to update selection indicators across all items
		log.Printf("Chroma image selected: %s (ID: %d)", name, itemID)
	})

	// Initial selection state
	if selectedID != nil && *selectedID == itemID {
		selectionIndicator.Show()
	}

	// Set min size for grid layout
	card.SetMinSize(fyne.NewSize(imgSize+20, imgSize+40))

	return card
}

// Helper parseURL (copied from utils.go for self-containment if needed, or ensure import)
// func parseURL(rawURL string) *url.URL { ... } // Already defined in utils.go

// --- End of skin_dialog.go ---
