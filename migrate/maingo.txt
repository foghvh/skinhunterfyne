// skinhunter/main.go
package main

import (
	"fmt"
	"log"
	"sync"

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
	// _ "net/http/pprof"
	// "net/http"
)

const appName = "Skin Hunter Fyne"

type skinHunterApp struct {
	fyneApp fyne.App
	window  fyne.Window

	headerContent *fyne.Container
	headerMutex   sync.Mutex
	footer        fyne.CanvasObject
	centerContent *fyne.Container
	background    fyne.CanvasObject
	navBackButton *widget.Button

	statusLabel      *widget.Label
	currentView      string
	selectedChampion data.ChampionSummary
	isDetailView     bool

	championsData      []data.ChampionSummary
	championsDataErr   error
	championsGridView  fyne.CanvasObject
	championDetailView *ui.ChampionView // Changed type to pointer
	installedView      fyne.CanvasObject
	profileView        fyne.CanvasObject
}

func main() {
	// go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	shApp := &skinHunterApp{
		currentView:   "loading",
		centerContent: container.NewMax(),
	}

	shApp.fyneApp = app.New()
	shApp.window = shApp.fyneApp.NewWindow(appName)
	shApp.window.Resize(fyne.NewSize(950, 720))
	shApp.window.CenterOnScreen()
	shApp.window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File", fyne.NewMenuItem("Quit", func() { shApp.fyneApp.Quit() })),
	))

	shApp.headerContent = container.NewHBox()
	header := shApp.createHeaderContainer(shApp.headerContent)
	shApp.footer = shApp.createFooter()
	shApp.navBackButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { shApp.goBack() })
	shApp.navBackButton.Hide()
	shApp.statusLabel = widget.NewLabel("Initializing...")
	bgColor := theme.BackgroundColor()
	shApp.background = canvas.NewRectangle(bgColor)
	layeredContent := container.NewStack(shApp.background, shApp.centerContent)
	mainAppLayout := container.NewBorder(header, shApp.footer, nil, nil, layeredContent)
	shApp.window.SetContent(mainAppLayout)
	shApp.window.SetMaster()
	shApp.showLoading()

	go func() {
		champions, err := data.FetchAllChampions()
		shApp.championsData = champions
		shApp.championsDataErr = err

		if err != nil {
			log.Printf("ERROR: Failed to get champion data for initial view: %v", err)
			fyne.Do(func() {
				if shApp.currentView == "loading" {
					errMsg := fmt.Sprintf("Failed to load champion list:\n%v\n\nPlease check connection or restart.", err)
					shApp.showError(errMsg)
				}
			})
			return
		}

		// Pre-create reusable views
		shApp.championDetailView = ui.NewChampionView(
			shApp.window,
			func(skin data.Skin, allChromas []data.Chroma) {
				ui.ShowSkinDialog(skin, allChromas, shApp.window)
			},
		)
		shApp.installedView = container.NewCenter(widget.NewLabel("Installed Skins View (Not Implemented)"))
		shApp.profileView = container.NewCenter(widget.NewLabel("User Profile View (Not Implemented)"))

		fyne.Do(func() {
			shApp.updateStatus("Ready")
			shApp.switchView("champions_grid")
		})
	}()

	shApp.window.ShowAndRun()
}

// Method implementations (createHeaderContainer, updateHeaderContent, etc.)
// use 'sh' as the receiver. switchView updated to use cache.

