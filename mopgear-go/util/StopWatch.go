package util

import "time"

type Stopwatch struct {
	running     bool
	startTime   time.Time
	accumulated time.Duration
}

func StopwatchMakeStopped() *Stopwatch {
	return &Stopwatch{}
}

func StopwatchMakeStarted() *Stopwatch {
	return &Stopwatch{running: true, startTime: time.Now()}
}

func (sw *Stopwatch) Start() {
	if sw.running {
		panic("Timer already started")
	}
	sw.startTime = time.Now()
	sw.running = true
}

func (sw *Stopwatch) Stop() {
	if !sw.running {
		panic("Timer not running")
	}
	sw.accumulated += time.Since(sw.startTime)
	sw.startTime = time.Time{}
	sw.running = false
}

func (sw *Stopwatch) Elapsed() time.Duration {
	if sw.running {
		total := sw.accumulated + time.Since(sw.startTime)
		return total
	} else {
		return sw.accumulated
	}
}

func (sw *Stopwatch) AddElapsedFrom(other *Stopwatch) {
	sw.accumulated += other.Elapsed()
}

type StopwatchNoisy struct {
	Stopwatch
	printer *PrintRecorder
}

func StopwatchNoisyStart(printer *PrintRecorder) *StopwatchNoisy {
	sw := &StopwatchNoisy{
		*StopwatchMakeStarted(),
		printer,
	}
	printer.Println("Started at " + sw.startTime.Format(time.DateTime))
	return sw
}

func (sw *StopwatchNoisy) Stop() {
	sw.Stopwatch.Stop()
	sw.printer.Println("Duration = " + sw.Elapsed().String())
	sw.printer.Println("Finished at " + time.Now().Format(time.DateTime))
}
