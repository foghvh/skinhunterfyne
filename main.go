// skinhunter/main.go
package main

import (
	"fmt"
	"log" // For basic setup if needed
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

	// Standard image formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	// Optional: WebP support (requires CGO, ensure setup if used)
	// _ "golang.org/x/image/webp" // Uncomment if WEBP support is needed and CGO is enabled
)

const appName = "Skin Hunter Fyne"

type skinHunterApp struct {
	fyneApp fyne.App
	window  fyne.Window

	headerContent *fyne.Container // Container for dynamic header elements
	headerMutex   sync.Mutex      // Mutex for safe header updates
	footer        fyne.CanvasObject
	centerContent *fyne.Container   // Main content area (Max layout)
	background    fyne.CanvasObject // Background layer
	navBackButton *widget.Button    // Back button for navigation

	// State tracking
	statusLabel      *widget.Label        // For internal status/debug
	currentView      string               // Name of the current view
	selectedChampion data.ChampionSummary // Champion for detail view
	isDetailView     bool                 // Flag for detail view context
}

func main() {
	// Basic setup (e.g., environment variables, logging config)
	// Example: Set Fyne theme (optional, default is usually light)
	// os.Setenv("FYNE_THEME", "dark")

	shApp := &skinHunterApp{
		currentView:   "loading",          // Start in loading state
		centerContent: container.NewMax(), // Use Max layout for central content switching
	}

	shApp.fyneApp = app.New()
	// Optional: Set App ID for packaging/desktop integration
	// shApp.fyneApp.SetAppID("com.example.skinhunter")

	shApp.window = shApp.fyneApp.NewWindow(appName)
	shApp.window.Resize(fyne.NewSize(950, 720)) // Adjusted size
	shApp.window.CenterOnScreen()
	shApp.window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File", fyne.NewMenuItem("Quit", func() { shApp.fyneApp.Quit() })),
		// Add other menus (View, Help) if needed
	))

	// --- Create UI Components ---
	shApp.headerContent = container.NewHBox() // HBox for header items (back button, title, spacer)
	header := shApp.createHeaderContainer(shApp.headerContent)
	shApp.footer = shApp.createFooter()
	shApp.navBackButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		shApp.goBack()
	})
	shApp.navBackButton.Hide() // Start hidden

	// Status label (primarily for debugging or future use)
	shApp.statusLabel = widget.NewLabel("Initializing...")

	// Background color
	// bgColor := color.NRGBA{R: 0x1e, G: 0x1e, B: 0x24, A: 0xff} // Dark Gray/Blue
	bgColor := theme.BackgroundColor() // Use theme background color for consistency
	shApp.background = canvas.NewRectangle(bgColor)

	// --- Main Layout ---
	// Stack background and center content area
	layeredContent := container.NewStack(shApp.background, shApp.centerContent)
	// Border layout for the main window structure
	mainAppLayout := container.NewBorder(header, shApp.footer, nil, nil, layeredContent)

	shApp.window.SetContent(mainAppLayout)
	shApp.window.SetMaster() // Ensure this is the main window

	// Initial UI state: Show loading indicator immediately
	shApp.showLoading()

	// --- Asynchronous Data Initialization ---
	go func() {
		err := data.InitData() // This fetches champion summaries and potentially all skins
		if err != nil {
			log.Printf("FATAL: Failed to initialize data: %v", err)
			// Show error in the UI. Updating UI components from a different goroutine
			// that are already visible is generally safe in Fyne after initial setup.
			shApp.showError(fmt.Sprintf("Failed to load required data:\n%v\n\nPlease check connection or restart.", err))

			// Send notification (optional)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Initialization Error",
				Content: "Failed to load required application data.",
			})
			return // Stop further processing if init fails
		}

		// Data loaded successfully, switch to the main view (Champions Grid)
		// Update UI directly from this goroutine is generally safe here.
		shApp.updateStatus("Ready") // Update internal status
		shApp.showChampionsGrid()   // Switch to the champion grid view

	}()

	// Start the application event loop
	shApp.window.ShowAndRun()
}

