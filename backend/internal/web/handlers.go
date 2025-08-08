package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"reconya-ai/db"
	"reconya-ai/internal/config"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/nicidentifier"
	"reconya-ai/internal/scan"
	"reconya-ai/internal/settings"
	"reconya-ai/internal/systemstatus"
	"reconya-ai/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Templates will be loaded from filesystem for now
// TODO: Embed templates in production build

type WebHandler struct {
	deviceService         *device.DeviceService
	eventLogService       *eventlog.EventLogService
	networkService        *network.NetworkService
	systemStatusService   *systemstatus.SystemStatusService
	scanManager           *scan.ScanManager
	geolocationRepository *db.GeolocationRepository
	settingsService       *settings.SettingsService
	nicIdentifierService  *nicidentifier.NicIdentifierService
	templates             *template.Template
	sessionStore          *sessions.CookieStore
	config                *config.Config
}

type PageData struct {
	Page             string
	User             *models.User
	Error            string
	Username         string
	Devices          []*models.Device
	EventLogs        []*models.EventLog
	SystemStatusData *SystemStatusTemplateData // Use the new struct for system status
	NetworkMap       *NetworkMapData
	Networks         []models.Network
	ScanState        *scan.ScanState
}

type NetworkMapData struct {
	BaseIP      string
	IPRange     []int
	Devices     map[string]*models.Device
	NetworkInfo *NetworkInfo
}

type NetworkInfo struct {
	OnlineDevices  int
	IdleDevices    int
	OfflineDevices int
}

// SystemStatusTemplateData holds all data required by the system-status.html template
type SystemStatusTemplateData struct {
	SystemStatus *models.SystemStatus
	NetworkCIDR  string
	NetworkInfo  *NetworkInfo
	DevicesCount int
	ScanState    *scan.ScanState
}

func NewWebHandler(
	deviceService *device.DeviceService,
	eventLogService *eventlog.EventLogService,
	networkService *network.NetworkService,
	systemStatusService *systemstatus.SystemStatusService,
	scanManager *scan.ScanManager,
	geolocationRepository *db.GeolocationRepository,
	settingsService *settings.SettingsService,
	nicIdentifierService *nicidentifier.NicIdentifierService,
	config *config.Config,
	sessionSecret string,
) *WebHandler {
	// Initialize template functions
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "Never"
			}
			return t.Format("2006-01-02 15:04:05")
		},
		"formatTimeAgo": func(t time.Time) string {
			if t.IsZero() {
				return "Never"
			}
			duration := time.Since(t)
			switch {
			case duration < time.Minute:
				return fmt.Sprintf("%ds ago", int(duration.Seconds()))
			case duration < time.Hour:
				return fmt.Sprintf("%dm ago", int(duration.Minutes()))
			case duration < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(duration.Hours()))
			default:
				return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
			}
		},
		"formatFileSize": func(bytes interface{}) string {
			var size float64
			switch v := bytes.(type) {
			case int:
				size = float64(v)
			case int64:
				size = float64(v)
			case float64:
				size = v
			default:
				return "N/A"
			}

			if size == 0 {
				return "N/A"
			}

			kb := size / 1024
			if kb < 1024 {
				return fmt.Sprintf("%.1f KB", kb)
			}
			mb := kb / 1024
			return fmt.Sprintf("%.1f MB", mb)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"deref": func(ptr interface{}) interface{} {
			if ptr == nil {
				return "-"
			}
			switch v := ptr.(type) {
			case *string:
				if v == nil {
					return "-"
				}
				return *v
			case *time.Time:
				if v == nil {
					return time.Time{}
				}
				return *v
			default:
				return ptr
			}
		},
		"formatEventType": func(eventType string) string {
			return strings.ReplaceAll(strings.Title(strings.ReplaceAll(eventType, "_", " ")), "_", " ")
		},
		"slice": func(items interface{}, start, end int) interface{} {
			switch v := items.(type) {
			case []*models.Port:
				if start >= len(v) {
					return []*models.Port{}
				}
				if end > len(v) {
					end = len(v)
				}
				return v[start:end]
			}
			return items
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"len": func(items interface{}) int {
			switch v := items.(type) {
			case []*models.Device:
				return len(v)
			case []*models.Port:
				return len(v)
			case []*models.WebService:
				return len(v)
			case []*models.EventLog:
				return len(v)
			}
			return 0
		},
		"or": func(args ...interface{}) interface{} {
			for _, arg := range args {
				if arg != nil && arg != "" {
					return arg
				}
			}
			if len(args) > 0 {
				return args[len(args)-1]
			}
			return nil
		},
		"where": func(slice interface{}, field, value string) interface{} {
			switch v := slice.(type) {
			case []*models.Device:
				var result []*models.Device
				for _, item := range v {
					var fieldValue string
					switch field {
					case "Status":
						fieldValue = string(item.Status)
					case "IPv4":
						fieldValue = item.IPv4
					}
					if fieldValue == value {
						result = append(result, item)
					}
				}
				return result
			}
			return slice
		},
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
		"last": func(slice []string) string {
			if len(slice) == 0 {
				return ""
			}
			return slice[len(slice)-1]
		},
		"add": func(a, b interface{}) interface{} {
			switch av := a.(type) {
			case int:
				if bv, ok := b.(int); ok {
					return av + bv
				}
				if bv, ok := b.(float64); ok {
					return float64(av) + bv
				}
			case float64:
				if bv, ok := b.(float64); ok {
					return av + bv
				}
				if bv, ok := b.(int); ok {
					return av + float64(bv)
				}
			}
			return a
		},
		"mul": func(a, b interface{}) interface{} {
			switch av := a.(type) {
			case int:
				if bv, ok := b.(int); ok {
					return av * bv
				}
				if bv, ok := b.(float64); ok {
					return float64(av) * bv
				}
			case float64:
				if bv, ok := b.(float64); ok {
					return av * bv
				}
				if bv, ok := b.(int); ok {
					return av * float64(bv)
				}
			}
			return a
		},
		"div": func(a, b interface{}) interface{} {
			switch av := a.(type) {
			case int:
				if bv, ok := b.(int); ok {
					if bv == 0 {
						return 0
					}
					return av / bv
				}
				if bv, ok := b.(float64); ok {
					if bv == 0 {
						return 0.0
					}
					return float64(av) / bv
				}
			case float64:
				if bv, ok := b.(float64); ok {
					if bv == 0 {
						return 0.0
					}
					return av / bv
				}
				if bv, ok := b.(int); ok {
					if bv == 0 {
						return 0.0
					}
					return av / float64(bv)
				}
			}
			return a
		},
		"sub": func(a, b interface{}) interface{} {
			switch av := a.(type) {
			case int:
				if bv, ok := b.(int); ok {
					return av - bv
				}
				if bv, ok := b.(float64); ok {
					return float64(av) - bv
				}
			case float64:
				if bv, ok := b.(float64); ok {
					return av - bv
				}
				if bv, ok := b.(int); ok {
					return av - float64(bv)
				}
			}
			return a
		},
		"cos": func(angle float64) float64 {
			return math.Cos(angle)
		},
		"sin": func(angle float64) float64 {
			return math.Sin(angle)
		},
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"string": func(v interface{}) string {
			switch val := v.(type) {
			case models.DeviceType:
				return string(val)
			case models.DeviceStatus:
				return string(val)
			case string:
				return val
			default:
				return fmt.Sprintf("%v", val)
			}
		},
	}

	// Get the absolute path to the directory containing the executable
	execPath, err := os.Executable()
	if err != nil {
		panic(fmt.Sprintf("Failed to get executable path: %v", err))
	}
	execDir := filepath.Dir(execPath)

	// Parse templates from filesystem
	tmpl := template.New("").Funcs(funcMap)

	// Parse templates with unique names to avoid conflicts
	baseFiles, err := filepath.Glob(filepath.Join(execDir, "templates/layouts/*.html"))
	if err != nil {
		panic(fmt.Sprintf("Failed to glob base templates: %v", err))
	}

	pageFiles, err := filepath.Glob(filepath.Join(execDir, "templates/pages/*.html"))
	if err != nil {
		panic(fmt.Sprintf("Failed to glob page templates: %v", err))
	}

	componentFiles, err := filepath.Glob(filepath.Join(execDir, "templates/components/*.html"))
	if err != nil {
		panic(fmt.Sprintf("Failed to glob component templates: %v", err))
	}

	indexFile := filepath.Join(execDir, "templates/index.html")

	files := append(baseFiles, append(pageFiles, componentFiles...)...)
	files = append(files, indexFile)
	log.Printf("Found template files: %v", files)

	if len(files) == 0 {
		panic("No template files found")
	}

	tmpl, err = tmpl.ParseFiles(files...)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse templates: %v", err))
	}

	// Log template names for debugging
	for _, t := range tmpl.Templates() {
		log.Printf("Loaded template: %s", t.Name())
	}

	// Debug: Try to find login.html specifically
	loginTmpl := tmpl.Lookup("login.html")
	if loginTmpl != nil {
		log.Printf("Found login.html template: %s", loginTmpl.Name())
	} else {
		log.Printf("ERROR: login.html template not found!")
	}

	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to false for HTTP (localhost development)
		SameSite: http.SameSiteStrictMode,
	}

	return &WebHandler{
		deviceService:         deviceService,
		eventLogService:       eventLogService,
		networkService:        networkService,
		systemStatusService:   systemStatusService,
		scanManager:           scanManager,
		geolocationRepository: geolocationRepository,
		settingsService:       settingsService,
		nicIdentifierService:  nicIdentifierService,
		templates:             tmpl,
		sessionStore:          store,
		config:                config,
	}
}

