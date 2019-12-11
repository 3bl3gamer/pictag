package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/ansel1/merry"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var imageDirpaths StringsFlag
	flag.Var(&imageDirpaths, "images", "path to directory with images, flag may be specified several times")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.ErrorStackMarshaler = func(err error) interface{} { return merry.Details(err) }
	zerolog.ErrorStackFieldName = "message" //TODO: https://github.com/rs/zerolog/issues/157
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05.000"})

	if len(imageDirpaths) == 0 {
		log.Fatal().Msg("no image directories, nothing to do")
	}
	for i, dirpath := range imageDirpaths {
		dirpath, err := filepath.Abs(dirpath)
		if err != nil {
			log.Fatal().Msg(merry.Details(err))
		}
		imageDirpaths[i] = dirpath
	}

	cacheDir, err := MakeCacheDir()
	if err != nil {
		log.Fatal().Msg(merry.Details(err))
	}

	configDir, err := MakeConfigDir()
	if err != nil {
		log.Fatal().Msg(merry.Details(err))
	}

	searcher := NewImageSearcher(cacheDir)
	for _, dirpath := range imageDirpaths {
		if err := searcher.ProcessFolder(dirpath); err != nil {
			log.Fatal().Msg(merry.Details(err))
		}
	}
	searcher.Sort()

	tags, err := loadImagesTags(configDir)
	if err != nil {
		log.Fatal().Msg(merry.Details(err))
	}

	if err := StartHTTPServer("127.0.0.1:9008", Env{"dev"}, searcher, tags); err != nil {
		log.Fatal().Msg(merry.Details(err))
	}
}
