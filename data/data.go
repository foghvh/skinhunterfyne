// skinhunter/data/data.go
package data

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cDragonBase         = "https://raw.communitydragon.org/latest"
	cDragonDataBE       = cDragonBase + "/plugins/rcp-be-lol-game-data/global/default"
	cDragonStaticAssets = cDragonBase + "/plugins/rcp-fe-lol-static-assets/global/default"
	cDragonAssetBase    = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default" // Used by asset()
)

var (
	httpClient          = &http.Client{Timeout: time.Second * 15} // Shared client
	championListCache   []ChampionSummary
	championDetailCache = make(map[int]*DetailedChampionData) // Use Champion ID (int) as key
	skinLinesCache      []SkinLine
	cacheMutex          sync.RWMutex
	cDragonVersion      string
	allSkinsMap         map[string]Skin // Map skin ID (string) to Skin struct
)

// --- Structs --- (Remain the same as previous correct version)
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
	Description          string             `json:"description"`
	Rarity               string             `json:"rarityGemPath"`
	IsLegacy             bool               `json:"isLegacy"`
	IsBase               bool               `json:"isBase"`
	Chromas              []Chroma           `json:"chromas"`
	SkinLines            []struct{ ID int } `json:"skinLines"`
	LoadScreenPath       string             `json:"loadScreenPath"`
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
	OriginSkinID int      `json:"-"`
}

// --- Initialization and Caching --- (Remain the same)
func InitData() error {
	log.Println("Initializing data...")
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if len(championListCache) > 0 && len(allSkinsMap) > 0 {
		log.Println("Data already initialized.")
		return nil
	}
	err := fetchCDragonVersion()
	if err != nil {
		log.Printf("WARN: Could not fetch CDragon version: %v. Using 'latest'.", err)
		cDragonVersion = "latest"
	} else {
		log.Printf("Using CDragon version: %s", cDragonVersion)
	}
	champions, err := fetchChampionSummary()
	if err != nil {
		return fmt.Errorf("failed to initialize champions: %w", err)
	}
	championListCache = champions
	log.Printf("Initialized %d champions.", len(championListCache))
	allSkins, err := fetchSkinsJSON()
	if err != nil {
		return fmt.Errorf("failed to initialize skins: %w", err)
	}
	allSkinsMap = allSkins
	log.Printf("Initialized %d skins.", len(allSkinsMap))
	log.Println("Data initialization complete.")
	return nil
}

func fetchCDragonVersion() error {
	url := fmt.Sprintf("%s/content-metadata.json", cDragonBase)
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status fetching version: %s", resp.Status)
	}
	var metadata struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return err
	}
	if metadata.Version == "" {
		return fmt.Errorf("version field missing in content-metadata.json")
	}
	cDragonVersion = metadata.Version
	return nil
}

func fetchChampionSummary() ([]ChampionSummary, error) {
	url := fmt.Sprintf("%s/v1/champion-summary.json", cDragonDataBE)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed http GET champion summary: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status fetching champion summary: %s", resp.Status)
	}
	var data []ChampionSummary
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode champion summary JSON: %w", err)
	}
	champions := make([]ChampionSummary, 0, len(data))
	for _, champ := range data {
		if champ.ID != -1 {
			champ.Key = strings.ToLower(champ.Alias)
			champions = append(champions, champ)
		}
	}
	sort.Slice(champions, func(i, j int) bool { return champions[i].Name < champions[j].Name })
	return champions, nil
}

func FetchAllChampions() ([]ChampionSummary, error) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	if len(championListCache) == 0 {
		log.Println("Champion list cache empty, attempting initialization...")
		cacheMutex.RUnlock()
		initErr := InitData()
		cacheMutex.RLock()
		if initErr != nil {
			return nil, fmt.Errorf("champion list not initialized and init failed: %w", initErr)
		}
		if len(championListCache) == 0 {
			return nil, fmt.Errorf("champion list still empty after initialization attempt")
		}
	}
	// log.Println("Serving champions from cache") // Less noisy
	return championListCache, nil
}

func fetchSkinsJSON() (map[string]Skin, error) {
	url := fmt.Sprintf("%s/v1/skins.json", cDragonDataBE)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed http GET skins.json: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status fetching skins.json: %s", resp.Status)
	}
	var data map[string]Skin
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode skins.json: %w", err)
	}
	processedData := make(map[string]Skin, len(data))
	for idStr, skin := range data {
		skinID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("WARN: Could not parse skin ID string '%s' from skins.json: %v", idStr, err)
			continue
		}
		skin.ID = skinID
		skin.IsBase = (skinID%1000 == 0)
		for i := range skin.Chromas {
			skin.Chromas[i].OriginSkinID = skin.ID
		}
		processedData[idStr] = skin
	}
	return processedData, nil
}

func GetAllSkinsMap() (map[string]Skin, error) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	if len(allSkinsMap) == 0 {
		return nil, fmt.Errorf("skins map not initialized, call data.InitData() first")
	}
	return allSkinsMap, nil
}

