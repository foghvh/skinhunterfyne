// skinhunter/main.go
package main

import (
	"fmt"
	"image/color"
	"log"
	"sync" // Necesario para actualizar header de forma segura

	"skinhunter/data"
	"skinhunter/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

const appName = "Skin Hunter Fyne"

type skinHunterApp struct {
	fyneApp fyne.App
	window  fyne.Window

	headerContent *fyne.Container // Contenedor para el contenido del header (actualizable)
	headerMutex   sync.Mutex      // Mutex para proteger actualizaciones del header
	footer        fyne.CanvasObject
	centerContent *fyne.Container
	background    fyne.CanvasObject
	navBackButton *widget.Button // Cambiado a *widget.Button para control más fino

	statusLabel      *widget.Label // Lo mantenemos pero no visible en header por defecto
	currentView      string
	selectedChampion data.ChampionSummary
	isDetailView     bool
}

func main() {
	shApp := &skinHunterApp{
		currentView:   "loading",
		centerContent: container.NewMax(), // Contenedor principal para las vistas
	}

	shApp.fyneApp = app.New()
	// shApp.fyneApp.Settings().SetTheme(theme.DarkTheme()) // Asegúrate de usar el tema oscuro si no es el default
	shApp.window = shApp.fyneApp.NewWindow(appName)
	// Ajusta el tamaño inicial si es necesario, pero el layout debería adaptarse
	shApp.window.Resize(fyne.NewSize(900, 700)) // Un poco más pequeño que antes
	shApp.window.CenterOnScreen()
	shApp.window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File", fyne.NewMenuItem("Quit", func() { shApp.fyneApp.Quit() })),
	))

	// Crear componentes
	shApp.headerContent = container.NewHBox()                  // Header vacío inicialmente
	header := shApp.createHeaderContainer(shApp.headerContent) // Contenedor que envuelve el contenido actualizable
	shApp.footer = shApp.createFooter()
	shApp.navBackButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { shApp.goBack() })
	shApp.navBackButton.Hide() // Oculto por defecto

	// Crear status label (no visible en header por defecto)
	shApp.statusLabel = widget.NewLabel("Initializing...") // Para debug o uso futuro

	// Definir fondo
	// bgColor := color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff} // Azul oscuro original
	bgColor := color.NRGBA{R: 0x1e, G: 0x1e, B: 0x24, A: 0xff} // Un gris/azul más oscuro como en la ref
	shApp.background = canvas.NewRectangle(bgColor)

	// --- Layout Principal Simplificado ---
	// Quitamos leftNav y rightNav del layout principal
	// El botón back se añadirá/quitará del headerContent dinámicamente
	// layeredContent contendrá el fondo y el contenido central
	layeredContent := container.NewStack(shApp.background, shApp.centerContent)
	// mainAppLayout usa Border para header, footer y layeredContent en el centro
	mainAppLayout := container.NewBorder(header, shApp.footer, nil, nil, layeredContent)

	shApp.window.SetContent(mainAppLayout)
	shApp.window.SetMaster() // Asegura que los diálogos se centren en esta ventana

	shApp.showLoading() // Muestra indicador de carga inicial

	// Carga de datos asíncrona
	go func() {
		err := data.InitData()
		if err != nil {
			log.Printf("FATAL: Failed to initialize data: %v", err)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Initialization Error",
				Content: "Failed to load required data.",
			})
			// --- CORRECCIÓN AQUÍ ---
			// Llamar directamente a showError (que actualiza la UI)
			shApp.showError(fmt.Sprintf("Failed to load required data:\n%v\n\nPlease check connection or restart.", err))
			// --- FIN CORRECCIÓN ---
			return
		}
		// --- CORRECCIÓN AQUÍ ---
		// Llamar directamente a las funciones que actualizan la UI
		shApp.updateStatus("Ready")
		shApp.showChampionsGrid()
		// --- FIN CORRECCIÓN ---
	}()

	shApp.window.ShowAndRun()
}

