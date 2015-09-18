package common

import (
	ned "github.com/ensonmj/NeoEditor/backend"
	"github.com/ensonmj/NeoEditor/iface"
)

type DrawSpliter func(direct string)
type DrawView func(v iface.View)

func DrawWindow(root *ned.Window, width, height int, ds DrawSpliter, dv DrawView) {
	if root.Type == "container" {
		switch root.Pos {
		case "horizontal":
			w := (width - 1) / 2
			ds("vertical")
			DrawWindow(root.Subs[0], w, height, ds, dv)
			DrawWindow(root.Subs[1], w, height, ds, dv)
		case "vertical":
			h := (height - 1) / 2
			ds("horizontal")
			DrawWindow(root.Subs[0], width, h, ds, dv)
			DrawWindow(root.Subs[1], width, h, ds, dv)
		}
	} else {
		dv()
	}
}
