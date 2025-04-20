// skinhunter/ui/utils.go
package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"

	// Import specific decoders in main.go
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings" // Added for Content-Type check
	"sync"
	"time"

	"skinhunter/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var imageClient = &http.Client{Timeout: time.Second * 15} // Client specifically for images
var imageCache = &sync.Map{}                              // Concurrent map for image caching

// Image Resource Cache Entry
type cachedResource struct {
	Resource fyne.Resource
	Error    error
}

// !! MODIFIED LoadResourceFromURLWithCache - Added Content-Type logging and SVG check !!
func LoadResourceFromURLWithCache(urlStr string) (fyne.Resource, error) {
	// Handle empty or placeholder URLs directly
	if urlStr == "" || urlStr == data.GetPlaceholderImageURL() {
		return nil, fmt.Errorf("empty or placeholder URL requested")
	}

	// Check cache first
	if cached, ok := imageCache.Load(urlStr); ok {
		cachedRes := cached.(cachedResource)
		return cachedRes.Resource, cachedRes.Error
	}

	// Fetch if not in cache
	log.Printf("DEBUG: Image cache miss, fetching: %s", urlStr) // Log URL on cache miss
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		log.Printf("ERROR: Failed create request for %s: %v", urlStr, err)
		cacheErr := fmt.Errorf("failed to create request: %w", err)
		imageCache.Store(urlStr, cachedResource{Resource: theme.BrokenImageIcon(), Error: cacheErr})
		return theme.BrokenImageIcon(), cacheErr
	}
	req.Header.Set("User-Agent", "SkinHunterFyneApp/1.0")

	resp, err := imageClient.Do(req)
	if err != nil {
		log.Printf("ERROR: Failed GET image %s: %v", urlStr, err)
		cacheErr := fmt.Errorf("failed http get: %w", err)
		imageCache.Store(urlStr, cachedResource{Resource: theme.BrokenImageIcon(), Error: cacheErr})
		return theme.BrokenImageIcon(), cacheErr
	}
	defer resp.Body.Close()

	// Log content type BEFORE reading body
	contentType := resp.Header.Get("Content-Type")
	log.Printf("DEBUG: Fetched %s - Status: %d, Content-Type: %s", urlStr, resp.StatusCode, contentType)

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("bad status: %d for %s", resp.StatusCode, urlStr)
		log.Println(errMsg)
		imageCache.Store(urlStr, cachedResource{Resource: theme.BrokenImageIcon(), Error: errMsg})
		return theme.BrokenImageIcon(), errMsg
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed read image body %s: %v", urlStr, err)
		cacheErr := fmt.Errorf("failed to read body: %w", err)
		imageCache.Store(urlStr, cachedResource{Resource: theme.BrokenImageIcon(), Error: cacheErr})
		return theme.BrokenImageIcon(), cacheErr
	}

	// --- Format Check ---
	// If it's SVG, treat it as an "error" for canvas.Image but cache the resource anyway.
	if strings.Contains(contentType, "svg") {
		log.Printf("INFO: Detected SVG for %s. Caching resource but returning 'format error' for canvas.Image.", urlStr)
		resource := fyne.NewStaticResource(urlStr, imageData)
		// Cache potentially usable SVG resource but signal incompatibility with canvas.Image
		svgErr := fmt.Errorf("unsupported format: SVG")
		imageCache.Store(urlStr, cachedResource{Resource: resource, Error: svgErr})
		// Return BrokenImageIcon visually, but the actual SVG resource is cached if needed elsewhere
		return theme.BrokenImageIcon(), svgErr
	}

	// Try to decode standard formats + webp
	_, _, decodeErr := image.DecodeConfig(bytes.NewReader(imageData))
	if decodeErr != nil {
		// Genuine decode error for expected raster formats
		log.Printf("ERROR: Failed image.DecodeConfig for %s (ContentType: %s): %v. Caching as broken.", urlStr, contentType, decodeErr)
		cacheErr := fmt.Errorf("image decode failed: %w", decodeErr)
		imageCache.Store(urlStr, cachedResource{Resource: theme.BrokenImageIcon(), Error: cacheErr})
		return theme.BrokenImageIcon(), cacheErr
	} else {
		// Successfully decoded (PNG, JPG, GIF, WebP)
		// log.Printf("DEBUG: Successfully decoded image %s", urlStr)
		resource := fyne.NewStaticResource(urlStr, imageData) // Use URL as name
		imageCache.Store(urlStr, cachedResource{Resource: resource, Error: nil})
		return resource, nil
	}
}