// createHeaderContainer: Sets up the static structure of the header (background, border).
func (sh *skinHunterApp) createHeaderContainer(content *fyne.Container) fyne.CanvasObject {
	// Use theme colors for adaptability
	headerBackground := theme.BackgroundColor()
	borderColor := theme.ShadowColor() // Or theme.SeparatorColor()

	bgRect := canvas.NewRectangle(headerBackground)
	// Thin bottom border line
	bottomBorder := canvas.NewRectangle(borderColor)
	bottomBorder.SetMinSize(fyne.NewSize(1, 1)) // Make it 1 pixel high

	// Add padding around the actual header content (HBox)
	paddedContent := container.NewPadded(content)
	// Place the border below the padded content
	headerLayout := container.NewBorder(nil, bottomBorder, nil, nil, paddedContent)

	// Stack background and the layout with border
	return container.NewStack(bgRect, headerLayout)
}

// updateHeaderContent: Safely updates the items within the header's HBox.
func (sh *skinHunterApp) updateHeaderContent(elements ...fyne.CanvasObject) {
	sh.headerMutex.Lock()
	defer sh.headerMutex.Unlock()

	if sh.headerContent == nil {
		log.Println("ERROR: updateHeaderContent called but headerContent is nil.")
		return
	}
	sh.headerContent.Objects = elements // Replace objects in the HBox
	// *** CORRECTION HERE ***
	sh.headerContent.Refresh() // Refresh the HBox (use 'sh', not 'shApp')
}

// updateStatus: Updates the internal status label (for logging/debugging).
func (sh *skinHunterApp) updateStatus(status string) {
	if sh.statusLabel != nil {
		sh.statusLabel.SetText(status)
		log.Printf("App Status: %s", status) // Log status changes
	}
}

// createFooter: Creates the bottom navigation bar using tab buttons.
func (sh *skinHunterApp) createFooter() fyne.CanvasObject {
	tabs := []struct {
		label  string
		icon   fyne.Resource
		view   string // Target view name for switchView
		action func() // Custom action (like opening search dialog)
	}{
		{"Champions", theme.HomeIcon(), "champions_grid", nil},
		{"Search", theme.SearchIcon(), "", func() { sh.showOmniSearch() }}, // Action-based tab
		{"Installed", theme.DownloadIcon(), "installed_view", nil},         // Placeholder view
		{"Profile", theme.AccountIcon(), "profile_view", nil},              // Placeholder view
	}

	buttons := make([]fyne.CanvasObject, len(tabs))
	for i := range tabs {
		t := tabs[i] // Capture loop variable for closure
		button := ui.NewTabButton(t.label, t.icon, func() {
			if t.action != nil {
				t.action() // Execute custom action if defined
			} else if t.view != "" {
				log.Printf("Tab '%s' tapped, switching to view '%s'", t.label, t.view)
				sh.switchView(t.view) // Switch view otherwise
			} else {
				log.Printf("Tab '%s' tapped, no action/view defined.", t.label)
			}
		})
		// Wrap button in Max to allow Grid layout to distribute space
		buttons[i] = container.NewMax(button)
	}

	// Use Grid layout for even distribution
	footerContent := container.NewGridWithColumns(len(tabs), buttons...)

	// Footer styling (similar to header)
	footerBackground := theme.BackgroundColor()
	borderColor := theme.ShadowColor()

	bgRect := canvas.NewRectangle(footerBackground)
	// Thin top border line
	topBorder := canvas.NewRectangle(borderColor)
	topBorder.SetMinSize(fyne.NewSize(1, 1))

	// Add padding around the footer buttons
	paddedFooterContent := container.NewPadded(footerContent)
	// Place border above the padded content
	footerLayout := container.NewBorder(topBorder, nil, nil, nil, paddedFooterContent)

	// Stack background and layout with border
	return container.NewStack(bgRect, footerLayout)
}

