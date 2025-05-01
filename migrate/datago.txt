// skinhunter/data/data.go
package data

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io" // Necesario para io.ReadAll
	"log"
	"net/http" // Necesario para http.Get (si no usamos Do)
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	// Import storage SOLO si se usa en OTRO LADO, para este fetch no es estrictamente necesario
	// pero lo mantenemos por si acaso y para inicializar el cliente que tu código usa
	storage "github.com/supabase-community/storage-go"
)

// ... (Constantes sin cambios) ...
const (
	cDragonBase         = "https://raw.communitydragon.org/latest"
	dataRootAPI         = cDragonBase + "/plugins/rcp-be-lol-game-data/global/default"
	cDragonStaticAssets = cDragonBase + "/plugins/rcp-fe-lol-static-assets/global/default"
	assetURLBase        = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default"
	jsonAssetPathPrefix = "/lol-game-data/assets"
	SupabaseURL         = "https://odlqwkgewzxxmbsqutja.supabase.co"
	SupabaseAPIKey      = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Im9kbHF3a2dld3p4eG1ic3F1dGphIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MzQyMTM2NzcsImV4cCI6MjA0OTc4OTY3N30.qka6a71bavDeUQgy_BKoVavaClRQa_gT36Au7oO9AF0"
	SupabaseBucket      = "api_json"
)

// ... (Variables Globales, mantenemos storageClient aunque no lo usemos *aquí*) ...
var (
	httpClient          = &http.Client{Timeout: time.Second * 15}
	championListCache   []ChampionSummary
	championDetailCache = make(map[int]*DetailedChampionData)
	skinLinesCache      []SkinLine
	cacheMutex          sync.RWMutex
	cDragonVersion      string
	allSkinsMap         map[string]Skin
	storageClient       *storage.Client
)

// ... (Loggers sin cambios) ...
var (
	debugLogger              = log.New(os.Stdout, "[DEBUG] ", log.Ltime|log.Lshortfile)
	imageLogFile, logOpenErr = os.OpenFile("c:\\Users\\dev\\buenoche\\image_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	imageLogger              *log.Logger
)

func init() {
	if logOpenErr != nil {
		fmt.Fprintf(os.Stderr, "[IMAGE_INIT_ERR] %v\n", logOpenErr)
		imageLogger = log.New(os.Stdout, "[IMAGE_ERR] ", log.Ltime|log.Lshortfile)
	} else if imageLogFile != nil {
		imageLogger = log.New(imageLogFile, "[IMAGE] ", log.Ltime|log.Lshortfile)
	} else {
		imageLogger = log.New(os.Stdout, "[IMAGE_WARN] ", log.Ltime|log.Lshortfile)
	}
	imageLogger.Println("--- Image Logger Initialized ---")
}

// ... (Structs sin cambios) ...
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
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	TilePath             string `json:"tilePath"`
	SplashPath           string `json:"splashPath"`
	UncenteredSplashPath string `json:"uncenteredSplashPath"`
	LoadScreenPath       string `json:"loadScreenPath"`
	Description          string `json:"description"`
	Rarity               string `json:"rarityGemPath"`
	IsLegacy             bool   `json:"isLegacy"`
	IsBase               bool
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
	OriginSkinID int      `json:"-"`
}

// --- Initialization and Caching ---
func InitData() error {
	log.Println("Initializing data...")
	cacheMutex.RLock()
	alreadyInitialized := len(championListCache) > 0 && len(allSkinsMap) > 0
	cacheMutex.RUnlock()
	if alreadyInitialized {
		return nil
	}
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if len(championListCache) > 0 && len(allSkinsMap) > 0 {
		return nil
	}

	// --- Inicializar Cliente Supabase Storage (aunque no se use para descargar JSON) ---
	storageClient = storage.NewClient(SupabaseURL+"/storage/v1", SupabaseAPIKey, nil)
	log.Println("Supabase Storage client initialized.")
	// ----------------------------------------------------------------------------------

	err := fetchCDragonVersion()
	if err != nil {
		log.Printf("WARN: Failed to fetch CDragon version: %v. Using 'latest'.", err)
		cDragonVersion = "latest"
	}
	log.Printf("Using CDragon version: %s", cDragonVersion)

	champions, err := fetchChampionSummary()
	if err != nil {
		return fmt.Errorf("failed to fetch champion summary: %w", err)
	}
	allSkins, err := fetchSkinsJSON()
	if err != nil {
		return fmt.Errorf("failed to fetch skins JSON: %w", err)
	}
	championListCache = champions
	allSkinsMap = allSkins
	log.Printf("Data initialized successfully: %d champions, %d skins.", len(championListCache), len(allSkinsMap))
	return nil
}

