package executor

type FakeExecutor struct {
	Handler func(string, ...string) error
}

func (e *FakeExecutor) Execute(name string, args ...string) error {
	return e.Handler(name, args...)
}

func NewFakeExecutor(h func(string, ...string) error) *FakeExecutor {
	return &FakeExecutor{
		Handler: h,
	}
}
