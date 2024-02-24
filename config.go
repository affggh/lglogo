package lglogo

import (
	"io"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Header    Header
	ImageData []ImageData `toml:"imagedata"`
}

type Header struct {
	Magic    string `toml:"magic"`
	Unknow   int    `toml:"unknow"`
	Metadata string `toml:"metadata"`
}

type ImageData struct {
	Name    string `toml:"name"`
	XOffset int    `toml:"xoffset"`
	YOffset int    `toml:"yoffset"`
}

func (c *Config) Read(rd io.Reader) error {
	data, err := io.ReadAll(rd)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(data, c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Saveto(fname string) error {
	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	fd, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.Write(data)
	if err != nil {
		return err
	}

	return nil
}
