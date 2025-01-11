package requestsgo

import (
	"errors"
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

// DownloadQueueResult holds the information of a completed or failed download
type DownloadQueueResult struct {
	url string
	dst string
	err error
}

// URL returns the URL
func (err DownloadQueueResult) URL() string { return err.url }

// Dst returns the destination
func (err DownloadQueueResult) Dst() string { return err.dst }

// Error returns the error
func (err DownloadQueueResult) Error() error { return err.err }

// Failed returns if the download failed
func (err DownloadQueueResult) Failed() bool { return err.err != nil }

// Succeeded returns if the download succeeded
func (err DownloadQueueResult) Succeeded() bool { return err.err == nil }

// DownloadQueueConfig holds the config for the DownloadQueue
type DownloadQueueConfig struct {
	WorkerCount     uint
	RecordSuccesses bool
	RecordFailures  bool
	Timeout         time.Duration
}

var (
	DefaultDownloadQueueConfig = DownloadQueueConfig{
		WorkerCount:     4,
		RecordSuccesses: false,
		RecordFailures:  true,
		Timeout:         0,
	}
)

// DownloadQueue manages a queue and workers for downloading files
type DownloadQueue struct {
	cfg      DownloadQueueConfig
	waitG    sync.WaitGroup
	jobs     []downloadQueueJob
	mJobs    sync.Mutex
	results  []DownloadQueueResult
	mResults sync.Mutex
	closed   bool
	mClosed  sync.Mutex
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
					queue.addResult(job, err)

					if cfg.Timeout > 0 {
						time.Sleep(cfg.Timeout)
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

func (queue *DownloadQueue) addResult(job downloadQueueJob, err error) {
	queue.mResults.Lock()
	defer queue.mResults.Unlock()

	if (queue.cfg.RecordSuccesses && err == nil) ||
		(queue.cfg.RecordFailures && err != nil) {
		queue.results = append(queue.results, DownloadQueueResult{
			url: job.URL,
			dst: job.Dst,
			err: err,
		})
	}
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

// Results returns the results collected by the queue
func (queue *DownloadQueue) Results() []DownloadQueueResult {
	queue.mResults.Lock()
	defer queue.mResults.Unlock()

	var errors []DownloadQueueResult

	for _, result := range queue.results {
		if result.Failed() {
			errors = append(errors, result)
		}
	}

	return errors
}

// Successes returns the uccesses collected by the queue
func (queue *DownloadQueue) Successes() []DownloadQueueResult {
	queue.mResults.Lock()
	defer queue.mResults.Unlock()

	var errors []DownloadQueueResult

	for _, result := range queue.results {
		if result.Succeeded() {
			errors = append(errors, result)
		}
	}

	return errors
}

// Errors returns the errors collected by the queue
func (queue *DownloadQueue) Errors() []DownloadQueueResult {
	queue.mResults.Lock()
	defer queue.mResults.Unlock()

	var errors []DownloadQueueResult

	for _, result := range queue.results {
		if result.Failed() {
			errors = append(errors, result)
		}
	}

	return errors
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