func (sh *skinHunterApp) switchView(viewName string) {
	if sh.currentView == viewName {
		log.Printf("Already in view '%s', no switch.", viewName)
		return
	}
	log.Printf("Switching view from '%s' to '%s'", sh.currentView, viewName)
	sh.isDetailView = false
	var newContent fyne.CanvasObject
	headerElements := []fyne.CanvasObject{layout.NewSpacer()}

	switch viewName {
	case "champions_grid":
		sh.navBackButton.Hide()
		titleLabel := widget.NewLabelWithStyle("Champions", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		if sh.championsGridView != nil {
			newContent = sh.championsGridView
		} else if sh.championsDataErr != nil {
			el := widget.NewLabel(fmt.Sprintf("Error loading champion list:\n%v", sh.championsDataErr))
			el.Wrapping = fyne.TextWrapWord
			el.Alignment = fyne.TextAlignCenter
			newContent = container.NewCenter(el)
		} else if sh.championsData == nil {
			newContent = container.NewCenter(widget.NewLabel("Champion data not available."))
		} else {
			newContent = ui.NewChampionGrid(sh.championsData, func(champ data.ChampionSummary) { sh.showChampionDetail(champ) })
			sh.championsGridView = newContent
		}

	case "champion_detail":
		sh.isDetailView = true
		sh.navBackButton.Show()
		titleText := "Champion Detail"
		if sh.selectedChampion.Name != "" {
			titleText = sh.selectedChampion.Name
		}
		titleLabel := widget.NewLabelWithStyle(titleText, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{sh.navBackButton, layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		if sh.championDetailView == nil { // Defensive creation
			sh.championDetailView = ui.NewChampionView(sh.window, func(s data.Skin, ac []data.Chroma) { ui.ShowSkinDialog(s, ac, sh.window) })
		}
		sh.championDetailView.UpdateContent(sh.selectedChampion)
		newContent = sh.championDetailView

	case "installed_view":
		sh.navBackButton.Hide()
		titleLabel := widget.NewLabelWithStyle("Installed", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		newContent = sh.installedView
	case "profile_view":
		sh.navBackButton.Hide()
		titleLabel := widget.NewLabelWithStyle("Profile", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		newContent = sh.profileView
	default:
		sh.navBackButton.Hide()
		log.Printf("Warning: Unknown view name requested: '%s'", viewName)
		titleLabel := widget.NewLabelWithStyle("Error", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		newContent = container.NewCenter(widget.NewLabel(fmt.Sprintf("Error: View '%s' not found.", viewName)))
	}

	if sh.centerContent == nil {
		log.Println("CRITICAL ERROR: centerContent container is nil during switchView!")
		return
	}
	if len(sh.centerContent.Objects) == 0 || sh.centerContent.Objects[0] != newContent {
		sh.centerContent.Objects = []fyne.CanvasObject{newContent}
		sh.centerContent.Refresh()
		log.Printf("Center content updated to show '%s'", viewName)
	} else {
		log.Printf("Skipping content update for view '%s', content is the same.", viewName)
		newContent.Refresh()
	}
	sh.currentView = viewName
	sh.updateHeaderContent(headerElements...)
}
func (sh *skinHunterApp) createHeaderContainer(content *fyne.Container) fyne.CanvasObject { /* ... as before ... */
	hb := theme.BackgroundColor()
	bc := theme.ShadowColor()
	bgr := canvas.NewRectangle(hb)
	btm := canvas.NewRectangle(bc)
	btm.SetMinSize(fyne.NewSize(1, 1))
	pc := container.NewPadded(content)
	hl := container.NewBorder(nil, btm, nil, nil, pc)
	return container.NewStack(bgr, hl)
}
func (sh *skinHunterApp) updateHeaderContent(elements ...fyne.CanvasObject) { /* ... as before ... */
	sh.headerMutex.Lock()
	defer sh.headerMutex.Unlock()
	if sh.headerContent == nil {
		return
	}
	sh.headerContent.Objects = elements
	sh.headerContent.Refresh()
}
func (sh *skinHunterApp) updateStatus(status string) { /* ... as before ... */
	if sh.statusLabel != nil {
		sh.statusLabel.SetText(status)
		log.Printf("App Status: %s", status)
	}
}
func (sh *skinHunterApp) createFooter() fyne.CanvasObject { /* ... as before ... */
	tabs := []struct {
		label  string
		icon   fyne.Resource
		view   string
		action func()
	}{{"Champions", theme.HomeIcon(), "champions_grid", nil}, {"Search", theme.SearchIcon(), "", func() { sh.showOmniSearch() }}, {"Installed", theme.DownloadIcon(), "installed_view", nil}, {"Profile", theme.AccountIcon(), "profile_view", nil}}
	btns := make([]fyne.CanvasObject, len(tabs))
	for i := range tabs {
		t := tabs[i]
		btn := ui.NewTabButton(t.label, t.icon, func() {
			if t.action != nil {
				t.action()
			} else if t.view != "" {
				sh.switchView(t.view)
			}
		})
		btns[i] = container.NewMax(btn)
	}
	fc := container.NewGridWithColumns(len(tabs), btns...)
	fbg := theme.BackgroundColor()
	bc := theme.ShadowColor()
	bgr := canvas.NewRectangle(fbg)
	tb := canvas.NewRectangle(bc)
	tb.SetMinSize(fyne.NewSize(1, 1))
	pfc := container.NewPadded(fc)
	fl := container.NewBorder(tb, nil, nil, nil, pfc)
	return container.NewStack(bgr, fl)
}
func (sh *skinHunterApp) showLoading() { /* ... as before ... */
	sh.updateStatus("Loading...")
	if sh.centerContent != nil {
		ld := container.NewCenter(container.NewVBox(widget.NewLabel("Loading Application Data..."), widget.NewProgressBarInfinite()))
		sh.centerContent.Objects = []fyne.CanvasObject{ld}
		sh.centerContent.Refresh()
		sh.currentView = "loading"
		log.Println("UI state: Loading")
		lt := widget.NewLabelWithStyle("Loading...", fyne.TextAlignCenter, fyne.TextStyle{})
		sh.updateHeaderContent(layout.NewSpacer(), lt, layout.NewSpacer())
	}
}
func (sh *skinHunterApp) showError(errMsg string) { /* ... as before ... */
	sh.updateStatus("Error")
	if sh.centerContent != nil {
		ei := widget.NewIcon(theme.ErrorIcon())
		el := widget.NewLabelWithStyle(errMsg, fyne.TextAlignCenter, fyne.TextStyle{})
		el.Wrapping = fyne.TextWrapWord
		eb := container.NewVBox(layout.NewSpacer(), container.NewCenter(ei), widget.NewSeparator(), el, layout.NewSpacer())
		pe := container.NewPadded(eb)
		sh.centerContent.Objects = []fyne.CanvasObject{pe}
		sh.centerContent.Refresh()
		sh.currentView = "error"
		log.Println("UI state: Error - ", errMsg)
		et := widget.NewLabelWithStyle("Error", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		sh.updateHeaderContent(layout.NewSpacer(), et, layout.NewSpacer())
	}
}
func (sh *skinHunterApp) showChampionsGrid() { sh.switchView("champions_grid") }
func (sh *skinHunterApp) showChampionDetail(champ data.ChampionSummary) { /* ... as before ... */
	log.Printf("Navigating to details for champion: %s (ID: %d)", champ.Name, champ.ID)
	sh.selectedChampion = champ
	sh.switchView("champion_detail")
}
func (sh *skinHunterApp) goBack() { /* ... as before ... */
	log.Printf("Go back requested from view: %s", sh.currentView)
	if sh.currentView != "champions_grid" {
		sh.switchView("champions_grid")
	} else {
		log.Println("Already at base view.")
	}
}
func (sh *skinHunterApp) showOmniSearch() { /* ... as before ... */
	log.Println("OmniSearch tab tapped.")
	si := widget.NewEntry()
	si.SetPlaceHolder("Search Champions, Skins...")
	dc := container.NewVBox(widget.NewLabel("OmniSearch (WIP)"), si, widget.NewLabel("Enter search term..."))
	dialog.ShowCustomConfirm("Search", "Search", "Cancel", dc, func(c bool) {
		if c {
			st := si.Text
			log.Printf("Search WIP: '%s'", st)
			dialog.ShowInformation("Search", fmt.Sprintf("Search for '%s' not implemented.", st), sh.window)
		} else {
			log.Println("Search cancelled.")
		}
	}, sh.window)
}

// --- End of main.go ---
