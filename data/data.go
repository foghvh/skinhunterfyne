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
	debugLogger     = log.New(os.Stdout, "[DEBUG] ", log.Ltime|log.Lshortfile)
	imageLogFile, _ = os.OpenFile("c:\\Users\\dev\\buenoche\\image_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) // Use user's path
	imageLogger     *log.Logger                                                                                           // Initialize in init or main
)

func init() {
	// Initialize imageLogger safely
	if imageLogFile != nil {
		imageLogger = log.New(imageLogFile, "[IMAGE] ", log.Ltime|log.Lshortfile)
	} else {
		// Fallback to stdout if file failed to open
		imageLogger = log.New(os.Stdout, "[IMAGE_ERR] ", log.Ltime|log.Lshortfile)
		imageLogger.Println("Failed to open image log file, logging to stdout.")
	}
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
	Rarity               string             `json:"rarityGemPath"`
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
func InitData() error {
	log.Println("Initializing data...")
	cacheMutex.Lock()
	if len(championListCache) > 0 && len(allSkinsMap) > 0 {
		cacheMutex.Unlock()
		log.Println("Data already initialized.")
		return nil
	}
	cacheMutex.Unlock()

	err := fetchCDragonVersion()
	if err != nil {
		log.Printf("WARN: Failed to fetch CDragon version: %v. Using 'latest'.", err)
		cDragonVersion = "latest"
	} else {
		log.Printf("Using CDragon version: %s", cDragonVersion)
	}

	champions, err := fetchChampionSummary()
	if err != nil {
		return fmt.Errorf("failed to fetch champion summary: %w", err)
	}
	allSkins, err := fetchSkinsJSON()
	if err != nil {
		return fmt.Errorf("failed to fetch skins JSON: %w", err)
	}

	cacheMutex.Lock()
	championListCache = champions
	allSkinsMap = allSkins
	cacheMutex.Unlock()

	log.Printf("Data initialized successfully: %d champions, %d skins.", len(championListCache), len(allSkinsMap))
	return nil
}

func fetchCDragonVersion() error { cDragonVersion = "latest"; return nil } // Hardcoded for now

func fetchChampionSummary() ([]ChampionSummary, error) {
	url := fmt.Sprintf("%s/v1/champion-summary.json", dataRootAPI)
	log.Printf("Fetching champion summary from: %s", url)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed for champion summary: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code for champion summary: %s", resp.Status)
	}
	var data []ChampionSummary
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode champion summary JSON: %w", err)
	}
	champs := make([]ChampionSummary, 0, len(data))
	for _, ch := range data {
		if ch.ID != -1 && ch.Name != "" && ch.Alias != "" {
			ch.Key = strings.ToLower(ch.Alias)
			ch.SquarePortraitPath = ensureLeadingSlash(ch.SquarePortraitPath)
			champs = append(champs, ch)
		} else {
			log.Printf("WARN: Skipping invalid champion entry: ID=%d, Name='%s'", ch.ID, ch.Name)
		}
	}
	sort.Slice(champs, func(i, j int) bool { return champs[i].Name < champs[j].Name })
	log.Printf("Successfully fetched and processed %d champions.", len(champs))
	return champs, nil
}

func FetchAllChampions() ([]ChampionSummary, error) {
	cacheMutex.RLock()
	if len(championListCache) > 0 {
		listCopy := make([]ChampionSummary, len(championListCache))
		copy(listCopy, championListCache)
		cacheMutex.RUnlock()
		return listCopy, nil
	}
	cacheMutex.RUnlock()
	log.Println("Champion cache empty, attempting re-initialization...")
	initErr := InitData()
	if initErr != nil {
		return nil, fmt.Errorf("champion cache empty and re-initialization failed: %w", initErr)
	}
	cacheMutex.RLock()
	if len(championListCache) == 0 {
		cacheMutex.RUnlock()
		return nil, fmt.Errorf("champion cache still empty after re-initialization")
	}
	listCopy := make([]ChampionSummary, len(championListCache))
	copy(listCopy, championListCache)
	cacheMutex.RUnlock()
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
		return nil, fmt.Errorf("bad status code for skins JSON: %s", resp.Status)
	}
	var data map[string]Skin
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
		s.ID = skinID
		s.IsBase = (skinID%1000 == 0)
		s.TilePath = ensureLeadingSlash(s.TilePath)
		s.SplashPath = ensureLeadingSlash(s.SplashPath)
		s.UncenteredSplashPath = ensureLeadingSlash(s.UncenteredSplashPath)
		s.LoadScreenPath = ensureLeadingSlash(s.LoadScreenPath)
		s.Rarity = ensureLeadingSlash(s.Rarity)
		if len(s.Chromas) > 0 {
			processedChromas := make([]Chroma, len(s.Chromas))
			for i, ch := range s.Chromas {
				ch.OriginSkinID = s.ID
				ch.ChromaPath = ensureLeadingSlash(ch.ChromaPath)
				processedChromas[i] = ch
			}
			s.Chromas = processedChromas
		}
		processedData[idStr] = s
	}
	log.Printf("Successfully fetched and processed %d skins.", len(processedData))
	return processedData, nil
}

