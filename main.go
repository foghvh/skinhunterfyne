// skinhunter/main.go
package main

import (
	"fmt"
	"image/color"
	"log"
	"strings"

	"skinhunter/data"
	"skinhunter/ui" // Import local ui package

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

	header          fyne.CanvasObject
	footer          fyne.CanvasObject
	leftNav         *fyne.Container
	rightNav        *fyne.Container
	centerContent   *fyne.Container
	background      fyne.CanvasObject
	navBackButton   fyne.Widget
	navToggleButton fyne.Widget

	statusLabel      *widget.Label
	currentView      string
	selectedChampion data.ChampionSummary
	isDetailView     bool
}

func main() {
	shApp := &skinHunterApp{
		currentView:   "loading",
		centerContent: container.NewMax(),
	}

	shApp.fyneApp = app.New()
	shApp.window = shApp.fyneApp.NewWindow(appName)
	shApp.window.Resize(fyne.NewSize(1150, 800))
	shApp.window.CenterOnScreen()
	shApp.window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File", fyne.NewMenuItem("Quit", func() { shApp.fyneApp.Quit() })),
	))

	shApp.header = shApp.createHeader()
	shApp.footer = shApp.createFooter()
	shApp.leftNav = shApp.createLeftNav()
	shApp.rightNav = shApp.createRightNav()

	bgColor := color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff}
	shApp.background = canvas.NewRectangle(bgColor)

	contentWithNav := container.NewBorder(nil, nil, shApp.leftNav, shApp.rightNav, shApp.centerContent)
	layeredContent := container.NewStack(shApp.background, contentWithNav)
	mainAppLayout := container.NewBorder(shApp.header, shApp.footer, nil, nil, layeredContent)

	shApp.window.SetContent(mainAppLayout)
	shApp.window.SetMaster()

	shApp.showLoading()
	go func() {
		err := data.InitData()
		if err != nil {
			log.Printf("FATAL: Failed to initialize data: %v", err)
			shApp.showError(fmt.Sprintf("Failed to load required data:\n%v\n\nPlease check connection or restart.", err))
			return
		}
		shApp.updateStatus("Ready")
		shApp.showChampionsGrid()
	}()

	shApp.window.ShowAndRun()
}

func (sh *skinHunterApp) createHeader() fyne.CanvasObject {
	logoMinSize := fyne.NewSize(180, 30) // Define size once
	logoRes, err := fyne.LoadResourceFromURLString("https://i.imgur.com/m40l0qA.png")
	var logo fyne.CanvasObject
	if err != nil {
		log.Println("WARN: Failed to load logo:", err)
		logo = widget.NewLabel("Skin Hunter")
	} else {
		logoImg := canvas.NewImageFromResource(logoRes)
		logoImg.FillMode = canvas.ImageFillContain
		logoImg.SetMinSize(logoMinSize)
		logo = logoImg
	}

	sh.statusLabel = widget.NewLabel("Loading...")
	sh.statusLabel.Alignment = fyne.TextAlignCenter
	statusBg := canvas.NewRectangle(color.NRGBA{R: 0x33, G: 0x33, B: 0x36, A: 0x90})
	paddedStatusLabel := container.NewPadded(sh.statusLabel)
	statusBox := container.NewStack(statusBg, paddedStatusLabel)

	headerContent := container.NewHBox(
		container.NewPadded(logo), // Padded logo on the left
		layout.NewSpacer(),        // Pushes status to center/right
		statusBox,                 // Status box takes its preferred size
		layout.NewSpacer(),        // Pushes any potential right elements away
	)

	headerBackground := canvas.NewRectangle(color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff})
	borderColor := color.NRGBA{R: 0x58, G: 0x6a, B: 0x9e, A: 0xff}
	bottomBorder := canvas.NewRectangle(borderColor)
	bottomBorder.SetMinSize(fyne.NewSize(1, 2))

	headerLayout := container.NewBorder(nil, bottomBorder, nil, nil, headerContent)
	return container.NewStack(headerBackground, headerLayout)
}

func (sh *skinHunterApp) updateStatus(status string) {
	if sh.statusLabel != nil {
		sh.statusLabel.SetText(status)
	}
}

func (sh *skinHunterApp) createFooter() fyne.CanvasObject {
	tabs := []struct {
		label  string
		icon   fyne.Resource
		view   string
		action func()
	}{
		{"Champions", theme.ListIcon(), "champions_grid", nil},
		{"Search", theme.SearchIcon(), "", func() { sh.showOmniSearch() }},
		{"Installed", theme.DownloadIcon(), "installed_view", nil},
		{"Profile", theme.HomeIcon(), "profile_view", nil},
	}
	buttons := []fyne.CanvasObject{}
	for i := range tabs {
		t := tabs[i]
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
		buttons = append(buttons, button)
	}
	footerContent := container.NewHBox(
		layout.NewSpacer(), buttons[0], layout.NewSpacer(), buttons[1], layout.NewSpacer(),
		buttons[2], layout.NewSpacer(), buttons[3], layout.NewSpacer(),
	)
	footerBackground := canvas.NewRectangle(color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff})
	borderColor := color.NRGBA{R: 0x58, G: 0x6a, B: 0x9e, A: 0xff}
	topBorder := canvas.NewRectangle(borderColor)
	topBorder.SetMinSize(fyne.NewSize(1, 2))
	footerLayout := container.NewBorder(topBorder, nil, nil, nil, footerContent)
	return container.NewStack(footerBackground, footerLayout)
}

