package module

import (
	"dirstat/module/internal/sys"
	"github.com/dustin/go-humanize"
	"io"
	"text/tabwriter"
	"text/template"
	"time"
)

type moduleTotal struct {
	total *totalInfo
	start time.Time
}

func (m *moduleTotal) init() {
}

func (m *moduleTotal) postScan() {
	m.total.ReadingTime = time.Since(m.start)
}

func (m *moduleTotal) handler() sys.FileHandler {
	return func(f *sys.FileEntry) {
		// Accumulate file statistic
		m.total.FilesTotal.Count++
		m.total.FilesTotal.Size += uint64(f.Size)
	}
}

func (m *moduleTotal) output(_ *tabwriter.Writer, w io.Writer) {
	const totalTemplate = `
Total files:            {{.FilesTotal.Count}} ({{.FilesTotal.Size | toBytesString }})
Total folders:          {{.CountFolders}}
Total file extensions:  {{.CountFileExts}}

Read taken:    {{.ReadingTime}}
`

	var report = template.Must(template.New("totalstat").Funcs(template.FuncMap{"toBytesString": humanize.IBytes}).Parse(totalTemplate))
	_ = report.Execute(w, m.total)
}