func ensureLeadingSlash(path string) string {
	if path != "" && !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func GetAllSkinsMap() (map[string]Skin, error) {
	cacheMutex.RLock()
	if len(allSkinsMap) == 0 {
		cacheMutex.RUnlock()
		log.Printf("WARN: Skin map cache was empty, attempting re-init...")
		err := InitData()
		cacheMutex.RLock()
		if err != nil {
			return nil, fmt.Errorf("skin map cache empty and re-init failed: %w", err)
		}
		if len(allSkinsMap) == 0 {
			return nil, fmt.Errorf("skin map cache still empty after re-init")
		}
	}
	mapCopy := make(map[string]Skin, len(allSkinsMap))
	for k, v := range allSkinsMap {
		mapCopy[k] = v
	}
	cacheMutex.RUnlock()
	return mapCopy, nil
}

func GetSkinsForChampion(championID int) ([]Skin, error) {
	cacheMutex.RLock()
	if len(allSkinsMap) == 0 {
		cacheMutex.RUnlock()
		return nil, fmt.Errorf("skins map not initialized")
	}
	skins := make([]Skin, 0)
	for _, s := range allSkinsMap {
		if GetChampionIDFromSkinID(s.ID) == championID {
			processedChromas := make([]Chroma, len(s.Chromas))
			for i, ch := range s.Chromas {
				ch.OriginSkinID = s.ID
				processedChromas[i] = ch
			}
			s.Chromas = processedChromas
			skins = append(skins, s)
		}
	}
	cacheMutex.RUnlock()
	sort.Slice(skins, func(i, j int) bool { return skins[i].ID < skins[j].ID })
	return skins, nil
}

func GetChampionIDFromSkinID(skinID int) int {
	if skinID < 1000 {
		return -1
	}
	return skinID / 1000
}

func FetchChampionDetails(championID int) (*DetailedChampionData, error) {
	cacheMutex.RLock()
	cachedData, found := championDetailCache[championID]
	cacheMutex.RUnlock()
	if found {
		log.Printf("Cache hit for champion details: %d", championID)
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
		return nil, fmt.Errorf("champion details not found (404) for ID: %d", championID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code for champion details %d: %s", championID, resp.Status)
	}
	var details DetailedChampionData
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode champion details JSON for %d: %w", championID, err)
	}
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
				ch.OriginSkinID = s.ID
				ch.ChromaPath = ensureLeadingSlash(ch.ChromaPath)
				processedChromas[j] = ch
			}
			s.Chromas = processedChromas
		}
		processedSkins[i] = s
	}
	details.Skins = processedSkins
	details.SquarePortraitPath = ensureLeadingSlash(details.SquarePortraitPath)
	cacheMutex.Lock()
	championDetailCache[championID] = &details
	cacheMutex.Unlock()
	log.Printf("Successfully fetched and cached details for champion %d (%s)", championID, details.Name)
	return &details, nil
}

