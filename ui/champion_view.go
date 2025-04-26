// skinhunter/ui/champion_view.go
package ui

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"unicode/utf8"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ChampionView represents the reusable champion detail view.
type ChampionView struct {
	widget.BaseWidget
	parentWindow fyne.Window
	onSkinSelect func(skin data.Skin, allChromas []data.Chroma)

	content         *fyne.Container // Stack(mainLayout, loading)
	mainLayout      *fyne.Container // Border(topSection, nil, nil, nil, skinsGridWidget)
	topSection      *fyne.Container
	champImage      *fyne.Container
	champNameLabel  *widget.Label
	champTitleLabel *widget.Label
	bioLabel        *widget.Label
	viewMoreButton  *widget.Button
	skinsTitleLabel *widget.Label

	// Reusable SkinsGrid Instance (using GridWrap internally now)
	skinsGridWidget *SkinsGrid // Holds the GridWrap based grid

	currentChampID int
	loading        *fyne.Container
	currentDetails *data.DetailedChampionData
}

// NewChampionView creates the initial structure of the champion view.
func NewChampionView(
	parentWindow fyne.Window,
	onSkinSelect func(skin data.Skin, allChromas []data.Chroma),
) *ChampionView {
	v := &ChampionView{
		parentWindow:   parentWindow,
		onSkinSelect:   onSkinSelect,
		currentChampID: -1,
	}
	v.ExtendBaseWidget(v)

	// --- Build initial static structure ---
	v.champImage = container.NewStack()
	v.champNameLabel = widget.NewLabelWithStyle(" ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	v.champTitleLabel = widget.NewLabel(" ")
	v.bioLabel = widget.NewLabel(" ")
	v.bioLabel.Wrapping = fyne.TextWrapWord
	v.viewMoreButton = widget.NewButton("View more", func() {})
	v.viewMoreButton.Disable()
	v.skinsTitleLabel = widget.NewLabelWithStyle(" Skins", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true})
	champTextInfo := container.NewVBox(v.champNameLabel, v.champTitleLabel)
	champHeader := container.NewHBox(v.champImage, container.NewPadded(champTextInfo))
	bioAndButton := container.NewVBox(v.bioLabel, container.NewHBox(layout.NewSpacer(), v.viewMoreButton))
	skinsIcon := widget.NewIcon(theme.ColorPaletteIcon())
	skinsTitleHeader := container.NewHBox(skinsIcon, v.skinsTitleLabel)
	topSectionContent := container.NewVBox(champHeader, widget.NewSeparator(), bioAndButton, widget.NewSeparator(), container.NewPadded(skinsTitleHeader))
	v.topSection = container.NewPadded(topSectionContent)

	// Create the reusable SkinsGrid instance ONCE
	v.skinsGridWidget = NewSkinsGrid(func(selectedSkin data.Skin) { // Pass the callback logic
		log.Printf("ChampionView received skin tap: %s", selectedSkin.Name)
		allChromasForChamp := make([]data.Chroma, 0)
		if v.currentDetails != nil {
			for _, s := range v.currentDetails.Skins {
				for _, ch := range s.Chromas {
					originID := ch.OriginSkinID
					if originID == 0 {
						originID = s.ID
					}
					if data.GetChampionIDFromSkinID(originID) == v.currentDetails.ID {
						chromaCopy := ch
						chromaCopy.OriginSkinID = originID
						allChromasForChamp = append(allChromasForChamp, chromaCopy)
					}
				}
			}
		}
		if v.onSkinSelect != nil {
			v.onSkinSelect(selectedSkin, allChromasForChamp)
		}
	})

	v.loading = container.NewCenter(container.NewVBox(widget.NewLabel("Loading..."), widget.NewProgressBarInfinite()))
	v.loading.Hide()
	// Main layout uses the SkinsGridWidget directly in the Center
	v.mainLayout = container.NewBorder(v.topSection, nil, nil, nil, v.skinsGridWidget) // Place widget directly
	v.content = container.NewStack(v.mainLayout, v.loading)

	return v
}

// CreateRenderer returns the renderer for the ChampionView.
func (v *ChampionView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.content)
}

// UpdateContent fetches details and updates the view's elements.
func (v *ChampionView) UpdateContent(championSummary data.ChampionSummary) {
	if championSummary.ID == v.currentChampID {
		log.Printf("ChampionView: Update requested for same champion ID %d, skipping.", championSummary.ID)
		return
	}
	log.Printf("ChampionView: Updating content for %s (ID: %d)", championSummary.Name, championSummary.ID)
	v.currentChampID = championSummary.ID
	v.currentDetails = nil

	v.loading.Show()
	v.updateTopSectionPlaceholders(championSummary.Name)
	v.skinsGridWidget.UpdateSkins(nil) // Clear grid / show placeholder immediately
	v.content.Refresh()

	go func(champID int, name string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered: %v\n%s", r, string(debug.Stack())) /* handle */
			}
		}()
		details, err := data.FetchChampionDetails(champID)
		updateUIFunc := func() {}

		if err != nil {
			log.Printf("ChampionView ERROR fetching details for %s (ID: %d): %v", name, champID, err)
			updateUIFunc = func() {
				v.skinsGridWidget.showPlaceholder(fmt.Sprintf("Error loading skins for %s", name)) // Show error in grid
				v.loading.Hide()
				v.content.Refresh()
			}
		} else {
			log.Printf("ChampionView: Details fetched successfully for %s", details.Name)
			v.currentDetails = details
			updateUIFunc = func() {
				v.updateTopSection(*details)
				// *** CORRECTION: Call UpdateSkins (not StartPopulatingSkins) ***
				v.skinsGridWidget.UpdateSkins(details.Skins) // Trigger incremental population
				v.loading.Hide()
				v.content.Refresh()
			}
		}
		fyne.Do(updateUIFunc)
	}(championSummary.ID, championSummary.Name)
}

