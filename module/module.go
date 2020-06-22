package module

import (
	"dirstat/module/internal/sys"
	"github.com/spf13/afero"
	"io"
)

// Module defines working modules interface
type Module interface {
	workers() []worker
	renderers() []renderer
}

type module struct {
	wks []worker
	rnd []renderer
}

func (m *module) workers() []worker {
	return m.wks
}

func (m *module) renderers() []renderer {
	return m.rnd
}

type worker interface {
	init()
	handler(evt *sys.ScanEvent)
	finalize()
}

type renderer interface {
	print(p printer)
}

// Context defines modules context
type Context struct {
	total *totalInfo
	top   int
}

// NewContext creates new module's context that needed to create new modules
func NewContext(top int) *Context {
	total := totalInfo{}

	ctx := Context{
		total: &total,
		top:   top,
	}
	return &ctx
}

// Execute runs modules over path specified
func Execute(path string, fs afero.Fs, w io.Writer, modules ...Module) {
	var renderers []renderer
	var workers []worker

	for _, m := range modules {
		renderers = append(renderers, m.renderers()...)
		workers = append(workers, m.workers()...)
	}

	var handlers []sys.ScanHandler
	for _, wo := range workers {
		wo.init()
		handlers = append(handlers, wo.handler)
	}

	sys.Scan(path, fs, handlers)

	for _, wo := range workers {
		wo.finalize()
	}

	render(w, renderers)
}

// NewFoldersModule creates new folders module
func NewFoldersModule(ctx *Context, hideOutput bool) Module {
	work := newFoldersWorker(ctx)
	if hideOutput {
		return newModule(work)
	}
	rend := newFoldersRenderer(work)
	return newModule(work, rend)
}

// NewTopFilesModule creates new top files statistic module
func NewTopFilesModule(ctx *Context) Module {
	work := newTopFilesWorker(ctx.top)
	rend := newTopFilesRenderer(work)
	m := newModule(work, rend)
	return m
}

// NewDetailFileModule creates new file statistic by file size range module
func NewDetailFileModule(verbose bool, enabledRanges []int) Module {
	// Do nothing if verbose not enabled
	if !verbose {
		return &module{
			[]worker{},
			[]renderer{},
		}
	}
	work := newDetailFileWorker(enabledRanges)
	rend := newDetailFileRenderer(work)
	m := newModule(work, rend)
	return m
}

// NewExtensionModule creates new file extensions statistic module
func NewExtensionModule(ctx *Context, hideOutput bool) Module {
	work := newExtWorker(ctx)
	if hideOutput {
		return newModule(work)
	}
	rend := newExtRenderer(work, ctx.top)
	return newModule(work, rend)
}

// NewAggregateFileModule creates new total file statistic module
func NewAggregateFileModule(ctx *Context) Module {
	work := newAggregateFileWorker()
	rend := newAggregateFileRenderer(ctx, work)

	m := newModule(work, rend)
	return m
}

// NewTotalModule creates new total statistic module
func NewTotalModule(ctx *Context) Module {
	rend := newTotalRenderer(ctx)

	m := module{
		[]worker{},
		[]renderer{rend},
	}
	return &m
}

func newModule(w worker, r ...renderer) Module {
	m := module{
		[]worker{w},
		[]renderer{},
	}
	m.rnd = append(m.rnd, r...)
	return &m
}