func GetSkinDetails(skinID int) (Skin, error) {
	cacheMutex.RLock()
	idStr := fmt.Sprintf("%d", skinID)
	cachedSkin, foundInMap := allSkinsMap[idStr]
	cacheMutex.RUnlock()
	if foundInMap {
		processedChromas := make([]Chroma, len(cachedSkin.Chromas))
		for i, ch := range cachedSkin.Chromas {
			ch.OriginSkinID = cachedSkin.ID
			processedChromas[i] = ch
		}
		cachedSkin.Chromas = processedChromas
		cachedSkin.IsBase = (cachedSkin.ID%1000 == 0)
		return cachedSkin, nil
	}
	log.Printf("WARN: Skin %d not found in initial map, attempting fetch via champion details.", skinID)
	championID := GetChampionIDFromSkinID(skinID)
	if championID <= 0 {
		return Skin{}, fmt.Errorf("invalid champion ID derived from skin ID %d", skinID)
	}
	details, err := FetchChampionDetails(championID)
	if err != nil {
		return Skin{}, fmt.Errorf("failed to fetch champion details for %d to find skin %d: %w", championID, skinID, err)
	}
	for _, s := range details.Skins {
		if s.ID == skinID {
			log.Printf("Found skin %d within details for champion %d (%s)", skinID, championID, details.Name)
			processedChromas := make([]Chroma, len(s.Chromas))
			for i, ch := range s.Chromas {
				ch.OriginSkinID = s.ID
				processedChromas[i] = ch
			}
			s.Chromas = processedChromas
			s.IsBase = (s.ID%1000 == 0)
			return s, nil
		}
	}
	return Skin{}, fmt.Errorf("skin ID %d not found in map or champion %d details", skinID, championID)
}

