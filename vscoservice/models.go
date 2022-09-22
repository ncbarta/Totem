package vscoservice

import "sync"

//// VSCO

type VSCOSite struct {
	SiteID               int    `json:"id"`
	SiteCollectionID     string `json:"site_collection_id"`
	ResponsiveURL        string `json:"responsive_url"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	UserID               int    `json:"user_id"`
	ProfileImageURL      string `json:"profile_image"`
	ProfileImageID       string `json:"profile_image_id"`
	RecentlyPublishedURL string `json:"recently_published"`
}

type VSCOMedia struct {
	ID                  string    `json:"_id"`
	SiteID              int       `json:"site_id"`
	PermaSubdomain      string    `json:"perma_subdomain"`
	Description         string    `json:"description"`
	CaptureDate         int64     `json:"capture_date"`
	UploadDate          int64     `json:"upload_date"`
	LastUpdated         int64     `json:"last_updated"`
	LocationCoordinates []float64 `json:"location_coords"`
	HasLocation         bool      `json:"has_location"`
	IsVideo             bool      `json:"is_video"`
	VideoURL            string    `json:"video_url"`
	ResponsiveURL       string    `json:"responsive_url"`
	ImageMetadata       struct {
		Make        string `json:"make"`
		Model       string `json:"model"`
		Orientation int    `json:"orientation"`
	} `json:"image_meta"`
	Tags []struct {
		Text string `json:"text"`
		Slug string `json:"slug"`
	} `json:"tags"`

	// /Collections Media
	CollectedDate int64 `json:"collected_date"`
	Favorited     bool  `json:"favorited"`
	Blacklisted   bool  `json:"blacklisted"`
}

//// INTERNAl

// VSCO tracking profile:
// - Multiple accounts can be owned by one person
// - You can reserve usernames in case that person creates an account
// - The SiteID's get saved, in case the person changes their name
type VSCOTrackingProfile struct {
	// The name of the person's whose accounts you are tracking
	TargetName string `json:"target_name"`

	// Currently tracking this target
	Active bool

	// Accounts you are tracking.
	Accounts []VSCOAccount
}

type VSCOAccount struct {
	UserID   int    `json:"user_id"`
	SiteID   int    `json:"site_id"`
	Username string `json:"username"`
}

type MediaData struct {
	Bytes     []byte
	Filename  string
	VSCOMedia VSCOMedia
}

type DataSink struct {
	wg        sync.WaitGroup
	semaphore chan int
	data      chan MediaData
}

// The User's bio, along with when it was recorded.
type BioRecord struct {
	Description string `json:"description"`
	CaptureDate int64  `json:"capture_date"`
}
