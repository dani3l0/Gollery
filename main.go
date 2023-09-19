package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	htmlTemplate "html/template"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/disintegration/imaging"
	"golang.org/x/exp/slices"
)

type Stats struct {
	Images   int
	Size     float32
	Cache    float32
	Free     float32
	LastScan int64
	Scanning bool
}

var stats Stats

const secret string = "mysecret2137"

var secretHash string
var stopScan = false

func auth(w http.ResponseWriter, r *http.Request) bool {
	token, err := r.Cookie("gollerysecret")
	if err != nil || token.Value != secretHash {
		http.Redirect(w, r, "/login.html", http.StatusSeeOther)
		return false
	}
	return true
}

func serveFullImage(w http.ResponseWriter, r *http.Request) {
	if !auth(w, r) {
		return
	}
	http.ServeFile(w, r, strings.Replace(r.URL.Path, "/f/", "images/", 1))
}

func serveThumbnail(w http.ResponseWriter, r *http.Request) {
	if !auth(w, r) {
		return
	}
	http.ServeFile(w, r, strings.Replace(r.URL.Path, "/t/", "cache/", 1))
}

func serveSettings(w http.ResponseWriter, r *http.Request) {
	if !auth(w, r) {
		return
	}
	http.ServeFile(w, r, "html/settings.html")
}

func listAllFiles(dir string) []string {
	var files []string
	realPath, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil
	}
	filepath.Walk(realPath, func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() {
			files = append(files, strings.Replace(path, realPath, dir, 1))
		}
		return nil
	})

	return files
}

func randomFileFrom(dir string) string {
	files := listAllFiles(dir)
	if len(files) > 0 {
		return files[rand.Intn(len(files))]
	}
	return ""
}

func thumbnail(filePath string) bool {
	src, err := imaging.Open(filePath)
	if err != nil {
		fmt.Println("failed to open image:", err)
		return false
	}
	dstThumbnail := imaging.Fill(src, 400, 400, imaging.Center, imaging.NearestNeighbor)
	cache := strings.Replace(filePath, "images", "cache", 1)
	os.MkdirAll(filepath.Dir(cache), 0750)
	err = imaging.Save(dstThumbnail, cache)
	if err != nil {
		fmt.Println("failed to save image:", err)
		return false
	}
	return true
}

func updateLibrary(fullScan bool) {
	if stats.Scanning {
		return
	}
	stopScan = false

	stats = Stats{
		Images:   0,
		Size:     0,
		Cache:    0,
		Free:     0,
		LastScan: 0,
		Scanning: true,
	}

	cached := listAllFiles("cache")
	gallery, err := os.ReadDir("images")
	if err != nil {
		return
	}

	var stored []string
	for _, e := range gallery {
		a := listAllFiles("images/" + e.Name())
		stored = append(stored, a...)
	}

	// Remove unused images in cache
	for e := 0; e < len(cached); e++ {
		f := strings.Replace(cached[e], "cache", "images", 1)
		if !slices.Contains(stored, f) && fullScan {
			os.Remove(cached[e])
		} else {
			stats.Images++
			cacheInfo, _ := os.Stat(cached[e])
			stats.Cache += float32(cacheInfo.Size()) / 1000.0 / 1000.0 / 1000.0
		}
	}

	// Add uncached images
	for e := 0; e < len(stored); e++ {
		if stopScan {
			break
		}
		f := strings.Replace(stored[e], "images", "cache", 1)
		if !slices.Contains(cached, f) && fullScan {
			ok := thumbnail(stored[e])
			if !ok {
				continue
			}
			stats.Images++
			cacheInfo, _ := os.Stat(f)
			stats.Cache += float32(cacheInfo.Size()) / 1000.0 / 1000.0 / 1000.0
		}
		fileInfo, _ := os.Stat(stored[e])
		stats.Size += float32(fileInfo.Size()) / 1000.0 / 1000.0 / 1000.0
	}

	if fullScan {
		stats.LastScan = time.Now().UnixMilli()
	}
	stats.Scanning = false
}

type galleryItem struct {
	Name   string
	Path   string
	Thumb  string
	Items  int
	IsFile bool
	Size   float32
}

type indexItems struct {
	Folders string
	Images  string
}

func galleryMain(w http.ResponseWriter, r *http.Request) {
	if !auth(w, r) {
		return
	}
	dir := strings.Replace(r.URL.Path, "/g/", "images/", 1)
	fi, _ := os.Stat(path.Clean(dir))
	if !fi.IsDir() {
		fullImage := strings.Replace(r.URL.Path, "/g/", "/f/", 1)
		http.Redirect(w, r, fullImage, http.StatusSeeOther)
		return
	}
	entries, _ := os.ReadDir(dir)
	itemT_, _ := os.ReadFile("html/item.html")
	itemT, _ := htmlTemplate.New("galleryItem").Parse(string(itemT_))
	T_, _ := os.ReadFile("html/gallery.html")
	T, _ := template.New("index").Parse(string(T_))

	var folders_ bytes.Buffer
	var images_ bytes.Buffer

	for _, e := range entries {
		name := e.Name()
		filePath := path.Clean(dir + "/" + name)
		isFile := !e.IsDir()
		thumb := filePath
		if !isFile {
			thumb = randomFileFrom(filePath)
		}
		thumb = path.Clean("/t/" + strings.Replace(thumb, "images", "", 1))
		item := galleryItem{
			Name:   name,
			Path:   name + "/",
			Thumb:  thumb,
			IsFile: isFile,
			Items:  0,
			Size:   20.0,
		}
		if isFile {
			itemT.Execute(&images_, item)
		} else {
			itemT.Execute(&folders_, item)
		}
	}

	result := indexItems{Folders: folders_.String(), Images: images_.String()}
	T.Execute(w, result)

}

func settingsApi(w http.ResponseWriter, r *http.Request) {
	if !auth(w, r) {
		return
	}

	var stat syscall.Statfs_t
	syscall.Statfs("images", &stat)
	free := float32(stat.Bavail*uint64(stat.Bsize)) / 1000.0 / 1000.0 / 1000.0
	stats.Free = float32(free)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func scan(w http.ResponseWriter, r *http.Request) {
	if !auth(w, r) {
		return
	}
	go updateLibrary(true)
}

func scanStop(w http.ResponseWriter, r *http.Request) {
	if !auth(w, r) {
		return
	}
	stopScan = true
}

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	sentSecret := r.Form.Get("secret")
	s_ := sha256.New()
	s_.Write([]byte(sentSecret))
	sentHash := hex.EncodeToString(s_.Sum(nil))
	if sentHash == secretHash {
		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "gollerysecret", Value: sentHash, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(-1)
	cookie := http.Cookie{Name: "gollerysecret", Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	// Init library
	os.MkdirAll("cache", 0750)
	os.MkdirAll("images", 0750)
	updateLibrary(false)

	// Init secret token
	secretHash_ := sha256.New()
	secretHash_.Write([]byte(secret))
	secretHash = hex.EncodeToString(secretHash_.Sum(nil))

	// Handle functions
	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", logout)
	http.HandleFunc("/f/", serveFullImage)
	http.HandleFunc("/t/", serveThumbnail)
	http.HandleFunc("/g/", galleryMain)
	http.HandleFunc("/settings", serveSettings)
	http.HandleFunc("/settings/api", settingsApi)
	http.HandleFunc("/settings/scan", scan)
	http.HandleFunc("/settings/scanStop", scanStop)
	http.Handle("/", http.FileServer(http.Dir("html")))
	http.ListenAndServe(":8080", nil)
}
