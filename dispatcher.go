package dispatcher

// Dispatcher define
// inspired by http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/
type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job

	maxWorkers int
	maxJobs    int
	JobQueue   chan Job
}

// NewDispatcher create dispatcher
func NewDispatcher(maxWorkers, maxJobs int) *Dispatcher {
	return &Dispatcher{
		WorkerPool: make(chan chan Job, maxWorkers),
		maxWorkers: maxWorkers,
		maxJobs:    maxJobs,
		JobQueue:   make(chan Job, maxJobs),
	}
}

// Add job to queue
func (d *Dispatcher) Add(job Job) bool {
	select {
	case d.JobQueue <- job:
		return true
	default:
		return false
	}
}

// SyncAdd sync add job, if job queue full will block
func (d *Dispatcher) SyncAdd(job Job) {
	d.JobQueue <- job
}

// Run dispatcher
func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

// Close dispatcher
func (d *Dispatcher) Close() {}

// CurrentJobs current job queue
func (d *Dispatcher) CurrentJobs() int {
	return len(d.JobQueue)
}

// CurrentWorkers current workers
func (d *Dispatcher) CurrentWorkers() int {
	return len(d.WorkerPool)
}

// MaxWorkers max workers
func (d *Dispatcher) MaxWorkers() int {
	return d.maxWorkers
}

// MaxJobs max jobs
func (d *Dispatcher) MaxJobs() int {
	return d.maxJobs
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.JobQueue:
			// try to obtain a worker job channel that is available.
			// this will block until a worker is idle
			jobChannel := <-d.WorkerPool

			// dispatch the job to the worker job channel
			jobChannel <- job
		}
	}
}
