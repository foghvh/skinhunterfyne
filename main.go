// skinhunter/main.go
package main

import (
	"fmt"
	"image/color"
	"log"     // Keep for potential future use like Mkdir
	"strings" // Import strings for ReplaceAll

	"skinhunter/data"
	"skinhunter/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	// "fyne.io/fyne/v2/driver/desktop" // No longer needed here
	"fyne.io/fyne/v2/dialog" // Needed for OmniSearch placeholder
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	// Import image decoders
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	// !!! ADD WEBP DECODER IMPORT !!!
	_ "golang.org/x/image/webp"
)

const appName = "Skin Hunter Fyne"

type skinHunterApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// UI Elements
	header        fyne.CanvasObject
	footer        fyne.CanvasObject
	leftNav       *fyne.Container
	rightNav      *fyne.Container
	centerContent *fyne.Container // Main area that changes
	background    fyne.CanvasObject
	// UI Element references needed for updates
	navBackButton   fyne.Widget // Direct reference to back button
	navToggleButton fyne.Widget // Direct reference to toggle button

	// State
	statusLabel      *widget.Label
	currentView      string               // e.g., "champions_grid", "champion_detail", "skin_line_detail"
	selectedChampion data.ChampionSummary // Store selected champ for detail view
	// selectedSkinLine data.SkinLine // Add later if needed
	isDetailView bool // Helper for back button visibility
}

func main() {
	// Set environment variable for Fyne development (optional)
	// os.Setenv("FYNE_THEME", "dark") // Enforce dark theme for testing

	shApp := &skinHunterApp{
		currentView:   "loading",          // Start in loading state
		centerContent: container.NewMax(), // Max layout allows easy swapping
	}

	shApp.fyneApp = app.New()
	// Set unique app ID via FyneApp.toml usually, removing runtime call

	// Consider setting custom theme here if needed
	// myTheme := &ui.MyTheme{} // Replace with your theme definition
	// shApp.fyneApp.Settings().SetTheme(myTheme)

	shApp.window = shApp.fyneApp.NewWindow(appName)

	// Initial size and positioning
	shApp.window.Resize(fyne.NewSize(1150, 800)) // Slightly wider
	shApp.window.CenterOnScreen()
	shApp.window.SetMainMenu(fyne.NewMainMenu( // Add default quit menu item
		fyne.NewMenu("File", fyne.NewMenuItem("Quit", func() { shApp.fyneApp.Quit() })),
	))

	// Create UI Components
	shApp.header = shApp.createHeader()
	shApp.footer = shApp.createFooter()
	// Create nav bars *and store references to buttons*
	shApp.leftNav = shApp.createLeftNav()
	shApp.rightNav = shApp.createRightNav()

	// Solid color background matching header/footer
	bgColor := color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff} // #0f1729
	shApp.background = canvas.NewRectangle(bgColor)

	// --- Layout Construction ---
	// Center area flanked by nav bars
	contentWithNav := container.NewBorder(nil, nil, shApp.leftNav, shApp.rightNav, shApp.centerContent)
	// Layer background and the main content+nav area
	layeredContent := container.NewStack(shApp.background, contentWithNav)
	// Final app layout with header and footer
	mainAppLayout := container.NewBorder(shApp.header, shApp.footer, nil, nil, layeredContent)

	shApp.window.SetContent(mainAppLayout)
	shApp.window.SetMaster()

	// --- Start Data Initialization and Show Initial View ---
	shApp.showLoading() // Show loading indicator first
	go func() {
		err := data.InitData() // Initialize champion list, skins map, etc.
		if err != nil {
			log.Printf("FATAL: Failed to initialize data: %v", err)
			// Show error state in UI - Showing widget content is generally safe from goroutine via Fyne's handling.
			shApp.showError(fmt.Sprintf("Failed to load required data:\n%v\n\nPlease check connection or restart.", err))
			return
		}
		// Data loaded successfully, switch to initial view (Champions Grid)
		shApp.updateStatus("Ready")
		shApp.showChampionsGrid() // Display the champion grid
	}()

	shApp.window.ShowAndRun()
}

