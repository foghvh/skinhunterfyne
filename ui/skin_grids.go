// skinhunter/ui/skins_grid.go
package ui

import (
	"log"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	// No se necesita "layout"
)

// NewSkinsGrid crea una cuadrícula de skins responsive y CENTRADA usando layout personalizado.
func NewSkinsGrid(skins []data.Skin, onSkinSelect func(skin data.Skin)) fyne.CanvasObject {

	// ... (código para filtrar nonBaseSkins y manejar casos vacíos) ...
	if skins == nil {
		log.Println("WARN: ...")
		return container.NewCenter(widget.NewLabel("..."))
	}
	nonBaseSkins := make([]data.Skin, 0, len(skins))
	for _, s := range skins {
		if !s.IsBase {
			nonBaseSkins = append(nonBaseSkins, s)
		}
	}
	if len(nonBaseSkins) == 0 {
		log.Println("INFO: ...") /* devuelve mensaje apropiado */
		return container.NewCenter(widget.NewLabel("..."))
	}

	// Definir tamaño de celda
	cellWidth := float32(210)
	cellHeight := float32(200)
	cellSize := fyne.NewSize(cellWidth, cellHeight)

	// Crear los widgets
	skinItems := make([]fyne.CanvasObject, 0, len(nonBaseSkins))
	for _, skin := range nonBaseSkins {
		item := SkinItem(skin, onSkinSelect) // Usa SkinItem con estilo bueno
		if item != nil {
			skinItems = append(skinItems, item)
		}
	}
	log.Printf("Created centered responsive skins grid container for %d displayable skins.", len(skinItems))

	// *** Usar el CONTENEDOR con LAYOUT PERSONALIZADO ***
	grid := NewCenteredGridWrap(cellSize, skinItems...) // Usa NUESTRA función helper

	// Envolver directamente en Scroll
	scrollableGrid := container.NewScroll(grid)

	return scrollableGrid
}
