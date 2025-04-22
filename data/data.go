// skinhunter/data/data.go
package data

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	// NO FYNE IMPORTS HERE
)

const (
	cDragonBase = "https://raw.communitydragon.org/latest"
	// !! dataRoot should be used for API calls like skins.json, champions.json !!
	dataRootAPI         = cDragonBase + "/plugins/rcp-be-lol-game-data/global/default"
	cDragonStaticAssets = cDragonBase + "/plugins/rcp-fe-lol-static-assets/global/default"
	// !! Base URL for constructing final ASSET URLs from relative paths !!
	// Corrected base URL for assets based on observed working URLs
	assetURLBase = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default" // Explicitly set full base
	// Prefix to remove from JSON paths before converting to lower case and appending
	jsonAssetPathPrefix = "/lol-game-data/assets" // The prefix from the JSON data itself
)

var (
	httpClient          = &http.Client{Timeout: time.Second * 15} // Shared client
	championListCache   []ChampionSummary
	championDetailCache = make(map[int]*DetailedChampionData) // Use Champion ID (int) as key
	skinLinesCache      []SkinLine                            // Currently unused but defined
	cacheMutex          sync.RWMutex
	cDragonVersion      string          // Store the fetched version
	allSkinsMap         map[string]Skin // Map skin ID (string) to Skin struct
)

// Define loggers (keep user's setup)
var (
	debugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ltime|log.Lshortfile)
	// Ensure this path is correct for your system or make it configurable
	imageLogFile, logOpenErr = os.OpenFile("c:\\Users\\dev\\buenoche\\image_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	imageLogger              *log.Logger
)

func init() {
	// Initialize imageLogger safely
	if logOpenErr != nil {
		// Fallback to stdout if file failed to open
		fmt.Fprintf(os.Stderr, "[IMAGE_INIT_ERR] Failed to open image log file 'c:\\Users\\dev\\buenoche\\image_debug.log': %v\n", logOpenErr)
		imageLogger = log.New(os.Stdout, "[IMAGE_ERR] ", log.Ltime|log.Lshortfile)
	} else if imageLogFile != nil {
		imageLogger = log.New(imageLogFile, "[IMAGE] ", log.Ltime|log.Lshortfile)
	} else {
		// Should not happen if logOpenErr is nil, but defensive coding
		imageLogger = log.New(os.Stdout, "[IMAGE_WARN] ", log.Ltime|log.Lshortfile)
		imageLogger.Println("Image log file was nil despite no open error.")
	}
	imageLogger.Println("--- Image Logger Initialized ---") // Add a marker for new runs
}

// --- Structs ---
type ChampionSummary struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Alias              string   `json:"alias"`
	SquarePortraitPath string   `json:"squarePortraitPath"`
	Roles              []string `json:"roles"`
	Key                string   `json:"key"`
}
type SkinLine struct {
	ID   int
	Name string
}
type Skin struct {
	ID                   int                `json:"id"`
	Name                 string             `json:"name"`
	TilePath             string             `json:"tilePath"`
	SplashPath           string             `json:"splashPath"`
	UncenteredSplashPath string             `json:"uncenteredSplashPath"`
	LoadScreenPath       string             `json:"loadScreenPath"`
	Description          string             `json:"description"`
	Rarity               string             `json:"rarityGemPath"` // Keep original JSON path
	IsLegacy             bool               `json:"isLegacy"`
	IsBase               bool               // Calculated
	Chromas              []Chroma           `json:"chromas"`
	SkinLines            []struct{ ID int } `json:"skinLines"`
}
type DetailedChampionData struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Alias              string   `json:"alias"`
	Title              string   `json:"title"`
	ShortBio           string   `json:"shortBio"`
	SquarePortraitPath string   `json:"squarePortraitPath"`
	Roles              []string `json:"roles"`
	Skins              []Skin   `json:"skins"`
}
type Chroma struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	ChromaPath   string   `json:"chromaPath"`
	Colors       []string `json:"colors"`
	OriginSkinID int      `json:"-"` // Calculated
}

