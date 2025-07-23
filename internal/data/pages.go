package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidCountPagesFormat returns error when we are unable to parse or convert a JSON string for Pages.
var ErrInvalidCountPagesFormat = errors.New("invalid count pages format")

type Pages int64

// MarshalJSON method on the Pages type so that it satisfies the
// json.Marshaler interface. This should return "<pages> pages".
func (p Pages) MarshalJSON() ([]byte, error) {
	jsVal := fmt.Sprintf("%d pages", p)
	return []byte(strconv.Quote(jsVal)), nil
}

// UnmarshalJSON ensures that Pages satisfies the
// json.Unmarshaler interface.
func (p *Pages) UnmarshalJSON(jsVal []byte) error {
	unquotedJsVal, err := strconv.Unquote(string(jsVal))
	if err != nil {
		return ErrInvalidCountPagesFormat
	}
	parts := strings.Split(unquotedJsVal, " ")
	if len(parts) != 2 || parts[1] != "pages" {
		return ErrInvalidCountPagesFormat
	}
	i, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return ErrInvalidCountPagesFormat
	}
	*p = Pages(i)
	return nil
}
