# cloudlab-template
A template repository for creating projects for CloudLab. 

Use this repository to create new projects for working with CloudLab. It comes packaged with the `rutu-sh/cloudlab-tools` repository as submodule and a Makefile for setting up CloudLab on your local machine. To learn what you can do with the `rutu-sh/cloudlab-tools` repository, check out the [GitHub Repository](https://github.com/rutu-sh/cloudlab-tools).

## Getting Started

1. Click the "Use this template" button to create a new repository from this template.
2. Clone the new repository to your local machine.
3. Configure CloudLab on your local machine. Run the following commands: 

```bash
make cl-setup
```

This creates a `.cloudlab` folder in your project's root directory. Within the folder there is a `cloudlab_config.mk` file. Add your CloudLab username, path to your SSH Key and Node IPs to this file. 

