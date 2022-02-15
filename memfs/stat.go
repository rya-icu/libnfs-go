package memfs

import (
	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/nfs"
)

type openedFile struct {
	f        fs.File
	pathName string
}

func (f *openedFile) File() fs.File {
	return f.f
}

func (f *openedFile) Path() string {
	return f.pathName
}

type Stat struct {
	// cwd         string
	current     nfs.FileHandle4
	handleStack []nfs.FileHandle4

	clientId uint64

	openedFiles map[uint32]*openedFile // stateid4.seqid => *openedFile

	seqId uint32
}

// func (t *Stat) Cwd() string {
// 	return fs.Abs(t.cwd)
// }

// func (t *Stat) SetCwd(p string) error {
// 	p = fs.Abs(p)
// 	t.cwd = p
// 	return nil
// }

func (t *Stat) SetCurrentHandle(fh nfs.FileHandle4) {
	t.current = fh
}

func (t *Stat) CurrentHandle() nfs.FileHandle4 {
	if t.current == nil {
		t.current = []byte{}
	}
	return t.current
}

func (t *Stat) PopHandle() (nfs.FileHandle4, bool) {
	if t.handleStack != nil {
		if len(t.handleStack) > 0 {
			size := len(t.handleStack)
			last := t.handleStack[size-1]
			t.handleStack = t.handleStack[:size-1]
			return last, true
		}
	}
	return nil, false
}

func (t *Stat) PushHandle(item nfs.FileHandle4) {
	if t.handleStack == nil {
		t.handleStack = []nfs.FileHandle4{}
	}
	t.handleStack = append(t.handleStack, item)
}

func (t *Stat) SetClientId(clientId uint64) {
	t.clientId = clientId
}

func (t *Stat) ClientId() (uint64, bool) {
	return t.clientId, t.clientId > 0
}

func (t *Stat) nextSeqId() uint32 {
	if t.seqId <= 0 {
		t.seqId = 1000
	}
	t.seqId++
	return t.seqId
}

func (t *Stat) AddOpenedFile(pathName string, f fs.File) uint32 {
	if t.openedFiles == nil {
		t.openedFiles = map[uint32]*openedFile{}
	}
	seqId := t.nextSeqId()
	t.openedFiles[seqId] = &openedFile{
		pathName: pathName,
		f:        f,
	}
	return seqId
}

func (t *Stat) GetOpenedFile(seqId uint32) fs.FileOpenState {
	if t.openedFiles != nil {
		if of, found := t.openedFiles[seqId]; found {
			return of
		}
	}
	return nil
}

func (t *Stat) RemoveOpenedFile(seqId uint32) fs.FileOpenState {
	if t.openedFiles != nil {
		if of, found := t.openedFiles[seqId]; found {
			delete(t.openedFiles, seqId)
			return of
		}
	}
	return nil
}

func (t *Stat) CleanUp() {
	log.Debugf("stat: cleanup()")
	// t.cwd = ""
	t.current = []byte{}
	if t.handleStack != nil {
		t.handleStack = t.handleStack[0:0]
	}
	if t.openedFiles != nil {
		for _, of := range t.openedFiles {
			of.f.Close()
		}
		t.openedFiles = nil
	}
}
