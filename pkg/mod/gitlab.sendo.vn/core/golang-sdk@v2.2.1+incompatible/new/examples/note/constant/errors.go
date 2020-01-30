package constant

type ErrID int

const (
	ERR_EMPTY_TEXT ErrID = iota + 410
	ERR_NOTE_NOT_FOUND
	ERR_NOTE_CANNOT_LIST
	ERR_NOTE_CANNOT_UPDATE
	ERR_NOTE_CANNOT_DELETE
	ERR_NOTE_CANNOT_ADD
)

func (s ErrID) String() string {
	return [...]string{
		"Text can not be blank",
		"Note is not found",
		"Can not list note",
		"Note updated failed",
		"NOte deleted failed",
		"NOte added failed",
	}[s-410]
}

func (s ErrID) Error() string {
	return s.String()
}
