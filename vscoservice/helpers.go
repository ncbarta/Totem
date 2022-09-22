package vscoservice

import (
	"net/http"
	"os"
	"time"
)

// MISC

func getEntriesDictForDir(dir string) map[string]struct{} {
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	s := make(map[string]struct{})
	for _, e := range entries {
		s[e.Name()] = struct{}{}
	}
	return s
}

func isInEntries(entry string, entries map[string]struct{}) bool {
	if _, doesContain := entries[entry]; doesContain {
		return true
	}
	return false
}

func parseMilliTimestamp(tm int64) time.Time {
	sec := tm / 1000
	msec := tm % 1000
	return time.Unix(sec, msec*int64(time.Millisecond))
}

// REQUEST

// A request for static data
func setInfoRequestHeaders(request *http.Request) {
	request.Header.Add("Content-Type", "*")
	request.Header.Set("Host", "vsco.co")
	request.Header.Set("Referer", "http://vsco.co/bob/images/1")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Accept-Encoding", "gzip, deflate")
	request.Header.Set("Accept-Language", "en-US,er;q=0.9")
	request.Header.Set("Authorization", "Bearer 7356455548d0a1d886db010883388d08be84d0c9")
}

// A request for media data
func setMediaRequestHeaders(request *http.Request) {
	setInfoRequestHeaders(request)
	request.Header.Set("X-Client-Build", "1")
	request.Header.Set("X-Client-Platform", "web")
}
