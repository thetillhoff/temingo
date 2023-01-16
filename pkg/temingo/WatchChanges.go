package temingo

import (
	"log"
	"path"
	"time"

	"github.com/radovskyb/watcher"
)

func WatchChanges(inputDir string, outputDir string, temingoignorePath string) error {
	log.Println("*** Starting to watch for file changes ... ***")

	// ignoring before adding, so the "to-be-ignored" paths won't be added
	w := watcher.New()

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	// If SetMaxEvents is not set, the default is to send all events.
	w.SetMaxEvents(1)

	w.Ignore(outputDir) // ignore the outputfolder

	w.Ignore(".git") // ignore the git-folder natively

	if err := w.AddRecursive(inputDir); err != nil { // watch the source-files-directory recursively
		return err

	}
	if err := w.Add(temingoignorePath); err != nil {
		return err
	}
	// for _, valuesFile := range valuesFilePaths { // for each valuesfilepath
	// 	if err := w.Add(valuesFile); err != nil { // watch the values-file
	// 		log.Fatalln(err)
	// 	}
	// }

	if verbose {
		log.Println("Watched paths/files:")
		// Print a list of all of the files and folders currently being watched and their paths.
		for watchedPath, f := range w.WatchedFiles() {
			log.Println(path.Join(watchedPath, f.Name()))
		}
	}

	go func() {
		for { // while true
			select {
			case event := <-w.Event: // receive events
				log.Println("*** Rebuilding because of a change in", event.Path, "***")
				Render(inputDir, outputDir, temingoignorePath)
				// TODO inform frontend via websocket connection
			case err := <-w.Error: // receive errors
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		return err
	}

	return nil
}