// --- Initialization and Caching ---
// InitData should ideally be called once at application startup.
func InitData() error {
	log.Println("Initializing data...")
	cacheMutex.RLock()
	alreadyInitialized := len(championListCache) > 0 && len(allSkinsMap) > 0
	cacheMutex.RUnlock()

	if alreadyInitialized {
		log.Println("Data already initialized.")
		return nil
	}

	// Use a single lock for the whole initialization process
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double check after acquiring write lock
	if len(championListCache) > 0 && len(allSkinsMap) > 0 {
		log.Println("Data already initialized (double check).")
		return nil
	}

	err := fetchCDragonVersion() // Currently hardcoded to latest
	if err != nil {
		log.Printf("WARN: Failed to fetch CDragon version: %v. Using 'latest'.", err)
		cDragonVersion = "latest" // Ensure it's set even on error
	} else {
		log.Printf("Using CDragon version: %s", cDragonVersion)
	}

	champions, err := fetchChampionSummary()
	if err != nil {
		log.Printf("ERROR: Failed to fetch champion summary during InitData: %v", err)
		return fmt.Errorf("failed to fetch champion summary: %w", err)
	}
	allSkins, err := fetchSkinsJSON()
	if err != nil {
		log.Printf("ERROR: Failed to fetch skins JSON during InitData: %v", err)
		return fmt.Errorf("failed to fetch skins JSON: %w", err)
	}

	// Assign to cache ONLY if successful
	championListCache = champions
	allSkinsMap = allSkins

	log.Printf("Data initialized successfully: %d champions, %d skins.", len(championListCache), len(allSkinsMap))
	return nil
}

// fetchCDragonVersion can be expanded later if needed
func fetchCDragonVersion() error { cDragonVersion = "latest"; return nil }

func fetchChampionSummary() ([]ChampionSummary, error) {
	url := fmt.Sprintf("%s/v1/champion-summary.json", dataRootAPI)
	log.Printf("Fetching champion summary from: %s", url)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed for champion summary: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code for champion summary: %s (%d)", resp.Status, resp.StatusCode)
	}
	var data []ChampionSummary
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode champion summary JSON: %w", err)
	}
	champs := make([]ChampionSummary, 0, len(data))
	for _, ch := range data {
		// Basic validation: ID should not be -1, Name and Alias should not be empty
		if ch.ID != -1 && ch.Name != "" && ch.Alias != "" {
			ch.Key = strings.ToLower(ch.Alias)
			// Ensure path consistency (useful for Asset function)
			ch.SquarePortraitPath = ensureLeadingSlash(ch.SquarePortraitPath)
			champs = append(champs, ch)
		} else {
			log.Printf("WARN: Skipping invalid champion entry: ID=%d, Name='%s', Alias='%s'", ch.ID, ch.Name, ch.Alias)
		}
	}
	sort.Slice(champs, func(i, j int) bool { return champs[i].Name < champs[j].Name })
	log.Printf("Successfully fetched and processed %d champions.", len(champs))
	return champs, nil
}

// FetchAllChampions retrieves the cached champion list.
// It will trigger InitData if the cache is empty.
func FetchAllChampions() ([]ChampionSummary, error) {
	cacheMutex.RLock()
	if len(championListCache) > 0 {
		listCopy := make([]ChampionSummary, len(championListCache))
		copy(listCopy, championListCache)
		cacheMutex.RUnlock()
		return listCopy, nil
	}
	cacheMutex.RUnlock()

	log.Println("Champion cache empty, triggering InitData...")
	// InitData handles its own locking
	initErr := InitData()
	if initErr != nil {
		log.Printf("ERROR: Failed to initialize data while fetching champions: %v", initErr)
		return nil, fmt.Errorf("champion cache empty and initialization failed: %w", initErr)
	}

	// Try reading again after initialization
	cacheMutex.RLock()
	if len(championListCache) == 0 {
		cacheMutex.RUnlock()
		log.Println("ERROR: Champion cache still empty after successful InitData call.")
		return nil, fmt.Errorf("champion cache still empty after initialization attempt")
	}
	listCopy := make([]ChampionSummary, len(championListCache))
	copy(listCopy, championListCache)
	cacheMutex.RUnlock()
	log.Printf("Returning %d champions after initialization.", len(listCopy))
	return listCopy, nil
}