func fetchCDragonVersion() error { cDragonVersion = "latest"; return nil }

// --- FetchChampionJsonFromSupabase (Usando HTTP GET a URL pública como en appgo.txt) ---
func FetchChampionJsonFromSupabase(champId int) (map[string]interface{}, error) {
	// Construir la URL pública directamente
	path := fmt.Sprintf("%d.json", champId)
	downloadURL := fmt.Sprintf("%s/object/public/%s/%s", SupabaseURL+"/storage/v1", SupabaseBucket, path)
	log.Printf("Fetching Supabase data via HTTP GET: %s", downloadURL)

	// Usar el httpClient global con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request for Supabase: %w", err)
	}

	resp, err := httpClient.Do(req) // Usar el cliente HTTP estándar
	if err != nil {
		return nil, fmt.Errorf("error executing HTTP GET for Supabase URL %s: %w", downloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Leer cuerpo del error si es posible
		return nil, fmt.Errorf("bad status code %d fetching Supabase URL %s: %s", resp.StatusCode, downloadURL, string(bodyBytes))
	}

	// Leer y parsear el JSON
	dataBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from Supabase URL %s: %w", downloadURL, err)
	}

	var championData map[string]interface{}
	if err := json.Unmarshal(dataBytes, &championData); err != nil {
		return nil, fmt.Errorf("invalid champion data format from Supabase for %d.json: %w", champId, err)
	}

	log.Printf("Successfully parsed Supabase JSON for champion %d via HTTP GET", champId)
	return championData, nil
}

// --- Resto de Funciones (sin cambios) ---
func fetchChampionSummary() ([]ChampionSummary, error) { /* ... */
	url := fmt.Sprintf("%s/v1/champion-summary.json", dataRootAPI)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	var data []ChampionSummary
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	champs := make([]ChampionSummary, 0, len(data))
	for _, ch := range data {
		if ch.ID != -1 && ch.Name != "" && ch.Alias != "" {
			ch.Key = strings.ToLower(ch.Alias)
			ch.SquarePortraitPath = ensureLeadingSlash(ch.SquarePortraitPath)
			champs = append(champs, ch)
		}
	}
	sort.Slice(champs, func(i, j int) bool { return champs[i].Name < champs[j].Name })
	return champs, nil
}
func FetchAllChampions() ([]ChampionSummary, error) { /* ... */
	cacheMutex.RLock()
	if len(championListCache) > 0 {
		lc := make([]ChampionSummary, len(championListCache))
		copy(lc, championListCache)
		cacheMutex.RUnlock()
		return lc, nil
	}
	cacheMutex.RUnlock()
	initErr := InitData()
	if initErr != nil {
		return nil, initErr
	}
	cacheMutex.RLock()
	if len(championListCache) == 0 {
		cacheMutex.RUnlock()
		return nil, fmt.Errorf("cache still empty")
	}
	lc := make([]ChampionSummary, len(championListCache))
	copy(lc, championListCache)
	cacheMutex.RUnlock()
	return lc, nil
}
func fetchSkinsJSON() (map[string]Skin, error) { /* ... */
	url := fmt.Sprintf("%s/v1/skins.json", dataRootAPI)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	var data map[string]Skin
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	pd := make(map[string]Skin, len(data))
	for idStr, s := range data {
		skinID, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if s.Name == "" {
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
			pcs := make([]Chroma, 0, len(s.Chromas))
			for _, ch := range s.Chromas {
				if ch.ID == 0 {
					continue
				}
				ch.OriginSkinID = s.ID
				ch.ChromaPath = ensureLeadingSlash(ch.ChromaPath)
				pcs = append(pcs, ch)
			}
			s.Chromas = pcs
		}
		pd[idStr] = s
	}
	return pd, nil
}
func ensureLeadingSlash(path string) string { /* ... */
	if path != "" && !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}
