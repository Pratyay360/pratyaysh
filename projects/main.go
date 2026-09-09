package projects

import (
	_ "embed"
	"fmt"

	"github.com/Pratyay360/pratyaysh/libs"
)

//go:embed project.md
var ProjectMD string

func RenderMarkdown(width int) (string, error) {
	return libs.RenderMarkdown(ProjectMD, width)
}

func ListProjects() {
	rendered, err := RenderMarkdown(80)
		if err != nil {
		fmt.Println(ProjectMD)
		return
	}
	fmt.Println(rendered)
}