func fetchSkinsJSON() (map[string]Skin, error) {
	url := fmt.Sprintf("%s/v1/skins.json", dataRootAPI)
	log.Printf("Fetching skins JSON from: %s", url)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed for skins JSON: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code for skins JSON: %s (%d)", resp.Status, resp.StatusCode)
	}

	var data map[string]Skin // Directly decode into map[string]Skin
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode skins JSON: %w", err)
	}

	processedData := make(map[string]Skin, len(data))
	for idStr, s := range data {
		skinID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("WARN: Skipping skin with invalid ID '%s': %v", idStr, err)
			continue
		}

		// Basic validation for essential fields
		if s.Name == "" {
			log.Printf("WARN: Skipping skin ID %d due to empty name.", skinID)
			continue
		}
		// TilePath or LoadScreenPath is usually required for display
		if s.TilePath == "" && s.LoadScreenPath == "" {
			log.Printf("WARN: Skin ID %d ('%s') has no TilePath or LoadScreenPath.", skinID, s.Name)
			// Decide if you want to skip or allow it
			// continue
		}

		s.ID = skinID
		s.IsBase = (skinID%1000 == 0)

		// Ensure paths have leading slashes if not empty
		s.TilePath = ensureLeadingSlash(s.TilePath)
		s.SplashPath = ensureLeadingSlash(s.SplashPath)
		s.UncenteredSplashPath = ensureLeadingSlash(s.UncenteredSplashPath)
		s.LoadScreenPath = ensureLeadingSlash(s.LoadScreenPath)
		s.Rarity = ensureLeadingSlash(s.Rarity) // Keep original path from JSON here

		// Process chromas, ensuring OriginSkinID and paths are set correctly
		if len(s.Chromas) > 0 {
			processedChromas := make([]Chroma, 0, len(s.Chromas)) // Avoid modifying original slice directly
			for _, ch := range s.Chromas {
				if ch.ID == 0 { // Basic validation for chroma ID
					log.Printf("WARN: Skipping chroma with ID 0 for skin %d (%s)", s.ID, s.Name)
					continue
				}
				ch.OriginSkinID = s.ID
				ch.ChromaPath = ensureLeadingSlash(ch.ChromaPath)
				processedChromas = append(processedChromas, ch)
			}
			s.Chromas = processedChromas
		}
		processedData[idStr] = s
	}
	log.Printf("Successfully fetched and processed %d skins.", len(processedData))
	return processedData, nil
}

