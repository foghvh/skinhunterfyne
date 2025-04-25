// skinhunter/ui/utils.go
package ui

import (
	"image/color"
	"log"
	"net/url" // Needed for parseURL helper if moved here (it's in skin_dialog)

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"skinhunter/data"
)

// SkinItem: Revised for lazy image loading and styling.
func SkinItem(skin data.Skin, onSelect func(skin data.Skin)) fyne.CanvasObject {
	if skin.IsBase {
		return nil // Still ignore base skins
	}

	// --- Sizing ---
	itemWidth := float32(210)
	imgHeight := float32(160)
	totalHeight := float32(200)
	imgSize := fyne.NewSize(itemWidth, imgHeight)
	iconSize := fyne.NewSize(18, 18) // Target size for icons when loaded as canvas.Image

	// --- Placeholder & Image Container ---
	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgSize)
	imageContainer := container.NewStack(placeholderRect, container.NewCenter(placeholderIcon))
	imageContainer.Refresh()

	// --- Asynchronous Image Loading (Main Tile) ---
	go func() {
		imgURL := data.GetSkinTileURL(skin)
		if imgURL == data.GetPlaceholderImageURL() {
			return // Don't try to load the placeholder URL itself
		}
		uri, err := storage.ParseURI(imgURL)
		if err != nil {
			log.Printf("ERROR: SkinItem failed to parse URI [%s] for skin %d: %v", imgURL, skin.ID, err)
			return // Exit goroutine on parse error
		}
		loadedImage := canvas.NewImageFromURI(uri)
		loadedImage.FillMode = canvas.ImageFillContain
		loadedImage.SetMinSize(imgSize)

		if imageContainer != nil {
			imageContainer.Objects = []fyne.CanvasObject{loadedImage}
			imageContainer.Refresh()
		} else {
			log.Printf("WARN: SkinItem imageContainer was nil when image loaded for skin %d", skin.ID)
		}
	}() // End of image loading goroutine

	// --- Rarity Icon ---
	rarityName, rarityIconURL := data.Rarity(skin)
	var rarityIconWidget fyne.CanvasObject = layout.NewSpacer() // Default spacer
	rarityIconSize := fyne.NewSize(16, 16)                      // Smaller size for rarity gem

	if rarityIconURL != "" && rarityIconURL != data.GetPlaceholderImageURL() {
		rarityPlaceholder := canvas.NewRectangle(color.Transparent)
		rarityPlaceholder.SetMinSize(rarityIconSize)
		// Use a Stack to hold the placeholder then the loaded icon image
		rarityIconContainer := container.NewStack(rarityPlaceholder)
		rarityIconWidget = rarityIconContainer // Assign container to layout

		go func(url string, container *fyne.Container, size fyne.Size) {
			uri, err := storage.ParseURI(url)
			if err != nil {
				log.Printf("WARN: Failed to parse rarity icon URI %s: %v", url, err)
				// Optional: Replace placeholder with an error icon in the stack
				return
			}
			// Use canvas.NewImageFromURI for network icons
			icon := canvas.NewImageFromURI(uri)
			icon.SetMinSize(size)
			icon.FillMode = canvas.ImageFillContain

			if container != nil {
				container.Objects = []fyne.CanvasObject{icon} // Replace placeholder
				container.Refresh()
			}
		}(rarityIconURL, rarityIconContainer, rarityIconSize)
	} else if rarityName != "Standard" && rarityName != "" {
		// Optional: Display text if name exists but no icon URL
		// rarityIconWidget = widget.NewLabel(rarityName)
	}

	// --- Top Icons (Legacy, Chroma) ---
	topIcons := []fyne.CanvasObject{}

	// Function to create an icon placeholder and load image into it
	createLazyIcon := func(iconURL string, size fyne.Size) *fyne.Container {
		placeholder := canvas.NewRectangle(color.Transparent)
		placeholder.SetMinSize(size)
		iconContainer := container.NewStack(placeholder)

		go func(url string, cont *fyne.Container) {
			if url == data.GetPlaceholderImageURL() || url == "" {
				// Don't load empty/placeholder URLs, maybe show different placeholder?
				placeholderIconWidget := widget.NewIcon(theme.QuestionIcon()) // Or hide container?
				cont.Objects = []fyne.CanvasObject{placeholderIconWidget}
				cont.Refresh()
				return
			}
			uri, err := storage.ParseURI(url)
			if err != nil {
				log.Printf("WARN: Error parsing icon URI %s: %v", url, err)
				// Replace placeholder with an error icon
				errorIconWidget := widget.NewIcon(theme.ErrorIcon())
				cont.Objects = []fyne.CanvasObject{errorIconWidget}
				cont.Refresh()
				return
			}
			// Create canvas.Image for network resource
			iconImage := canvas.NewImageFromURI(uri)
			iconImage.SetMinSize(size)
			iconImage.FillMode = canvas.ImageFillContain
			cont.Objects = []fyne.CanvasObject{iconImage} // Replace placeholder
			cont.Refresh()

		}(iconURL, iconContainer)

		return iconContainer
	}

	if skin.IsLegacy {
		legacyURL := data.LegacyIconURL()
		legacyIconContainer := createLazyIcon(legacyURL, iconSize)
		topIcons = append(topIcons, legacyIconContainer) // Add the container
	}
	if len(skin.Chromas) > 0 {
		chromaURL := data.ChromaIconURL()
		chromaIconContainer := createLazyIcon(chromaURL, iconSize)
		if len(topIcons) > 0 {
			topIcons = append(topIcons, widget.NewSeparator()) // Visual separator
		}
		topIcons = append(topIcons, chromaIconContainer) // Add the container
	}

	topIconsContainer := container.NewHBox(layout.NewSpacer())
	if len(topIcons) > 0 {
		// Add the icons/containers directly to the HBox
		topIconsContainer.Add(container.NewHBox(topIcons...))
	}

	// --- Bottom Bar (Name & Rarity) ---
	nameLabel := widget.NewLabel(skin.Name)
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	bottomBarContent := container.NewBorder(
		nil, nil,
		container.NewPadded(rarityIconWidget),
		nil,
		nameLabel,
	)
	bottomBgColor := color.NRGBA{R: 0x10, G: 0x10, B: 0x15, A: 0xD0}
	bottomBgRect := canvas.NewRectangle(bottomBgColor)
	bottomBar := container.NewStack(bottomBgRect, bottomBarContent)

	// --- Assemble Card Content ---
	cardContentLayout := container.NewBorder(
		container.NewPadded(topIconsContainer),
		bottomBar,
		nil, nil,
		imageContainer, // Center: The stack containing placeholder/image
	)

	card := widget.NewCard("", "", cardContentLayout)
	tappableItem := NewTappableCard(card, func() {
		log.Printf("Skin selected: %s (ID: %d)", skin.Name, skin.ID)
		onSelect(skin)
	})
	tappableItem.SetMinSize(fyne.NewSize(itemWidth, totalHeight))

	return tappableItem
}

