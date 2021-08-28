package benchcheck

// GetModule will download a specific version of a module and
// return a directory where you can find the module code.
// It leverages "go get" to do the job, so the returned directory
// should be considered read only (it is part of the Go cache).
//
// A non-nil error is returned if it fails.
func GetModule(name string, version string) (string, error) {
	return "", nil
}
