package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"io"
	"os"
)

type options struct {
	vrange []int
	path   string
}

type conf interface {
	// fs defines app file system abstraction
	fs() afero.Fs

	// w defines app output
	w() io.Writer
}

type appConf struct {
	filesystem afero.Fs
	writer     io.Writer
}

func (a *appConf) fs() afero.Fs {
	return a.filesystem
}

func (a *appConf) w() io.Writer {
	return a.writer
}

func newAppConf() conf {
	c := appConf{
		filesystem: afero.NewOsFs(),
		writer:     os.Stdout,
	}
	return &c
}

func configure(cmd *cobra.Command, opt *options) {
	configurePath(cmd, &opt.path)
	confRange(cmd, &opt.vrange)
}

func confRange(cmd *cobra.Command, rn *[]int) {
	cmd.Flags().IntSliceVarP(rn, "range", "r", []int{}, "Output verbose files info for range specified. Range is the number between 1 and 10")
}

func configurePath(cmd *cobra.Command, path *string) {
	cmd.Flags().StringVarP(path, "path", "p", "", "REQUIRED. Directory path to show info.")
}
