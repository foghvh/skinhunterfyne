// skinhunter/ui/utils.go
package ui

import (
	"image/color" // Necesario para el color de fondo de la barra
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout" // Necesario para layout.NewSpacer
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"skinhunter/data"
)

// SkinItem: Versión revisada para parecerse a la referencia (imagen Irelia)
func SkinItem(skin data.Skin, onSelect func(skin data.Skin)) fyne.CanvasObject {
	if skin.IsBase {
		return nil // No mostrar skins base
	}

	// --- Sizing (Ajustar para un look tipo tarjeta/tile ~4:3 o similar) ---
	itemWidth := float32(350)                     // Ancho similar al anterior
	imgHeight := float32(350)                     // Altura mayor para mejor proporción
	imgSize := fyne.NewSize(itemWidth, imgHeight) // Tamaño del área de la imagen
	iconSize := fyne.NewSize(18, 18)              // Iconos pequeños en overlay

	var imageWidget fyne.CanvasObject
	var rarityIconWidget fyne.CanvasObject = layout.NewSpacer() // Placeholder si no hay icono
	topIcons := []fyne.CanvasObject{}                           // Slice para iconos superiores (legado, chroma)

	// --- Carga de Imagen (Tile o Fallback) ---
	imgURL := data.GetSkinTileURL(skin)
	uri, err := storage.ParseURI(imgURL)

	if err != nil || imgURL == data.GetPlaceholderImageURL() {
		// Placeholder si hay error o URL vacía/placeholder
		log.Printf("WARN: Placeholder/Error for skin tile %s (ID: %d) URL [%s]: %v", skin.Name, skin.ID, imgURL, err)
		placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
		placeholderRect.SetMinSize(imgSize)
		placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
		imageWidget = container.NewStack(placeholderRect, container.NewCenter(placeholderIcon))
		imageWidget.Refresh()
	} else {
		img := canvas.NewImageFromURI(uri)
		// *** CLAVE: Usar Contain para mantener el aspect ratio ***
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(imgSize)
		imageWidget = img
	}

	// --- Icono de Rareza ---
	_, rarityIconURL := data.Rarity(skin) // Obtiene URL del icono de rareza
	if rarityIconURL != "" && rarityIconURL != data.GetPlaceholderImageURL() {
		rarityUri, err := storage.ParseURI(rarityIconURL)
		if err == nil {
			rarityIcon := canvas.NewImageFromURI(rarityUri)
			// Ajustar tamaño del icono de rareza si es necesario
			rarityIcon.SetMinSize(fyne.NewSize(16, 16))
			rarityIcon.FillMode = canvas.ImageFillContain
			rarityIconWidget = rarityIcon // Asignar al widget
		} else {
			log.Printf("WARN: Error parsing rarity icon URI %s: %v", rarityIconURL, err)
		}
	}

	// --- Iconos Superiores (Legado y Chroma) ---
	if skin.IsLegacy {
		legacyURL := data.LegacyIconURL()
		legacyUri, err := storage.ParseURI(legacyURL)
		if err == nil {
			legacyIcon := canvas.NewImageFromURI(legacyUri)
			legacyIcon.SetMinSize(iconSize)
			legacyIcon.FillMode = canvas.ImageFillContain
			topIcons = append(topIcons, legacyIcon)
		} else {
			log.Printf("WARN: Error parsing legacy icon URI: %v", err)
		}
	}
	if len(skin.Chromas) > 0 {
		chromaURL := data.ChromaIconURL()
		chromaUri, err := storage.ParseURI(chromaURL)
		if err == nil {
			chromaIcon := canvas.NewImageFromURI(chromaUri)
			chromaIcon.SetMinSize(iconSize)
			chromaIcon.FillMode = canvas.ImageFillContain
			// Añadir espacio si ya hay otro icono
			if len(topIcons) > 0 {
				topIcons = append(topIcons, layout.NewSpacer()) // O un rect transparente pequeño
			}
			topIcons = append(topIcons, chromaIcon)
		} else {
			log.Printf("WARN: Error parsing chroma icon URI: %v", err)
		}
	}
	// Contenedor para los iconos superiores, alineados a la derecha
	topIconsContainer := container.NewHBox(layout.NewSpacer()) // Empuja a la derecha
	if len(topIcons) > 0 {
		topIconsContainer.Add(container.NewHBox(topIcons...)) // Añade los iconos existentes
	}

	// --- Barra Inferior (Nombre y Rareza) ---
	nameLabel := widget.NewLabel(skin.Name)
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	// Layout para nombre y rareza: Rareza a la izquierda, Nombre en el centro (con padding)
	bottomBarContent := container.NewBorder(
		nil, nil, // Top, Bottom
		rarityIconWidget,               // Left (Icono de rareza)
		nil,                            // Right
		container.NewPadded(nameLabel), // Center (Nombre con padding)
	)
	// Fondo semi-transparente para la barra
	// Usar un color del tema o definir uno custom
	// bottomBgColor := theme.HoverBackgroundColor() // Ejemplo usando color del tema
	bottomBgColor := color.NRGBA{R: 0x10, G: 0x10, B: 0x15, A: 0xD0} // Negro/gris oscuro translúcido
	bottomBgRect := canvas.NewRectangle(bottomBgColor)
	// Stack para poner el contenido sobre el fondo
	bottomBar := container.NewStack(bottomBgRect, bottomBarContent)

	// --- Ensamblaje dentro del Card con Border Layout ---
	cardContentLayout := container.NewBorder(
		container.NewPadded(topIconsContainer), // Top: Iconos superiores con padding
		bottomBar,                              // Bottom: Barra con nombre y rareza
		nil,                                    // Left: nil
		nil,                                    // Right: nil
		imageWidget,                            // Center: La imagen principal
	)

	// --- Crear el Card y el Tappable ---
	// Usar Card para el fondo y estructura visual
	card := widget.NewCard("", "", cardContentLayout) // Sin título/subtítulo en el Card

	// Envolver en TappableCard para la interacción
	tappableItem := NewTappableCard(card, func() {
		onSelect(skin)
	})

	return tappableItem
}

// --- Helpers (Asegúrate que NewTappableCard y NewTabButton estén aquí) ---

// TappableCard makes any canvas object tappable.
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
	// Devolver un renderer simple que solo dibuja el contenido
	// Esto evita el padding/fondo por defecto del renderer de Card si envolvieses un Card
	// ¡PERO! aquí el contenido YA es un Card, así que esto está bien.
	return widget.NewSimpleRenderer(c.content)
}
func (c *TappableCard) Tapped(_ *fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped()
	}
}
func (c *TappableCard) Cursor() desktop.Cursor { return desktop.PointerCursor }

// NewIconButton crea un botón con solo un icono.
func NewIconButton(iconRes fyne.Resource, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon("", iconRes, onTap)
	return btn
}

// NewTabButton crea un vertical button con icono y label.
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