// ensureLeadingSlash adds a leading slash if the path is not empty and doesn't have one.
func ensureLeadingSlash(path string) string {
	if path != "" && !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// GetAllSkinsMap retrieves the cached map of all skins.
// It will attempt to initialize data if the map is empty.
func GetAllSkinsMap() (map[string]Skin, error) {
	cacheMutex.RLock()
	if len(allSkinsMap) > 0 {
		// Return a shallow copy to prevent external modification of the cache
		mapCopy := make(map[string]Skin, len(allSkinsMap))
		for k, v := range allSkinsMap {
			mapCopy[k] = v
		}
		cacheMutex.RUnlock()
		return mapCopy, nil
	}
	cacheMutex.RUnlock()

	log.Printf("WARN: Skin map cache was empty, attempting InitData...")
	err := InitData() // InitData handles its own locking
	if err != nil {
		return nil, fmt.Errorf("skin map cache empty and initialization failed: %w", err)
	}

	// Try reading again
	cacheMutex.RLock()
	if len(allSkinsMap) == 0 {
		cacheMutex.RUnlock()
		return nil, fmt.Errorf("skin map cache still empty after initialization attempt")
	}
	mapCopy := make(map[string]Skin, len(allSkinsMap))
	for k, v := range allSkinsMap {
		mapCopy[k] = v
	}
	cacheMutex.RUnlock()
	return mapCopy, nil
}

// GetSkinsForChampion retrieves skins for a specific champion ID from the cached map.
func GetSkinsForChampion(championID int) ([]Skin, error) {
	cacheMutex.RLock()
	// Check if the map is initialized before proceeding
	if len(allSkinsMap) == 0 {
		cacheMutex.RUnlock()
		// Don't try to InitData here, let FetchAllChampions/GetAllSkinsMap handle it.
		// This function assumes the cache should be populated by the time it's called.
		log.Printf("ERROR: GetSkinsForChampion called but allSkinsMap is not initialized.")
		return nil, fmt.Errorf("skins map not initialized")
	}

	skins := make([]Skin, 0)
	for _, s := range allSkinsMap {
		if GetChampionIDFromSkinID(s.ID) == championID {
			// Ensure chromas have the correct OriginSkinID when retrieved
			// Create a copy to avoid modifying the cached version if necessary,
			// though current implementation seems safe.
			skinCopy := s
			if len(skinCopy.Chromas) > 0 {
				processedChromas := make([]Chroma, len(skinCopy.Chromas))
				for i, ch := range skinCopy.Chromas {
					processedChromas[i] = ch                       // Copy struct
					processedChromas[i].OriginSkinID = skinCopy.ID // Ensure it's set
				}
				skinCopy.Chromas = processedChromas
			}
			skins = append(skins, skinCopy)
		}
	}
	cacheMutex.RUnlock() // Unlock before sorting

	// Sort skins by their ID
	sort.Slice(skins, func(i, j int) bool { return skins[i].ID < skins[j].ID })
	return skins, nil
}

// GetChampionIDFromSkinID derives the champion ID from a skin ID.
func GetChampionIDFromSkinID(skinID int) int {
	if skinID < 1000 { // Base skins (e.g., Annie 1) or invalid IDs
		return -1 // Or handle specific cases like champion ID 1 if needed
	}
	return skinID / 1000 // Integer division gives the champion ID
}

// FetchChampionDetails fetches detailed data for a specific champion, using cache first.
func FetchChampionDetails(championID int) (*DetailedChampionData, error) {
	cacheMutex.RLock()
	cachedData, found := championDetailCache[championID]
	cacheMutex.RUnlock()

	if found {
		log.Printf("Cache hit for champion details: %d", championID)
		// Return a copy or ensure the cached data is not modified externally if necessary
		// For now, returning the pointer assumes read-only usage or careful handling.
		return cachedData, nil
	}

	log.Printf("Cache miss, fetching details for champion: %d", championID)
	url := fmt.Sprintf("%s/v1/champions/%d.json", dataRootAPI, championID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed for champion details %d: %w", championID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("ERROR: Champion details not found (404) for ID: %d", championID)
		return nil, fmt.Errorf("champion details not found (404) for ID: %d", championID)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Bad status code for champion details %d: %s (%d)", championID, resp.Status, resp.StatusCode)
		return nil, fmt.Errorf("bad status code for champion details %d: %s", championID, resp.Status)
	}

	var details DetailedChampionData
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode champion details JSON for %d: %w", championID, err)
	}

	// Post-process fetched details before caching
	details.SquarePortraitPath = ensureLeadingSlash(details.SquarePortraitPath)
	if len(details.Skins) > 0 {
		processedSkins := make([]Skin, len(details.Skins))
		for i, s := range details.Skins {
			s.IsBase = (s.ID%1000 == 0)
			s.TilePath = ensureLeadingSlash(s.TilePath)
			s.SplashPath = ensureLeadingSlash(s.SplashPath)
			s.UncenteredSplashPath = ensureLeadingSlash(s.UncenteredSplashPath)
			s.LoadScreenPath = ensureLeadingSlash(s.LoadScreenPath)
			s.Rarity = ensureLeadingSlash(s.Rarity)

			if len(s.Chromas) > 0 {
				processedChromas := make([]Chroma, len(s.Chromas))
				for j, ch := range s.Chromas {
					ch.OriginSkinID = s.ID // Ensure OriginSkinID is set correctly
					ch.ChromaPath = ensureLeadingSlash(ch.ChromaPath)
					processedChromas[j] = ch
				}
				s.Chromas = processedChromas
			}
			processedSkins[i] = s
		}
		details.Skins = processedSkins
	}

	// Cache the processed details
	cacheMutex.Lock()
	championDetailCache[championID] = &details
	cacheMutex.Unlock()

	log.Printf("Successfully fetched and cached details for champion %d (%s)", championID, details.Name)
	return &details, nil // Return the pointer to the newly cached data
}