func GetSkinsForChampion(championID int) ([]Skin, error) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	if len(allSkinsMap) == 0 {
		return nil, fmt.Errorf("skins map not initialized")
	}
	skins := make([]Skin, 0)
	for _, skin := range allSkinsMap {
		if GetChampionIDFromSkinID(skin.ID) == championID {
			for i := range skin.Chromas {
				skin.Chromas[i].OriginSkinID = skin.ID
			}
			skins = append(skins, skin)
		}
	}
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
	if cachedData, found := championDetailCache[championID]; found {
		cacheMutex.RUnlock()
		log.Printf("Serving detailed champion %d from cache", championID)
		return cachedData, nil
	}
	cacheMutex.RUnlock()
	log.Printf("Fetching details for champion %d from CDragon...", championID)
	url := fmt.Sprintf("%s/v1/champions/%d.json", cDragonDataBE, championID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed http GET for champion %d details: %w", championID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("champion %d not found (404)", championID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status %d for champion %d details", resp.StatusCode, championID)
	}
	var champDetails DetailedChampionData
	if err := json.NewDecoder(resp.Body).Decode(&champDetails); err != nil {
		return nil, fmt.Errorf("failed to decode JSON for champion %d details: %w", championID, err)
	}
	for i := range champDetails.Skins {
		originSkinID := champDetails.Skins[i].ID
		champDetails.Skins[i].IsBase = (originSkinID%1000 == 0)
		for j := range champDetails.Skins[i].Chromas {
			champDetails.Skins[i].Chromas[j].OriginSkinID = originSkinID
		}
	}
	cacheMutex.Lock()
	championDetailCache[championID] = &champDetails
	cacheMutex.Unlock()
	log.Printf("Fetched and cached details for champion %d (%s)", championID, champDetails.Name)
	return &champDetails, nil
}

func GetSkinDetails(skinID int) (Skin, error) {
	cacheMutex.RLock()
	skinIDStr := fmt.Sprintf("%d", skinID)
	cachedSkin, found := allSkinsMap[skinIDStr]
	cacheMutex.RUnlock()
	if found {
		for i := range cachedSkin.Chromas {
			cachedSkin.Chromas[i].OriginSkinID = skinID
		}
		cachedSkin.IsBase = (skinID%1000 == 0)
		return cachedSkin, nil
	} else {
		log.Printf("WARN: Skin %d not in initial map. Trying champ detail fetch...", skinID)
		champID := GetChampionIDFromSkinID(skinID)
		if champID <= 0 {
			return Skin{}, fmt.Errorf("invalid champion ID for skin %d", skinID)
		}
		champDetails, err := FetchChampionDetails(champID)
		if err != nil {
			return Skin{}, fmt.Errorf("skin %d not in map, champ %d fetch failed: %w", skinID, champID, err)
		}
		for _, s := range champDetails.Skins {
			if s.ID == skinID {
				log.Printf("Found skin %d in champ %d details", skinID, champID)
				s.IsBase = (skinID%1000 == 0)
				for i := range s.Chromas {
					s.Chromas[i].OriginSkinID = skinID
				}
				return s, nil
			}
		}
		return Skin{}, fmt.Errorf("skin %d not found in map or champ %d details", skinID, champID)
	}
}

// --- URL Helper Functions (like data.js) ---

// !!! MODIFIED Asset function - Removed lowercasing on path construction !!!
func Asset(path string) string {
	if path == "" {
		return GetPlaceholderImageURL() // Return placeholder for empty paths
	}
	lowerPathCheck := strings.ToLower(path) // Use lowercase ONLY for checks

	// Check if already a full URL
	if strings.HasPrefix(lowerPathCheck, "http://") || strings.HasPrefix(lowerPathCheck, "https://") {
		return path // Assume it's correct
	}

	// Standardize path separators just in case
	// Use original path for concatenation as case might matter
	correctedPath := strings.ReplaceAll(path, "\\", "/")

	// Handle paths relative to plugin root
	// Use lowercase for prefix check, but concatenate original path
	if strings.HasPrefix(lowerPathCheck, "/plugins/") {
		// Ensure no double slash if cDragonBase already ends with one (it doesn't)
		return cDragonBase + correctedPath
	}

	// Handle paths seemingly relative to lol-game-data/assets
	if strings.HasPrefix(lowerPathCheck, "/lol-game-data/assets") {
		// Strip prefix based on lowercase check, get relative part from original path
		// This assumes the casing of the prefix itself doesn't matter, but suffix might
		prefixLen := len("/lol-game-data/assets")
		if len(correctedPath) > prefixLen {
			relativePath := correctedPath[prefixLen:]
			// Ensure relativePath starts with a single '/' if needed
			if !strings.HasPrefix(relativePath, "/") {
				relativePath = "/" + relativePath
			}
			return cDragonAssetBase + relativePath
		} else {
			log.Printf("WARN: Path '%s' matches asset prefix but is too short.", path)
			return GetPlaceholderImageURL() // Avoid constructing bad URL
		}
	}

	// If path looks like just the suffix (e.g., v1/champion-icons/1.png)
	// Assume it's relative to AssetBase - *Keep original casing*
	if !strings.HasPrefix(correctedPath, "/") && (strings.HasSuffix(lowerPathCheck, ".png") || strings.HasSuffix(lowerPathCheck, ".jpg") || strings.HasSuffix(lowerPathCheck, ".webp") || strings.HasSuffix(lowerPathCheck, ".svg")) {
		// log.Printf("DEBUG: Assuming relative asset path '%s' starts from CDragon Asset Base", correctedPath)
		return cDragonAssetBase + "/" + correctedPath
	}

	// Log the problematic path and return placeholder
	log.Printf("WARN: Could not determine absolute URL for asset path '%s', using placeholder.", path)
	return GetPlaceholderImageURL()
}

