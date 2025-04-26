// skinhunter/ui/centered_gridwrap_layout.go
package ui

import (
	"log"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// CenteredGridWrapLayout organiza los objetos en una cuadrícula responsive
// y centra horizontalmente el bloque de la cuadrícula.
type CenteredGridWrapLayout struct {
	CellSize fyne.Size
}

// NewCenteredGridWrapLayout crea una nueva instancia del layout personalizado.
func NewCenteredGridWrapLayout(cellSize fyne.Size) fyne.Layout {
	if cellSize.Width <= 0 || cellSize.Height <= 0 {
		log.Printf("WARN: NewCenteredGridWrapLayout creado con CellSize inválido: %v. Usando fallback.", cellSize)
		cellSize = fyne.NewSize(100, 100)
	}
	return &CenteredGridWrapLayout{CellSize: cellSize}
}

// Layout posiciona los objetos. (Lógica de centrado correcta)
func (l *CenteredGridWrapLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	if l.CellSize.Width <= 0 || l.CellSize.Height <= 0 || len(objects) == 0 {
		for _, o := range objects {
			o.Hide()
		}
		return
	}
	for _, o := range objects {
		o.Show()
	}

	colsF := float64(containerSize.Width) / float64(l.CellSize.Width)
	cols := int(math.Max(1.0, math.Floor(colsF)))

	actualGridWidth := float32(cols) * l.CellSize.Width
	offsetX := (containerSize.Width - actualGridWidth) / 2.0
	if offsetX < 0 {
		offsetX = 0
	}

	row, col := 0, 0
	for _, obj := range objects {
		x := offsetX + float32(col)*l.CellSize.Width
		y := float32(row) * l.CellSize.Height
		obj.Resize(l.CellSize)
		obj.Move(fyne.NewPos(x, y))
		col++
		if col >= cols {
			col = 0
			row++
		}
	}
}

// MinSize calcula el tamaño mínimo necesario para este layout.
// *** USANDO ALTURA TOTAL CALCULADA (para que Scroll sepa el rango) ***
func (l *CenteredGridWrapLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if l.CellSize.Width <= 0 || l.CellSize.Height <= 0 || len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}

	// Ancho mínimo es el de una celda.
	minWidth := l.CellSize.Width

	// Altura mínima es la necesaria si todos los objetos estuvieran en una columna.
	// Esto le informa al Scroll cuál es la altura total del contenido potencial.
	// CORRECTION: Calculate rows based on a reasonable minimum width (e.g., 1 column)
	// This estimate helps the scroll container.
	rows := len(objects) // Estimate rows for single column
	minHeight := float32(rows) * l.CellSize.Height

	return fyne.NewSize(minWidth, minHeight)
}

// --- Helper NewCenteredGridWrap ---
// (Sin cambios)
func NewCenteredGridWrap(cellSize fyne.Size, objects ...fyne.CanvasObject) *fyne.Container {
	// Usar NewWithoutLayout y asignar nuestro layout personalizado
	container := container.NewWithoutLayout(objects...)
	container.Layout = NewCenteredGridWrapLayout(cellSize)
	return container
}

// --- End of centered_gridwrap_layout.go ---
