package health

type ComponentResult struct {
	Name     string
	Status   Status
	ErrorMsg string
	Details  map[string]interface{}
}

type Result struct {
	Status  Status
	Details []ComponentResult
}