func (sh *skinHunterApp) createHeader() fyne.CanvasObject {
	// Use resource loading for the logo
	logoRes, err := fyne.LoadResourceFromURLString("https://i.imgur.com/m40l0qA.png")
	var logo fyne.CanvasObject
	if err != nil {
		log.Println("WARN: Failed to load logo:", err)
		logo = widget.NewLabel("Skin Hunter") // Fallback text
	} else {
		logoImg := canvas.NewImageFromResource(logoRes)
		logoImg.FillMode = canvas.ImageFillContain
		logoImg.SetMinSize(fyne.NewSize(180, 30)) // Adjust size as needed
		logo = logoImg
	}

	// Status Label (Box)
	sh.statusLabel = widget.NewLabel("Loading...")
	sh.statusLabel.Alignment = fyne.TextAlignCenter
	// Add padding/background for status label like screenshot
	statusBg := canvas.NewRectangle(color.NRGBA{R: 0x33, G: 0x33, B: 0x36, A: 0x90}) // Semi-transparent grey
	// statusBorder removed
	statusBox := container.NewStack(
		statusBg,
		container.NewPadded(sh.statusLabel),
	)

	// Header Content Layout
	headerContent := container.NewHBox(
		container.NewPadded(logo), // Pad logo slightly
		layout.NewSpacer(),        // Pushes status label towards center
		statusBox,                 // Centered status
		layout.NewSpacer(),        // Pushes right content away
		// Placeholder for potential right-aligned header items (size matches logo padding approx)
		container.NewPadded(canvas.NewRectangle(color.Transparent)),
	)

	// Header Background and Bottom Border
	headerBackground := canvas.NewRectangle(color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff}) // #0f1729
	// Simulate bottom border with a thin rectangle
	borderColor := color.NRGBA{R: 0x58, G: 0x6a, B: 0x9e, A: 0xff} // #586a9e from React border
	bottomBorder := canvas.NewRectangle(borderColor)
	bottomBorder.SetMinSize(fyne.NewSize(1, 2)) // Very thin height for border line

	headerLayout := container.NewBorder(
		nil,           // Top
		bottomBorder,  // Bottom (acts as border)
		nil,           // Left
		nil,           // Right
		headerContent, // Center content
	)

	// Stack background and the layout with border
	return container.NewStack(headerBackground, headerLayout)
}

func (sh *skinHunterApp) updateStatus(status string) {
	if sh.statusLabel != nil {
		// log.Printf("Updating status: %s", status) // Can be noisy, uncomment if needed
		sh.statusLabel.SetText(status)
	}
}

func (sh *skinHunterApp) createFooter() fyne.CanvasObject {
	tabs := []struct {
		label  string
		icon   fyne.Resource
		view   string // Target view identifier for switchView
		action func() // Optional direct action
	}{
		{"Champions", theme.ListIcon(), "champions_grid", nil},
		{"Search", theme.SearchIcon(), "", func() { sh.showOmniSearch() }}, // Example action
		{"Installed", theme.DownloadIcon(), "installed_view", nil},
		{"Profile", theme.HomeIcon(), "profile_view", nil}, // Simplified for now
	}

	buttons := []fyne.CanvasObject{}
	for i := range tabs {
		t := tabs[i] // Capture loop variable correctly
		button := ui.NewTabButton(t.label, t.icon, func() {
			if t.action != nil {
				t.action()
			} else if t.view != "" {
				log.Printf("Tab '%s' tapped, switching to view '%s'", t.label, t.view)
				sh.switchView(t.view) // Use switchView to change centerContent
			} else {
				log.Printf("Tab '%s' tapped, no action/view defined.", t.label)
			}
		})
		buttons = append(buttons, button)
	}

	// Arrange buttons with spacers for centered grouping
	footerContent := container.NewHBox(
		layout.NewSpacer(), // Push group from left edge
		buttons[0],
		layout.NewSpacer(), // Space between buttons
		buttons[1],
		layout.NewSpacer(), // Space between buttons
		buttons[2],
		layout.NewSpacer(), // Space between buttons
		buttons[3],
		layout.NewSpacer(), // Push group from right edge
	)

	// Background and Top Border (similar to header)
	footerBackground := canvas.NewRectangle(color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0xff}) // #0f1729
	borderColor := color.NRGBA{R: 0x58, G: 0x6a, B: 0x9e, A: 0xff}                           // #586a9e
	topBorder := canvas.NewRectangle(borderColor)
	topBorder.SetMinSize(fyne.NewSize(1, 2)) // Thin height

	footerLayout := container.NewBorder(
		topBorder,                          // Top (acts as border)
		nil,                                // Bottom
		nil,                                // Left
		nil,                                // Right
		container.NewPadded(footerContent), // Center content with padding
	)

	return container.NewStack(footerBackground, footerLayout) // Stack background and layout
}

