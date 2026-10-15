package openbw

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/marianogappa/openbw-go/bwgame"
)

// ErrNoData is returned when no data directory is given and OPENBW_DATA is not set.
var ErrNoData = errors.New("openbw: no game data: set Options.Data or OPENBW_DATA")

// Data is the game data: .dat files, iscript, GRP frame sizes, tilesets.
// One Data is safe for concurrent use by any number of games.
type Data struct {
	global    *bwgame.GlobalState
	filesHash string
}

// LoadData reads the game data from dir. An empty dir means $OPENBW_DATA.
func LoadData(dir string) (*Data, error) {
	if dir == "" {
		dir = os.Getenv("OPENBW_DATA")
	}
	if dir == "" {
		return nil, ErrNoData
	}
	readFile := bwgame.DataFilesDirectory(dir)
	hash := sha256.New()
	global, err := bwgame.GlobalInit(func(path string) ([]byte, error) {
		data, err := readFile(path)
		if err == nil {
			hash.Write([]byte(bwgame.DataPathToLowerSlash(path)))
			hash.Write(binary.LittleEndian.AppendUint64([]byte{0}, uint64(len(data))))
			hash.Write(data)
		}
		return data, err
	})
	if err != nil {
		return nil, fmt.Errorf("openbw: load game data from %s: %w", dir, err)
	}
	return &Data{global: global, filesHash: hex.EncodeToString(hash.Sum(nil))}, nil
}

// FilesHash is the SHA-256 of every data file's path, size and content, in load order, as hex.
func (d *Data) FilesHash() string { return d.filesHash }

// Global gives the engine data (L0).
func (d *Data) Global() *bwgame.GlobalState { return d.global }

var defaultData struct {
	once sync.Once
	data *Data
	err  error
}

func dataOrDefault(d *Data) (*Data, error) {
	if d != nil {
		return d, nil
	}
	defaultData.once.Do(func() {
		defaultData.data, defaultData.err = LoadData("")
	})
	return defaultData.data, defaultData.err
}

// EngineError is the error that stops a game when the engine fails.
type EngineError struct {
	Frame int32
	Value any
	Stack []byte
}

func (e *EngineError) Error() string {
	return fmt.Sprintf("openbw: engine error at frame %d: %v", e.Frame, e.Value)
}

// Unwrap gives Value if it is an error.
func (e *EngineError) Unwrap() error {
	err, _ := e.Value.(error)
	return err
}
