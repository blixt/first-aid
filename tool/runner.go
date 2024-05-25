package tool

var (
	// NopRunner is a runner that does nothing extra.
	NopRunner = NewRunner(func(status string) {})
)

type Runner interface {
	Report(status string)
}

type runner struct {
	report func(status string)
}

// NewRunner returns a new Runner. Tools run with this Runner will report status
// updates to the provided function.
func NewRunner(report func(status string)) Runner {
	return &runner{report: report}
}

func (r *runner) Report(status string) {
	r.report(status)
}
