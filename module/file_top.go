package module

import (
	"dirstat/module/internal/sys"
	"fmt"
	"github.com/aegoroff/godatastruct/rbtree"
)

func newTopFilesRenderer(work *topFilesWorker) renderer {
	return &topFilesRenderer{work}
}

type topFilesWorker struct {
	tree rbtree.RbTree
	top  int
}

type topFilesRenderer struct {
	work *topFilesWorker
}

func newTopFilesWorker(top int) *topFilesWorker {
	return &topFilesWorker{rbtree.NewRbTree(), top}
}

// Worker methods

func (*topFilesWorker) init()     {}
func (*topFilesWorker) finalize() {}

func (m *topFilesWorker) handler(evt *sys.ScanEvent) {
	if evt.File == nil {
		return
	}
	f := evt.File

	fileContainer := file{size: f.Size, path: f.Path}
	insertTo(m.tree, m.top, &fileContainer)
}

// Renderer method

func (m *topFilesRenderer) print(p printer) {
	p.cprint("\n<gray>TOP %d files by size:</>\n\n", m.work.top)

	p.print("%v\t%v\n", "File", "Size")
	p.print("%v\t%v\n", "------", "----")

	i := 1

	m.work.tree.Descend(func(n rbtree.Node) bool {
		file := n.Key().(*file)
		h := fmt.Sprintf("%2d. %s", i, file)

		i++

		p.print("%v\t%v\n", h, human(file.size))

		return true
	})

	p.flush()
}