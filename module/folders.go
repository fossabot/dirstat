package module

import (
	"dirstat/module/internal/sys"
	"fmt"
	"github.com/aegoroff/godatastruct/rbtree"
	"github.com/akutz/sortfold"
	"strings"
)

// Folder interface
type folderI interface {
	Path() string
	Size() int64
	Count() int64
}

// folder represents file system container that described by path
// and has size and the number of elements in it (count field).
type folder struct {
	path  string
	size  int64
	count int64
}

// Count sortable folder
type folderC struct {
	folder
}

// Size sortable folder
type folderS struct {
	folder
}

// Path sortable folder methods

func (f *folder) LessThan(y interface{}) bool {
	if sys.RunUnderWindows() {
		return sortfold.CompareFold(f.String(), y.(*folder).String()) < 0
	}

	return f.String() < y.(*folder).String()
}

func (f *folder) EqualTo(y interface{}) bool {
	if sys.RunUnderWindows() {
		return strings.EqualFold(f.String(), y.(*folder).String())
	}

	return f.String() == y.(*folder).String()
}

func (f *folder) String() string { return f.path }
func (f *folder) Path() string   { return f.path }
func (f *folder) Size() int64    { return f.size }
func (f *folder) Count() int64   { return f.count }

// Count sortable folder methods

func (fc *folderC) LessThan(y interface{}) bool { return fc.count < y.(*folderC).count }
func (fc *folderC) EqualTo(y interface{}) bool  { return fc.count == y.(*folderC).count }

// Size sortable folder methods

func (fs *folderS) LessThan(y interface{}) bool { return fs.size < y.(*folderS).size }
func (fs *folderS) EqualTo(y interface{}) bool  { return fs.size == y.(*folderS).size }

type foldersWorker struct {
	voidInit
	total   *totalInfo
	folders rbtree.RbTree
	bySize  *fixedTree
	byCount *fixedTree
}

type foldersRenderer struct {
	*foldersWorker
}

func newFoldersWorker(ctx *Context) *foldersWorker {
	return &foldersWorker{
		total:   ctx.total,
		folders: rbtree.NewRbTree(),
		bySize:  newFixedTree(ctx.top),
		byCount: newFixedTree(ctx.top),
	}
}

func newFoldersRenderer(work *foldersWorker) renderer {
	return &foldersRenderer{work}
}

// Worker methods

func (m *foldersWorker) finalize() {
	m.folders.WalkInorder(func(node rbtree.Node) {
		fn := node.Key().(*folder)

		fs := folderS{*fn}
		m.bySize.insert(&fs)

		fc := folderC{*fn}
		m.byCount.insert(&fc)
	})

	m.total.CountFolders = m.folders.Len()
}

func (m *foldersWorker) handler(evt *sys.ScanEvent) {
	if evt.Folder == nil {
		return
	}
	fe := evt.Folder

	fn := folder{
		path:  fe.Path,
		count: fe.Count,
		size:  fe.Size,
	}
	m.folders.Insert(&fn)
}

// Renderer method

type folderCast func(c rbtree.Comparable) folderI

func castSize(c rbtree.Comparable) folderI  { return c.(*folderS) }
func castCount(c rbtree.Comparable) folderI { return c.(*folderC) }

func (f *foldersRenderer) print(p printer) {
	const format = "%v\t%v\t%v\t%v\t%v\n"

	p.cprint("\n<gray>TOP %d folders by size:</>\n\n", f.bySize.size)

	f.printTop(f.bySize, p, format, castSize)

	p.cprint("\n<gray>TOP %d folders by count:</>\n\n", f.byCount.size)

	f.printTop(f.byCount, p, format, castCount)
}

func (f *foldersRenderer) printTop(ft *fixedTree, p printer, format string, cast folderCast) {
	p.print(format, "Folder", "Files", "%", "Size", "%")
	p.print(format, "------", "-----", "------", "----", "------")

	i := 1

	ft.tree.Descend(func(n rbtree.Node) bool {
		f.printTableRow(&i, cast(n.Key()), p)

		return true
	})

	p.flush()
}

func (f *foldersRenderer) printTableRow(i *int, fi folderI, p printer) {
	h := fmt.Sprintf("%2d. %s", *i, fi.Path())

	*i++

	count := fi.Count()
	sz := uint64(fi.Size())

	f.total.printCountAndSizeStatLine(p, count, sz, h)
}
