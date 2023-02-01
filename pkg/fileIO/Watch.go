package fileIO

import (
	"io/fs"
	"log"
	"path"
	"time"

	"github.com/radovskyb/watcher"
)

// Will watch all included paths recusively for changes regularly
// and run the provided actionEvent
func Watch(inclusionPaths []string, exclusionPaths []string, verbose bool, interval time.Duration, actionEvent func(watcher.Event) error) error {
	var (
		err         error
		w           *watcher.Watcher
		f           fs.FileInfo
		watchedPath string
	)

	// ignoring before adding, so the "to-be-ignored" paths won't be added
	w = watcher.New()

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	// If SetMaxEvents is not set, the default is to send all events.
	w.SetMaxEvents(1)

	for _, exclusionPath := range exclusionPaths { // Set exclusions before inclusions, so they will be already taken into consideration
		w.Ignore(exclusionPath)
	}

	for _, recursivePath := range inclusionPaths {
		if err = w.AddRecursive(recursivePath); err != nil {
			return err
		}
	}

	if verbose {
		log.Println("Watched paths/files:")
		// Print a list of all of the files and folders currently being watched and their paths.
		for watchedPath, f = range w.WatchedFiles() {
			log.Println(path.Join(watchedPath, f.Name()))
		}
	}

	go func() {
		for { // while true
			select {
			case event := <-w.Event: // receive events
				err = actionEvent(event)
				if err != nil {
					log.Fatalln(err)
				}
			case err = <-w.Error: // receive errors
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Start the watching process - it'll check for changes every time the interval is over.
	if err = w.Start(interval); err != nil {
		return err
	}

	return nil
}
