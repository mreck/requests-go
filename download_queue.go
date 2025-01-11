package requestsgo

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrQueueClosed = errors.New("download queue is closed")
)

type downloadQueueJob struct {
	URL string
	Dst string
}

// DownloadQueueError hold the information of a queued download that failed
type DownloadQueueError struct {
	url string
	dst string
	err error
}

// URL returns the URL
func (err DownloadQueueError) URL() string { return err.url }

// Dst returns the destination
func (err DownloadQueueError) Dst() string { return err.dst }

// Err returns the error
func (err DownloadQueueError) Err() error { return err.err }

// Error returns the formated error string
func (err DownloadQueueError) Error() string {
	return fmt.Sprintf("downloading %s -> %s failed: %s", err.url, err.dst, err.err)
}

// DownloadQueueConfig holds the config for the DownloadQueue
type DownloadQueueConfig struct {
	WorkerCount uint
}

// DownloadQueue manages a queue and workers for downloading files
type DownloadQueue struct {
	cfg     DownloadQueueConfig
	waitG   sync.WaitGroup
	jobs    []downloadQueueJob
	mJobs   sync.Mutex
	errors  []DownloadQueueError
	mErrors sync.Mutex
	closed  bool
	mClosed sync.Mutex
}

// CreateDownloadQueue creates a new DownloadQueue
func (client *Client) CreateDownloadQueue(cfg DownloadQueueConfig) *DownloadQueue {
	queue := &DownloadQueue{cfg: cfg}

	for i := 0; i < int(cfg.WorkerCount); i++ {
		go func() {
			for {
				job, hasNext := queue.nextJob()

				if !hasNext {
					if queue.IsClosed() {
						queue.waitG.Done()
						return
					}
					time.Sleep(200 * time.Millisecond)
				} else {
					err := client.GetDownload(job.URL, job.Dst)
					if err != nil {
						queue.addError(job, err)
					}
				}
			}
		}()

		queue.waitG.Add(1)
	}

	return queue
}

func (queue *DownloadQueue) nextJob() (downloadQueueJob, bool) {
	queue.mJobs.Lock()
	defer queue.mJobs.Unlock()

	if len(queue.jobs) == 0 {
		return downloadQueueJob{}, false
	}

	var job downloadQueueJob
	job, queue.jobs = queue.jobs[0], queue.jobs[1:]

	return job, true
}

func (queue *DownloadQueue) addError(job downloadQueueJob, err error) {
	queue.mErrors.Lock()
	defer queue.mErrors.Unlock()

	queue.errors = append(queue.errors, DownloadQueueError{url: job.URL, dst: job.Dst, err: err})
}

// IsClosed return if the DownloadQueue is closed
func (queue *DownloadQueue) IsClosed() bool {
	queue.mClosed.Lock()
	defer queue.mClosed.Unlock()

	return queue.closed
}

// Close closes the DownloadQueue
func (queue *DownloadQueue) Close() {
	queue.mClosed.Lock()
	defer queue.mClosed.Unlock()

	queue.closed = true
}

// Errors returns the errors that occured in the DownloadQueue
func (queue *DownloadQueue) Errors() []DownloadQueueError {
	queue.mErrors.Lock()
	defer queue.mErrors.Unlock()

	return queue.errors
}

// Enqueue adds a new download to the queue
func (queue *DownloadQueue) Enqueue(url string, dst string) error {
	if queue.IsClosed() {
		return ErrQueueClosed
	}

	queue.mJobs.Lock()
	defer queue.mJobs.Unlock()

	queue.jobs = append(queue.jobs, downloadQueueJob{URL: url, Dst: dst})

	return nil
}

// WaitUntilDone closes the DownloadQueue and waits all workers are done
func (queue *DownloadQueue) WaitUntilDone() {
	queue.Close()
	queue.waitG.Wait()
}