// switchView: Handles changing the main content area and updating header.
func (sh *skinHunterApp) switchView(viewName string) {
	// Avoid unnecessary reloads if already in the target view (unless it's the base grid)
	if sh.currentView == viewName && viewName != "champions_grid" {
		log.Printf("Already in view '%s', no switch.", viewName)
		return
	}
	log.Printf("Switching view from '%s' to '%s'", sh.currentView, viewName)

	sh.isDetailView = false // Reset detail view flag
	var newContent fyne.CanvasObject
	headerElements := []fyne.CanvasObject{layout.NewSpacer()} // Default: empty header

	// --- Determine Content and Header based on viewName ---
	switch viewName {
	case "champions_grid":
		sh.navBackButton.Hide() // Hide back button on main grid
		// Title for champion grid (optional)
		titleLabel := widget.NewLabelWithStyle("Champions", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		// Create the grid view (now with lazy loading)
		newContent = ui.NewChampionGrid(func(champ data.ChampionSummary) {
			sh.showChampionDetail(champ) // Callback to show detail view
		})

	case "champion_detail":
		sh.isDetailView = true
		sh.navBackButton.Show() // Show back button for detail view
		// Use champion name in header title
		titleText := "Champion Detail"
		if sh.selectedChampion.Name != "" {
			titleText = sh.selectedChampion.Name // Use actual champ name
		}
		titleLabel := widget.NewLabelWithStyle(titleText, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		// Header: Back button, Spacer, Title, Spacer
		headerElements = []fyne.CanvasObject{
			sh.navBackButton,
			layout.NewSpacer(),
			titleLabel,
			layout.NewSpacer(),
		}
		// Create the champion detail view (now uses lazy-loading skins grid)
		newContent = ui.NewChampionView(
			sh.selectedChampion,
			sh.window,
			func(skin data.Skin, allChromas []data.Chroma) {
				// Callback when a skin is selected in the detail view's grid
				ui.ShowSkinDialog(skin, allChromas, sh.window)
			},
		)

	case "installed_view":
		sh.navBackButton.Hide()
		titleLabel := widget.NewLabelWithStyle("Installed", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		newContent = container.NewCenter(widget.NewLabel("Installed Skins View (Not Implemented)")) // Placeholder

	case "profile_view":
		sh.navBackButton.Hide()
		titleLabel := widget.NewLabelWithStyle("Profile", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		newContent = container.NewCenter(widget.NewLabel("User Profile View (Not Implemented)")) // Placeholder

	default: // Unknown view
		sh.navBackButton.Hide()
		log.Printf("Warning: Unknown view name requested: '%s'", viewName)
		titleLabel := widget.NewLabelWithStyle("Error", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerElements = []fyne.CanvasObject{layout.NewSpacer(), titleLabel, layout.NewSpacer()}
		newContent = container.NewCenter(widget.NewLabel(fmt.Sprintf("Error: View '%s' not found.", viewName)))
	}

	// --- Update Center Content ---
	if sh.centerContent == nil {
		log.Println("CRITICAL ERROR: centerContent container is nil during switchView!")
		return // Cannot proceed
	}
	sh.centerContent.Objects = []fyne.CanvasObject{newContent} // Replace content
	sh.centerContent.Refresh()
	sh.currentView = viewName // Update current view state
	log.Printf("Center content updated to show '%s'", viewName)

	// --- Update Header Content ---
	sh.updateHeaderContent(headerElements...) // Update header elements dynamically
}

// showLoading: Displays a loading indicator in the center content area.
func (sh *skinHunterApp) showLoading() {
	sh.updateStatus("Loading...") // Update internal status
	if sh.centerContent != nil {
		loadingIndicator := widget.NewProgressBarInfinite()
		loadingLabel := widget.NewLabel("Loading Application Data...")
		loadingBox := container.NewVBox(loadingLabel, loadingIndicator)
		sh.centerContent.Objects = []fyne.CanvasObject{container.NewCenter(loadingBox)}
		sh.centerContent.Refresh()
		sh.currentView = "loading" // Set state
		log.Println("UI state: Loading")
		// Set a generic loading header
		loadingTitle := widget.NewLabelWithStyle("Loading...", fyne.TextAlignCenter, fyne.TextStyle{})
		sh.updateHeaderContent(layout.NewSpacer(), loadingTitle, layout.NewSpacer())
	} else {
		log.Println("ERROR: centerContent is nil during showLoading!")
	}
}

// showError: Displays an error message in the center content area.
func (sh *skinHunterApp) showError(errMsg string) {
	sh.updateStatus("Error") // Update internal status
	if sh.centerContent != nil {
		errorIcon := widget.NewIcon(theme.ErrorIcon())
		errorLabel := widget.NewLabelWithStyle(errMsg, fyne.TextAlignCenter, fyne.TextStyle{})
		errorLabel.Wrapping = fyne.TextWrapWord // Wrap long error messages
		// Center the icon and label vertically
		errorBox := container.NewVBox(
			layout.NewSpacer(), // Push down
			container.NewCenter(errorIcon),
			widget.NewSeparator(),
			errorLabel,
			layout.NewSpacer(), // Push up
		)
		// Add padding around the error box
		paddedError := container.NewPadded(errorBox)
		sh.centerContent.Objects = []fyne.CanvasObject{paddedError}
		sh.centerContent.Refresh()
		sh.currentView = "error" // Set state
		log.Println("UI state: Error - ", errMsg)
		// Update header to show error state
		errorTitle := widget.NewLabelWithStyle("Error", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		sh.updateHeaderContent(layout.NewSpacer(), errorTitle, layout.NewSpacer())
	} else {
		log.Println("ERROR: centerContent is nil during showError!")
	}
}

// showChampionsGrid: Convenience function to switch to the champion grid.
func (sh *skinHunterApp) showChampionsGrid() {
	sh.switchView("champions_grid")
}

// showChampionDetail: Stores selected champion and switches to detail view.
func (sh *skinHunterApp) showChampionDetail(champ data.ChampionSummary) {
	log.Printf("Navigating to details for champion: %s (ID: %d)", champ.Name, champ.ID)
	sh.selectedChampion = champ      // Store the selected champion's summary
	sh.switchView("champion_detail") // Trigger the view switch
}

// goBack: Handles the action for the navigation back button.
func (sh *skinHunterApp) goBack() {
	log.Printf("Go back requested from view: %s", sh.currentView)
	// If we are in any view that is not the main grid, go back to the grid
	if sh.currentView != "champions_grid" {
		sh.switchView("champions_grid")
	} else {
		// If already on the main grid, log it (or potentially exit app)
		log.Println("Already at base view (Champions Grid), cannot go back further.")
		// Optional: sh.fyneApp.Quit() // Uncomment if back on main grid should quit
	}
}

// showOmniSearch: Displays a placeholder search dialog.
func (sh *skinHunterApp) showOmniSearch() {
	log.Println("OmniSearch tab tapped.")
	searchInput := widget.NewEntry()
	searchInput.SetPlaceHolder("Search Champions, Skins...")
	// Dialog content
	dialogContent := container.NewVBox(
		widget.NewLabel("OmniSearch (Work in Progress)"),
		searchInput,
		widget.NewLabel("Enter search term and press 'Search'."),
	)
	// Show a confirmation dialog for search input
	dialog.ShowCustomConfirm(
		"Search",      // Title
		"Search",      // Confirm button text
		"Cancel",      // Dismiss button text
		dialogContent, // Content
		func(confirmed bool) { // Callback
			if confirmed {
				searchText := searchInput.Text
				log.Printf("Search submitted (WIP): '%s'", searchText)
				// Placeholder for actual search logic
				dialog.ShowInformation("Search", fmt.Sprintf("Search for '%s' is not yet implemented.", searchText), sh.window)
			} else {
				log.Println("Search cancelled.")
			}
		},
		sh.window, // Parent window
	)
}

// --- End of main.go ---