// GetSkinDetails retrieves details for a single skin, trying the map first, then champion details.
func GetSkinDetails(skinID int) (Skin, error) {
	idStr := fmt.Sprintf("%d", skinID)

	cacheMutex.RLock()
	cachedSkin, foundInMap := allSkinsMap[idStr]
	cacheMutex.RUnlock() // Unlock before potentially fetching details

	if foundInMap {
		// Ensure chromas have the correct OriginSkinID when retrieved
		// Create a copy to avoid modifying the cached version
		skinCopy := cachedSkin
		skinCopy.IsBase = (skinCopy.ID%1000 == 0) // Recalculate just in case
		if len(skinCopy.Chromas) > 0 {
			processedChromas := make([]Chroma, len(skinCopy.Chromas))
			for i, ch := range skinCopy.Chromas {
				processedChromas[i] = ch                       // Copy struct
				processedChromas[i].OriginSkinID = skinCopy.ID // Ensure it's set
			}
			skinCopy.Chromas = processedChromas
		}
		return skinCopy, nil
	}

	// Skin not found in the main map, try fetching via champion details
	log.Printf("WARN: Skin %d not found in initial map, attempting fetch via champion details.", skinID)
	championID := GetChampionIDFromSkinID(skinID)
	if championID <= 0 {
		return Skin{}, fmt.Errorf("invalid champion ID (%d) derived from skin ID %d", championID, skinID)
	}

	// FetchChampionDetails handles caching internally
	details, err := FetchChampionDetails(championID)
	if err != nil {
		// Check if the error was specifically a 404 for the champion
		if strings.Contains(err.Error(), "404") {
			return Skin{}, fmt.Errorf("champion %d not found while searching for skin %d", championID, skinID)
		}
		return Skin{}, fmt.Errorf("failed to fetch champion details for %d to find skin %d: %w", championID, skinID, err)
	}

	// Search for the specific skin within the fetched champion details
	for _, s := range details.Skins {
		if s.ID == skinID {
			log.Printf("Found skin %d within details for champion %d (%s)", skinID, championID, details.Name)
			// The skins within 'details' should already be processed by FetchChampionDetails
			// Return a copy of the found skin
			skinCopy := s
			skinCopy.IsBase = (skinCopy.ID%1000 == 0) // Recalculate just in case
			// Chromas should already have OriginSkinID set correctly by FetchChampionDetails
			return skinCopy, nil
		}
	}

	// If the skin wasn't found even after fetching details
	log.Printf("ERROR: Skin ID %d not found in map or in details for champion %d (%s)", skinID, championID, details.Name)
	return Skin{}, fmt.Errorf("skin ID %d not found in map or champion %d details", skinID, championID)
}

// --- URL Helper Functions ---

