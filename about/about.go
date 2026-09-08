package about

import (
	_ "embed"
	"fmt"

	"github.com/Pratyay360/pratyaysh/libs"
)

var AboutMD string

func RenderMarkdown(width int) (string, error) {
	return libs.RenderMarkdown(AboutMD, width)
}

func About() {
	rendered, err := RenderMarkdown(80)
	if err != nil {
		fmt.Println(AboutMD)
		return
	}
	fmt.Println(rendered)
}
