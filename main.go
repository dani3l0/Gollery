package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	htmlTemplate "html/template"
	"image"
	"image/jpeg"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/image/draw"
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

func serveFullImage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, strings.Replace(r.URL.Path, "/f/", "images/", 1))
}

func serveThumbnail(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, strings.Replace(r.URL.Path, "/t/", "cache/", 1))
}

func serveSettings(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/settings.html")
}

func listAllFiles(dir string) []string {
	var files []string
	realPath, _ := filepath.EvalSymlinks(dir)
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

func thumbnail(filePath string) {
	inputFile, _ := os.Open(filePath)
	defer inputFile.Close()

	inputImage, _, _ := image.Decode(inputFile)
	b := inputImage.Bounds()
	w, h := float32(b.Dx()), float32(b.Dy())
	divideBy := 1.0 / (440.0 / w)

	w = w / divideBy
	h = h / divideBy
	thumbnail := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	draw.ApproxBiLinear.Scale(thumbnail, thumbnail.Bounds(), inputImage, inputImage.Bounds(), draw.Over, nil)

	cache := strings.Replace(filePath, "images", "cache", 1)
	os.MkdirAll(filepath.Dir(cache), 0750)
	outputFile, _ := os.Create(cache)
	defer outputFile.Close()
	jpeg.Encode(outputFile, thumbnail, &jpeg.Options{Quality: 60})
}

func updateLibrary(fullScan bool) {
	if stats.Scanning {
		return
	}
	stats = Stats{
		Images:   0,
		Size:     0,
		Cache:    0,
		Free:     0,
		LastScan: 0,
		Scanning: true,
	}
	cached := listAllFiles("cache")
	gallery, _ := os.ReadDir("images")
	var stored []string
	for _, e := range gallery {
		for _, f := range listAllFiles("images/" + e.Name()) {
			stored = append(stored, f)
		}
	}

	// Remove unused images in cache
	for e := 0; e < len(cached); e++ {
		f := strings.Replace(cached[e], "cache", "images", 1)
		if !slices.Contains(stored, f) && fullScan {
			fmt.Print("Removing: ")
			fmt.Println(cached[e])
			os.Remove(cached[e])
		} else {
			stats.Images++
			cacheInfo, _ := os.Stat(cached[e])
			stats.Cache += float32(cacheInfo.Size()) / 1000.0 / 1000.0 / 1000.0
		}
	}

	// Add uncached images
	for e := 0; e < len(stored); e++ {
		f := strings.Replace(stored[e], "images", "cache", 1)
		if !slices.Contains(cached, f) && fullScan {
			fmt.Print("Generating: ")
			fmt.Println(stored[e])
			thumbnail(stored[e])
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
	Items string
}

func galleryMain(w http.ResponseWriter, r *http.Request) {
	dir := strings.Replace(r.URL.Path, "/g/", "images/", 1)
	entries, _ := os.ReadDir(dir)
	itemT_, _ := os.ReadFile("html/item.html")
	itemT, _ := htmlTemplate.New("galleryItem").Parse(string(itemT_))
	T_, _ := os.ReadFile("html/gallery.html")
	T, _ := template.New("index").Parse(string(T_))

	var tpl bytes.Buffer

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
		itemT.Execute(&tpl, item)
	}

	result := indexItems{Items: tpl.String()}
	T.Execute(w, result)

}

func settingsApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func scan(w http.ResponseWriter, r *http.Request) {
	go updateLibrary(true)
}

func main() {
	os.MkdirAll(filepath.Dir("cache"), 0750)
	os.MkdirAll(filepath.Dir("images"), 0750)
	updateLibrary(false)
	http.HandleFunc("/f/", serveFullImage)
	http.HandleFunc("/t/", serveThumbnail)
	http.HandleFunc("/g/", galleryMain)
	http.HandleFunc("/settings", serveSettings)
	http.HandleFunc("/settings/api", settingsApi)
	http.HandleFunc("/settings/scan", scan)
	http.Handle("/", http.FileServer(http.Dir("html")))
	http.ListenAndServe(":8080", nil)
}
