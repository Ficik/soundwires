package pipewire

import (
	"encoding/json"
	"os/exec"
)

// Dump runs pw-dump and returns all PipeWire objects as a slice.
func Dump(pwDumpBin string) ([]PWObject, error) {
	out, err := exec.Command(pwDumpBin).Output()
	if err != nil {
		return nil, err
	}
	var objects []PWObject
	if err := json.Unmarshal(out, &objects); err != nil {
		return nil, err
	}
	return objects, nil
}
