package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cs3238-tsuzu/popcon-sc/ppjc/client"
)

func encodeFileName(category, name string) string {
	return category + name + category
}

type PPDownloader struct {
	client *ppjc.Client

	dirToSave             string
	encodedNameToFileName map[string]string
	lastUpdateTime        map[string]time.Time
	counter               map[string]int64
	mutex                 sync.RWMutex
	fileNameCounter       int64
}

func NewPPDownloader(client *ppjc.Client, dirToSave string) (*PPDownloader, error) {
	path := filepath.Join(dirToSave, "download")

	if err := os.RemoveAll(path); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}

	return &PPDownloader{
		client:                client,
		dirToSave:             path,
		encodedNameToFileName: make(map[string]string),
		lastUpdateTime:        make(map[string]time.Time),
		counter:               make(map[string]int64),
	}, nil
}

func (ppd *PPDownloader) mutexProcess(f func()) {
	ppd.mutex.Lock()
	f()
	ppd.mutex.Unlock()
}

func (ppd *PPDownloader) mutexProcessReadOnly(f func()) {
	ppd.mutex.RLock()
	f()
	ppd.mutex.RUnlock()
}

func (ppd *PPDownloader) newFileName() string {
	return "ppj_" + strconv.FormatInt(atomic.AddInt64(&ppd.fileNameCounter, 1), 10) + ".txt"
}

func (ppd *PPDownloader) Download(category, name string) (string, *PPLocker, error) {
	enc := encodeFileName(category, name)
	var path string
	ppd.mutexProcess(func() {
		if cnt, ok := ppd.counter[enc]; ok {
			ppd.counter[enc] = cnt + 1
			path = ppd.encodedNameToFileName[enc]
			ppd.lastUpdateTime[enc] = time.Now()
		}
	})

	if len(path) != 0 {
		a, b := ppd.NewLocker(category, name)

		return filepath.Join(ppd.dirToSave, path), a, b
	}

	readCloser, err := ppd.client.FileDownload(category, name)

	if err != nil {
		return "", nil, err
	}
	defer readCloser.Close()

	path = ppd.newFileName()
	fp, err := os.Create(filepath.Join(ppd.dirToSave, path))

	if err != nil {
		return "", nil, err
	}
	defer fp.Close()

	_, err = io.Copy(fp, readCloser)

	if err != nil {
		return "", nil, err
	}

	fp.Close()
	readCloser.Close()

	ppd.mutexProcess(func() {
		if _, ok := ppd.counter[enc]; !ok {
			ppd.counter[enc] = 0
			ppd.encodedNameToFileName[enc] = path
			ppd.lastUpdateTime[enc] = time.Now()
		} else {
			os.Remove(filepath.Join(ppd.dirToSave, path))
			path = ppd.encodedNameToFileName[enc]
		}
	})

	a, b := ppd.NewLocker(category, name)

	return filepath.Join(ppd.dirToSave, path), a, b
}

func (ppd *PPDownloader) RunAutomaticallyDeleter(ctx context.Context, expire time.Duration) {
	if expire < 1*time.Minute {
		expire = 1 * time.Minute
	}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var fileNames []string
				ppd.mutexProcess(func() {
					now := time.Now()
					targets := make([]string, 0, 20)
					for k, v := range ppd.lastUpdateTime {
						if now.Sub(v) > expire {
							if ppd.counter[k] == 0 {
								targets = append(targets, k)
							}
						}
					}

					fileNames = make([]string, len(targets))
					for i := range targets {
						fileNames[i] = ppd.encodedNameToFileName[targets[i]]

						delete(ppd.counter, targets[i])
						delete(ppd.encodedNameToFileName, targets[i])
						delete(ppd.lastUpdateTime, targets[i])
					}
				})

				for i := range fileNames {
					if err := os.Remove(filepath.Join(ppd.dirToSave, fileNames[i])); err != nil {
						FSLog().WithError(err).WithField("path", filepath.Join(ppd.dirToSave, fileNames[i])).Error("file deletion error")
					}
				}
			case <-ctx.Done():
				break
			}
		}
	}()
}

type PPFile struct {
	*os.File
	Category, Name string
	downloader     *PPDownloader
}

func (ppd *PPDownloader) OpenFile(category, name string) (*PPFile, error) {
	var err error
	var fileName string
	ppd.mutexProcess(func() {
		if file, ok := ppd.encodedNameToFileName[encodeFileName(category, name)]; !ok {
			err = os.ErrNotExist
		} else {
			fileName = file
			ppd.counter[encodeFileName(category, name)]++
		}
	})

	if err != nil {
		return nil, err
	}

	fp, err := os.OpenFile(filepath.Join(ppd.dirToSave, fileName), os.O_RDONLY, 0777)
	if err != nil {
		return nil, err
	}

	return &PPFile{
		File:       fp,
		Category:   category,
		Name:       name,
		downloader: ppd,
	}, nil
}

func (f *PPFile) Close() error {
	f.downloader.mutexProcess(func() {
		f.downloader.counter[encodeFileName(f.Category, f.Name)]--
	})

	return f.File.Close()
}

type PPLocker struct {
	unlocker func()
}

// Unlock can be called concurrently
func (locker *PPLocker) Unlock() {
	locker.unlocker()
}

func (ppd *PPDownloader) NewLocker(category, name string) (*PPLocker, error) {
	var err error
	ppd.mutexProcess(func() {
		if cnt, ok := ppd.counter[encodeFileName(category, name)]; ok {
			ppd.counter[encodeFileName(category, name)] = cnt + 1
		} else {
			err = os.ErrNotExist
		}
	})

	if err != nil {
		return nil, err
	}

	var status int32

	unlocker := func() {
		ppd.mutexProcess(func() {
			if atomic.AddInt32(&status, 1) == 1 {
				ppd.counter[encodeFileName(category, name)]--
			}
		})
	}

	return &PPLocker{
		unlocker: unlocker,
	}, nil
}
