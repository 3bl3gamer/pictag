package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ansel1/merry"
	"github.com/nfnt/resize"
	"github.com/rwcarlsen/goexif/exif"
)

type ImageFile struct {
	Path          string    `json:"-"`
	RelativePath  string    `json:"relativePath"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"createdAt"`
	Stamp         string    `json:"stamp"`
	Key           string    `json:"key"`
	Thumb         string    `json:"-"`
	RelativeThumb string    `json:"relativeThumb"`
}

type ImagesDir struct {
	Path   string
	Prefix string
	Suffix string
}

func NewImageFile(fpath, relativeFpath, thumbDir string, createdAt time.Time) *ImageFile {
	fname := filepath.Base(fpath)
	stamp := createdAt.Local().Format("2006-01-02 15:04:05")
	key := createdAt.In(time.UTC).Format("2006-01-02T15:04:05Z") + " " + fname
	thumbName := createdAt.In(time.UTC).Format("2006-01-02_15-04-05") +
		"_" + strings.Replace(relativeFpath, "/", "_", -1)
	thumb := thumbDir + "/" + thumbName
	return &ImageFile{fpath, relativeFpath, fname, createdAt, stamp, key, thumb, thumbName}
}

type ImageCacheInfo struct {
	CreatedAt time.Time
}

type ImageSearcher struct {
	images     []*ImageFile
	imagesDirs []*ImagesDir
	cache      map[string]*ImageCacheInfo
	cacheDir   string
	thumbDir   string
}

func NewImageSearcher(cacheDir string) *ImageSearcher {
	return &ImageSearcher{cacheDir: cacheDir, thumbDir: cacheDir + "/thumbnails"}
}

func (s *ImageSearcher) ProcessFolder(dirpath string) error {
	if err := s.loadCache(); err != nil {
		return merry.Wrap(err)
	}
	dirpath = filepath.Clean(dirpath)
	imagesDir := &ImagesDir{Path: dirpath, Prefix: filepath.Dir(dirpath), Suffix: filepath.Base(dirpath)}
	s.imagesDirs = append(s.imagesDirs, imagesDir)

	thumbChan, thumbWG := s.startThumbRoutines()

	stt := time.Now()
	imgCount := 0
	newImgCount := 0
	err := filepath.Walk(imagesDir.Path, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return merry.Wrap(err)
		}
		if !strings.HasPrefix(fpath, imagesDir.Path) {
			return merry.New("unexpected file path: " + fpath + ", need it no be in " + imagesDir.Path)
		}
		relativeFpath := fpath[len(imagesDir.Prefix)+1:]
		fmt.Printf("\r\033[0K%s, %.1f fps", relativeFpath, float64(imgCount)/float64(time.Now().Sub(stt)/time.Second))

		var imgFile *ImageFile
		cachedInfo, ok := s.cache[relativeFpath]
		if ok {
			imgFile = NewImageFile(fpath, relativeFpath, s.thumbDir, cachedInfo.CreatedAt)
		} else if !info.IsDir() && s.shouldProcessFile(fpath) {
			caStr, err := s.createdAt(fpath)
			if err != nil {
				return merry.Wrap(err)
			}
			if caStr == "" {
				caStr = "1970:01:01 00:00:00"
			}
			createdAt, err := time.Parse("2006:01:02 15:04:05", caStr)
			if err != nil {
				return merry.Wrap(err)
			}
			imgFile = NewImageFile(fpath, relativeFpath, s.thumbDir, createdAt)
			s.cache[relativeFpath] = &ImageCacheInfo{createdAt}
			newImgCount++

			if newImgCount%100 == 0 {
				if err := s.saveCache(); err != nil {
					return merry.Wrap(err)
				}
			}
		}

		if imgFile != nil {
			s.images = append(s.images, imgFile)
			thumbChan <- imgFile
			imgCount++
		}

		return nil
	})
	print("\033[0K")
	if err != nil {
		return merry.Wrap(err)
	}
	if err := merry.Wrap(s.saveCache()); err != nil {
		return merry.Wrap(err)
	}

	close(thumbChan)
	thumbWG.Wait()
	return nil
}

func (s *ImageSearcher) Sort() {
	sort.Slice(s.images, func(i, j int) bool {
		return s.images[i].CreatedAt.After(s.images[j].CreatedAt)
	})
}

