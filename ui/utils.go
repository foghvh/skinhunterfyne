// skinhunter/ui/utils.go
package ui

import (
	"image/color"
	"log"
	"strings"

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

// ChampionGridItem: Usa la lógica del ejemplo simple.
const cdragonBase = "https://raw.communitydragon.org/latest"

func ChampionGridItem(champ data.ChampionSummary, onSelect func(champ data.ChampionSummary)) fyne.CanvasObject {
	// log.Println("WARN: ChampionGridItem called directly, NewChampionGrid uses GridWrap now.") // Quitado warning redundante
	imgMinSize := fyne.NewSize(80, 80)
	var imageWidget fyne.CanvasObject

	rawPath := champ.SquarePortraitPath
	if rawPath == "" {
		placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
		imageWidget = container.NewStack(container.NewCenter(placeholderIcon))
	} else {
		fixedPath := cdragonBase + strings.Replace(rawPath, "/lol-game-data/assets", "/plugins/rcp-be-lol-game-data/global/default", 1)
		uri, err := storage.ParseURI(fixedPath)
		if err != nil {
			placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
			imageWidget = container.NewStack(container.NewCenter(placeholderIcon))
		} else {
			img := canvas.NewImageFromURI(uri)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(imgMinSize)
			imageWidget = img
		}
	}
	nameLabel := widget.NewLabel(champ.Name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	itemContent := container.NewVBox(imageWidget, container.NewPadded(nameLabel))
	card := widget.NewCard("", "", itemContent)
	tapButton := widget.NewButton("", func() { onSelect(champ) })
	stack := container.NewStack(card, tapButton)
	return stack
}

// SkinItem: Usa data.Asset + NewImageFromURI.
func SkinItem(skin data.Skin, onSelect func(skin data.Skin)) fyne.CanvasObject {
	if skin.IsBase {
		return nil
	}

	itemWidth := float32(190)
	itemHeight := float32(240)
	imgHeight := itemHeight * 0.75
	imgSize := fyne.NewSize(itemWidth, imgHeight)
	// itemTotalSize := fyne.NewSize(itemWidth, itemHeight) // Comentado

	var finalImgWidget fyne.CanvasObject
	var rarityIconWidget fyne.CanvasObject = layout.NewSpacer()
	var topIcons = []fyne.CanvasObject{}
	iconSize := fyne.NewSize(20, 20)

	imgURL := data.GetSkinTileURL(skin)
	uri, err := storage.ParseURI(imgURL)
	if err != nil {
		log.Printf("Error parsing skin tile URI [%s]: %v", imgURL, err)
		placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
		placeholderRect.SetMinSize(imgSize)
		finalImgWidget = placeholderRect
	} else {
		imgWidget := canvas.NewImageFromURI(uri)
		imgWidget.FillMode = canvas.ImageFillStretch
		imgWidget.SetMinSize(imgSize)
		finalImgWidget = imgWidget
	}

	_, rarityIconURL := data.Rarity(skin)
	if rarityIconURL != "" {
		rarityUri, err := storage.ParseURI(rarityIconURL)
		if err == nil {
			rarityIcon := canvas.NewImageFromURI(rarityUri)
			rarityIcon.SetMinSize(fyne.NewSize(16, 16))
			rarityIcon.FillMode = canvas.ImageFillContain
			rarityIconWidget = rarityIcon
		} else {
			log.Printf("Error parsing rarity icon URI %s: %v", rarityIconURL, err)
		}
	}

	if skin.IsLegacy {
		legacyURL := data.LegacyIconURL()
		legacyUri, err := storage.ParseURI(legacyURL)
		if err == nil {
			legacyIcon := canvas.NewImageFromURI(legacyUri)
			legacyIcon.SetMinSize(iconSize)
			legacyIcon.FillMode = canvas.ImageFillContain
			topIcons = append(topIcons, legacyIcon)
		} else {
			log.Printf("Error parsing legacy icon URI: %v", err)
		}
	}
	if len(skin.Chromas) > 0 {
		chromaURL := data.ChromaIconURL()
		chromaUri, err := storage.ParseURI(chromaURL)
		if err == nil {
			chromaIcon := canvas.NewImageFromURI(chromaUri)
			chromaIcon.SetMinSize(iconSize)
			chromaIcon.FillMode = canvas.ImageFillContain
			if len(topIcons) > 0 {
				spacerRect := canvas.NewRectangle(color.Transparent)
				spacerRect.SetMinSize(fyne.NewSize(5, 1))
				topIcons = append(topIcons, spacerRect)
			}
			topIcons = append(topIcons, chromaIcon)
		} else {
			log.Printf("Error parsing chroma icon URI: %v", err)
		}
	}

	nameLabel := widget.NewLabel(skin.Name)
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	bottomContent := container.NewBorder(nil, nil, rarityIconWidget, nil, container.NewPadded(nameLabel))
	bgRect := canvas.NewRectangle(color.NRGBA{R: 0x0A, G: 0x0E, B: 0x19, A: 0xD0})
	bottomBar := container.NewStack(bgRect, bottomContent)
	topIconsHBox := container.NewHBox(layout.NewSpacer(), container.NewHBox(topIcons...))
	contentLayout := container.NewBorder(container.NewPadded(topIconsHBox), bottomBar, nil, nil, finalImgWidget)
	card := widget.NewCard("", "", contentLayout)
	tapButton := widget.NewButton("", func() { onSelect(skin) })
	finalItem := container.NewStack(card, tapButton)

	// *** FIX: ELIMINAR SetMinSize en finalItem ***
	// finalItem.SetMinSize(itemTotalSize)

	return finalItem
}

// --- Helpers sin cambios ---
type TappableCard struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	onTapped func()
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
func (c *TappableCard) Cursor() desktop.Cursor { return desktop.PointerCursor }
func NewIconButton(iconRes fyne.Resource, onTap func()) *widget.Button {
	return widget.NewButtonWithIcon("", iconRes, onTap)
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