// Asset constructs the absolute URL for a given asset path from the JSON data.
// Input `path`: Raw path from JSON (e.g., "/lol-game-data/assets/v1/...")
// It expects paths to start with jsonAssetPathPrefix.
func Asset(path string) string {
	if path == "" {
		imageLogger.Printf("❌ Asset path is empty, returning placeholder.")
		return GetPlaceholderImageURL()
	}

	// Ensure path starts with a slash for consistent processing
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
		imageLogger.Printf("➡️ Input Asset Path (slash added): '%s'", path)
	} else {
		imageLogger.Printf("➡️ Input Asset Path: '%s'", path)
	}

	// Check if it's already a full URL
	lowerPath := strings.ToLower(path)
	if strings.HasPrefix(lowerPath, "http://") || strings.HasPrefix(lowerPath, "https://") {
		imageLogger.Printf("⚠️ Path looks like a full URL, returning as-is: '%s'", path)
		return path // Return the full URL directly
	}

	// Check if the path starts with the expected prefix
	lowerJsonAssetPathPrefix := strings.ToLower(jsonAssetPathPrefix)
	if strings.HasPrefix(lowerPath, lowerJsonAssetPathPrefix) {
		// Extract the relative path *after* the prefix
		// Use original path casing for the relative part initially
		relativePath := path[len(jsonAssetPathPrefix):]

		// Convert the relative path part to lowercase for the final URL construction
		lowerRelativePath := strings.ToLower(relativePath)

		// Ensure the base URL doesn't have a trailing slash
		correctedBase := strings.TrimSuffix(assetURLBase, "/")

		// Ensure the relative path starts with a slash
		if !strings.HasPrefix(lowerRelativePath, "/") {
			lowerRelativePath = "/" + lowerRelativePath
		}

		resultURL := correctedBase + lowerRelativePath
		imageLogger.Printf("✅ Generated Asset URL: '%s' (Base: '%s', Rel Orig Case: '%s', Rel Lower Case: '%s')", resultURL, correctedBase, relativePath, lowerRelativePath)
		return resultURL
	} else {
		// Path doesn't start with the expected prefix. Log a warning and try a fallback.
		imageLogger.Printf("⚠️ Asset path '%s' missing expected prefix '%s'. Attempting direct append fallback.", path, jsonAssetPathPrefix)

		// Convert the whole path to lowercase for the fallback URL
		lowerPathForFallback := strings.ToLower(path)
		// Ensure it starts with a slash
		if !strings.HasPrefix(lowerPathForFallback, "/") {
			lowerPathForFallback = "/" + lowerPathForFallback
		}

		correctedBase := strings.TrimSuffix(assetURLBase, "/")
		resultURL := correctedBase + lowerPathForFallback // Append the lowercased original path

		imageLogger.Printf("⚠️ Fallback URL (Appended Lower Case Path): '%s'", resultURL)
		return resultURL
	}
}

// GetChampionSquarePortraitURL uses Asset to get the champion portrait URL.
func GetChampionSquarePortraitURL(champ ChampionSummary) string {
	imageLogger.Printf("Getting champion portrait for: %s (Path: %s)", champ.Name, champ.SquarePortraitPath)
	return Asset(champ.SquarePortraitPath)
}

// GetSkinTileURL uses Asset to get the skin tile URL, falling back to LoadScreenPath.
func GetSkinTileURL(skin Skin) string {
	path := skin.TilePath
	source := "TilePath"
	if path == "" && skin.LoadScreenPath != "" {
		imageLogger.Printf("Skin tile path empty for ID %d (%s), using LoadScreenPath fallback.", skin.ID, skin.Name)
		path = skin.LoadScreenPath
		source = "LoadScreenPath (Fallback)"
	} else if path == "" {
		imageLogger.Printf("❌ Skin tile path AND LoadScreenPath empty for ID %d (%s). Returning placeholder.", skin.ID, skin.Name)
		return GetPlaceholderImageURL() // Return placeholder if both are empty
	}
	imageLogger.Printf("Getting skin tile/loadscreen for ID %d (%s) (Source: %s, Path: %s)", skin.ID, skin.Name, source, path)
	return Asset(path)
}

