package main

type MountCreationError struct {
	arg int
	err error
}

func (e *MountCreationError) Error() string {
	return "Error creating mount point for new namespace: " + e.err.Error()
}