// createHeaderContainer: Crea el contenedor ESTRUCTURAL del header (fondo, borde)
// Recibe el contenedor HBox donde irá el contenido dinámico.
func (sh *skinHunterApp) createHeaderContainer(content *fyne.Container) fyne.CanvasObject {
	// headerBackground := color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff} // Azul oscuro original
	headerBackground := color.NRGBA{R: 0x1e, G: 0x1e, B: 0x24, A: 0xff} // Gris/azul oscuro ref
	// borderColor := color.NRGBA{R: 0x58, G: 0x6a, B: 0x9e, A: 0xff} // Borde azulado original
	borderColor := color.NRGBA{R: 0x44, G: 0x44, B: 0x4a, A: 0xff} // Borde grisáceo ref

	bgRect := canvas.NewRectangle(headerBackground)
	bottomBorder := canvas.NewRectangle(borderColor)
	bottomBorder.SetMinSize(fyne.NewSize(1, 1)) // Borde más fino

	// Usamos NewPadded para dar espacio interno al contenido del header
	paddedContent := container.NewPadded(content)
	headerLayout := container.NewBorder(nil, bottomBorder, nil, nil, paddedContent) // Pone el borde debajo del contenido

	return container.NewStack(bgRect, headerLayout) // Fondo detrás del layout con borde
}

// updateHeaderContent: Actualiza los elementos DENTRO del HBox del header
func (sh *skinHunterApp) updateHeaderContent(elements ...fyne.CanvasObject) {
	sh.headerMutex.Lock() // Proteger acceso concurrente
	defer sh.headerMutex.Unlock()

	if sh.headerContent != nil {
		sh.headerContent.Objects = elements // Reemplaza los objetos
		sh.headerContent.Refresh()          // Refresca el HBox interno
	} else {
		log.Println("ERROR: headerContent es nil, no se puede actualizar.")
	}
}

func (sh *skinHunterApp) updateStatus(status string) {
	// Esta función ahora solo actualiza la variable interna
	// Podrías usarla para logs o mostrarla en otro lugar si quieres
	if sh.statusLabel != nil {
		sh.statusLabel.SetText(status)
		log.Printf("Internal Status Updated: %s", status)
	}
}

func (sh *skinHunterApp) createFooter() fyne.CanvasObject {
	tabs := []struct {
		label  string
		icon   fyne.Resource
		view   string
		action func()
	}{
		{"Champions", theme.HomeIcon(), "champions_grid", nil}, // Icono cambiado
		{"Search", theme.SearchIcon(), "", func() { sh.showOmniSearch() }},
		{"Installed", theme.DownloadIcon(), "installed_view", nil},
		{"Profile", theme.AccountIcon(), "profile_view", nil}, // Icono cambiado
	}

	buttons := make([]fyne.CanvasObject, len(tabs)) // Crear slice con tamaño exacto
	for i := range tabs {
		t := tabs[i] // Capturar variable de rango correctamente para el closure
		// Usar NewTabButton que crea un VBox con icono y texto
		button := ui.NewTabButton(t.label, t.icon, func() {
			if t.action != nil {
				t.action()
			} else if t.view != "" {
				log.Printf("Tab '%s' tapped, switching to view '%s'", t.label, t.view)
				sh.switchView(t.view)
			} else {
				log.Printf("Tab '%s' tapped, no action/view defined.", t.label)
			}
		})
		// Añadir botón a la lista de botones
		buttons[i] = container.NewMax(button) // Envuelve en Max para que ocupen espacio equitativamente
	}

	// Usar Grid con 4 columnas para distribución equitativa
	footerContent := container.NewGridWithColumns(4, buttons...)

	// footerBackground := color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff} // Azul oscuro original
	footerBackground := color.NRGBA{R: 0x1e, G: 0x1e, B: 0x24, A: 0xff} // Gris/azul oscuro ref
	// borderColor := color.NRGBA{R: 0x58, G: 0x6a, B: 0x9e, A: 0xff} // Borde azulado original
	borderColor := color.NRGBA{R: 0x44, G: 0x44, B: 0x4a, A: 0xff} // Borde grisáceo ref

	bgRect := canvas.NewRectangle(footerBackground)
	topBorder := canvas.NewRectangle(borderColor)
	topBorder.SetMinSize(fyne.NewSize(1, 1)) // Borde más fino

	// Añadir Padding al contenido del footer para espaciado interno
	paddedFooterContent := container.NewPadded(footerContent)

	footerLayout := container.NewBorder(topBorder, nil, nil, nil, paddedFooterContent) // Pone borde arriba
	return container.NewStack(bgRect, footerLayout)                                    // Fondo detrás
}

