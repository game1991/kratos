package health

type ComponentResult struct {
	Status   Status
	Err error
	Details  map[string]interface{}
}

type Result struct {
	Status     Status
	Components map[string]ComponentResult
}
