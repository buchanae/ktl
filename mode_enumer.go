// Code generated by "enumer -type=Mode -text"; DO NOT EDIT

package ktl

import (
	"fmt"
)

const _ModeName = "BestEffortFailFast"

var _ModeIndex = [...]uint8{0, 10, 18}

func (i Mode) String() string {
	if i < 0 || i >= Mode(len(_ModeIndex)-1) {
		return fmt.Sprintf("Mode(%d)", i)
	}
	return _ModeName[_ModeIndex[i]:_ModeIndex[i+1]]
}

var _ModeValues = []Mode{0, 1}

var _ModeNameToValueMap = map[string]Mode{
	_ModeName[0:10]:  0,
	_ModeName[10:18]: 1,
}

// ModeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ModeString(s string) (Mode, error) {
	if val, ok := _ModeNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Mode values", s)
}

// ModeValues returns all values of the enum
func ModeValues() []Mode {
	return _ModeValues
}

// IsAMode returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Mode) IsAMode() bool {
	for _, v := range _ModeValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalText implements the encoding.TextMarshaler interface for Mode
func (i Mode) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Mode
func (i *Mode) UnmarshalText(text []byte) error {
	var err error
	*i, err = ModeString(string(text))
	return err
}