// GetChampionSquarePortraitURL uses the path from ChampionSummary.
func GetChampionSquarePortraitURL(champ ChampionSummary) string {
	// Example champ.SquarePortraitPath: "/lol-game-data/assets/v1/champion-icons/1.png"
	return Asset(champ.SquarePortraitPath)
}

// GetSkinTileURL uses the TilePath or LoadScreenPath from the Skin struct.
func GetSkinTileURL(skin Skin) string {
	// Example TilePath: "/lol-game-data/assets/v1/champion-tiles/103/103001.jpg"
	// Example LoadScreenPath: "/lol-game-data/assets/v1/champion-loadscreens/103/103001.jpg"
	path := skin.TilePath
	if path == "" {
		path = skin.LoadScreenPath // Common fallback for tiles
	}
	// Log the path being used for tiles
	// log.Printf("DEBUG: Tile path for skin %s (%d): %s", skin.Name, skin.ID, path)
	return Asset(path)
}

// GetSkinSplashURL uses the SplashPath or UncenteredSplashPath from the Skin struct.
func GetSkinSplashURL(skin Skin) string {
	// Example SplashPath: "/lol-game-data/assets/v1/champion-splashes/103/103001.jpg"
	// Example Uncentered: "/lol-game-data/assets/v1/champion-splashes/uncentered/103/103001.jpg"
	path := skin.SplashPath
	if skin.UncenteredSplashPath != "" { // Prefer uncentered if exists? Usually larger.
		path = skin.UncenteredSplashPath
	}
	// log.Printf("DEBUG: Splash path for skin %s (%d): %s", skin.Name, skin.ID, path)
	return Asset(path)
}

// GetChromaImageURL uses the ChromaPath from the Chroma struct.
func GetChromaImageURL(chroma Chroma) string {
	// Example ChromaPath: "/lol-game-data/assets/v1/champion-chroma-images/103/103001.png"
	// log.Printf("DEBUG: Chroma path for chroma %s (%d): %s", chroma.Name, chroma.ID, chroma.ChromaPath)
	return Asset(chroma.ChromaPath)
}

// KhadaUrl, Placeholder, Rarity, LegacyIconURL, ChromaIconURL, ParseHexColor - Remain the same
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

var rarityMap = map[string][2]string{
	"raritygem_ultimate.png":     {"Ultimate", "ultimate.png"},
	"raritygem_mythic.png":       {"Mythic", "mythic.png"},
	"raritygem_legendary.png":    {"Legendary", "legendary.png"},
	"raritygem_epic.png":         {"Epic", "epic.png"},
	"raritygem_transcendent.png": {"Transcendent", "transcendent.png"},
	"raritygem_exalted.png":      {"Exalted", "exalted.png"},
}

func Rarity(skin Skin) (name string, iconURL string) {
	if skin.Rarity == "" {
		return "", ""
	}
	lowerRarityPath := strings.ToLower(skin.Rarity)
	for pathSuffix, data := range rarityMap {
		if strings.HasSuffix(lowerRarityPath, pathSuffix) {
			name = data[0]
			iconFilename := data[1]
			iconURL = fmt.Sprintf("%s/v1/rarity-gem-icons/%s", cDragonDataBE, iconFilename)
			return name, iconURL
		}
	}
	return "", ""
}
func LegacyIconURL() string {
	return fmt.Sprintf("%s/images/summoner-icon/icon-legacy.png", cDragonStaticAssets)
}
func ChromaIconURL() string {
	return fmt.Sprintf("%s/images/skin-viewer/icon-chroma-default.png", cDragonStaticAssets)
}

func ParseHexColor(s string) (color.NRGBA, error) {
	if s == "" {
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}, fmt.Errorf("empty color string")
	}
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	}
	c := color.NRGBA{A: 255}
	var r, g, b uint64
	var err error
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
		var a uint64
		if err == nil {
			a, err = strconv.ParseUint(s[6:8], 16, 8)
			c.A = uint8(a)
		}
	default:
		err = fmt.Errorf("invalid hex length: %d", len(s))
	}
	if err != nil {
		log.Printf("WARN: Error parsing hex color '%s': %v. Using gray.", s, err)
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}, err
	}
	c.R = uint8(r)
	c.G = uint8(g)
	c.B = uint8(b)
	return c, nil
}
