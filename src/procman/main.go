package main

import (
	"fmt"
	"os"
	"procman/common"
	"procman/images"
	"text/tabwriter"
)

func listImages() common.ImageListErr {
	images, err := images.ListImages()
	if err.Code != 0 {
		return err
	}

	// id name tag created

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)

	fmt.Fprintln(writer, "Id\tName\tTag\tCreated\t")
	for _, img := range images {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t\n", img.Id, img.Name, img.Tag, img.Created)
	}
	writer.Flush()
	return common.ImageListErr{Code: 0, Message: "success"}
}

func main() {
	// _logger := common.GetLogger()
	// _, err := images.BuildImage("test-1", "/users/rutu_sh/src/linux-ns/image-setup-scripts/alpine-basic", "0.0.1")
	// if err.Code != 0 {
	// 	_logger.Error().Msgf("error: %v", err)
	// 	os.Exit(1)
	// }
	listImages()
}
