package main

import (
	"log"
	"os"
	"path"
	"text/template"

	"github.com/ansel1/merry"
)

// backup:
// a=[]; for (let i in localStorage) if (i!='length') a.push([i,localStorage[i]]); a.sort((a,b) => (a+'').localeCompare(b)); h={}; a.forEach(([k,v]) => h[k]=v); copy(JSON.stringify(h, null, "\t"))

func saveIndex(images []*ImageFile) error {
	f, err := os.Create("index.html")
	if err != nil {
		return merry.Wrap(err)
	}
	exPath, err := os.Executable()
	if err != nil {
		return merry.Wrap(err)
	}
	tmpl, err := template.ParseFiles(path.Dir(exPath) + "/template.html")
	if err != nil {
		return merry.Wrap(err)
	}
	err = tmpl.Execute(f, map[string]interface{}{"images": images})
	if err != nil {
		return merry.Wrap(err)
	}
	return merry.Wrap(f.Close())
}

func main() {
	searcher := &ImageSearcher{}
	for i, dirname := range os.Args {
		if i == 0 {
			continue
		}
		if err := searcher.ProcessFolder(dirname); err != nil { // /home/zblzgamer/Pictures/T7/102MSDCF
			log.Fatal(merry.Details(err))
		}
	}
	searcher.Sort()

	if err := saveIndex(searcher.images); err != nil {
		log.Fatal(merry.Details(err))
	}
}
