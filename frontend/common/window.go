package common

import (
	ned "github.com/ensonmj/NeoEditor/backend"
)

type DrawSpliter func(l ned.Line)
type DrawView func(v ned.View)

func DrawWindow(root *ned.Window, r ned.Rect, ds DrawSpliter, dv DrawView) {
	if root.Type == "container" {
		switch root.Direct {
		case ned.Horizontal: // split window to left and right
			w := (r.Width - 1) / 2
			l := ned.Line{ned.Point{r.X + w, r.Y}, ned.Vertical, r.Height}
			ds(l)
			sub := r
			sub.Width = w
			DrawWindow(root.Subs[0], sub, ds, dv)
			sub.X = r.X + w + 1
			DrawWindow(root.Subs[1], sub, ds, dv)
		case ned.Vertical: // split window to above and under
			h := (r.Height - 1) / 2
			l := ned.Line{ned.Point{r.X, r.Y + h}, ned.Horizontal, r.Width}
			ds(l)
			sub := r
			sub.Height = h
			DrawWindow(root.Subs[0], sub, ds, dv)
			sub.Y = r.Y + h + 1
			DrawWindow(root.Subs[1], sub, ds, dv)
		}
	} else {
		//dv()
	}
}