// updateTopSection updates the widgets in the top part of the view.
func (v *ChampionView) updateTopSection(details data.DetailedChampionData) { /* ... as before ... */
	log.Printf("Updating top section for %s", details.Name)
	v.updateChampionImage(details)
	v.champNameLabel.SetText(details.Name)
	title := ""
	if details.Title != "" {
		title = strings.Title(strings.ToLower(details.Title))
	}
	v.champTitleLabel.SetText(title)
	v.skinsTitleLabel.SetText(fmt.Sprintf("%s Skins", details.Name))
	bioExcerpt := details.ShortBio
	const maxBioLen = 180
	if utf8.RuneCountInString(bioExcerpt) > maxBioLen { /* truncate */
		count, cutoff := 0, 0
		for i := range bioExcerpt {
			if count >= maxBioLen {
				cutoff = i
				ls := strings.LastIndex(bioExcerpt[:cutoff], " ")
				if ls > maxBioLen-30 {
					cutoff = ls
				}
				break
			}
			count++
		}
		if cutoff == 0 {
			cutoff = len(bioExcerpt)
		}
		bioExcerpt = bioExcerpt[:cutoff] + "..."
	}
	v.bioLabel.SetText(bioExcerpt)
	v.viewMoreButton.Enable()
	v.viewMoreButton.OnTapped = func() {
		fullBioLabel := widget.NewLabel(details.ShortBio)
		fullBioLabel.Wrapping = fyne.TextWrapWord
		scrollBio := container.NewScroll(fullBioLabel)
		scrollBio.SetMinSize(fyne.NewSize(450, 350))
		dialog.ShowCustom(fmt.Sprintf("%s - Biography", details.Name), "Close", scrollBio, v.parentWindow)
	}
}

// updateTopSectionPlaceholders sets default/error state text.
func (v *ChampionView) updateTopSectionPlaceholders(name string) { /* ... as before ... */
	log.Printf("Updating top section placeholders for %s", name)
	if v.champImage == nil {
		v.champImage = container.NewStack()
	}
	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	imgSize := float32(64)
	imgAreaSize := fyne.NewSize(imgSize, imgSize)
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgAreaSize)
	v.champImage.Objects = []fyne.CanvasObject{placeholderRect, container.NewCenter(placeholderIcon)}
	v.champImage.Refresh()
	if v.champNameLabel != nil {
		v.champNameLabel.SetText(name)
	}
	if v.champTitleLabel != nil {
		v.champTitleLabel.SetText(" ")
	}
	if v.skinsTitleLabel != nil {
		v.skinsTitleLabel.SetText(fmt.Sprintf("%s Skins", name))
	}
	if v.bioLabel != nil {
		v.bioLabel.SetText("Loading bio...")
	}
	if v.viewMoreButton != nil {
		v.viewMoreButton.Disable()
		v.viewMoreButton.OnTapped = nil
	}
}

// updateChampionImage handles lazy-loading the champion portrait.
func (v *ChampionView) updateChampionImage(details data.DetailedChampionData) { /* ... as before ... */
	if v.champImage == nil {
		return
	}
	placeholderIcon := widget.NewIcon(theme.BrokenImageIcon())
	imgSize := float32(64)
	imgAreaSize := fyne.NewSize(imgSize, imgSize)
	placeholderRect := canvas.NewRectangle(theme.InputBorderColor())
	placeholderRect.SetMinSize(imgAreaSize)
	v.champImage.Objects = []fyne.CanvasObject{placeholderRect, container.NewCenter(placeholderIcon)}
	v.champImage.Refresh()
	go func(c data.DetailedChampionData, stack *fyne.Container) {
		imageUrl := data.Asset(c.SquarePortraitPath)
		if imageUrl == data.GetPlaceholderImageURL() {
			return
		}
		imgUri, parseErr := storage.ParseURI(imageUrl)
		if parseErr != nil {
			return
		}
		imgWidget := canvas.NewImageFromURI(imgUri)
		imgWidget.SetMinSize(imgAreaSize)
		imgWidget.FillMode = canvas.ImageFillContain
		fyne.Do(func() {
			if v.currentChampID == c.ID && stack != nil && stack.Visible() {
				stack.Objects = []fyne.CanvasObject{imgWidget}
				stack.Refresh()
			}
		})
	}(details, v.champImage)
}

// --- End of champion_view.go ---
