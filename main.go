package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type Config struct {
	ListenAddress string
	SavegameDir   string
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	root := cobra.Command{
		Use:   "satisfactory_latest_savegame",
		Short: "A tool to find the latest savegame for Satisfactory",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := Config{
				ListenAddress: cmd.Flag("listen.address").Value.String(),
				SavegameDir:   cmd.Flag("savegame.dir").Value.String(),
			}
			startServer(cfg)
		},
	}
	root.Flags().String("listen.address", ":8080", "Address to listen on")
	root.Flags().String("savegame.dir", "/home/steam/.config/Epic/FactoryGame/Saved/SaveGames/server/", "Directory to search for savegames")

	err := root.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("executing command failed")
	}
}

func startServer(cfg Config) {
	mux := setupMux(cfg.SavegameDir)

	log.Info().Str("address", cfg.ListenAddress).Str("dir", cfg.SavegameDir).Msg("starting server")
	err := http.ListenAndServe(cfg.ListenAddress, mux)
	if err != nil {
		log.Fatal().Err(err).Msg("starting server failed")
	}
}

func setupMux(savegameDir string) (mux *http.ServeMux) {
	mux = http.NewServeMux()
	mux.HandleFunc("/latest", errWrapper(handleLatest(savegameDir)))

	return mux
}

func errWrapper(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			if os.IsNotExist(err) {
				log.Info().Err(err).Msg("no savegame found")
				http.Error(w, "No savegame found", http.StatusNotFound)
				return
			}

			log.Error().Err(err).Msg("handling request failed")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func handleLatest(savegameDir string) func(w http.ResponseWriter, r *http.Request) (err error) {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		reader, info, err := openLatestSavegame(savegameDir)
		if err != nil {
			return err
		}
		defer func() {
			if err := reader.Close(); err != nil {
				log.Warn().Err(err).Msg("closing savegame failed. ignoring...")
			}
		}()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("ETag", buildSavegameETag(info))
		http.ServeContent(w, r, info.Name(), info.ModTime(), reader)
		return nil
	}
}

func openLatestSavegame(savegameDir string) (*os.File, os.FileInfo, error) {
	latest, err := findLatestSavegame(savegameDir)
	if err != nil {
		return nil, nil, err
	}
	log.Debug().Str("path", latest).Msg("found latest savegame")

	file, err := os.OpenFile(latest, os.O_RDONLY, 0)
	if err != nil {
		return nil, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}

	return file, info, nil
}

func buildSavegameETag(info os.FileInfo) string {
	return fmt.Sprintf(`"%s-%x-%x"`, info.Name(), info.Size(), info.ModTime().UnixNano())
}

func findLatestSavegame(savegameDir string) (string, error) {
	files, err := os.ReadDir(savegameDir)
	if err != nil {
		return "", err
	}

	var latest string
	var latestTime int64
	for _, file := range files {
		if file.IsDir() || path.Ext(file.Name()) != ".sav" {
			continue
		}
		info, err := file.Info()
		if err != nil {
			log.Warn().Err(err).Str("file", file.Name()).Msg("getting file info failed. ignoring...")
			continue
		}
		modTime := info.ModTime().Unix()
		if modTime > latestTime {
			latest = path.Join(savegameDir, file.Name())
			latestTime = modTime
		}
	}
	if latest == "" {
		log.Info().Msg("no savegame found")
		return "", os.ErrNotExist
	}
	return latest, nil
}
