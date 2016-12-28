package jobdata

// Data is all of the meta data for a job, but not defined by steps
type Data struct {
	Name string
	ID   int

	LocalPath string
	Origin    string

	TimesRun int
}
