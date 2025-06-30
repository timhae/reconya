package web

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"reconya-ai/internal/config"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/systemstatus"
	"reconya-ai/models"
)

// Templates will be loaded from filesystem for now
// TODO: Embed templates in production build

type WebHandler struct {
	deviceService       *device.DeviceService
	eventLogService     *eventlog.EventLogService
	networkService      *network.NetworkService
	systemStatusService *systemstatus.SystemStatusService
	templates           *template.Template
	sessionStore        *sessions.CookieStore
	config              *config.Config
}

type PageData struct {
	Page         string
	User         *models.User
	Error        string
	Username     string
	Devices      []*models.Device
	EventLogs    []*models.EventLog
	SystemStatus *models.SystemStatus
	NetworkMap   *NetworkMapData
}

type NetworkMapData struct {
	BaseIP         string
	IPRange        []int
	Devices        map[string]*models.Device
	NetworkInfo    *NetworkInfo
}

type NetworkInfo struct {
	OnlineDevices  int
	IdleDevices    int
	OfflineDevices int
}

func NewWebHandler(
	deviceService *device.DeviceService,
	eventLogService *eventlog.EventLogService,
	networkService *network.NetworkService,
	systemStatusService *systemstatus.SystemStatusService,
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
				return ""
			}
			switch v := ptr.(type) {
			case *string:
				if v == nil {
					return ""
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

	// Parse templates from filesystem
	tmpl := template.New("").Funcs(funcMap)
	
	// Parse templates with unique names to avoid conflicts
	baseFiles, err := filepath.Glob("templates/layouts/*.html")
	if err != nil {
		panic(fmt.Sprintf("Failed to glob base templates: %v", err))
	}
	
	pageFiles, err := filepath.Glob("templates/pages/*.html")
	if err != nil {
		panic(fmt.Sprintf("Failed to glob page templates: %v", err))
	}
	
	componentFiles, err := filepath.Glob("templates/components/*.html")
	if err != nil {
		panic(fmt.Sprintf("Failed to glob component templates: %v", err))
	}

	indexFile := "templates/index.html"
	
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
		deviceService:       deviceService,
		eventLogService:     eventLogService,
		networkService:      networkService,
		systemStatusService: systemStatusService,
		templates:           tmpl,
		sessionStore:        store,
		config:              config,
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

	systemStatus, err := h.systemStatusService.GetLatest()
	if err != nil {
		log.Printf("Error getting system status for home page: %v", err)
		// Fallback to mock data or handle gracefully
		systemStatus = &models.SystemStatus{
			NetworkID: "N/A",
			PublicIP:  nil,
		}
	}

	devicesSlice, err := h.deviceService.FindAll()
	if err != nil {
		log.Printf("Error getting devices for home page: %v", err)
		devicesSlice = []models.Device{} // Ensure it's an empty slice, not nil
	}
	
	// Convert to pointer slice for template
	devices := make([]*models.Device, len(devicesSlice))
	for i := range devicesSlice {
		devices[i] = &devicesSlice[i]
	}

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

	data := PageData{
		Page:         "dashboard",
		User:         user,
		SystemStatus: systemStatus,
		Devices:      devices,
		EventLogs:    eventLogs,
	}

	if err := h.templates.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Use standalone login template to avoid conflicts
		loginTmpl, err := template.ParseFiles("templates/standalone/login.html")
		if err != nil {
			log.Printf("Failed to parse standalone login template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		data := PageData{Page: "login"}
		log.Printf("Executing standalone login template with data: %+v", data)
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
		data := PageData{
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

	devicesSlice, err := h.deviceService.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter to show only idle and online devices
	var devices []*models.Device
	for i := range devicesSlice {
		if devicesSlice[i].Status == models.DeviceStatusOnline || devicesSlice[i].Status == models.DeviceStatusIdle {
			devices = append(devices, &devicesSlice[i])
		}
	}

	viewMode := r.URL.Query().Get("view")
	data := struct {
		Devices  []*models.Device
		ViewMode string
	}{
		Devices:  devices,
		ViewMode: viewMode,
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

	if err := h.templates.ExecuteTemplate(w, "components/device-modal.html", device); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APIUpdateDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["id"]

	hostname := r.FormValue("hostname")
	comment := r.FormValue("comment")

	var hostnamePtr, commentPtr *string
	if hostname != "" {
		hostnamePtr = &hostname
	}
	if comment != "" {
		commentPtr = &comment
	}

	device, err := h.deviceService.UpdateDevice(deviceID, hostnamePtr, commentPtr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "components/device-modal.html", device); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APISystemStatus(w http.ResponseWriter, r *http.Request) {
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

	devicesSlice, err := h.deviceService.FindAll()
	if err != nil {
		log.Printf("Error getting devices for system status: %v", err)
		devicesSlice = []models.Device{} // Ensure it's an empty slice, not nil
	}

	devices := make([]*models.Device, len(devicesSlice))
	for i := range devicesSlice {
		devices[i] = &devicesSlice[i]
	}

	networkMapData := h.buildNetworkMap(devices)

	// Get the network CIDR if we have a NetworkID
	var networkCIDR string = "N/A"
	if status != nil && status.NetworkID != "" {
		network, err := h.networkService.FindByID(status.NetworkID)
		if err == nil && network != nil {
			networkCIDR = network.CIDR
		}
	}

	data := struct {
		SystemStatus *models.SystemStatus
		NetworkCIDR  string
		NetworkInfo  *NetworkInfo
		DevicesCount int
	}{
		SystemStatus: status,
		NetworkCIDR:  networkCIDR,
		NetworkInfo:  networkMapData.NetworkInfo,
		DevicesCount: len(devices),
	}

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

func (h *WebHandler) APINetworkMap(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "reconya-session")
	user := h.getUserFromSession(session)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	devicesSlice, err := h.deviceService.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		case "online":
			online++
		case "idle":
			idle++
		default:
			offline++
		}
	}

	// Parse network CIDR from config
	baseIP, ipRange := h.parseNetworkCIDR(h.config.NetworkCIDR)

	return &NetworkMapData{
		BaseIP:  baseIP,
		IPRange: ipRange,
		Devices: deviceMap,
		NetworkInfo: &NetworkInfo{
			OnlineDevices:  online,
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
	// Same as APIDevices - targets are devices
	devicesSlice, err := h.deviceService.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	devices := make([]*models.Device, len(devicesSlice))
	for i := range devicesSlice {
		devices[i] = &devicesSlice[i]
	}

	viewMode := r.URL.Query().Get("view")
	data := struct {
		Devices  []*models.Device
		ViewMode string
	}{
		Devices:  devices,
		ViewMode: viewMode,
	}

	if err := h.templates.ExecuteTemplate(w, "components/device-grid.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) APITrafficCore(w http.ResponseWriter, r *http.Request) {
	devicesSlice, err := h.deviceService.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Show all devices (online, idle, and offline) with visual indicators
	devices := make([]*models.Device, len(devicesSlice))
	for i := range devicesSlice {
		devices[i] = &devicesSlice[i]
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
	devicesSlice, err := h.deviceService.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Show all devices (online, idle, and offline) with visual indicators
	devices := make([]*models.Device, len(devicesSlice))
	for i := range devicesSlice {
		devices[i] = &devicesSlice[i]
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