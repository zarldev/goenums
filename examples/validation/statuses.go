package validation

func (s Status) IsCompleted() bool {
	return s == Statuses.FAILED || s == Statuses.PASSED || s == Statuses.SKIPPED
}

func (s Status) IsRunning() bool {
	return s == Statuses.SCHEDULED || s == Statuses.RUNNING
}
