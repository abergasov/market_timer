package stopper

type Stopper interface {
	Stop()
}

var stoppers = make([]Stopper, 0, 10)

func AddStopper(s Stopper) {
	stoppers = append(stoppers, s)
}

func Stop() {
	for i := len(stoppers) - 1; i >= 0; i-- {
		stoppers[i].Stop()
	}
}