func (s *ImageSearcher) shouldProcessFile(fpath string) bool {
	fpath = strings.ToLower(fpath)
	if !strings.HasSuffix(fpath, ".jpeg") && !strings.HasSuffix(fpath, ".jpg") {
		return false
	}
	if filepath.Base(filepath.Dir(fpath)) == "thumbnails" {
		return false
	}
	return true
}

func (s *ImageSearcher) createdAt(fpath string) (string, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return "", merry.Wrap(err)
	}

	x, err := exif.Decode(f)
	if err == io.EOF {
		return "", nil
	}
	if err != nil {
		log.Print("[WARN] processing " + fpath + ": " + err.Error())
		return "", nil
	}

	return s.createdAtFromExif(x), nil
}

func (s *ImageSearcher) createdAtFromExif(x *exif.Exif) string {
	dto, _ := x.Get("DateTimeOriginal")
	dtd, _ := x.Get("DateTimeDigitized")
	min := ""
	if dto != nil {
		min, _ = dto.StringVal()
	}
	if dtd != nil {
		dtdStr, _ := dtd.StringVal()
		if dtdStr != "" && dtdStr < min {
			min = dtdStr
		}
	}
	return min
}

func (s *ImageSearcher) loadCache() error {
	f, err := os.Open(s.cacheDir + "/cache.json")
	if os.IsNotExist(err) {
		s.cache = make(map[string]*ImageCacheInfo)
		return nil
	}
	if err != nil {
		return merry.Wrap(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&s.cache); err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func (s *ImageSearcher) saveCache() error {
	f, err := os.Create(s.cacheDir + "/cache_tmp.json")
	if err != nil {
		return merry.Wrap(err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(s.cache); err != nil {
		return merry.Wrap(err)
	}
	if err := os.Rename(s.cacheDir+"/cache_tmp.json", s.cacheDir+"/cache.json"); err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func (s ImageSearcher) updateThumb(imgFile *ImageFile) error {
	if _, err := os.Stat(imgFile.Thumb); !os.IsNotExist(err) {
		return nil
	}
	if err := os.MkdirAll(path.Dir(imgFile.Thumb), 0755); err != nil {
		return merry.Wrap(err)
	}

	fThumb, err := os.Create(imgFile.Thumb + "_tmp")
	if err != nil {
		return merry.Wrap(err)
	}
	defer fThumb.Close()

	fSrc, err := os.Open(imgFile.Path)
	if err != nil {
		return merry.Wrap(err)
	}
	defer fSrc.Close()

	var imgThumb image.Image
	img, _, err := image.Decode(fSrc)
	if _, ok := err.(jpeg.FormatError); ok {
		fmt.Printf("\r\033[0KWARN: broken image: %s (%s)\n", imgFile.Path, err)
		imgThumb = s.makeBrokenImageThumb(imgFile, fThumb)
	} else if err != nil {
		return merry.Wrap(err)
	} else {
		imgThumb = resize.Thumbnail(192, 192, img, resize.Bicubic)
	}

	if err := jpeg.Encode(fThumb, imgThumb, &jpeg.Options{Quality: 80}); err != nil {
		return merry.Wrap(err)
	}
	if err := os.Rename(imgFile.Thumb+"_tmp", imgFile.Thumb); err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func (s *ImageSearcher) makeBrokenImageThumb(imgFile *ImageFile, f *os.File) image.Image {
	img := image.NewYCbCr(image.Rect(0, 0, 192, 192), image.YCbCrSubsampleRatio420)
	for i := range img.Y {
		img.Y[i] = 0
	}
	for i := range img.Cr {
		if (i%16 < 8) != ((i/96)%16 < 8) {
			img.Cr[i] = 255
			img.Cb[i] = 255
		} else {
			img.Cr[i] = 150
			img.Cb[i] = 0
		}
	}
	return img
}

func (s *ImageSearcher) startThumbRoutines() (chan *ImageFile, *sync.WaitGroup) {
	n := runtime.NumCPU()
	c := make(chan *ImageFile, n)
	wg := &sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			for imgFile := range c {
				if err := s.updateThumb(imgFile); err != nil {
					panic(merry.Details(merry.Prepend(err, imgFile.Path)))
				}
			}
			wg.Done()
		}()
	}
	return c, wg
}
