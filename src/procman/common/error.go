package common

import "fmt"

type ImageBuildErr struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func (e *ImageBuildErr) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

type ImageListErr struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func (e *ImageListErr) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}
