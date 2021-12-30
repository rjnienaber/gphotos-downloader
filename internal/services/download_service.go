package services

import (
	"github.com/rjnienaber/gphotos_downloader/internal/database"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
	"github.com/rjnienaber/gphotos_downloader/pkg/workerpool"
)

type RetryTracker interface {
	ShouldRetry(err error) bool
	Wait()
}

type RetryFactory interface {
	Create() RetryTracker
}

type DownloaderQueuer interface {
	QueueDownload(ids ...string)
}

type DownloadService struct {
	api          googlephotos.Downloader
	db           database.PhotoDatabase
	logger       utils.Logger
	queue        *workerpool.JobQueue
	retryFactory RetryFactory
	rootDir      string
	maxWorkers   int
}

type Option func(svc *DownloadService)

func NewDownloadService(api googlephotos.Downloader, db database.PhotoDatabase, rootDir string, opts ...Option) DownloadService {
	service := DownloadService{api: api, db: db, rootDir: rootDir, logger: utils.NewLogger(utils.Silent)}
	for _, opt := range opts {
		opt(&service)
	}

	if service.retryFactory == nil {
		service.retryFactory = NoRetryFactory{}
	}

	service.queue = workerpool.NewJobQueue(service.maxWorkers)
	service.queue.Start()

	return service
}

func (s *DownloadService) QueueDownload(ids ...string) {
	for _, id := range ids {
		job := DownloadJob{Id: id, api: s.api, db: s.db, logger: s.logger, rootDir: s.rootDir, retryFactory: s.retryFactory}
		s.queue.Submit(&job)
	}
}

func (s *DownloadService) Finish() {
	s.queue.Stop()
}

func WithRetryFactory(factory RetryFactory) Option {
	return func(service *DownloadService) {
		service.retryFactory = factory
	}
}

func WithMaxWorkers(maxWorkers int) Option {
	return func(service *DownloadService) {
		service.maxWorkers = maxWorkers
	}
}

func WithLogger(logger utils.Logger) Option {
	return func(service *DownloadService) {
		service.logger = logger
	}
}
