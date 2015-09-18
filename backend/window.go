package neoeditor

import "errors"

type Window struct {
	Type string // "container" or "window"
	// "full": the root window
	// "horizontal", "vertical": the container which has 2 subwindows
	// "above", "under", "left", "right": subwindow
	Pos    string
	Parent *Window    // root window has no parent
	Subs   [2]*Window // 2 subwindows at most
}

func NewWindow() *Window {
	return &Window{Type: "window", Pos: "full"}
}

func (w *Window) Split(orientation string) error {
	if w.Type != "window" {
		return errors.New("Can't split non-window type window")
	}

	w.Type = "container"
	w.Pos = orientation
	w.Subs[0] = &Window{Type: "window"}
	w.Subs[0].Parent = w
	w.Subs[1] = &Window{Type: "window"}
	w.Subs[1].Parent = w
	// new windown always under or on the right of the orig window w
	// and always stored in the second positon of Subs
	switch orientation {
	case "horizontal":
		w.Subs[0].Pos = "left"
		w.Subs[0].Pos = "right"
	case "vertical":
		w.Subs[0].Pos = "above"
		w.Subs[0].Pos = "under"
	}

	return nil
}

// Delete self, copy sibling window to parent and cache all two sibling and self
func (w *Window) Delete() (*Window, error) {
	if w.Type != "container" {
		return nil, errors.New("Can't delete non-container type window")
	}

	p := w.Parent
	if p == nil {
		return nil, errors.New("Can't delete root window")
	}

	var sibling *Window
	if w == p.Subs[0] {
		sibling = p.Subs[1]
	} else {
		sibling = p.Subs[0]
	}

	*p = *sibling

	return p, nil
}

// Command