// createLeftNav y createRightNav ya no se usan en el layout principal

func (sh *skinHunterApp) switchView(viewName string) {
	if sh.currentView == viewName && viewName != "champions_grid" { // Evita recargar la misma vista (excepto grid)
		log.Printf("Already in view '%s', no change.", viewName)
		return
	}
	log.Printf("Switching view from '%s' to '%s'", sh.currentView, viewName)
	sh.isDetailView = false // Resetear flag por defecto
	var newContent fyne.CanvasObject
	headerElements := []fyne.CanvasObject{layout.NewSpacer()} // Header vacío por defecto

	switch viewName {
	case "champions_grid":
		// El header de la grid de campeones estará vacío (o con un título genérico si quieres)
		sh.navBackButton.Hide() // Asegura que el botón back esté oculto
		newContent = ui.NewChampionGrid(func(champ data.ChampionSummary) { sh.showChampionDetail(champ) })
		// headerElements = []fyne.CanvasObject{widget.NewLabel("Champions"), layout.NewSpacer()} // Ejemplo título

	case "champion_detail":
		sh.isDetailView = true
		sh.navBackButton.Show() // Mostrar botón back para esta vista

		// Título específico para la vista de detalle
		titleLabel := widget.NewLabel("Champion Detail")
		titleLabel.Alignment = fyne.TextAlignCenter // Centrar el título
		titleLabel.TextStyle = fyne.TextStyle{Bold: true}

		// Añadir botón back a la izquierda y título centrado
		headerElements = []fyne.CanvasObject{
			sh.navBackButton,   // Botón a la izquierda
			layout.NewSpacer(), // Empuja título al centro
			titleLabel,
			layout.NewSpacer(), // Empuja cualquier cosa a la derecha
		}

		// Crear contenido de la vista de detalle
		newContent = ui.NewChampionView(
			sh.selectedChampion,
			sh.window, // Pasar la ventana principal
			func(skin data.Skin, allChromas []data.Chroma) {
				ui.ShowSkinDialog(skin, allChromas, sh.window)
			},
		)

	case "installed_view":
		sh.navBackButton.Hide()
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), widget.NewLabel("Installed Skins"), layout.NewSpacer()}
		newContent = container.NewCenter(widget.NewLabel("Installed Skins (Not Implemented)"))
	case "profile_view":
		sh.navBackButton.Hide()
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), widget.NewLabel("Profile"), layout.NewSpacer()}
		newContent = container.NewCenter(widget.NewLabel("User Profile (Not Implemented)"))
	default:
		sh.navBackButton.Hide()
		log.Printf("Warning: Unknown view name '%s'", viewName)
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), widget.NewLabel("Error"), layout.NewSpacer()}
		newContent = container.NewCenter(widget.NewLabel(fmt.Sprintf("Error: View '%s' not found", viewName)))
	}

	// Actualizar contenido central
	if sh.centerContent != nil {
		sh.centerContent.Objects = []fyne.CanvasObject{newContent}
		sh.centerContent.Refresh()
		sh.currentView = viewName
		log.Printf("Center content updated to '%s'", viewName)
	} else {
		log.Println("ERROR: centerContent is nil during switchView!")
	}

	// Actualizar el contenido del header dinámicamente
	sh.updateHeaderContent(headerElements...)
}