// --- URL Helper Functions ---
func Asset(path string) string {
	if path == "" {
		imageLogger.Printf("❌ Asset path is empty, returning placeholder.")
		return GetPlaceholderImageURL()
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	imageLogger.Printf("➡️ Input Asset Path: '%s'", path)
	if strings.HasPrefix(strings.ToLower(path), "http://") || strings.HasPrefix(strings.ToLower(path), "https://") {
		imageLogger.Printf("⚠️ Path looks like a full URL, returning as-is: '%s'", path)
		return path
	}
	if strings.HasPrefix(strings.ToLower(path), jsonAssetPathPrefix) {
		relativePath := path[len(jsonAssetPathPrefix):]
		lowerRelativePath := strings.ToLower(relativePath)
		correctedBase := strings.TrimSuffix(assetURLBase, "/")
		if !strings.HasPrefix(lowerRelativePath, "/") {
			lowerRelativePath = "/" + lowerRelativePath
		}
		resultURL := correctedBase + lowerRelativePath
		imageLogger.Printf("✅ Generated Asset URL: '%s' (Base: '%s', Rel: '%s')", resultURL, correctedBase, lowerRelativePath)
		return resultURL
	} else {
		imageLogger.Printf("⚠️ Asset path '%s' missing expected prefix '%s'. Attempting direct append fallback.", path, jsonAssetPathPrefix)
		lowerPath := strings.ToLower(path)
		if !strings.HasPrefix(lowerPath, "/") {
			lowerPath = "/" + lowerPath
		}
		resultURL := strings.TrimSuffix(assetURLBase, "/") + lowerPath
		imageLogger.Printf("⚠️ Fallback URL: '%s'", resultURL)
		return resultURL
	}
}

func GetChampionSquarePortraitURL(champ ChampionSummary) string {
	imageLogger.Printf("Getting champion portrait for: %s (Path: %s)", champ.Name, champ.SquarePortraitPath)
	return Asset(champ.SquarePortraitPath)
}
func GetSkinTileURL(skin Skin) string {
	path := skin.TilePath
	if path == "" && skin.LoadScreenPath != "" {
		imageLogger.Printf("Skin tile path empty for ID %d (%s), using LoadScreenPath fallback.", skin.ID, skin.Name)
		path = skin.LoadScreenPath
	} else if path == "" {
		imageLogger.Printf("❌ Skin tile path AND LoadScreenPath empty for ID %d (%s).", skin.ID, skin.Name)
		return GetPlaceholderImageURL()
	}
	imageLogger.Printf("Getting skin tile/loadscreen for ID %d (%s) (Path: %s)", skin.ID, skin.Name, path)
	return Asset(path)
}
func GetSkinSplashURL(skin Skin) string {
	path := skin.SplashPath
	if skin.UncenteredSplashPath != "" {
		imageLogger.Printf("Using uncentered splash for skin ID %d (%s)", skin.ID, skin.Name)
		path = skin.UncenteredSplashPath
	} else if path == "" {
		imageLogger.Printf("❌ Skin splash path AND UncenteredSplashPath empty for ID %d (%s).", skin.ID, skin.Name)
		fallbackPath := skin.TilePath
		if fallbackPath == "" {
			fallbackPath = skin.LoadScreenPath
		}
		if fallbackPath != "" {
			imageLogger.Printf("Falling back to Tile/LoadScreen path for splash: %s", fallbackPath)
			path = fallbackPath
		} else {
			return GetPlaceholderImageURL()
		}
	}
	imageLogger.Printf("Getting skin splash for ID %d (%s) (Path: %s)", skin.ID, skin.Name, path)
	return Asset(path)
}
func GetChromaImageURL(chroma Chroma) string {
	imageLogger.Printf("Getting chroma image for ID %d (%s) (Path: %s)", chroma.ID, chroma.Name, chroma.ChromaPath)
	return Asset(chroma.ChromaPath)
}

// --- Other Helper Functions ---
func KhadaUrl(skinID int, chromaID int) string {
	baseURL := "https://modelviewer.lol/model-viewer?id="
	if chromaID > 0 && skinID != chromaID && GetChampionIDFromSkinID(skinID) == GetChampionIDFromSkinID(chromaID) {
		return fmt.Sprintf("%s%d&chroma=%d", baseURL, skinID, chromaID)
	}
	return fmt.Sprintf("%s%d", baseURL, skinID)
}
func GetPlaceholderImageURL() string {
	return "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
}

var rarityMap = map[string][2]string{"raritygem_ultimate.png": {"Ultimate", "ultimate.png"}, "raritygem_mythic.png": {"Mythic", "mythic.png"}, "raritygem_legendary.png": {"Legendary", "legendary.png"}, "raritygem_epic.png": {"Epic", "epic.png"}, "raritygem_transcendent.png": {"Transcendent", "transcendent.png"}, "raritygem_exalted.png": {"Exalted", "exalted.png"}}

func Rarity(skin Skin) (string, string) {
	if skin.Rarity == "" {
		return "Standard", ""
	}
	lowerPath := strings.ToLower(skin.Rarity)
	for suffix, data := range rarityMap {
		if strings.HasSuffix(lowerPath, suffix) {
			displayName := data[0]
			iconFilename := data[1]
			iconPath := fmt.Sprintf("%s/v1/rarity-gem-icons/%s", jsonAssetPathPrefix, iconFilename)
			iconURL := Asset(iconPath)
			return displayName, iconURL
		}
	}
	log.Printf("WARN: Unknown rarity gem path format: %s", skin.Rarity)
	return "Unknown", ""
}
func LegacyIconURL() string {
	return fmt.Sprintf("%s/images/summoner-icon/icon-legacy.png", cDragonStaticAssets)
}
func ChromaIconURL() string {
	return fmt.Sprintf("%s/images/skin-viewer/icon-chroma-default.png", cDragonStaticAssets)
}
func ParseHexColor(s string) (color.NRGBA, error) {
	if s == "" {
		return color.NRGBA{128, 128, 128, 255}, fmt.Errorf("empty hex string")
	}
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	}
	var r, g, b, a uint64
	var err error
	a = 255
	switch len(s) {
	case 6:
		r, err = strconv.ParseUint(s[0:2], 16, 8)
		if err == nil {
			g, err = strconv.ParseUint(s[2:4], 16, 8)
		}
		if err == nil {
			b, err = strconv.ParseUint(s[4:6], 16, 8)
		}
	case 3:
		r, err = strconv.ParseUint(s[0:1]+s[0:1], 16, 8)
		if err == nil {
			g, err = strconv.ParseUint(s[1:2]+s[1:2], 16, 8)
		}
		if err == nil {
			b, err = strconv.ParseUint(s[2:3]+s[2:3], 16, 8)
		}
	case 8:
		r, err = strconv.ParseUint(s[0:2], 16, 8)
		if err == nil {
			g, err = strconv.ParseUint(s[2:4], 16, 8)
		}
		if err == nil {
			b, err = strconv.ParseUint(s[4:6], 16, 8)
		}
		if err == nil {
			a, err = strconv.ParseUint(s[6:8], 16, 8)
		}
	default:
		err = fmt.Errorf("invalid hex string length: %d", len(s))
	}
	if err != nil {
		log.Printf("WARN: Failed to parse hex color string '%s': %v", s, err)
		return color.NRGBA{128, 128, 128, 255}, err
	}
	return color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}, nil
}
