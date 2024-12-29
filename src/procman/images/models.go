package images

type Image struct {
	Id             string `yaml:"id"`
	Name           string `yaml:"name"`
	ContextTempDir string `yaml:"context_temp_dir,omitempty"`
	ImgPath        string `yaml:"imgpath,omitempty"`
	Tag            string `yaml:"tag,omitempty"`
	Created        string `yaml:"created,omitempty"`
}

// image spec structs

type Step struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Source      string   `yaml:"source,omitempty"`
	Destination string   `yaml:"destination,omitempty"`
	Workdir     string   `yaml:"workdir,omitempty"`
	Command     []string `yaml:"command,omitempty"`
}

type Job struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Command []string `yaml:"command"`
}

type ImageSpec struct {
	Base  string `yaml:"base"`
	Steps []Step `yaml:"steps"`
	Job   Job    `yaml:"job"`
}