// --- UI Element Creation Functions --- (Remain the same as previous version)
func NewAsyncImage(width, height float32) (*fyne.Container, *canvas.Image) {
	imgWidget := canvas.NewImageFromResource(nil)
	imgWidget.FillMode = canvas.ImageFillContain
	if width > 0 && height > 0 {
		imgWidget.SetMinSize(fyne.NewSize(width, height))
	} else {
		imgWidget.SetMinSize(fyne.NewSize(32, 32))
	}
	imgWidget.Hide()
	placeholderColor := color.NRGBA{R: 0x33, G: 0x33, B: 0x36, A: 0xff}
	placeholder := canvas.NewRectangle(placeholderColor)
	placeholder.SetMinSize(imgWidget.MinSize())
	loading := widget.NewProgressBarInfinite()
	loadingCenter := container.NewCenter(loading)
	stackContainer := container.NewStack(placeholder, loadingCenter, imgWidget)
	return stackContainer, imgWidget
}

func SetImageURL(imgWidget *canvas.Image, stack *fyne.Container, urlStr string) {
	if stack == nil || imgWidget == nil {
		log.Println("ERROR: SetImageURL nil stack or imgWidget")
		return
	}
	if len(stack.Objects) < 3 {
		log.Printf("ERROR: SetImageURL stack structure invalid: %d objects", len(stack.Objects))
		return
	}
	var placeholder, loading, img fyne.CanvasObject
	foundPlaceholder, foundLoading, foundImgWidget := false, false, false
	for _, obj := range stack.Objects { /* ... find components ... */
		if _, ok := obj.(*canvas.Rectangle); ok && !foundPlaceholder {
			placeholder = obj
			foundPlaceholder = true
		} else if centerContainer, ok := obj.(*fyne.Container); ok && !foundLoading {
			if len(centerContainer.Objects) == 1 {
				if _, isProgress := centerContainer.Objects[0].(*widget.ProgressBarInfinite); isProgress {
					loading = obj
					foundLoading = true
				}
			}
		} else if obj == imgWidget && !foundImgWidget {
			img = obj
			foundImgWidget = true
		}
	}
	if placeholder == nil || loading == nil || img == nil {
		log.Printf("WARN: Could not find elements in stack for SetImageURL (P:%t, L:%t, I:%t)", foundPlaceholder, foundLoading, foundImgWidget)
		return
	}

	imgWidget.Hide()
	loading.Show()
	placeholder.Show()
	stack.Refresh()
	go func() {
		resource, err := LoadResourceFromURLWithCache(urlStr)
		resourceToSet := theme.BrokenImageIcon() // Default to broken
		placeholderHidden := false               // Track if placeholder should be hidden
		if err == nil && resource != nil {
			resourceToSet = resource
			placeholderHidden = true // Hide placeholder only on success
		} else {
			log.Printf("ERROR: SetImageURL failed for %s: %v", urlStr, err)
		}
		imgWidget.Resource = resourceToSet
		imgWidget.Refresh()
		imgWidget.Show()
		loading.Hide()
		if placeholderHidden {
			placeholder.Hide()
		} else {
			placeholder.Show()
		}
		stack.Refresh()
	}()
}

func ChampionGridItem(champ data.ChampionSummary, onSelect func(champ data.ChampionSummary)) fyne.CanvasObject {
	imgSize := float32(80)
	imgContainer, imgWidget := NewAsyncImage(imgSize, imgSize)
	SetImageURL(imgWidget, imgContainer, data.GetChampionSquarePortraitURL(champ))
	imgWidget.FillMode = canvas.ImageFillContain
	nameLabel := widget.NewLabel(champ.Name)
	nameLabel.Alignment = fyne.TextAlignCenter
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	itemContent := container.NewVBox(imgContainer, container.NewPadded(nameLabel))
	tapButton := widget.NewButton("", func() { onSelect(champ) })
	return container.NewStack(itemContent, tapButton)
}