// --- Helpers (TappableCard, NewIconButton, NewTabButton) ---

// TappableCard makes any canvas object tappable.
type TappableCard struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	onTapped func()
	minSize  fyne.Size // Added to store minimum size
}

func NewTappableCard(content fyne.CanvasObject, onTap func()) *TappableCard {
	c := &TappableCard{content: content, onTapped: onTap}
	c.ExtendBaseWidget(c)
	return c
}

func (c *TappableCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

func (c *TappableCard) Tapped(_ *fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped()
	}
}

func (c *TappableCard) MinSize() fyne.Size {
	if c.minSize.IsZero() {
		return c.BaseWidget.MinSize()
	}
	return c.minSize
}

func (c *TappableCard) SetMinSize(size fyne.Size) {
	c.minSize = size
	c.Refresh() // Refresh widget when min size changes
}

func (c *TappableCard) Cursor() desktop.Cursor { return desktop.PointerCursor }

// NewIconButton creates a button with only an icon.
func NewIconButton(iconRes fyne.Resource, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon("", iconRes, onTap)
	// btn.Importance = widget.LowImportance // Optional styling
	return btn
}

// NewTabButton creates a vertical button with icon and label for footer.
func NewTabButton(label string, icon fyne.Resource, tapped func()) fyne.CanvasObject {
	btnIcon := canvas.NewImageFromResource(icon)
	btnIcon.SetMinSize(fyne.NewSize(24, 24)) // Standard icon size
	btnIcon.FillMode = canvas.ImageFillContain

	btnLabel := widget.NewLabel(label)
	btnLabel.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		container.NewCenter(btnIcon),
		btnLabel,
	)

	cardContent := container.NewPadded(content)
	return NewTappableCard(cardContent, tapped)
}

// Helper parseURL
func parseURL(rawURL string) *url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Error parsing URL '%s': %v", rawURL, err)
		return &url.URL{} // Return empty URL on error
	}
	return parsed
}

// --- End of utils.go ---