// createLeftNav creates the left navigation area and stores button references
func (sh *skinHunterApp) createLeftNav() *fyne.Container {
	// Create buttons and store references in shApp struct
	sh.navBackButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		sh.goBack()
	})
	sh.navBackButton.(*widget.Button).Importance = widget.MediumImportance // Style it
	sh.navBackButton.Hide()                                                // Start hidden

	sh.navToggleButton = widget.NewButtonWithIcon("", theme.ListIcon(), func() {
		log.Println("INFO: Champion/Skinline toggle placeholder clicked.")
		// TODO: Implement toggle logic
	})
	sh.navToggleButton.Hide() // Initially hidden

	// Use VBox for arrangement
	navBox := container.NewVBox(
		layout.NewSpacer(),
		sh.navBackButton,   // Add reference
		sh.navToggleButton, // Add reference
		layout.NewSpacer(),
	)

	// Return the padded container (FixedWidth removed)
	return container.NewPadded(navBox)
}

// createRightNav creates the right navigation area
func (sh *skinHunterApp) createRightNav() *fyne.Container {
	startButton := widget.NewButton("Start", func() {
		log.Println("INFO: 'Start' button placeholder clicked.")
		// TODO: Implement 'Start' functionality
	})
	// Use VBox similar to leftNav
	navBox := container.NewVBox(
		layout.NewSpacer(),
		startButton,
		layout.NewSpacer(),
	)
	// Return the padded container (FixedWidth removed)
	return container.NewPadded(navBox)
}

// switchView changes the content of the center area and manages nav visibility
func (sh *skinHunterApp) switchView(viewName string) {
	// Avoid reloading the same view unless navigating back from detail
	if sh.currentView == viewName && !sh.isDetailView {
		log.Printf("Already in view '%s', no change.", viewName)
		return
	}

	log.Printf("Switching view from '%s' to '%s'", sh.currentView, viewName)
	sh.updateStatus("Loading " + strings.ReplaceAll(viewName, "_", " ") + "...") // User feedback

	var newContent fyne.CanvasObject
	wasDetailView := sh.isDetailView // Store previous state
	sh.isDetailView = false          // Reset detail flag for new view

	switch viewName {
	case "champions_grid":
		// Create the grid, passing the champion selection callback
		newContent = ui.NewChampionGrid(func(champ data.ChampionSummary) {
			sh.showChampionDetail(champ) // This triggers another switchView call
		})

	case "champion_detail":
		sh.isDetailView = true // Mark this as a detail view
		// Create champion view, passing selected champion and callbacks
		newContent = ui.NewChampionView(
			sh.selectedChampion,
			func() { sh.goBack() }, // Back button callback
			func(skin data.Skin, allChromas []data.Chroma) { // Skin selection callback
				ui.ShowSkinDialog(skin, allChromas, sh.window)
			},
		)

	case "installed_view":
		newContent = container.NewCenter(widget.NewLabel("Installed Skins (Not Implemented)"))
	case "profile_view":
		newContent = container.NewCenter(widget.NewLabel("User Profile (Not Implemented)"))

	default:
		log.Printf("Warning: Unknown view name '%s'", viewName)
		newContent = container.NewCenter(widget.NewLabel("Error: View not found"))
	}

	// Safely update central content
	if sh.centerContent != nil {
		sh.centerContent.Objects = []fyne.CanvasObject{newContent} // Replace content
		sh.centerContent.Refresh()
	} else {
		log.Println("ERROR: centerContent is nil during switchView!")
	}
	sh.currentView = viewName

	// Update Left Nav Button Visibility *after* view is set
	sh.updateLeftNav(sh.isDetailView, wasDetailView)

	// Update status - consider letting views update status when fully loaded if needed
	sh.updateStatus("Ready")
}