func GetAllSkinsMap() (map[string]Skin, error) { /* ... */
	cacheMutex.RLock()
	if len(allSkinsMap) > 0 {
		mc := make(map[string]Skin, len(allSkinsMap))
		for k, v := range allSkinsMap {
			mc[k] = v
		}
		cacheMutex.RUnlock()
		return mc, nil
	}
	cacheMutex.RUnlock()
	err := InitData()
	if err != nil {
		return nil, err
	}
	cacheMutex.RLock()
	if len(allSkinsMap) == 0 {
		cacheMutex.RUnlock()
		return nil, fmt.Errorf("cache still empty")
	}
	mc := make(map[string]Skin, len(allSkinsMap))
	for k, v := range allSkinsMap {
		mc[k] = v
	}
	cacheMutex.RUnlock()
	return mc, nil
}
func GetSkinsForChampion(championID int) ([]Skin, error) { /* ... */
	cacheMutex.RLock()
	if len(allSkinsMap) == 0 {
		cacheMutex.RUnlock()
		return nil, fmt.Errorf("skins map not initialized")
	}
	skins := make([]Skin, 0)
	for _, s := range allSkinsMap {
		if GetChampionIDFromSkinID(s.ID) == championID {
			skinCopy := s
			if len(skinCopy.Chromas) > 0 {
				pcs := make([]Chroma, len(skinCopy.Chromas))
				for i, ch := range skinCopy.Chromas {
					pcs[i] = ch
					if pcs[i].OriginSkinID == 0 {
						pcs[i].OriginSkinID = skinCopy.ID
					}
				}
				skinCopy.Chromas = pcs
			}
			skins = append(skins, skinCopy)
		}
	}
	cacheMutex.RUnlock()
	sort.Slice(skins, func(i, j int) bool { return skins[i].ID < skins[j].ID })
	return skins, nil
}
func GetChampionIDFromSkinID(skinID int) int { /* ... */
	if skinID < 1000 {
		return -1
	}
	return skinID / 1000
}
func DeriveOriginSkinID(chromaID int) int { /* ... */
	if chromaID < 1000 {
		return -1
	}
	champID := chromaID / 1000
	return champID * 1000
}
func FetchChampionDetails(championID int) (*DetailedChampionData, error) { /* ... */
	cacheMutex.RLock()
	cd, f := championDetailCache[championID]
	cacheMutex.RUnlock()
	if f {
		return cd, nil
	}
	url := fmt.Sprintf("%s/v1/champions/%d.json", dataRootAPI, championID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	var details DetailedChampionData
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}
	details.SquarePortraitPath = ensureLeadingSlash(details.SquarePortraitPath)
	if len(details.Skins) > 0 {
		ps := make([]Skin, len(details.Skins))
		for i, s := range details.Skins {
			s.IsBase = (s.ID%1000 == 0)
			s.TilePath = ensureLeadingSlash(s.TilePath)
			s.SplashPath = ensureLeadingSlash(s.SplashPath)
			s.UncenteredSplashPath = ensureLeadingSlash(s.UncenteredSplashPath)
			s.LoadScreenPath = ensureLeadingSlash(s.LoadScreenPath)
			s.Rarity = ensureLeadingSlash(s.Rarity)
			if len(s.Chromas) > 0 {
				pcs := make([]Chroma, len(s.Chromas))
				for j, ch := range s.Chromas {
					pcs[j] = ch
					pcs[j].OriginSkinID = s.ID
					pcs[j].ChromaPath = ensureLeadingSlash(ch.ChromaPath)
				}
				s.Chromas = pcs
			}
			ps[i] = s
		}
		details.Skins = ps
	}
	cacheMutex.Lock()
	championDetailCache[championID] = &details
	cacheMutex.Unlock()
	return &details, nil
}
func GetSkinDetails(skinID int) (Skin, error) { /* ... */
	idStr := fmt.Sprintf("%d", skinID)
	cacheMutex.RLock()
	cs, fim := allSkinsMap[idStr]
	cacheMutex.RUnlock()
	if fim {
		sc := cs
		sc.IsBase = (sc.ID%1000 == 0)
		if len(sc.Chromas) > 0 {
			pcs := make([]Chroma, len(sc.Chromas))
			for i, ch := range sc.Chromas {
				pcs[i] = ch
				pcs[i].OriginSkinID = sc.ID
			}
			sc.Chromas = pcs
		}
		return sc, nil
	}
	champID := GetChampionIDFromSkinID(skinID)
	if champID <= 0 {
		return Skin{}, fmt.Errorf("invalid champID from skinID %d", skinID)
	}
	details, err := FetchChampionDetails(champID)
	if err != nil {
		return Skin{}, err
	}
	for _, s := range details.Skins {
		if s.ID == skinID {
			return s, nil
		}
	}
	return Skin{}, fmt.Errorf("skin %d not found for champ %d", skinID, champID)
}
func Asset(path string) string { /* ... */
	if path == "" {
		return GetPlaceholderImageURL()
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	lp := strings.ToLower(path)
	if strings.HasPrefix(lp, "http://") || strings.HasPrefix(lp, "https://") {
		return path
	}
	ljp := strings.ToLower(jsonAssetPathPrefix)
	if strings.HasPrefix(lp, ljp) {
		rp := path[len(jsonAssetPathPrefix):]
		lrp := strings.ToLower(rp)
		cb := strings.TrimSuffix(assetURLBase, "/")
		if !strings.HasPrefix(lrp, "/") {
			lrp = "/" + lrp
		}
		return cb + lrp
	} else {
		lpfb := strings.ToLower(path)
		if !strings.HasPrefix(lpfb, "/") {
			lpfb = "/" + lpfb
		}
		cb := strings.TrimSuffix(assetURLBase, "/")
		return cb + lpfb
	}
}
func GetChampionSquarePortraitURL(champ ChampionSummary) string {
	return Asset(champ.SquarePortraitPath)
}
func GetSkinTileURL(skin Skin) string {
	path := skin.TilePath
	if path == "" {
		path = skin.LoadScreenPath
	}
	if path == "" {
		return GetPlaceholderImageURL()
	}
	return Asset(path)
}
func GetSkinSplashURL(skin Skin) string {
	path := skin.UncenteredSplashPath
	if path == "" {
		path = skin.SplashPath
	}
	if path == "" {
		path = skin.TilePath
	}
	if path == "" {
		path = skin.LoadScreenPath
	}
	if path == "" {
		return GetPlaceholderImageURL()
	}
	return Asset(path)
}
func GetChromaImageURL(chroma Chroma) string { return Asset(chroma.ChromaPath) }
func KhadaUrl(skinID int, chromaID int) string {
	bu := "https://modelviewer.lol/model-viewer?id="
	if chromaID > 0 && skinID != chromaID && GetChampionIDFromSkinID(skinID) == GetChampionIDFromSkinID(chromaID) {
		return fmt.Sprintf("%s%d&chroma=%d", bu, skinID, chromaID)
	}
	return fmt.Sprintf("%s%d", bu, skinID)
}
func GetPlaceholderImageURL() string {
	return "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
}

var rarityMap = map[string][2]string{"raritygem_ultimate.png": {"Ultimate", "ultimate.png"}, "raritygem_mythic.png": {"Mythic", "mythic.png"}, "raritygem_legendary.png": {"Legendary", "legendary.png"}, "raritygem_epic.png": {"Epic", "epic.png"} /* ... */}

func Rarity(skin Skin) (string, string) {
	if skin.Rarity == "" {
		return "Standard", ""
	}
	lp := strings.ToLower(skin.Rarity)
	for sfx, d := range rarityMap {
		if strings.HasSuffix(lp, sfx) {
			ip := fmt.Sprintf("%s/v1/rarity-gem-icons/%s", jsonAssetPathPrefix, d[1])
			iu := Asset(ip)
			return d[0], iu
		}
	}
	return "Unknown", ""
}
func LegacyIconURL() string {
	return fmt.Sprintf("%s/images/summoner-icon/icon-legacy.png", cDragonStaticAssets)
}
func ChromaIconURL() string {
	return fmt.Sprintf("%s/images/skin-viewer/icon-chroma-default.png", cDragonStaticAssets)
}
func ParseHexColor(s string) (color.NRGBA, error) { /* ... */
	if s == "" {
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}, fmt.Errorf("empty hex")
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
		err = fmt.Errorf("invalid hex length %d", len(s))
	}
	if err != nil {
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}, err
	}
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}, nil
}

// --- End of data.go ---
