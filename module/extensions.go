package module

import (
	"dirstat/module/internal/sys"
	"path/filepath"
	"sort"
)

type extWorker struct {
	total      *totalInfo
	aggregator map[string]countSizeAggregate
	top        int
}

type extRenderer struct {
	work *extWorker
}

func newExtWorker(ctx *Context) *extWorker {
	return &extWorker{
		total:      ctx.total,
		aggregator: make(map[string]countSizeAggregate),
		top:        ctx.top,
	}
}

func newExtRenderer(work *extWorker) renderer {
	return &extRenderer{work}
}

// Worker methods

func (m *extWorker) init() {
}

func (m *extWorker) finalize() {
	m.total.CountFileExts = len(m.aggregator)
}

func (m *extWorker) handler(evt *sys.ScanEvent) {
	if evt.File == nil {
		return
	}
	f := evt.File

	ext := filepath.Ext(f.Path)
	a := m.aggregator[ext]
	a.Size += uint64(f.Size)
	a.Count++
	m.aggregator[ext] = a
}

// Renderer method

func (e *extRenderer) print(p printer) {
	extBySize := e.evolventMap(func(agr countSizeAggregate) int64 {
		return int64(agr.Size)
	})

	extByCount := e.evolventMap(func(agr countSizeAggregate) int64 {
		return agr.Count
	})

	sort.Sort(sort.Reverse(extBySize))
	sort.Sort(sort.Reverse(extByCount))

	const format = "%v\t%v\t%v\t%v\t%v\n"

	p.print("\nTOP %d file extensions by size:\n\n", e.work.top)

	e.printTableHead(p, format)

	e.printTopTen(p, extBySize, func(data containers, item *container) (int64, uint64) {
		count := e.work.aggregator[item.name].Count
		sz := uint64(item.size)
		return count, sz
	})

	p.flush()

	p.print("\nTOP %d file extensions by count:\n\n", e.work.top)

	e.printTableHead(p, format)

	e.printTopTen(p, extByCount, func(data containers, item *container) (int64, uint64) {
		count := item.size
		sz := e.work.aggregator[item.name].Size
		return count, sz
	})

	p.flush()
}

func (e *extRenderer) printTableHead(p printer, format string) {
	p.printtab(format, "Extension", "Count", "%", "Size", "%")
	p.printtab(format, "---------", "-----", "------", "----", "------")
}

func (e *extRenderer) printTopTen(p printer, data containers, selector func(data containers, item *container) (int64, uint64)) {
	for i := 0; i < e.work.top && i < len(data); i++ {
		h := data[i].name

		count, sz := selector(data, data[i])

		e.work.total.printCountAndSizeStatLine(p, count, sz, h)
	}
}

func (e *extRenderer) evolventMap(mapper func(countSizeAggregate) int64) containers {
	var result = make(containers, len(e.work.aggregator))
	i := 0
	for k, v := range e.work.aggregator {
		result[i] = &container{size: mapper(v), name: k}
		i++
	}
	return result
}