// updateLeftNav safely updates the visibility of nav buttons using stored references
func (sh *skinHunterApp) updateLeftNav(isDetail, wasDetail bool) {
	// Use the stored references directly
	if sh.navBackButton == nil || sh.navToggleButton == nil {
		log.Println("WARN: Nav buttons not initialized, cannot update visibility.")
		return
	}

	if isDetail {
		sh.navBackButton.Show()
		sh.navToggleButton.Hide()
	} else {
		sh.navBackButton.Hide()
		// Show toggle only in champion grid for now
		if sh.currentView == "champions_grid" {
			// sh.navToggleButton.Show() // Keep hidden until fully implemented
			sh.navToggleButton.Hide()
		} else {
			sh.navToggleButton.Hide()
		}
	}
	// No need to refresh the leftNav container explicitly,
	// Show/Hide on the widgets trigger necessary refresh.
}

// showLoading displays a loading indicator in the central area.
func (sh *skinHunterApp) showLoading() {
	sh.updateStatus("Loading...")
	if sh.centerContent != nil {
		loadingIndicator := widget.NewProgressBarInfinite()
		sh.centerContent.Objects = []fyne.CanvasObject{container.NewCenter(loadingIndicator)}
		sh.centerContent.Refresh()
	} else {
		log.Println("ERROR: centerContent is nil during showLoading!")
	}
	sh.currentView = "loading"
	// Ensure nav is updated *after* creating the nav buttons
	if sh.navBackButton != nil { // Check if nav is created before updating
		sh.updateLeftNav(false, sh.isDetailView)
	}
}

// showError displays an error message in the central area.
func (sh *skinHunterApp) showError(errMsg string) {
	sh.updateStatus("Error")
	if sh.centerContent != nil {
		errorLabel := widget.NewLabelWithStyle(errMsg, fyne.TextAlignCenter, fyne.TextStyle{})
		errorLabel.Wrapping = fyne.TextWrapWord
		sh.centerContent.Objects = []fyne.CanvasObject{container.NewPadded(container.NewCenter(errorLabel))}
		sh.centerContent.Refresh()
	} else {
		log.Println("ERROR: centerContent is nil during showError!")
	}
	sh.currentView = "error"
	if sh.navBackButton != nil { // Check if nav is created before updating
		sh.updateLeftNav(false, sh.isDetailView)
	}
}

// --- Navigation Actions ---

// showChampionsGrid initiates switching to the main champion grid view.
func (sh *skinHunterApp) showChampionsGrid() {
	sh.switchView("champions_grid")
}

// showChampionDetail stores the selected champion and initiates switching to the detail view.
func (sh *skinHunterApp) showChampionDetail(champ data.ChampionSummary) {
	// log.Printf("Navigating to details for champion: %s", champ.Name)
	sh.selectedChampion = champ // Store selected champion data
	sh.switchView("champion_detail")
}

// goBack handles navigating back, typically from a detail view to a list view.
func (sh *skinHunterApp) goBack() {
	// log.Printf("Go back requested from view: %s", sh.currentView)
	// Simple logic: always go back to champions grid for now.
	// Could be enhanced with a navigation stack later.
	if sh.isDetailView {
		sh.switchView("champions_grid")
	} else {
		log.Println("Already at base view, cannot go back further.")
	}
}

// showOmniSearch handles the Search tab action (placeholder dialog).
func (sh *skinHunterApp) showOmniSearch() {
	// Use v2 dialog parent type
	dialog.ShowInformation("Search", "Omnisearch not implemented yet.", sh.window) // Corrected parent type
}
