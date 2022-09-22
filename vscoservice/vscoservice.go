package vscoservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type VSCOService struct {
	Username string
	Site     VSCOSite

	// Path: ./Totem/{TrackingProfile.TargetName}
	TargetDir string

	// Path: ./Totem/{TrackingProfile.TargetName}/{Account}
	AccountDir string

	// Path: ./Totem/{TrackingProfile.TargetName}/{Account}/Profile
	ProfileDir string

	// Path: ./Totem/{TrackingProfile.TargetName}/{Account}/Gallery
	GalleryDir string

	// Path: ./Totem/{TrackingProfile.TargetName}/{Account}/Collection
	CollectionDir string
}

var (
	httpClient = http.DefaultClient
)

const (
	concurrentMediaDownloads = 3 // Turning this too high will get you rate limited.
)

func New(account *VSCOAccount, targetPath string) VSCOService {

	n := VSCOService{
		Username:      account.Username,
		Site:          VSCOSite{},
		TargetDir:     targetPath,
		AccountDir:    targetPath + "/" + account.Username,
		ProfileDir:    targetPath + "/" + account.Username + "/Profile",
		GalleryDir:    targetPath + "/" + account.Username + "/Gallery",
		CollectionDir: targetPath + "/" + account.Username + "/Collection",
	}
	n.setSite(account)

	if n.Site.Name != "" {
		// Create profile directory
		if _, err := os.Stat(n.ProfileDir); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(n.ProfileDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		// Create gallery directory
		if _, err := os.Stat(n.GalleryDir); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(n.GalleryDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		// Create collection directory
		if _, err := os.Stat(n.CollectionDir); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(n.CollectionDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
	}

	return n
}

// Sets the VSCOService site & updates VSCOAccount information if necessary
func (v *VSCOService) setSite(account *VSCOAccount) {
	if account.SiteID != -1 {
		request, err := http.NewRequest("GET", "https://vsco.co/api/2.0/sites/"+strconv.Itoa(account.SiteID), nil)
		if err != nil {
			panic(err)
		}
		setInfoRequestHeaders(request)

		response, err := httpClient.Do(request)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		var VSCOSiteShim struct {
			Site VSCOSite `json:"site"`
		}

		err = json.Unmarshal(body, &VSCOSiteShim)
		if err != nil {
			panic(err)
		}

		if VSCOSiteShim.Site.Name != "" {
			v.Site = VSCOSiteShim.Site
			account.UserID = v.Site.UserID

			// Update directory name if Username has changed
			if account.Username != v.Site.Name {
				account.Username = v.Site.Name
				v.Username = account.Username
				err := os.Rename(v.AccountDir, v.TargetDir+"/"+v.Username)
				if err != nil {
					panic(err)
				}

				// Update all of the paths
				v.AccountDir = v.TargetDir + "/" + v.Username
				v.ProfileDir = v.AccountDir + "/Profile"
				v.GalleryDir = v.AccountDir + "/Gallery"
				v.CollectionDir = v.AccountDir + "/Collection"
			}
		} else {
			fmt.Println("User", account.Username, "has deleted their account.")
		}
	} else {
		request, err := http.NewRequest("GET", "https://vsco.co/api/2.0/sites?subdomain="+account.Username, nil)
		if err != nil {
			panic(err)
		}
		setInfoRequestHeaders(request)

		response, err := httpClient.Do(request)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		var VSCOSiteShim struct {
			Sites []VSCOSite `json:"sites"`
		}

		err = json.Unmarshal(body, &VSCOSiteShim)
		if err != nil {
			panic(err)
		}

		if len(VSCOSiteShim.Sites) > 0 {
			v.Site = VSCOSiteShim.Sites[0]

			account.SiteID = v.Site.SiteID
			account.UserID = v.Site.UserID
		} else {
			fmt.Println("User", account.Username, "does not currently exist.")
		}
	}
}

// Records changes to the user's bio
func (v *VSCOService) CheckBio() {
	if v.Site.Name == "" {
		return
	}

	var bioRecords []BioRecord

	bioRecordFile := v.ProfileDir + "/.BioRecord.json"

	// Creates/Replaces .BioRecord.json file
	update := func() {
		bytes, err := json.Marshal(bioRecords)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(bioRecordFile, bytes, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	// Initialize file if it does not exist, and return
	if _, err := os.Stat(bioRecordFile); errors.Is(err, os.ErrNotExist) {
		bioRecords = []BioRecord{
			{v.Site.Description, time.Now().Unix()},
		}
		update()
		return
	}

	// Otherwise, read it & update it if it has changed
	bytes, err := os.ReadFile(bioRecordFile)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &bioRecords)
	if err != nil {
		panic(err)
	}

	bioRecordsLen := len(bioRecords)
	if bioRecordsLen > 0 {
		if v.Site.Description != bioRecords[bioRecordsLen-1].Description {
			bioRecords = append(bioRecords, BioRecord{v.Site.Description, time.Now().Unix()})
			update()
		}
	}
}

func (v *VSCOService) CheckProfileImage() {
	if v.Site.Name == "" {
		return
	}

	// Read contents of /Profile
	entries := getEntriesDictForDir(v.ProfileDir)

	// If the current profileImageID is not in /Profile, download it.
	if !isInEntries(v.Site.ProfileImageID, entries) {
		imageRequest, err := http.NewRequest("GET", "https://"+v.Site.ResponsiveURL, nil)
		if err != nil {
			panic(err)
		}
		setMediaRequestHeaders(imageRequest)

		imageResponse, err := httpClient.Do(imageRequest)
		if err != nil {
			panic(err)
		}
		defer imageResponse.Body.Close()

		data, err := io.ReadAll(imageResponse.Body)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(v.ProfileDir+"/"+v.Site.ProfileImageID+".jpg", data, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

// Collects all of the media on the user's profile /Gallery that has not been stored locally
// Stored by upload date
func (v *VSCOService) CheckGalleryMedia() {
	if v.Site.Name == "" {
		return
	}

	entries := getEntriesDictForDir(v.GalleryDir)

	sink := DataSink{
		semaphore: make(chan int, concurrentMediaDownloads),
		data:      make(chan MediaData),
	}

	var page = 1
	for {
		media := v.galleryMediaRequest(page)

		if len(media) == 0 {
			break
		}

		for _, m := range media {
			t := parseMilliTimestamp(m.UploadDate)
			ts := t.Format("Mon, Jan 2, 15h04m05s, MST 2006")

			if !isInEntries(ts, entries) {
				sink.wg.Add(1)
				go func(m VSCOMedia, ts string) {
					sink.semaphore <- 1
					defer func() {
						<-sink.semaphore
						entries[ts] = struct{}{}
						sink.wg.Done()
					}()

					var mediaRequest *http.Request
					var err error

					if m.IsVideo {
						mediaRequest, err = http.NewRequest("GET", "http://"+m.VideoURL, nil)
						if err != nil {
							panic(err)
						}
					} else {
						mediaRequest, err = http.NewRequest("GET", "http://"+m.ResponsiveURL, nil)
						if err != nil {
							panic(err)
						}
					}

					setMediaRequestHeaders(mediaRequest)

					mediaResponse, err := httpClient.Do(mediaRequest)
					if err != nil {
						panic(err)
					}
					defer mediaResponse.Body.Close()

					data, err := io.ReadAll(mediaResponse.Body)
					if err != nil {
						panic(err)
					}
					sink.data <- MediaData{data, ts, m}
				}(m, ts)
			}
		}
		page++
	}

	go func() {
		sink.wg.Wait()
		close(sink.data)
	}()

	for md := range sink.data {
		mediaDir := v.GalleryDir + "/" + md.Filename

		os.Mkdir(mediaDir, os.ModePerm)

		extension := ".jpg"
		if md.VSCOMedia.IsVideo {
			extension = ".mp4"
		}

		// Write image/video
		err := os.WriteFile(mediaDir+"/"+md.VSCOMedia.ID+extension, md.Bytes, os.ModePerm)
		if err != nil {
			panic(err)
		}

		jsonBytes, err := json.Marshal(md.VSCOMedia)
		if err != nil {
			panic(err)
		}

		// Write info file
		err = os.WriteFile(mediaDir+"/"+"Info.json", jsonBytes, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func (v *VSCOService) galleryMediaRequest(page int) []VSCOMedia {
	request, err := http.NewRequest("GET", "https://vsco.co/api/2.0/medias?site_id="+strconv.Itoa(v.Site.SiteID)+"&size=30"+"&page="+strconv.Itoa(page), nil)
	if err != nil {
		panic(err)
	}
	setMediaRequestHeaders(request)

	response, err := httpClient.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var VSCOMediaShim struct {
		Media []VSCOMedia `json:"media"`
	}

	err = json.Unmarshal(body, &VSCOMediaShim)
	if err != nil {
		panic(err)
	}

	return VSCOMediaShim.Media
}

// Collects all of the media on the user's profile /Gallery that has not been stored locally
// Stored by collected date
func (v *VSCOService) CheckCollectionMedia() {
	if v.Site.Name == "" {
		return
	}

	entries := getEntriesDictForDir(v.CollectionDir)

	sink := DataSink{
		semaphore: make(chan int, concurrentMediaDownloads),
		data:      make(chan MediaData),
	}

	var page = 1
	for {
		media := v.collectionMediaRequest(page)

		if len(media) == 0 {
			break
		}

		for _, m := range media {
			t := parseMilliTimestamp(m.UploadDate + m.CollectedDate)
			ts := t.Format("Mon, Jan 2, 15h04m05s, MST 2006")

			if !isInEntries(ts, entries) {
				sink.wg.Add(1)
				go func(m VSCOMedia, ts string) {
					sink.semaphore <- 1
					defer func() {
						<-sink.semaphore
						entries[ts] = struct{}{}
						sink.wg.Done()
					}()

					var mediaRequest *http.Request
					var err error

					if m.IsVideo {
						mediaRequest, err = http.NewRequest("GET", "http://"+m.VideoURL, nil)
						if err != nil {
							panic(err)
						}
					} else {
						mediaRequest, err = http.NewRequest("GET", "http://"+m.ResponsiveURL, nil)
						if err != nil {
							panic(err)
						}
					}

					setMediaRequestHeaders(mediaRequest)

					mediaResponse, err := httpClient.Do(mediaRequest)
					if err != nil {
						panic(err)
					}
					defer mediaResponse.Body.Close()

					data, err := io.ReadAll(mediaResponse.Body)
					if err != nil {
						panic(err)
					}
					sink.data <- MediaData{data, ts, m}
				}(m, ts)

			}
		}
		page++
	}

	go func() {
		sink.wg.Wait()
		close(sink.data)
	}()

	for md := range sink.data {
		mediaDir := v.CollectionDir + "/" + md.Filename

		os.Mkdir(mediaDir, os.ModePerm)

		extension := ".jpg"
		if md.VSCOMedia.IsVideo {
			extension = ".mp4"
		}

		// Write image/video
		err := os.WriteFile(mediaDir+"/"+md.VSCOMedia.ID+extension, md.Bytes, os.ModePerm)
		if err != nil {
			panic(err)
		}

		jsonBytes, err := json.Marshal(md.VSCOMedia)
		if err != nil {
			panic(err)
		}

		// Write info file
		err = os.WriteFile(mediaDir+"/"+"Info.json", jsonBytes, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func (v *VSCOService) collectionMediaRequest(page int) []VSCOMedia {
	request, err := http.NewRequest("GET", "https://vsco.co/api/2.0/collections/"+v.Site.SiteCollectionID+"/medias?size=30"+"&page="+strconv.Itoa(page), nil)
	if err != nil {
		panic(err)
	}
	setMediaRequestHeaders(request)

	response, err := httpClient.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var VSCOMediaShim struct {
		Media []VSCOMedia `json:"medias"`
	}

	err = json.Unmarshal(body, &VSCOMediaShim)
	if err != nil {
		panic(err)
	}

	return VSCOMediaShim.Media
}

// Prints out the user's bio records & optionally checks if there is a new one.
func (v *VSCOService) PrintBio(withUpdate bool) {
	bioRecordFile := v.ProfileDir + "/.BioRecord.json"
	var bioRecords []BioRecord

	if withUpdate {
		v.CheckBio()
	}

	if _, err := os.Stat(bioRecordFile); errors.Is(err, os.ErrNotExist) {
		fmt.Println("This user has no bio record")
		return
	}

	bytes, err := os.ReadFile(bioRecordFile)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &bioRecords)
	if err != nil {
		panic(err)
	}

	fmt.Println("User:", v.Username)

	for _, r := range bioRecords {
		t := time.Unix(r.CaptureDate, 0)
		ts := t.Format("Mon, Jan 2, 15h04m05s, MST 2006")

		fmt.Println(ts, "|", r.Description)
	}
}
