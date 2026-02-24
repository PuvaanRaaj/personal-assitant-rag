package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuvaanRaaj/personal-rag-agent/internal/logger"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/service"
	"github.com/fsnotify/fsnotify"
)

// Watcher monitors a local directory for changes
type Watcher struct {
	path            string
	userID          string
	documentService *service.DocumentService
	watcher         *fsnotify.Watcher
}

// NewWatcher creates a new watcher service
func NewWatcher(path, userID string, documentService *service.DocumentService) (*Watcher, error) {
	// Create folder if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create knowledge base directory: %w", err)
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	return &Watcher{
		path:            path,
		userID:          userID,
		documentService: documentService,
		watcher:         fsWatcher,
	}, nil
}

// Start begins monitoring the directory
func (w *Watcher) Start(ctx context.Context) error {
	// Add root path and subdirectories recursively
	err := filepath.Walk(w.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return w.watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk knowledge base path: %w", err)
	}

	logger.Info("Watcher started", "path", w.path, "user_id", w.userID)

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				// Process write and create events
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					// Check if it's a file we care about
					info, err := os.Stat(event.Name)
					if err != nil {
						continue
					}
					if info.IsDir() {
						// Add new directory to watcher
						w.watcher.Add(event.Name)
						continue
					}

					// Process new file with a small delay to ensure write is complete
					go func(path string) {
						time.Sleep(500 * time.Millisecond)
						logger.Info("Processing file change", "file", path)
						_, err := w.documentService.ProcessLocalFile(context.Background(), w.userID, path)
						if err != nil {
							logger.Error("Failed to process local file", "file", path, "error", err)
						} else {
							logger.Info("Successfully indexed local file", "file", path)
						}
					}(event.Name)
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				logger.Error("Watcher error", "error", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Sync performs a full scan of the directory
func (w *Watcher) Sync(ctx context.Context) error {
	logger.Info("Starting manual sync of knowledge base", "path", w.path)
	
	err := filepath.Walk(w.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only process supported files
		ext := strings.ToLower(filepath.Ext(path))
		allowedTypes := map[string]bool{
			".pdf": true, ".txt": true, ".md": true,
			".json": true, ".csv": true,
		}
		if !allowedTypes[ext] {
			return nil
		}

		logger.Info("Syncing file", "file", path)
		_, err = w.documentService.ProcessLocalFile(ctx, w.userID, path)
		if err != nil {
			// If it's already there or other errors, log and continue
			logger.Debug("Sync skipped file", "file", path, "reason", err.Error())
		} else {
			logger.Info("Sync indexed file", "file", path)
		}

		return nil
	})

	return err
}

// Close stops the watcher
func (w *Watcher) Close() error {
	return w.watcher.Close()
}
