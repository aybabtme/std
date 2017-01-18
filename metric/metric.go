package metric

import "time"

type Node interface {
	Counter(name, desc string, labels ...string) Counter
	Sampler(name, desc string, min, max float64, labels ...string) Sampler
	Lbl(k, v string) Node
}

type Counter func(float64, ...string)
type Sampler func(float64, ...string)

func StopWatch(sampler Sampler) func(string) {
	start := time.Now()
	return func(step string) {
		sampler(time.Since(start).Seconds(), step)
	}
}