func (sh *skinHunterApp) createLeftNav() *fyne.Container {
	sh.navBackButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { sh.goBack() })
	sh.navBackButton.Hide()
	sh.navToggleButton = widget.NewButtonWithIcon("", theme.ListIcon(), func() { log.Println("INFO: Champion/Skinline toggle placeholder clicked.") })
	sh.navToggleButton.Hide()
	navBox := container.NewVBox(layout.NewSpacer(), sh.navBackButton, sh.navToggleButton, layout.NewSpacer())
	return container.NewPadded(navBox)
}

func (sh *skinHunterApp) createRightNav() *fyne.Container {
	startButton := widget.NewButton("Start", func() {
		log.Println("INFO: 'Start' button placeholder clicked.")
		dialog.ShowInformation("Start Game", "Placeholder: Game starting...", sh.window)
	})
	navBox := container.NewVBox(layout.NewSpacer(), startButton, layout.NewSpacer())
	return container.NewPadded(navBox)
}

func (sh *skinHunterApp) switchView(viewName string) {
	if sh.currentView == viewName && viewName != "champions_grid" {
		log.Printf("Already in view '%s', no change.", viewName)
		return
	}
	log.Printf("Switching view from '%s' to '%s'", sh.currentView, viewName)
	sh.updateStatus("Loading " + strings.ReplaceAll(viewName, "_", " ") + "...")
	var newContent fyne.CanvasObject
	wasDetailView := sh.isDetailView
	sh.isDetailView = false
	switch viewName {
	case "champions_grid":
		newContent = ui.NewChampionGrid(func(champ data.ChampionSummary) { sh.showChampionDetail(champ) })
	case "champion_detail":
		sh.isDetailView = true
		newContent = ui.NewChampionView(sh.selectedChampion, func() { sh.goBack() },
			func(skin data.Skin, allChromas []data.Chroma) { ui.ShowSkinDialog(skin, allChromas, sh.window) },
		)
	case "installed_view":
		newContent = container.NewCenter(widget.NewLabel("Installed Skins (Not Implemented)"))
	case "profile_view":
		newContent = container.NewCenter(widget.NewLabel("User Profile (Not Implemented)"))
	default:
		log.Printf("Warning: Unknown view name '%s'", viewName)
		newContent = container.NewCenter(widget.NewLabel(fmt.Sprintf("Error: View '%s' not found", viewName)))
	}
	if sh.centerContent != nil {
		sh.centerContent.Objects = []fyne.CanvasObject{newContent}
		sh.centerContent.Refresh()
		sh.currentView = viewName
		log.Printf("Center content updated to '%s'", viewName)
	} else {
		log.Println("ERROR: centerContent is nil during switchView!")
	}
	sh.updateLeftNav(sh.isDetailView, wasDetailView)
	sh.updateStatus(strings.Title(strings.ReplaceAll(viewName, "_", " ")))
}

func (sh *skinHunterApp) updateLeftNav(isDetail, wasDetail bool) {
	if sh.navBackButton == nil || sh.navToggleButton == nil {
		log.Println("WARN: Nav buttons not initialized.")
		return
	}
	if isDetail {
		sh.navBackButton.Show()
		sh.navToggleButton.Hide()
		log.Println("Nav updated: Showing Back, Hiding Toggle")
	} else {
		sh.navBackButton.Hide()
		if sh.currentView == "champions_grid" {
			sh.navToggleButton.Hide() // Keep hidden until implemented
			log.Println("Nav updated: Hiding Back, Hiding Toggle (for champions_grid)")
		} else {
			sh.navToggleButton.Hide()
			log.Println("Nav updated: Hiding Back, Hiding Toggle")
		}
	}
}

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
	if sh.navBackButton != nil && sh.navToggleButton != nil {
		sh.updateLeftNav(false, sh.isDetailView)
	}
}

func (sh *skinHunterApp) showError(errMsg string) {
	sh.updateStatus("Error")
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
	if sh.navBackButton != nil && sh.navToggleButton != nil {
		sh.updateLeftNav(false, sh.isDetailView)
	}
}

func (sh *skinHunterApp) showChampionsGrid() { sh.switchView("champions_grid") }
func (sh *skinHunterApp) showChampionDetail(champ data.ChampionSummary) {
	log.Printf("Navigating to details for champion: %s (ID: %d)", champ.Name, champ.ID)
	sh.selectedChampion = champ
	sh.switchView("champion_detail")
}
func (sh *skinHunterApp) goBack() {
	log.Printf("Go back requested from view: %s", sh.currentView)
	if sh.isDetailView || sh.currentView != "champions_grid" {
		sh.switchView("champions_grid")
	} else {
		log.Println("Already at base view (Champions Grid), cannot go back further.")
	}
}
func (sh *skinHunterApp) showOmniSearch() {
	searchInput := widget.NewEntry()
	searchInput.SetPlaceHolder("Search Champions, Skins...")
	dialog.ShowCustomConfirm("OmniSearch (WIP)", "Search", "Cancel", container.NewVBox(
		widget.NewLabel("Search feature is not yet implemented."), searchInput,
	), func(ok bool) {
		if ok {
			log.Printf("Search submitted (WIP): %s", searchInput.Text)
		}
	}, sh.window)
}
