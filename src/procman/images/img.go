package images

import (
	"fmt"
	"os"
	"procman/common"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func getImageContextDir(img_id string) string {
	return fmt.Sprintf("/var/lib/procman/img/%v/rootfs", img_id)
}

func getImageDir(img_id string) string {
	return fmt.Sprintf("/var/lib/procman/img/%v", img_id)
}

func getParentImgDir() string {
	return "/var/lib/procman/img"
}

/*
Takes the image name and the absolute path to the setup script as argument and builds a tar.gz with dependencies injected.

	Args:
		- name string: the name of the image
		- img_context string: the absolute path to the image context dir

	Returns:
		- Image object with attributes containing details for the container image
		- ImageBuildErr
*/
func BuildImage(name string, img_context string, tag string) (Image, common.ImageBuildErr) {

	_logger := common.GetLogger()
	_logger.Info().Msgf("creating image '%v' using setup script %v", name, img_context)

	img_id := strings.Split(uuid.New().String(), "-")[0]
	img := Image{
		Id:             img_id,
		Name:           name,
		ContextTempDir: getImageContextDir(img_id),
		Tag:            tag,
		ImgPath:        getImageDir(img_id),
	}

	if err := os.MkdirAll(img.ContextTempDir, 0755); err != nil {
		return Image{}, common.ImageBuildErr{
			Code:    500,
			Message: fmt.Sprintf("context creation failed with error: %v", err),
		}
	}

	img, err := buildImage(img, img_context)
	if err.Code != 0 {
		return Image{}, err
	}

	return img, common.ImageBuildErr{Code: 0, Message: "success"}
}

func ListImages() ([]Image, common.ImageListErr) {

	_logger := common.GetLogger()

	_logger.Info().Msgf("listing images")

	img_dir := getParentImgDir()

	images := []Image{}

	_, err := os.Stat(img_dir)
	if !os.IsExist(err) {
		dirs, err := os.ReadDir(img_dir)
		if err != nil {
			_logger.Error().Msgf("error reading dir: %v", err)
			return []Image{}, common.ImageListErr{Code: 500, Message: fmt.Sprintf("error reading dir: %v", err)}
		}
		for _, dir := range dirs {
			dirName := dir.Name()
			imgMetadataFile := fmt.Sprintf("%v/%v/img.yaml", img_dir, dirName)
			if _, err := os.Stat(imgMetadataFile); err != nil {
				_logger.Error().Msgf("error statfile: %v", err)
				continue
			}

			data, err := os.ReadFile(imgMetadataFile)
			if err != nil {
				_logger.Error().Msgf("error readfile: %v", err)
				continue
			}
			parsedImgMetadata := Image{}
			err = yaml.Unmarshal(data, &parsedImgMetadata)

			if err != nil {
				_logger.Error().Msgf("error unmarshall: %v", err)
				continue
			}

			images = append(images, parsedImgMetadata)
		}
	}
	return images, common.ImageListErr{Code: 0, Message: "success"}
}
