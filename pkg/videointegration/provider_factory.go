package videointegration

import (
	"log"
	"os"
	"strings"
)

// ProviderFactory creates video providers based on configuration
type ProviderFactory struct {
	logger    *log.Logger
	providers map[Provider]VideoProvider
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(logger *log.Logger) *ProviderFactory {
	return &ProviderFactory{
		logger:    logger,
		providers: make(map[Provider]VideoProvider),
	}
}

// InitializeProviders initializes all configured providers based on environment variables
func (f *ProviderFactory) InitializeProviders() error {
	// Always register Jitsi (free, no credentials needed)
	jitsiURL := os.Getenv("JITSI_SERVER_URL")
	if jitsiURL == "" {
		jitsiURL = "https://meet.jit.si"
	}
	f.providers[ProviderJitsi] = NewJitsiProvider(jitsiURL, f.logger)
	f.logger.Printf("Registered Jitsi provider (server: %s)", jitsiURL)

	// Register Google Meet if credentials are available
	if os.Getenv("GOOGLE_CLIENT_ID") != "" && os.Getenv("GOOGLE_REFRESH_TOKEN") != "" {
		googleProvider, err := NewGoogleMeetProviderFull(f.logger)
		if err != nil {
			f.logger.Printf("Warning: Failed to initialize Google Meet provider: %v", err)
		} else {
			f.providers[ProviderGoogleMeet] = googleProvider
			f.logger.Println("Registered Google Meet provider")
		}
	} else {
		f.logger.Println("Google Meet provider not configured (missing credentials)")
	}

	// Register Zoom if credentials are available
	if os.Getenv("ZOOM_ACCOUNT_ID") != "" && os.Getenv("ZOOM_CLIENT_ID") != "" {
		zoomProvider, err := NewZoomProvider(f.logger)
		if err != nil {
			f.logger.Printf("Warning: Failed to initialize Zoom provider: %v", err)
		} else {
			f.providers[ProviderZoom] = zoomProvider
			f.logger.Println("Registered Zoom provider")
		}
	} else {
		f.logger.Println("Zoom provider not configured (missing credentials)")
	}

	return nil
}

// GetProvider returns a provider by type
func (f *ProviderFactory) GetProvider(providerType Provider) (VideoProvider, bool) {
	provider, ok := f.providers[providerType]
	return provider, ok
}

// GetProviderByString returns a provider by string name
func (f *ProviderFactory) GetProviderByString(name string) (VideoProvider, bool) {
	return f.GetProvider(Provider(strings.ToLower(name)))
}

// GetDefaultProvider returns the default provider (Jitsi)
func (f *ProviderFactory) GetDefaultProvider() VideoProvider {
	return f.providers[ProviderJitsi]
}

// GetAvailableProviders returns a list of available provider types
func (f *ProviderFactory) GetAvailableProviders() []Provider {
	types := make([]Provider, 0, len(f.providers))
	for pt := range f.providers {
		types = append(types, pt)
	}
	return types
}

// IsProviderAvailable checks if a provider is available
func (f *ProviderFactory) IsProviderAvailable(providerType Provider) bool {
	_, ok := f.providers[providerType]
	return ok
}

// RegisterAllWithService registers all providers with the video integration service
func (f *ProviderFactory) RegisterAllWithService(service *Service) {
	for pt, provider := range f.providers {
		service.RegisterProvider(pt, provider)
	}
}