// GetSkinSplashURL uses Asset to get the skin splash URL, preferring UncenteredSplashPath.
func GetSkinSplashURL(skin Skin) string {
	path := skin.SplashPath
	source := "SplashPath"
	if skin.UncenteredSplashPath != "" {
		imageLogger.Printf("Using uncentered splash for skin ID %d (%s)", skin.ID, skin.Name)
		path = skin.UncenteredSplashPath
		source = "UncenteredSplashPath"
	} else if path == "" {
		// Fallback if both SplashPath and UncenteredSplashPath are empty
		imageLogger.Printf("❌ Skin splash path AND UncenteredSplashPath empty for ID %d (%s).", skin.ID, skin.Name)
		// Try falling back to TilePath or LoadScreenPath as a last resort for *some* image
		fallbackPath := skin.TilePath
		fallbackSource := "TilePath (Splash Fallback)"
		if fallbackPath == "" {
			fallbackPath = skin.LoadScreenPath
			fallbackSource = "LoadScreenPath (Splash Fallback)"
		}
		if fallbackPath != "" {
			imageLogger.Printf("⚠️ Falling back to %s for splash: %s", fallbackSource, fallbackPath)
			path = fallbackPath
			source = fallbackSource
		} else {
			imageLogger.Printf("❌ No fallback paths available for splash for skin ID %d (%s). Returning placeholder.", skin.ID, skin.Name)
			return GetPlaceholderImageURL() // Return placeholder if absolutely no path is found
		}
	}
	imageLogger.Printf("Getting skin splash for ID %d (%s) (Source: %s, Path: %s)", skin.ID, skin.Name, source, path)
	return Asset(path)
}

// GetChromaImageURL uses Asset to get the chroma image URL.
func GetChromaImageURL(chroma Chroma) string {
	imageLogger.Printf("Getting chroma image for ID %d (%s) (Origin Skin: %d, Path: %s)", chroma.ID, chroma.Name, chroma.OriginSkinID, chroma.ChromaPath)
	return Asset(chroma.ChromaPath)
}

// --- Other Helper Functions ---

// KhadaUrl generates a URL for the LoL Model Viewer.
func KhadaUrl(skinID int, chromaID int) string {
	baseURL := "https://modelviewer.lol/model-viewer?id="
	// Check if chromaID is valid (not 0) and different from skinID,
	// and if they belong to the same champion (basic check).
	if chromaID > 0 && skinID != chromaID {
		// Add a check to ensure the chroma actually belongs to the skin
		// This requires fetching skin details, which might be too slow here.
		// A simpler check is based on champion ID derivation.
		if GetChampionIDFromSkinID(skinID) == GetChampionIDFromSkinID(chromaID) {
			return fmt.Sprintf("%s%d&chroma=%d", baseURL, skinID, chromaID)
		} else {
			log.Printf("WARN: KhadaUrl called with skin %d and chroma %d from different champions?", skinID, chromaID)
			// Fallback to just the skin ID if champion IDs don't match
		}
	}
	// Default to just the skin ID if no valid chroma or if chroma is the same as skin ID
	return fmt.Sprintf("%s%d", baseURL, skinID)
}

// GetPlaceholderImageURL returns a transparent 1x1 pixel GIF data URI.
func GetPlaceholderImageURL() string {
	// This is a standard 1x1 transparent GIF
	return "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
}

// rarityMap maps the suffix of the Rarity path to a display name and icon filename.
// Keys are the EXPECTED suffixes in lowercase.
var rarityMap = map[string][2]string{
	"raritygem_ultimate.png":     {"Ultimate", "ultimate.png"},
	"raritygem_mythic.png":       {"Mythic", "mythic.png"},
	"raritygem_legendary.png":    {"Legendary", "legendary.png"},
	"raritygem_epic.png":         {"Epic", "epic.png"},
	"raritygem_transcendent.png": {"Transcendent", "transcendent.png"}, // Check if this is still used
	"raritygem_exalted.png":      {"Exalted", "exalted.png"},           // Check if this is still used
	// Add default/standard if needed, though empty Rarity usually implies Standard
}