// updateLeftNav ya no es necesaria de la misma forma

func (sh *skinHunterApp) showLoading() {
	sh.updateStatus("Loading...")
	if sh.centerContent != nil {
		loadingIndicator := widget.NewProgressBarInfinite()
		loadingLabel := widget.NewLabel("Loading Application Data...")
		loadingBox := container.NewVBox(loadingLabel, loadingIndicator)
		sh.centerContent.Objects = []fyne.CanvasObject{container.NewCenter(loadingBox)}
		sh.centerContent.Refresh()
		sh.currentView = "loading"
		log.Println("UI state: Loading")
	} else {
		log.Println("ERROR: centerContent is nil during showLoading!")
	}
	// Asegurarse que el header esté vacío o con estado de carga al inicio
	sh.updateHeaderContent(layout.NewSpacer(), widget.NewLabel("Loading..."), layout.NewSpacer())
}

func (sh *skinHunterApp) showError(errMsg string) {
	sh.updateStatus("Error") // Actualiza estado interno
	if sh.centerContent != nil {
		errorIcon := widget.NewIcon(theme.ErrorIcon())
		errorLabel := widget.NewLabelWithStyle(errMsg, fyne.TextAlignCenter, fyne.TextStyle{})
		errorLabel.Wrapping = fyne.TextWrapWord
		errorBox := container.NewVBox(container.NewCenter(errorIcon), widget.NewSeparator(), errorLabel)
		sh.centerContent.Objects = []fyne.CanvasObject{container.NewPadded(container.NewCenter(errorBox))}
		sh.centerContent.Refresh()
		sh.currentView = "error"
		log.Println("UI state: Error - ", errMsg)
	} else {
		log.Println("ERROR: centerContent is nil during showError!")
	}
	// Actualizar header para mostrar error (opcional)
	errorTitle := widget.NewLabelWithStyle("Error", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	sh.updateHeaderContent(layout.NewSpacer(), errorTitle, layout.NewSpacer())
}

// showChampionsGrid: Llama a switchView para mostrar la cuadrícula
func (sh *skinHunterApp) showChampionsGrid() {
	sh.switchView("champions_grid")
}

// showChampionDetail: Guarda el campeón y llama a switchView
func (sh *skinHunterApp) showChampionDetail(champ data.ChampionSummary) {
	log.Printf("Navigating to details for champion: %s (ID: %d)", champ.Name, champ.ID)
	sh.selectedChampion = champ
	sh.switchView("champion_detail") // Cambia a la vista de detalle
}

// goBack: Navega a la vista anterior (probablemente siempre a la grid de campeones)
func (sh *skinHunterApp) goBack() {
	log.Printf("Go back requested from view: %s", sh.currentView)
	// Si estamos en detalle o cualquier otra vista que no sea la grid, volvemos a la grid
	if sh.currentView != "champions_grid" {
		sh.switchView("champions_grid")
	} else {
		// Opcional: Salir de la app o no hacer nada si ya estamos en la vista principal
		log.Println("Already at base view (Champions Grid), cannot go back further.")
		// sh.fyneApp.Quit() // Descomentar si quieres que "back" aquí cierre la app
	}
}

// showOmniSearch: Muestra un diálogo placeholder
func (sh *skinHunterApp) showOmniSearch() {
	searchInput := widget.NewEntry()
	searchInput.SetPlaceHolder("Search Champions, Skins...")
	// Crear contenido del diálogo
	dialogContent := container.NewVBox(
		widget.NewLabel("Search feature is not yet fully implemented."),
		searchInput,
	)
	dialog.ShowCustomConfirm("OmniSearch", "Search", "Cancel", dialogContent, func(ok bool) {
		if ok {
			log.Printf("Search submitted (WIP): %s", searchInput.Text)
			// Aquí iría la lógica de búsqueda real
		}
	}, sh.window)
}
