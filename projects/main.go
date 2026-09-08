package projects

import (
	_ "embed"
	"fmt"

	"charm.land/glamour/v2"
	"github.com/Pratyay360/pratyaysh/libs"
)

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
	glamour.Render(rendered, err.Error())
}
