// skinhunter/ui/utils.go
package ui

import (
	"image/color"
	"log"
	"net/url"

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

// SkinItem crea el objeto visual para una skin, usando TappableCard.
// Ya no necesita parámetro 'app' o 'window', usa fyne.Do internamente.
func SkinItem(skin data.Skin, onSelect func(skin data.Skin)) fyne.CanvasObject {
	if skin.IsBase {
		return nil
	}

	itemWidth := float32(210)
	imgHeight := float32(160)
	totalHeight := float32(200)
	imgSize := fyne.NewSize(itemWidth, imgHeight)
	iconSize := fyne.NewSize(18, 18)
	rarityIconSize := fyne.NewSize(16, 16)

	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgSize)
	imageContainer := container.NewStack(placeholderRect, container.NewCenter(placeholderIcon))

	go func() { // Load main image
		imgURL := data.GetSkinTileURL(skin)
		if imgURL == data.GetPlaceholderImageURL() {
			return
		}
		uri, err := storage.ParseURI(imgURL)
		if err != nil {
			return
		}
		loadedImage := canvas.NewImageFromURI(uri)
		loadedImage.FillMode = canvas.ImageFillContain
		loadedImage.SetMinSize(imgSize)
		fyne.Do(func() {
			if imageContainer != nil && imageContainer.Visible() {
				imageContainer.Objects = []fyne.CanvasObject{loadedImage}
				imageContainer.Refresh()
			}
		})
	}()

	_, rarityIconURL := data.Rarity(skin)
	var rarityIconWidget fyne.CanvasObject = layout.NewSpacer()
	if rarityIconURL != "" && rarityIconURL != data.GetPlaceholderImageURL() {
		rarityPlaceholder := canvas.NewRectangle(color.Transparent)
		rarityPlaceholder.SetMinSize(rarityIconSize)
		rarityIconContainer := container.NewStack(rarityPlaceholder)
		rarityIconWidget = rarityIconContainer
		go func(url string, cont *fyne.Container, size fyne.Size) { // Load rarity icon
			uri, err := storage.ParseURI(url)
			if err != nil {
				return
			}
			icon := canvas.NewImageFromURI(uri)
			icon.SetMinSize(size)
			icon.FillMode = canvas.ImageFillContain
			fyne.Do(func() {
				if cont != nil && cont.Visible() {
					cont.Objects = []fyne.CanvasObject{icon}
					cont.Refresh()
				}
			})
		}(rarityIconURL, rarityIconContainer, rarityIconSize)
	}

	topIcons := []fyne.CanvasObject{}
	createLazyIcon := func(iconURL string, size fyne.Size) *fyne.Container {
		placeholder := canvas.NewRectangle(color.Transparent)
		placeholder.SetMinSize(size)
		iconContainer := container.NewStack(placeholder)
		go func(url string, cont *fyne.Container) { // Load top icons
			if url == data.GetPlaceholderImageURL() || url == "" {
				fyne.Do(func() {
					if cont != nil && cont.Visible() {
						ph := widget.NewIcon(theme.QuestionIcon())
						cont.Objects = []fyne.CanvasObject{ph}
						cont.Refresh()
					}
				})
				return
			}
			uri, err := storage.ParseURI(url)
			if err != nil {
				fyne.Do(func() {
					if cont != nil && cont.Visible() {
						eh := widget.NewIcon(theme.ErrorIcon())
						cont.Objects = []fyne.CanvasObject{eh}
						cont.Refresh()
					}
				})
				return
			}
			iconImage := canvas.NewImageFromURI(uri)
			iconImage.SetMinSize(size)
			iconImage.FillMode = canvas.ImageFillContain
			fyne.Do(func() {
				if cont != nil && cont.Visible() {
					cont.Objects = []fyne.CanvasObject{iconImage}
					cont.Refresh()
				}
			})
		}(iconURL, iconContainer)
		return iconContainer
	}
	if skin.IsLegacy {
		lic := createLazyIcon(data.LegacyIconURL(), iconSize)
		topIcons = append(topIcons, lic)
	}
	if len(skin.Chromas) > 0 {
		cic := createLazyIcon(data.ChromaIconURL(), iconSize)
		if len(topIcons) > 0 {
			topIcons = append(topIcons, widget.NewSeparator())
		}
		topIcons = append(topIcons, cic)
	}
	topIconsContainer := container.NewHBox(layout.NewSpacer())
	if len(topIcons) > 0 {
		topIconsContainer.Add(container.NewHBox(topIcons...))
	}

	nameLabel := widget.NewLabel(skin.Name)
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	bottomBarContent := container.NewBorder(nil, nil, container.NewPadded(rarityIconWidget), nil, nameLabel)
	bottomBgColor := color.NRGBA{R: 0x10, G: 0x10, B: 0x15, A: 0xD0}
	bottomBgRect := canvas.NewRectangle(bottomBgColor)
	bottomBar := container.NewStack(bottomBgRect, bottomBarContent)
	cardContentLayout := container.NewBorder(container.NewPadded(topIconsContainer), bottomBar, nil, nil, imageContainer)
	card := widget.NewCard("", "", cardContentLayout)                // Use Card for potential styling/background
	tappableItem := NewTappableCard(card, func() { onSelect(skin) }) // Wrap the Card
	tappableItem.SetMinSize(fyne.NewSize(itemWidth, totalHeight))    // Set size on the TappableCard
	return tappableItem
}

// --- Helpers (TappableCard, NewIconButton, NewTabButton) ---
// !!! TappableCard DEFINED HERE !!!
type TappableCard struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	onTapped func()
	minSize  fyne.Size
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
		if c.content != nil {
			return c.content.MinSize()
		} else {
			return c.BaseWidget.MinSize()
		}
	}
	return c.minSize
}
func (c *TappableCard) SetMinSize(size fyne.Size) { c.minSize = size; c.Refresh() }
func (c *TappableCard) Cursor() desktop.Cursor    { return desktop.PointerCursor }

func NewIconButton(iconRes fyne.Resource, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon("", iconRes, onTap)
	return btn
}
func NewTabButton(label string, icon fyne.Resource, tapped func()) fyne.CanvasObject {
	btnIcon := canvas.NewImageFromResource(icon)
	btnIcon.SetMinSize(fyne.NewSize(24, 24))
	btnIcon.FillMode = canvas.ImageFillContain
	btnLabel := widget.NewLabel(label)
	btnLabel.Alignment = fyne.TextAlignCenter
	content := container.NewVBox(container.NewCenter(btnIcon), btnLabel)
	cardContent := container.NewPadded(content)
	return NewTappableCard(cardContent, tapped)
}
func parseURL(rawURL string) *url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Error parsing URL '%s': %v", rawURL, err)
		return &url.URL{}
	}
	return parsed
}

// --- End of utils.go ---
