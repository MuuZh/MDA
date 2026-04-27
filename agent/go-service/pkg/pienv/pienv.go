package pienv

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
)

// PI v2.5.0 environment variable keys.
const (
	EnvInterfaceVersion   = "PI_INTERFACE_VERSION"
	EnvClientName         = "PI_CLIENT_NAME"
	EnvClientVersion      = "PI_CLIENT_VERSION"
	EnvClientLanguage     = "PI_CLIENT_LANGUAGE"
	EnvClientMaaFWVersion = "PI_CLIENT_MAAFW_VERSION"
	EnvVersion            = "PI_VERSION"
	EnvController         = "PI_CONTROLLER"
	EnvResource           = "PI_RESOURCE"
)

// Win32Config holds Win32 controller-specific fields.
type Win32Config struct {
	ClassRegex  string `json:"class_regex,omitempty"`
	WindowRegex string `json:"window_regex,omitempty"`
	Screencap   string `json:"screencap,omitempty"`
	Mouse       string `json:"mouse,omitempty"`
	Keyboard    string `json:"keyboard,omitempty"`
}

// Controller is the parsed PI_CONTROLLER single-line JSON.
type Controller struct {
	Name               string          `json:"name"`
	Label              string          `json:"label,omitempty"`
	Description        string          `json:"description,omitempty"`
	Icon               string          `json:"icon,omitempty"`
	Type               string          `json:"type"`
	DisplayShortSide   *int            `json:"display_short_side,omitempty"`
	DisplayLongSide    *int            `json:"display_long_side,omitempty"`
	DisplayRaw         *bool           `json:"display_raw,omitempty"`
	PermissionRequired bool            `json:"permission_required,omitempty"`
	AttachResourcePath []string        `json:"attach_resource_path,omitempty"`
	Option             []string        `json:"option,omitempty"`
	Win32              *Win32Config    `json:"win32,omitempty"`
	Adb                json.RawMessage `json:"adb,omitempty"`
}

// Resource is the parsed PI_RESOURCE single-line JSON.
type Resource struct {
	Name        string   `json:"name"`
	Label       string   `json:"label,omitempty"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Path        []string `json:"path"`
	Controller  []string `json:"controller,omitempty"`
	Option      []string `json:"option,omitempty"`
}

// Env holds all parsed PI_* environment variables (PI v2.5.0).
type Env struct {
	InterfaceVersion   string
	ClientName         string
	ClientVersion      string
	ClientLanguage     string
	ClientMaaFWVersion string
	Version            string

	Controller    *Controller
	ControllerRaw string
	Resource      *Resource
	ResourceRaw   string
}

var (
	global *Env
	once   sync.Once
)

func doInit() {
	env := &Env{
		InterfaceVersion:   os.Getenv(EnvInterfaceVersion),
		ClientName:         os.Getenv(EnvClientName),
		ClientVersion:      os.Getenv(EnvClientVersion),
		ClientLanguage:     os.Getenv(EnvClientLanguage),
		ClientMaaFWVersion: os.Getenv(EnvClientMaaFWVersion),
		Version:            os.Getenv(EnvVersion),
		ControllerRaw:      os.Getenv(EnvController),
		ResourceRaw:        os.Getenv(EnvResource),
	}

	if env.ControllerRaw != "" {
		var ctrl Controller
		if err := json.Unmarshal([]byte(env.ControllerRaw), &ctrl); err != nil {
			log.Warn().Err(err).
				Str("component", "pienv").
				Str("env_key", EnvController).
				Msg("failed to parse env")
		} else {
			env.Controller = &ctrl
		}
	}

	if env.ResourceRaw != "" {
		var res Resource
		if err := json.Unmarshal([]byte(env.ResourceRaw), &res); err != nil {
			log.Warn().Err(err).
				Str("component", "pienv").
				Str("env_key", EnvResource).
				Msg("failed to parse env")
		} else {
			env.Resource = &res
		}
	}

	global = env

	le := log.Info().
		Str("component", "pienv").
		Str("interface_version", env.InterfaceVersion).
		Str("client_name", env.ClientName).
		Str("client_version", env.ClientVersion).
		Str("client_language", env.ClientLanguage).
		Str("pi_version", env.Version).
		Bool("controller_ok", env.Controller != nil).
		Bool("resource_ok", env.Resource != nil)

	if env.Controller != nil {
		le = le.Str("ctrl_name", env.Controller.Name).
			Str("ctrl_type", env.Controller.Type)
	}
	if env.Resource != nil {
		le = le.Str("res_name", env.Resource.Name)
	}

	le.Msg("PI environment initialized")
}

// Init reads and parses all PI_* environment variables into the global singleton.
func Init() {
	once.Do(doInit)
}

// Get returns the global Env, initializing on first access if needed.
func Get() *Env {
	once.Do(doInit)
	return global
}

// Convenience accessors

// InterfaceVersion returns the PI_INTERFACE_VERSION value.
func InterfaceVersion() string { return Get().InterfaceVersion }

// ClientName returns the PI_CLIENT_NAME value.
func ClientName() string { return Get().ClientName }

// ClientVersion returns the PI_CLIENT_VERSION value.
func ClientVersion() string { return Get().ClientVersion }

// ClientLanguage returns the PI_CLIENT_LANGUAGE value.
func ClientLanguage() string { return Get().ClientLanguage }

// ClientMaaFWVersion returns the PI_CLIENT_MAAFW_VERSION value.
func ClientMaaFWVersion() string { return Get().ClientMaaFWVersion }

// ProjectVersion returns the PI_VERSION value.
func ProjectVersion() string { return Get().Version }

// GetController returns the parsed Controller, or nil if absent.
func GetController() *Controller { return Get().Controller }

// GetResource returns the parsed Resource, or nil if absent.
func GetResource() *Resource { return Get().Resource }

// ControllerType returns the controller type (e.g. "Win32", "Adb"), or empty.
func ControllerType() string {
	if c := GetController(); c != nil {
		return c.Type
	}
	return ""
}

// ControllerName returns the controller name identifier, or empty.
func ControllerName() string {
	if c := GetController(); c != nil {
		return c.Name
	}
	return ""
}

// ResourceName returns the resource name identifier, or empty.
func ResourceName() string {
	if r := GetResource(); r != nil {
		return r.Name
	}
	return ""
}