// Rarity determines the display name and icon URL for a skin's rarity.
func Rarity(skin Skin) (string, string) {
	if skin.Rarity == "" {
		// Consider "Standard" or "Common" for skins without a specific rarity gem path
		return "Standard", "" // No icon for standard rarity
	}

	lowerPath := strings.ToLower(skin.Rarity)

	for suffix, data := range rarityMap {
		if strings.HasSuffix(lowerPath, suffix) {
			displayName := data[0]
			iconFilename := data[1]
			// Construct the *relative* path for the icon asset
			iconAssetPath := fmt.Sprintf("%s/v1/rarity-gem-icons/%s", jsonAssetPathPrefix, iconFilename)
			// Use the Asset function to get the full URL
			iconURL := Asset(iconAssetPath)
			imageLogger.Printf("Rarity found for '%s': Name='%s', IconPath='%s', IconURL='%s'", skin.Rarity, displayName, iconAssetPath, iconURL)
			return displayName, iconURL
		}
	}

	// If no known suffix matches, log a warning and return a default
	log.Printf("WARN: Unknown rarity gem path format: '%s' for skin %d (%s)", skin.Rarity, skin.ID, skin.Name)
	// You might want a specific "Unknown" category or default to "Standard"
	return "Unknown", "" // Or return "Standard", ""
}

// LegacyIconURL provides the URL for the legacy content icon.
func LegacyIconURL() string {
	// This path seems to be from a different plugin (static assets)
	return fmt.Sprintf("%s/images/summoner-icon/icon-legacy.png", cDragonStaticAssets)
}

// ChromaIconURL provides the URL for the default chroma availability icon.
func ChromaIconURL() string {
	// This path also seems to be from static assets
	return fmt.Sprintf("%s/images/skin-viewer/icon-chroma-default.png", cDragonStaticAssets)
}

// ParseHexColor parses a hex color string (e.g., #RRGGBB, #RGB, RRGGBB, RGB, RRGGBBAA) into a color.NRGBA.
func ParseHexColor(s string) (color.NRGBA, error) {
	if s == "" {
		// Return a default (e.g., grey) and an error for empty strings
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}, fmt.Errorf("empty hex string")
	}

	// Remove leading '#' if present
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	}

	var r, g, b, a uint64
	var err error
	a = 255 // Default alpha to fully opaque

	switch len(s) {
	case 6: // RRGGBB
		r, err = strconv.ParseUint(s[0:2], 16, 8)
		if err == nil {
			g, err = strconv.ParseUint(s[2:4], 16, 8)
		}
		if err == nil {
			b, err = strconv.ParseUint(s[4:6], 16, 8)
		}
	case 3: // RGB
		// Expand to RRGGBB
		r, err = strconv.ParseUint(s[0:1]+s[0:1], 16, 8)
		if err == nil {
			g, err = strconv.ParseUint(s[1:2]+s[1:2], 16, 8)
		}
		if err == nil {
			b, err = strconv.ParseUint(s[2:3]+s[2:3], 16, 8)
		}
	case 8: // RRGGBBAA
		r, err = strconv.ParseUint(s[0:2], 16, 8)
		if err == nil {
			g, err = strconv.ParseUint(s[2:4], 16, 8)
		}
		if err == nil {
			b, err = strconv.ParseUint(s[4:6], 16, 8)
		}
		if err == nil {
			a, err = strconv.ParseUint(s[6:8], 16, 8) // Parse alpha
		}
	default:
		err = fmt.Errorf("invalid hex string length: %d (expected 3, 6, or 8)", len(s))
	}

	if err != nil {
		log.Printf("WARN: Failed to parse hex color string '%s': %v", s, err)
		// Return a default color (e.g., grey) and the error
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}, err
	}

	// If parsing was successful, return the NRGBA color
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}, nil
}