func SkinItem(skin data.Skin, onSelect func(skin data.Skin)) fyne.CanvasObject {
	if skin.IsBase {
		return nil
	}
	imgSize := float32(200)
	imgURL := data.GetSkinTileURL(skin)
	imgContainer, imgWidget := NewAsyncImage(imgSize, imgSize)
	SetImageURL(imgWidget, imgContainer, imgURL)
	imgWidget.FillMode = canvas.ImageFillContain
	nameLabel := widget.NewLabel(skin.Name)
	nameLabel.TextStyle = fyne.TextStyle{Bold: false}
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Wrapping = fyne.TextWrapOff
	var rarityIconURL string
	_, rarityIconURL = data.Rarity(skin)
	var rarityContainer fyne.CanvasObject = widget.NewLabel("")
	if rarityIconURL != "" {
		rarityRes, err := LoadResourceFromURLWithCache(rarityIconURL)
		if err == nil && rarityRes != nil {
			rarityIcon := canvas.NewImageFromResource(rarityRes)
			rarityIcon.SetMinSize(fyne.NewSize(16, 16))
			rarityIcon.FillMode = canvas.ImageFillContain
			rarityContainer = rarityIcon
		} else {
			log.Printf("WARN: Failed load rarity icon %s: %v", rarityIconURL, err)
		}
	}
	bottomContent := container.NewHBox(rarityContainer, widget.NewSeparator(), nameLabel)
	bgColor := color.NRGBA{R: 0x0A, G: 0x0E, B: 0x19, A: 0xD0}
	bgRect := canvas.NewRectangle(bgColor)
	bottomBar := container.NewMax(bgRect, container.NewPadded(bottomContent))
	topIcons := []fyne.CanvasObject{}
	iconSize := fyne.NewSize(24, 24)
	if skin.IsLegacy {
		legacyRes, err := LoadResourceFromURLWithCache(data.LegacyIconURL())
		if err == nil && legacyRes != nil {
			legacyIcon := canvas.NewImageFromResource(legacyRes)
			legacyIcon.SetMinSize(iconSize)
			legacyIcon.FillMode = canvas.ImageFillContain
			topIcons = append(topIcons, legacyIcon)
		} else {
			log.Printf("WARN: Failed load legacy icon: %v", err)
		}
	}
	if len(skin.Chromas) > 0 {
		chromaRes, err := LoadResourceFromURLWithCache(data.ChromaIconURL())
		if err == nil && chromaRes != nil {
			chromaIcon := canvas.NewImageFromResource(chromaRes)
			chromaIcon.SetMinSize(iconSize)
			chromaIcon.FillMode = canvas.ImageFillContain
			if len(topIcons) > 0 {
				topIcons = append(topIcons, layout.NewSpacer())
			}
			topIcons = append(topIcons, chromaIcon)
		} else {
			log.Printf("WARN: Failed load chroma icon: %v", err)
		}
	}
	topIconsContainer := container.NewHBox(topIcons...)
	contentLayout := container.NewBorder(container.NewPadded(topIconsContainer), bottomBar, nil, nil, imgContainer)
	tapButton := widget.NewButton("", func() { onSelect(skin) })
	return container.NewStack(contentLayout, tapButton)
}

// --- Tappable Card Helper --- (Remain the same)
type tappableCard struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	onTapped func()
}

func newTappableCard(content fyne.CanvasObject, onTap func()) *tappableCard {
	c := &tappableCard{content: content, onTapped: onTap}
	c.ExtendBaseWidget(c)
	return c
}
func (c *tappableCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}
func (c *tappableCard) Tapped(_ *fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped()
	}
}
func (c *tappableCard) Cursor() desktop.Cursor { return desktop.PointerCursor }

// --- Other Helpers --- (Remain the same)
func parseURL(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Error parsing URL '%s': %v", urlStr, err)
		errorURL, _ := url.Parse("https://example.com/invalid-url")
		return errorURL
	}
	return u
}
func newIconButton(iconRes fyne.Resource, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon("", iconRes, onTap)
	return btn
}
func NewTabButton(label string, icon fyne.Resource, tapped func()) fyne.CanvasObject {
	btnIcon := canvas.NewImageFromResource(icon)
	btnIcon.SetMinSize(fyne.NewSize(24, 24))
	btnIcon.FillMode = canvas.ImageFillContain
	btnLabel := widget.NewLabel(label)
	btnLabel.Alignment = fyne.TextAlignCenter
	btnLabel.TextStyle = fyne.TextStyle{Monospace: false}
	content := container.NewVBox(btnIcon, btnLabel)
	return newTappableCard(container.NewPadded(content), tapped)
}
