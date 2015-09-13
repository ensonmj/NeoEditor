package key

type Key rune

// KeyPress describes a key press event.
// Note that Key does not distinguish between capital and non-capital letters;
// use the Text property for this purpose.
type KeyPress struct {
	Text                    string // the text representation of the key
	Key                     Key    // the code for the key that was pressed
	Shift, Super, Alt, Ctrl bool   // true if modifier key was pressed
}