// Page Handlers
func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated before showing the main page
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := PageData{
		Page: "dashboard", // This is the page name for the layout
		User: user,
	}

	log.Printf("Index: Attempting to execute index.html template")
	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Index: Template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) Home(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			// For HTMX requests, return a redirect header instead of HTTP redirect
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get system status from service
	status, err := h.systemStatusService.GetLatest()
	if err != nil {
		log.Printf("Error getting system status for home page: %v", err)
		// Fallback to mock data or handle gracefully
		status = &models.SystemStatus{
			NetworkID: "N/A",
			PublicIP:  nil,
		}
	} else if status == nil {
		log.Printf("No system status found in database for home page, using fallback")
		status = &models.SystemStatus{
			NetworkID: "N/A",
			PublicIP:  nil,
		}
	}

	// Get current or selected network to determine which network to show
	currentNetwork := h.scanManager.GetSelectedOrCurrentNetwork()
	scanState := h.scanManager.GetState()
	var devices []*models.Device
	var networkCIDR string = "N/A"

	if currentNetwork != nil {
		log.Printf("Home: currentNetwork is not nil, ID: %s", currentNetwork.ID)
		// Show devices from the currently selected/scanning network
		devicesSlice, err := h.deviceService.FindByNetworkID(currentNetwork.ID)
		if err != nil {
			log.Printf("Error getting devices for home page system status %s: %v", currentNetwork.ID, err)
			devices = []*models.Device{}
		} else {
			// Convert []models.Device to []*models.Device
			devices = make([]*models.Device, len(devicesSlice))
			for i := range devicesSlice {
				devices[i] = &devicesSlice[i]
			}
		}
		networkCIDR = currentNetwork.CIDR
	} else {
		log.Println("Home: currentNetwork is nil, falling back to all devices")
		// If no network is selected, show all devices
		devices, err = h.deviceService.FindAll()
		if err != nil {
			log.Printf("Error getting all devices for home page system status: %v", err)
			devices = []*models.Device{}
		}
	}

	networkMapData := h.buildNetworkMap(devices)

	systemStatusData := &SystemStatusTemplateData{
		SystemStatus: status,
		NetworkCIDR:  networkCIDR,
		NetworkInfo:  networkMapData.NetworkInfo,
		DevicesCount: len(devices),
		ScanState:    &scanState,
	}

	// Get recent event logs
	eventLogSlice, err := h.eventLogService.GetAll(20)
	if err != nil {
		log.Printf("Error getting event logs for home page: %v", err)
		eventLogSlice = []models.EventLog{} // Ensure it's an empty slice, not nil
	}

	// Convert to pointer slice for template
	eventLogs := make([]*models.EventLog, len(eventLogSlice))
	for i := range eventLogSlice {
		eventLogs[i] = &eventLogSlice[i]
	}

	// Get networks list
	networksSlice, err := h.networkService.FindAll()
	if err != nil {
		log.Printf("Error getting networks for home page: %v", err)
		networksSlice = []models.Network{} // Ensure it's an empty slice, not nil
	}

	data := PageData{
		Page:             "dashboard",
		User:             user,
		SystemStatusData: systemStatusData,
		Devices:          devices,
		EventLogs:        eventLogs,
		Networks:         networksSlice,
		ScanState:        &scanState,
	}

	if err := h.templates.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) About(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := struct {
		Page    string
		User    *models.User
		Version string
	}{
		Page:    "about",
		User:    user,
		Version: "0.14",
	}

	if err := h.templates.ExecuteTemplate(w, "about.html", data); err != nil {
		log.Printf("About template execution error: %v", err)
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

func (h *WebHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Get the absolute path to the directory containing the executable
		execPath, err := os.Executable()
		if err != nil {
			panic(fmt.Sprintf("Failed to get executable path: %v", err))
		}
		execDir := filepath.Dir(execPath)
		// Use standalone login template to avoid conflicts
		loginTmpl, err := template.ParseFiles(filepath.Join(execDir, ("templates/standalone/login.html")))
		if err != nil {
			log.Printf("Failed to parse standalone login template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			Page     string
			Error    string
			Username string
		}{
			Page:     "login",
			Error:    "",
			Username: "",
		}
		if err := loginTmpl.Execute(w, data); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Handle POST login
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Simple authentication (replace with your auth logic)
	if h.authenticate(username, password) {
		session, _ := h.sessionStore.Get(r, "reconya-session")
		session.Values["user_id"] = username
		session.Values["username"] = username
		session.Save(r, w)

		// Redirect to home page after successful login
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		data := struct {
			Page     string
			Error    string
			Username string
		}{
			Page:     "login",
			Error:    "Invalid username or password",
			Username: username,
		}
		if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *WebHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	session.Values = make(map[interface{}]interface{})
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// API Handlers for HTMX
func (h *WebHandler) APIDevices(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get current or selected network to determine which network to show
	currentNetwork := h.scanManager.GetSelectedOrCurrentNetwork()
	var devicesSlice []models.Device
	var err error

	if currentNetwork != nil {
		// Show devices from the currently selected/scanning network
		devicesSlice, err = h.deviceService.FindByNetworkID(currentNetwork.ID)
		if err != nil {
			log.Printf("Error getting devices for network %s: %v", currentNetwork.ID, err)
			devicesSlice = []models.Device{}
		}
	} else {
		// If no network is selected, show empty list
		devicesSlice = []models.Device{}
	}

	// Show devices with visual status indicators
	devices := make([]*models.Device, len(devicesSlice))
	for i := range devicesSlice {
		devices[i] = &devicesSlice[i]
	}

	// Get user's screenshot setting
	screenshotsEnabled := h.settingsService.AreScreenshotsEnabled(fmt.Sprintf("%d", user.ID))

	viewMode := r.URL.Query().Get("view")
	data := struct {
		Devices            []*models.Device
		ViewMode           string
		ScreenshotsEnabled bool
	}{
		Devices:            devices,
		ViewMode:           viewMode,
		ScreenshotsEnabled: screenshotsEnabled,
	}

	log.Printf("APIDevices: Found %d devices, viewMode: %s", len(devices), viewMode)
	if len(devices) > 0 {
		log.Printf("First device: ID=%s, IPv4=%s, Status=%s", devices[0].ID, devices[0].IPv4, devices[0].Status)
	}

	if err := h.templates.ExecuteTemplate(w, "components/device-grid.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APIDeviceModal(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["id"]

	device, err := h.deviceService.FindByID(deviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if device == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Get user's screenshot setting
	screenshotsEnabled := h.settingsService.AreScreenshotsEnabled(fmt.Sprintf("%d", user.ID))

	// Debug logging for IPv6 fields
	log.Printf("Device %s IPv6 data: LinkLocal=%v, UniqueLocal=%v, Global=%v, Addresses=%v",
		device.ID, device.IPv6LinkLocal, device.IPv6UniqueLocal, device.IPv6Global, device.IPv6Addresses)

	// Create template data with device and settings
	data := struct {
		*models.Device
		ScreenshotsEnabled bool
	}{
		Device:             device,
		ScreenshotsEnabled: screenshotsEnabled,
	}

	if err := h.templates.ExecuteTemplate(w, "components/device-modal.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APIUpdateDevice(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["id"]

	name := r.FormValue("hostname")
	comment := r.FormValue("comment")

	log.Printf("Updating device %s: name='%s', comment='%s'", deviceID, name, comment)

	var namePtr, commentPtr *string
	if name != "" {
		namePtr = &name
	}
	if comment != "" {
		commentPtr = &comment
	}

	device, err := h.deviceService.UpdateDevice(deviceID, namePtr, commentPtr)
	if err != nil {
		log.Printf("Failed to update device %s: %v", deviceID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated device %s", deviceID)

	// Get user's screenshot setting to match APIDeviceModal template data structure
	screenshotsEnabled := h.settingsService.AreScreenshotsEnabled(fmt.Sprintf("%d", user.ID))

	// Create template data with device and settings (matching APIDeviceModal structure)
	data := struct {
		*models.Device
		ScreenshotsEnabled bool
	}{
		Device:             device,
		ScreenshotsEnabled: screenshotsEnabled,
	}

	if err := h.templates.ExecuteTemplate(w, "components/device-modal.html", data); err != nil {
		log.Printf("Template execution error for device %s: %v", deviceID, err)
		http.Error(w, "Failed to render device modal", http.StatusInternalServerError)
	}
}

func (h *WebHandler) APIDeleteDevice(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["id"]

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get device info before deletion for logging
	device, err := h.deviceService.FindByID(deviceID)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	if device == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Delete the device
	err = h.deviceService.Delete(deviceID)
	if err != nil {
		log.Printf("Failed to delete device %s: %v", deviceID, err)
		http.Error(w, fmt.Sprintf("Failed to delete device: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the event
	h.eventLogService.Log(models.DeviceDeleted, fmt.Sprintf("Device %s deleted", device.IPv4), "")

	log.Printf("Successfully deleted device %s (%s)", device.IPv4, deviceID)

	// Return empty response to remove the table row
	w.WriteHeader(http.StatusOK)
}

// Test endpoint to add IPv6 data to a device (for debugging)
func (h *WebHandler) APITestIPv6(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceID := r.FormValue("device_id")
	if deviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	// Add test IPv6 data to the device
	ipv6Addresses := map[string]string{
		"link_local":   "fe80::1234:5678:90ab:cdef",
		"unique_local": "fd00::1234:5678:90ab:cdef",
		"global":       "2001:db8::1234:5678:90ab:cdef",
	}

	err := h.deviceService.UpdateDeviceIPv6Addresses(deviceID, ipv6Addresses)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update device IPv6: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("IPv6 addresses added successfully"))
}

func (h *WebHandler) APISystemStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("APISystemStatus called")
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get system status from service
	status, err := h.systemStatusService.GetLatest()
	if err != nil {
		log.Printf("Error getting system status: %v", err)
		// If service fails, create mock data for now
		status = &models.SystemStatus{
			NetworkID: "N/A",
			PublicIP:  nil,
		}
	} else if status == nil {
		log.Printf("No system status found in database, using fallback")
		// If no system status exists yet, create mock data
		status = &models.SystemStatus{
			NetworkID: "N/A",
			PublicIP:  nil,
		}
	} else {
		log.Printf("SystemStatus found: NetworkID=%s", status.NetworkID)
	}

	// Get current or selected network to determine which network to show
	currentNetwork := h.scanManager.GetSelectedOrCurrentNetwork()
	scanState := h.scanManager.GetState()
	var devices []*models.Device
	var networkCIDR string = "N/A"

	if currentNetwork != nil {
		log.Printf("APISystemStatus: currentNetwork is not nil, ID: %s", currentNetwork.ID)
		// Show devices from the currently selected/scanning network
		devicesSlice, err := h.deviceService.FindByNetworkID(currentNetwork.ID)
		if err != nil {
			log.Printf("Error getting devices for system status %s: %v", currentNetwork.ID, err)
			devices = []*models.Device{}
		} else {
			// Convert []models.Device to []*models.Device
			devices = make([]*models.Device, len(devicesSlice))
			for i := range devicesSlice {
				devices[i] = &devicesSlice[i]
			}
		}
		networkCIDR = currentNetwork.CIDR
	} else {
		log.Println("APISystemStatus: currentNetwork is nil, falling back to all devices")
		// If no network is selected, show all devices
		devices, err = h.deviceService.FindAll()
		if err != nil {
			log.Printf("Error getting all devices for system status: %v", err)
			devices = []*models.Device{}
		}
	}

	networkMapData := h.buildNetworkMap(devices)

	data := SystemStatusTemplateData{
		SystemStatus: status,
		NetworkCIDR:  networkCIDR,
		NetworkInfo:  networkMapData.NetworkInfo,
		DevicesCount: len(devices),
		ScanState:    &scanState,
	}

	log.Printf("APISystemStatus: returning data: %+v", data)

	if err := h.templates.ExecuteTemplate(w, "components/system-status.html", data); err != nil {
		log.Printf("Error executing template for system status: %v", err)
	}
}

func (h *WebHandler) APIEventLogs(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get recent event logs
	eventLogSlice, err := h.eventLogService.GetAll(20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to pointer slice
	eventLogs := make([]*models.EventLog, len(eventLogSlice))
	for i := range eventLogSlice {
		eventLogs[i] = &eventLogSlice[i]
	}

	data := struct {
		EventLogs []*models.EventLog
	}{
		EventLogs: eventLogs,
	}

	if err := h.templates.ExecuteTemplate(w, "components/event-logs.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APIEventLogsTable(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get more event logs for the table view (100 instead of 20)
	eventLogSlice, err := h.eventLogService.GetAll(100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to pointer slice
	eventLogs := make([]*models.EventLog, len(eventLogSlice))
	for i := range eventLogSlice {
		eventLogs[i] = &eventLogSlice[i]
	}

	data := struct {
		EventLogs []*models.EventLog
	}{
		EventLogs: eventLogs,
	}

	if err := h.templates.ExecuteTemplate(w, "components/event-logs-table.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APINetworkMap(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get current or selected network to determine which network to show
	currentNetwork := h.scanManager.GetSelectedOrCurrentNetwork()
	var devicesSlice []models.Device
	var err error

	if currentNetwork != nil {
		// Show devices from the currently selected/scanning network
		devicesSlice, err = h.deviceService.FindByNetworkID(currentNetwork.ID)
		if err != nil {
			log.Printf("Error getting devices for network map %s: %v", currentNetwork.ID, err)
			devicesSlice = []models.Device{}
		}
	} else {
		// If no network is selected, show empty map
		devicesSlice = []models.Device{}
	}

	devices := make([]*models.Device, len(devicesSlice))
	for i := range devicesSlice {
		devices[i] = &devicesSlice[i]
	}

	networkMap := h.buildNetworkMap(devices)
	if err := h.templates.ExecuteTemplate(w, "components/network-map.html", networkMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Helper methods
func (h *WebHandler) getUserFromSession(session *sessions.Session) *models.User {
	if userID, ok := session.Values["user_id"].(string); ok {
		return &models.User{
			Username: userID,
		}
	}
	return nil
}

func (h *WebHandler) authenticate(username, password string) bool {
	// Simple authentication - replace with your logic
	return username == "admin" && password == "password"
}

func (h *WebHandler) buildNetworkMap(devices []*models.Device) *NetworkMapData {
	// Build device map by IP
	deviceMap := make(map[string]*models.Device)
	online, idle, offline := 0, 0, 0

	for _, device := range devices {
		deviceMap[device.IPv4] = device
		switch device.Status {
		case models.DeviceStatusOnline:
			online++
		case models.DeviceStatusIdle:
			idle++
		default:
			offline++
		}
	}

	// Parse network CIDR from current scan network
	var baseIP string
	var ipRange []int
	currentNetwork := h.scanManager.GetCurrentNetwork()
	if currentNetwork != nil {
		baseIP, ipRange = h.parseNetworkCIDR(currentNetwork.CIDR)
	} else {
		// Fallback if no network is selected
		baseIP = "192.168.1"
		ipRange = make([]int, 254)
		for i := range ipRange {
			ipRange[i] = i + 1
		}
	}

	return &NetworkMapData{
		BaseIP:  baseIP,
		IPRange: ipRange,
		Devices: deviceMap,
		NetworkInfo: &NetworkInfo{
			OnlineDevices:  online + idle, // Count both online and idle as "online" for dashboard
			IdleDevices:    idle,
			OfflineDevices: offline,
		},
	}
}

// parseNetworkCIDR parses a CIDR string and returns base IP and host range
func (h *WebHandler) parseNetworkCIDR(cidr string) (string, []int) {
	// Default fallback
	defaultBaseIP := "192.168.1"
	defaultRange := make([]int, 254)
	for i := 1; i <= 254; i++ {
		defaultRange[i-1] = i
	}

	if cidr == "" {
		return defaultBaseIP, defaultRange
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Printf("Error parsing CIDR %s: %v", cidr, err)
		return defaultBaseIP, defaultRange
	}

	// Get network address
	networkIP := ipNet.IP

	// Calculate subnet mask bits
	ones, bits := ipNet.Mask.Size()
	if bits != 32 {
		log.Printf("Invalid network mask in CIDR %s", cidr)
		return defaultBaseIP, defaultRange
	}

	// Calculate number of host addresses
	hostBits := bits - ones
	totalHosts := 1 << hostBits // 2^hostBits

	// Subtract network and broadcast addresses
	usableHosts := totalHosts - 2
	if usableHosts <= 0 {
		usableHosts = 1
	}

	// Generate base IP (network portion)
	parts := strings.Split(networkIP.String(), ".")
	if len(parts) < 3 {
		return defaultBaseIP, defaultRange
	}

	// For /23 networks (like 192.168.10.0/23), we need to handle the range properly
	// For /24 networks, it's simpler
	var baseIP string
	var ipRange []int

	if ones >= 24 {
		// /24 or smaller subnet - use the first 3 octets as base
		baseIP = strings.Join(parts[:3], ".")
		// Generate host range for the last octet
		maxHosts := usableHosts
		if maxHosts > 254 {
			maxHosts = 254
		}
		ipRange = make([]int, maxHosts)
		for i := 1; i <= maxHosts; i++ {
			ipRange[i-1] = i
		}
	} else {
		// Larger subnet (like /23) - more complex range calculation
		baseIP = strings.Join(parts[:3], ".")

		// For /23, we have 512 addresses total, 510 usable
		// This spans two /24 networks (e.g., 192.168.10.0-192.168.11.255)
		maxHosts := usableHosts
		if maxHosts > 510 {
			maxHosts = 510
		}

		// Generate a reasonable range for visualization (limit to avoid UI issues)
		visualHosts := maxHosts
		if visualHosts > 254 {
			visualHosts = 254
		}

		ipRange = make([]int, visualHosts)
		for i := 1; i <= visualHosts; i++ {
			ipRange[i-1] = i
		}
	}

	return baseIP, ipRange
}

func (h *WebHandler) APITargets(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Same as APIDevices - targets are devices
	devices, err := h.deviceService.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user's screenshot setting
	screenshotsEnabled := h.settingsService.AreScreenshotsEnabled(fmt.Sprintf("%d", user.ID))

	viewMode := r.URL.Query().Get("view")
	data := struct {
		Devices            []*models.Device
		ViewMode           string
		ScreenshotsEnabled bool
	}{
		Devices:            devices,
		ViewMode:           viewMode,
		ScreenshotsEnabled: screenshotsEnabled,
	}

	if err := h.templates.ExecuteTemplate(w, "components/device-grid.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APITrafficCore(w http.ResponseWriter, r *http.Request) {
	devices, err := h.deviceService.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Devices []*models.Device
	}{
		Devices: devices,
	}

	if err := h.templates.ExecuteTemplate(w, "components/traffic-core.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APIDeviceList(w http.ResponseWriter, r *http.Request) {
	devices, err := h.deviceService.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Devices []*models.Device
	}{
		Devices: devices,
	}

	if err := h.templates.ExecuteTemplate(w, "components/device-list.html", data); err != nil {
		log.Printf("Error executing template for device list: %v", err)
	}
}

// APICleanupDeviceNames clears all device names
func (h *WebHandler) APICleanupNetworkBroadcastDevices(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.deviceService.CleanupNetworkBroadcastDevices()
	if err != nil {
		log.Printf("Error cleaning up network/broadcast devices: %v", err)
		http.Error(w, "Failed to cleanup network/broadcast devices", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Network/broadcast devices cleaned up successfully"))
}

func (h *WebHandler) APICleanupDeviceNames(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.deviceService.CleanupAllDeviceNames()
	if err != nil {
		log.Printf("Device name cleanup failed: %v", err)
		http.Error(w, fmt.Sprintf("Cleanup failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "All device names have been cleared successfully"}`))
}

// Network API handlers
func (h *WebHandler) APINetworks(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("APINetworks: Fetching networks for display")
	// Get all networks from service
	networksSlice, err := h.networkService.FindAll()
	if err != nil {
		log.Printf("APINetworks: Error getting networks: %v", err)
		networksSlice = []models.Network{} // Ensure it's an empty slice, not nil
	}

	log.Printf("APINetworks: Retrieved %d networks from service", len(networksSlice))

	// Convert to pointer slice for template
	networks := make([]*models.Network, len(networksSlice))
	for i := range networksSlice {
		// Get device count for each network
		deviceCount, _ := h.networkService.GetDeviceCount(networksSlice[i].ID)
		networksSlice[i].DeviceCount = deviceCount
		networks[i] = &networksSlice[i]
	}

	// Get scan state for network selection highlighting
	scanState := h.scanManager.GetState()

	data := struct {
		Networks  []*models.Network
		ScanState *scan.ScanState
	}{
		Networks:  networks,
		ScanState: &scanState,
	}

	if err := h.templates.ExecuteTemplate(w, "components/network-list.html", data); err != nil {
		log.Printf("Error executing template for networks: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APINetworkModal(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	networkID := vars["id"]

	data := struct {
		Network *models.Network
		Error   string
	}{
		Network: &models.Network{},
	}

	// If editing existing network, load it
	if networkID != "" {
		network, err := h.networkService.FindByID(networkID)
		if err != nil {
			data.Error = "Network not found"
		} else if network != nil {
			data.Network = network
		}
	}

	if err := h.templates.ExecuteTemplate(w, "components/network-modal.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APICreateNetwork(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	cidr := strings.TrimSpace(r.FormValue("cidr"))
	description := strings.TrimSpace(r.FormValue("description"))

	log.Printf("APICreateNetwork: Received request - name=%s, cidr=%s, description=%s", name, cidr, description)

	data := struct {
		Network *models.Network
		Error   string
	}{
		Network: &models.Network{
			Name:        name,
			CIDR:        cidr,
			Description: description,
		},
	}

	// Validate CIDR
	if cidr == "" {
		data.Error = "CIDR address is required"
		if err := h.templates.ExecuteTemplate(w, "components/network-modal.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Validate CIDR format
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		data.Error = "Invalid CIDR format. Please use format like 192.168.1.0/24"
		if err := h.templates.ExecuteTemplate(w, "components/network-modal.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Create network
	log.Printf("APICreateNetwork: Calling networkService.Create")
	network, err := h.networkService.Create(name, cidr, description)
	if err != nil {
		log.Printf("APICreateNetwork: Error creating network: %v", err)
		data.Error = fmt.Sprintf("Failed to create network: %v", err)
		if err := h.templates.ExecuteTemplate(w, "components/network-modal.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	log.Printf("APICreateNetwork: Network created successfully: ID=%s, CIDR=%s", network.ID, network.CIDR)

	// Log the event
	h.eventLogService.Log(models.NetworkCreated, fmt.Sprintf("Network %s (%s) created", network.CIDR, network.Name), "")

	// Return success indicator that will trigger the frontend to handle the response
	w.Header().Set("HX-Trigger", "network-saved")
	w.WriteHeader(http.StatusOK)
	// Return empty response - frontend will handle the success message
	w.Write([]byte(""))
}

func (h *WebHandler) APIUpdateNetwork(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	networkID := vars["id"]

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	cidr := strings.TrimSpace(r.FormValue("cidr"))
	description := strings.TrimSpace(r.FormValue("description"))

	// Validate CIDR
	if cidr == "" {
		http.Error(w, "CIDR address is required", http.StatusBadRequest)
		return
	}

	// Validate CIDR format
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		http.Error(w, "Invalid CIDR format. Please use format like 192.168.1.0/24", http.StatusBadRequest)
		return
	}

	// Update network
	network, err := h.networkService.Update(networkID, name, cidr, description)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update network: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the event
	h.eventLogService.Log(models.NetworkUpdated, fmt.Sprintf("Network %s (%s) updated", network.CIDR, network.Name), "")

	// Return success indicator that will trigger the frontend to handle the response
	w.Header().Set("HX-Trigger", "network-saved")
	w.WriteHeader(http.StatusOK)
	// Return empty response - frontend will handle the success message
	w.Write([]byte(""))
}

func (h *WebHandler) APIDeleteNetwork(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	networkID := vars["id"]

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if a scan is currently running on this network
	if h.scanManager.IsRunning() {
		currentNetwork := h.scanManager.GetCurrentNetwork()
		if currentNetwork != nil && currentNetwork.ID == networkID {
			http.Error(w, "Cannot delete network: a scan is currently running on this network. Please stop the scan first.", http.StatusConflict)
			return
		}
	}

	// Get network info before deletion for logging
	network, err := h.networkService.FindByID(networkID)
	if err != nil {
		http.Error(w, "Network not found", http.StatusNotFound)
		return
	}

	// Check if network has devices before deletion
	deviceCount, err := h.networkService.GetDeviceCount(networkID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check network devices: %v", err), http.StatusInternalServerError)
		return
	}

	if deviceCount > 0 {
		http.Error(w, fmt.Sprintf("Cannot delete network: %d devices are still using this network. Please remove or reassign devices first.", deviceCount), http.StatusBadRequest)
		return
	}

	// Delete network
	err = h.networkService.Delete(networkID)
	if err != nil {
		// Check if this is a foreign key constraint error
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			http.Error(w, "Cannot delete network: devices are still using this network. Please remove or reassign devices first.", http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete network: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Log the event
	if network != nil {
		h.eventLogService.Log(models.NetworkDeleted, fmt.Sprintf("Network %s (%s) deleted", network.CIDR, network.Name), "")
	}

	// Return empty response to remove the table row
	w.WriteHeader(http.StatusOK)
}

// APINetworkDeleteInfo returns information about network deletion including affected devices
func (h *WebHandler) APINetworkDeleteInfo(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	networkID := vars["id"]

	// Get network info
	network, err := h.networkService.FindByID(networkID)
	if err != nil {
		http.Error(w, "Network not found", http.StatusNotFound)
		return
	}

	// Check if a scan is currently running on this network
	isScanning := false
	if h.scanManager.IsRunning() {
		currentNetwork := h.scanManager.GetCurrentNetwork()
		if currentNetwork != nil && currentNetwork.ID == networkID {
			isScanning = true
		}
	}

	// Get device count
	deviceCount, err := h.networkService.GetDeviceCount(networkID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check network devices: %v", err), http.StatusInternalServerError)
		return
	}

	// Get devices for this network to show in confirmation
	devices, err := h.deviceService.FindByNetworkID(networkID)
	if err != nil {
		log.Printf("Error fetching devices for network %s: %v", networkID, err)
		devices = []models.Device{} // Empty slice if error
	}

	deleteInfo := struct {
		Network     *models.Network `json:"network"`
		DeviceCount int             `json:"deviceCount"`
		Devices     []models.Device `json:"devices"`
		IsScanning  bool            `json:"isScanning"`
		CanDelete   bool            `json:"canDelete"`
		Message     string          `json:"message"`
	}{
		Network:     network,
		DeviceCount: deviceCount,
		Devices:     devices,
		IsScanning:  isScanning,
		CanDelete:   !isScanning,
		Message:     "",
	}

	if isScanning {
		deleteInfo.Message = "Cannot delete network: a scan is currently running on this network. Please stop the scan first."
	} else if deviceCount > 0 {
		deleteInfo.Message = fmt.Sprintf("This network contains %d device(s). Deleting the network will also remove these devices from the system.", deviceCount)
	} else {
		deleteInfo.Message = "Are you sure you want to delete this network?"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deleteInfo)
}

// APIForceDeleteNetwork deletes a network and all its devices with confirmation
func (h *WebHandler) APIForceDeleteNetwork(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	networkID := vars["id"]

	// Check if a scan is currently running on this network
	if h.scanManager.IsRunning() {
		currentNetwork := h.scanManager.GetCurrentNetwork()
		if currentNetwork != nil && currentNetwork.ID == networkID {
			http.Error(w, "Cannot delete network: a scan is currently running on this network. Please stop the scan first.", http.StatusConflict)
			return
		}
	}

	// Get network info before deletion for logging
	network, err := h.networkService.FindByID(networkID)
	if err != nil {
		http.Error(w, "Network not found", http.StatusNotFound)
		return
	}

	// Get device count for logging
	deviceCount, err := h.networkService.GetDeviceCount(networkID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check network devices: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete devices first if they exist
	if deviceCount > 0 {
		err = h.deviceService.DeleteByNetworkID(networkID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete network devices: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Deleted %d devices from network %s before network deletion", deviceCount, networkID)
	}

	// Now delete the network
	err = h.networkService.Delete(networkID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete network: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the event
	if network != nil {
		message := fmt.Sprintf("Network %s (%s) deleted", network.CIDR, network.Name)
		if deviceCount > 0 {
			message += fmt.Sprintf(" along with %d device(s)", deviceCount)
		}
		h.eventLogService.Log(models.NetworkDeleted, message, "")
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Network deleted successfully",
	})
}

// APINetworkDeleteModal returns the network deletion confirmation modal
func (h *WebHandler) APINetworkDeleteModal(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	networkID := vars["id"]

	// Get network info
	network, err := h.networkService.FindByID(networkID)
	if err != nil {
		http.Error(w, "Network not found", http.StatusNotFound)
		return
	}

	// Check if a scan is currently running on this network
	isScanning := false
	if h.scanManager.IsRunning() {
		currentNetwork := h.scanManager.GetCurrentNetwork()
		if currentNetwork != nil && currentNetwork.ID == networkID {
			isScanning = true
		}
	}

	// Get device count
	deviceCount, err := h.networkService.GetDeviceCount(networkID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check network devices: %v", err), http.StatusInternalServerError)
		return
	}

	// Get devices for this network to show in confirmation
	devices, err := h.deviceService.FindByNetworkID(networkID)
	if err != nil {
		log.Printf("Error fetching devices for network %s: %v", networkID, err)
		devices = []models.Device{} // Empty slice if error
	}

	deleteInfo := struct {
		Network     *models.Network `json:"network"`
		DeviceCount int             `json:"deviceCount"`
		Devices     []models.Device `json:"devices"`
		IsScanning  bool            `json:"isScanning"`
		CanDelete   bool            `json:"canDelete"`
		Message     string          `json:"message"`
	}{
		Network:     network,
		DeviceCount: deviceCount,
		Devices:     devices,
		IsScanning:  isScanning,
		CanDelete:   !isScanning,
		Message:     "",
	}

	if isScanning {
		deleteInfo.Message = "Cannot delete network: a scan is currently running on this network. Please stop the scan first."
	} else if deviceCount > 0 {
		deleteInfo.Message = fmt.Sprintf("This network contains %d device(s). Deleting the network will also remove these devices from the system.", deviceCount)
	} else {
		deleteInfo.Message = "Are you sure you want to delete this network?"
	}

	if err := h.templates.ExecuteTemplate(w, "components/network-delete-modal.html", deleteInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APIScanStatus returns the current scan status
func (h *WebHandler) APIScanStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	scanState := h.scanManager.GetState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scanState)
}

// APIScanStart starts scanning a network
func (h *WebHandler) APIScanStart(w http.ResponseWriter, r *http.Request) {
	log.Printf("APIScanStart: Request received, method=%s", r.Method)

	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		log.Printf("APIScanStart: Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	networkID := r.FormValue("network-selector")
	log.Printf("APIScanStart: Network ID from form: '%s'", networkID)

	if networkID == "" {
		log.Printf("APIScanStart: No network ID provided")
		// Return scan control component with error message
		h.APIScanControlWithError(w, r, "Please select a network to scan")
		return
	}

	err := h.scanManager.StartScan(networkID)
	if err != nil {
		if scanErr, ok := err.(*scan.ScanError); ok {
			switch scanErr.Type {
			case scan.AlreadyRunning:
				http.Error(w, scanErr.Message, http.StatusConflict)
			case scan.NetworkNotFound:
				http.Error(w, scanErr.Message, http.StatusNotFound)
			default:
				http.Error(w, scanErr.Message, http.StatusBadRequest)
			}
		} else {
			http.Error(w, fmt.Sprintf("Failed to start scan: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Note: Scan started event is logged by scan_manager.go to avoid duplicates

	// Return updated scan control component
	h.APIScanControl(w, r)
}

// APIScanStop stops the current scan
func (h *WebHandler) APIScanStop(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.scanManager.StopScan()
	if err != nil {
		if scanErr, ok := err.(*scan.ScanError); ok {
			switch scanErr.Type {
			case scan.NotRunning:
				http.Error(w, scanErr.Message, http.StatusConflict)
			default:
				http.Error(w, scanErr.Message, http.StatusBadRequest)
			}
		} else {
			http.Error(w, fmt.Sprintf("Failed to stop scan: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Log the event
	h.eventLogService.Log(models.ScanStopped, "Network scan stopped", "")

	// Return updated scan control component
	h.APIScanControl(w, r)
}

// APIScanControl returns the scan control component
func (h *WebHandler) APIScanControl(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set cache control headers to prevent browser caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Get networks and scan state
	networksSlice, err := h.networkService.FindAll()
	if err != nil {
		log.Printf("Error getting networks for scan control: %v", err)
		networksSlice = []models.Network{}
	}

	scanState := h.scanManager.GetState()

	data := PageData{
		Networks:  networksSlice,
		ScanState: &scanState,
	}

	if err := h.templates.ExecuteTemplate(w, "components/scan-control-inner.html", data); err != nil {
		log.Printf("Error rendering scan control template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APIScanControlWithError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set cache control headers to prevent browser caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Get networks and scan state
	networksSlice, err := h.networkService.FindAll()
	if err != nil {
		log.Printf("Error getting networks for scan control: %v", err)
		networksSlice = []models.Network{}
	}

	scanState := h.scanManager.GetState()

	data := PageData{
		Networks:  networksSlice,
		ScanState: &scanState,
		Error:     errorMsg,
	}

	if err := h.templates.ExecuteTemplate(w, "components/scan-control-inner.html", data); err != nil {
		log.Printf("Error rendering scan control template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// APIScanSelectNetwork sets the selected network (without starting scan)
// APIDashboardMetrics returns JSON data for dashboard metrics
func (h *WebHandler) APIDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get current or selected network to determine which network to show
	currentNetwork := h.scanManager.GetSelectedOrCurrentNetwork()
	var devices []*models.Device
	var networkCIDR string = "N/A"
	var err error

	if currentNetwork != nil {
		// Show devices from the currently selected/scanning network
		devicesSlice, err := h.deviceService.FindByNetworkID(currentNetwork.ID)
		if err != nil {
			log.Printf("Error getting devices for dashboard metrics %s: %v", currentNetwork.ID, err)
			devices = []*models.Device{}
		} else {
			// Convert []models.Device to []*models.Device
			devices = make([]*models.Device, len(devicesSlice))
			for i := range devicesSlice {
				devices[i] = &devicesSlice[i]
			}
		}
		networkCIDR = currentNetwork.CIDR
	} else {
		// If no network is selected, show all devices
		devices, err = h.deviceService.FindAll()
		if err != nil {
			log.Printf("Error getting all devices for dashboard metrics: %v", err)
			devices = []*models.Device{}
		}
	}

	networkMapData := h.buildNetworkMap(devices)

	// Get system status for public IP
	status, err := h.systemStatusService.GetLatest()
	var publicIP string = "N/A"
	if err == nil && status != nil && status.PublicIP != nil {
		publicIP = *status.PublicIP
	}

	metrics := map[string]interface{}{
		"networkRange":   networkCIDR,
		"publicIP":       publicIP,
		"devicesFound":   len(devices),
		"devicesOnline":  networkMapData.NetworkInfo.OnlineDevices,
		"devicesOffline": networkMapData.NetworkInfo.OfflineDevices,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *WebHandler) APIScanSelectNetwork(w http.ResponseWriter, r *http.Request) {
	log.Println("APIScanSelectNetwork called")
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	networkID := r.FormValue("network-id")
	if networkID == "" {
		http.Error(w, "Network ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Setting selected network to: %s", networkID)
	err := h.scanManager.SetSelectedNetwork(networkID)
	if err != nil {
		if scanErr, ok := err.(*scan.ScanError); ok {
			switch scanErr.Type {
			case scan.NetworkNotFound:
				http.Error(w, scanErr.Message, http.StatusNotFound)
			default:
				http.Error(w, scanErr.Message, http.StatusBadRequest)
			}
		} else {
			http.Error(w, fmt.Sprintf("Failed to select network: %v", err), http.StatusInternalServerError)
		}
		return
	}

	log.Println("APIScanSelectNetwork completed successfully")
	w.Header().Set("HX-Trigger", "network-selected")
	w.WriteHeader(http.StatusOK)
}

// APIAbout returns the about page content for the SPA
func (h *WebHandler) APIAbout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	data := struct {
		Version string
	}{
		Version: "0.14",
	}

	if err := h.templates.ExecuteTemplate(w, "components/about.html", data); err != nil {
		log.Printf("About component template execution error: %v", err)
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

// WorldMapData holds geolocation information for the world map component
type WorldMapData struct {
	PublicIP  string
	Location  string
	Latitude  string
	Longitude string
	Country   string
	City      string
	Timezone  string
	ISP       string
	PinX      int // X coordinate for pin position on map (0-400)
	PinY      int // Y coordinate for pin position on map (0-200)
}

// GeolocationResponse represents the response from ipapi.co
type GeolocationResponse struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Country     string  `json:"country_name"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"org"`
}

// APIWorldMap handles the world map component data with caching
func (h *WebHandler) APIWorldMap(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	// Get system status to access public IP
	systemStatus, err := h.systemStatusService.GetLatest()
	if err != nil {
		log.Printf("Error getting system status for world map: %v", err)
		// Provide fallback data
		unknown := "Unknown"
		systemStatus = &models.SystemStatus{
			PublicIP: &unknown,
		}
	}

	// Extract public IP string
	publicIP := "Unknown"
	if systemStatus.PublicIP != nil && *systemStatus.PublicIP != "" {
		publicIP = *systemStatus.PublicIP
	}

	// Initialize with default values
	worldMapData := &WorldMapData{
		PublicIP:  publicIP,
		Location:  "Unknown Location",
		Latitude:  "0.0000",
		Longitude: "0.0000",
		Country:   "Unknown",
		City:      "Unknown",
		Timezone:  "UTC",
		ISP:       "Unknown ISP",
		PinX:      200, // Center of map
		PinY:      100, // Center of map
	}

	// Get geolocation data if we have a valid public IP
	if h.isValidPublicIP(publicIP) {
		geoData := h.getCachedGeolocationData(ctx, publicIP)
		if geoData != nil && h.geolocationRepository.IsValidCache(geoData) {
			worldMapData = h.buildWorldMapDataFromCache(geoData)
		} else {
			log.Printf("Using fallback data for IP: %s (cache invalid or missing)", publicIP)
		}
	}

	if err := h.templates.ExecuteTemplate(w, "components/world-map.html", worldMapData); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getGeolocationData fetches geolocation data with fallback mechanisms
func (h *WebHandler) getGeolocationData(ip string) *GeolocationResponse {
	// Try ipapi.co first
	if geoData := h.tryIPApiCo(ip); geoData != nil {
		return geoData
	}

	// Fallback to hardcoded data for common IPs
	return h.getFallbackGeolocation(ip)
}

// tryIPApiCo attempts to get geolocation from ipapi.co
func (h *WebHandler) tryIPApiCo(ip string) *GeolocationResponse {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	url := fmt.Sprintf("https://ipapi.co/%s/json/", ip)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error fetching from ipapi.co: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ipapi.co returned status: %d", resp.StatusCode)
		return nil
	}

	var geoResponse GeolocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResponse); err != nil {
		log.Printf("Error parsing ipapi.co response: %v", err)
		return nil
	}

	// Check for error in response
	if geoResponse.City == "" && geoResponse.Country == "" {
		return nil
	}

	return &geoResponse
}

// isValidPublicIP checks if the IP is a valid public IP address
func (h *WebHandler) isValidPublicIP(ip string) bool {
	return ip != "Unknown" && ip != "" &&
		!strings.HasPrefix(ip, "192.168.") &&
		!strings.HasPrefix(ip, "10.") &&
		!strings.HasPrefix(ip, "172.16.") &&
		!strings.HasPrefix(ip, "127.") &&
		!strings.HasPrefix(ip, "169.254.")
}

// getCachedGeolocationData retrieves geolocation data with caching logic
func (h *WebHandler) getCachedGeolocationData(ctx context.Context, ip string) *models.GeolocationCache {
	// First try to get from cache
	cachedData, err := h.geolocationRepository.FindByIP(ctx, ip)
	if err == nil && h.geolocationRepository.IsValidCache(cachedData) {
		log.Printf("Using cached geolocation data for IP: %s", ip)
		return cachedData
	}

	// Cache miss or invalid data - fetch fresh data
	log.Printf("Cache miss or invalid for IP: %s, fetching fresh data", ip)
	geoResponse := h.getGeolocationData(ip)
	if geoResponse == nil {
		log.Printf("Failed to fetch geolocation data for IP: %s", ip)
		return nil
	}

	// Convert response to cache model
	cache := &models.GeolocationCache{
		IP:          ip,
		City:        geoResponse.City,
		Region:      geoResponse.Region,
		Country:     geoResponse.Country,
		CountryCode: geoResponse.CountryCode,
		Latitude:    geoResponse.Latitude,
		Longitude:   geoResponse.Longitude,
		Timezone:    geoResponse.Timezone,
		ISP:         geoResponse.ISP,
		Source:      "api",
	}

	// Determine source type based on data quality
	if geoResponse.City == "" || geoResponse.Country == "" || geoResponse.CountryCode == "XX" {
		cache.Source = "fallback"
	}

	// Save to cache
	if err := h.geolocationRepository.Upsert(ctx, cache); err != nil {
		log.Printf("Failed to cache geolocation data: %v", err)
	} else {
		log.Printf("Cached geolocation data for IP: %s", ip)
	}

	return cache
}

// buildWorldMapDataFromCache converts cached geolocation data to world map data
func (h *WebHandler) buildWorldMapDataFromCache(cache *models.GeolocationCache) *WorldMapData {
	// Convert lat/lon to map coordinates (400x200 SVG)
	// Longitude: -180 to 180 -> 0 to 400
	// Latitude: 90 to -90 -> 0 to 200 (inverted)
	pinX := int((cache.Longitude + 180) * 400 / 360)
	pinY := int((90 - cache.Latitude) * 200 / 180)

	// Clamp values to stay within bounds
	if pinX < 0 {
		pinX = 0
	}
	if pinX > 400 {
		pinX = 400
	}
	if pinY < 0 {
		pinY = 0
	}
	if pinY > 200 {
		pinY = 200
	}

	location := fmt.Sprintf("%s, %s", cache.City, cache.Country)
	if cache.City == "" || cache.Country == "" {
		location = "Unknown Location"
	}

	return &WorldMapData{
		PublicIP:  cache.IP,
		Location:  location,
		Latitude:  fmt.Sprintf("%.4f", cache.Latitude),
		Longitude: fmt.Sprintf("%.4f", cache.Longitude),
		Country:   cache.CountryCode,
		City:      cache.City,
		Timezone:  cache.Timezone,
		ISP:       cache.ISP,
		PinX:      pinX,
		PinY:      pinY,
	}
}

// getFallbackGeolocation provides hardcoded geolocation for common IPs
func (h *WebHandler) getFallbackGeolocation(ip string) *GeolocationResponse {
	switch {
	case strings.HasPrefix(ip, "8.8."):
		// Google DNS
		return &GeolocationResponse{
			IP:          ip,
			City:        "Mountain View",
			Region:      "California",
			Country:     "United States",
			CountryCode: "US",
			Latitude:    37.4419,
			Longitude:   -122.1430,
			Timezone:    "America/Los_Angeles",
			ISP:         "Google LLC",
		}
	case strings.HasPrefix(ip, "1.1."):
		// Cloudflare DNS
		return &GeolocationResponse{
			IP:          ip,
			City:        "San Francisco",
			Region:      "California",
			Country:     "United States",
			CountryCode: "US",
			Latitude:    37.7749,
			Longitude:   -122.4194,
			Timezone:    "America/Los_Angeles",
			ISP:         "Cloudflare Inc",
		}
	case strings.HasPrefix(ip, "208.67."):
		// OpenDNS
		return &GeolocationResponse{
			IP:          ip,
			City:        "San Francisco",
			Region:      "California",
			Country:     "United States",
			CountryCode: "US",
			Latitude:    37.7749,
			Longitude:   -122.4194,
			Timezone:    "America/Los_Angeles",
			ISP:         "Cisco OpenDNS",
		}
	default:
		// Try to guess based on IP range patterns
		if strings.HasPrefix(ip, "5.") || strings.HasPrefix(ip, "31.") || strings.HasPrefix(ip, "46.") {
			// European IP ranges
			return &GeolocationResponse{
				IP:          ip,
				City:        "London",
				Region:      "England",
				Country:     "United Kingdom",
				CountryCode: "GB",
				Latitude:    51.5074,
				Longitude:   -0.1278,
				Timezone:    "Europe/London",
				ISP:         "European ISP",
			}
		}
		// Default location (center of map)
		return &GeolocationResponse{
			IP:          ip,
			City:        "Unknown",
			Region:      "Unknown",
			Country:     "Unknown",
			CountryCode: "XX",
			Latitude:    0.0,
			Longitude:   0.0,
			Timezone:    "UTC",
			ISP:         "Internet Service Provider",
		}
	}
}

// APISettings returns the settings page
func (h *WebHandler) APISettings(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user settings
	settings, err := h.settingsService.GetUserSettings(fmt.Sprintf("%d", user.ID))
	if err != nil {
		log.Printf("Error getting user settings: %v", err)
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	data := struct {
		Settings *models.Settings
	}{
		Settings: settings,
	}

	if err := h.templates.ExecuteTemplate(w, "components/settings.html", data); err != nil {
		log.Printf("Error executing settings template: %v", err)
		http.Error(w, "Failed to render settings", http.StatusInternalServerError)
		return
	}
}

// APISettingsScreenshots handles screenshot settings updates
func (h *WebHandler) APISettingsScreenshots(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the enabled parameter - checkbox sends value when checked, nothing when unchecked
	enabledStr := r.FormValue("enabled")
	enabled := enabledStr == "true"

	log.Printf("Screenshot settings update: enabled=%s, parsed=%v", enabledStr, enabled)

	// Update settings
	updates := map[string]interface{}{
		"screenshots_enabled": enabled,
	}

	_, err := h.settingsService.UpdateUserSettings(fmt.Sprintf("%d", user.ID), updates)
	if err != nil {
		log.Printf("Error updating screenshot settings: %v", err)
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated screenshot settings for user %d: enabled=%v", user.ID, enabled)
	w.WriteHeader(http.StatusOK)
}

// APIDetectedNetworks returns detected networks that don't exist in the database
func (h *WebHandler) APIDetectedNetworks(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	detectedNetworks := h.nicIdentifierService.GetDetectedNetworks()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(detectedNetworks); err != nil {
		http.Error(w, "Failed to encode detected networks", http.StatusInternalServerError)
	}
}

// APINetworkSuggestion creates a network from a suggestion
func (h *WebHandler) APINetworkSuggestion(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cidr := strings.TrimSpace(r.FormValue("cidr"))
	if cidr == "" {
		http.Error(w, "CIDR is required", http.StatusBadRequest)
		return
	}

	// Validate CIDR format
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		http.Error(w, "Invalid CIDR format", http.StatusBadRequest)
		return
	}

	// Create network with auto-generated name
	name := fmt.Sprintf("Network %s", cidr)
	description := "Auto-detected network"

	network, err := h.networkService.Create(name, cidr, description)
	if err != nil {
		log.Printf("Failed to create suggested network %s: %v", cidr, err)
		http.Error(w, fmt.Sprintf("Failed to create network: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the event
	h.eventLogService.Log(models.NetworkCreated, fmt.Sprintf("Network %s created from suggestion", cidr), "")

	log.Printf("Created network from suggestion: %s (ID: %s)", cidr, network.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"network": network,
		"message": fmt.Sprintf("Network %s created successfully", cidr),
	})
}

// APIDetectedNetworksDebug returns detected networks without authentication (for testing)
func (h *WebHandler) APIDetectedNetworksDebug(w http.ResponseWriter, r *http.Request) {
	detectedNetworks := h.nicIdentifierService.GetDetectedNetworks()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"detected_networks": detectedNetworks,
		"count":             len(detectedNetworks),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode detected networks", http.StatusInternalServerError)
	}
}

// APINetworksDebug returns all networks in database (for testing)
func (h *WebHandler) APINetworksDebug(w http.ResponseWriter, r *http.Request) {
	networks, err := h.networkService.FindAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get networks: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"existing_networks": networks,
		"count":             len(networks),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode networks", http.StatusInternalServerError)
	}
}
