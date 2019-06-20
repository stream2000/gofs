package virtualFileSystem

import (
	"../LruCache"
)

type SuperBlock interface {
	NewSuperBlock()
	ReadInode(number int) Inode            // when create an vInode,read it from disk
	WriteInode(number int, data InodeAttr) // write an vInode to disk
	DestroyInode(num int) bool             // when there is no refference to an vInode, call this function to drop an vInode in memory
	RecoverFromDisk()
	Init()
	GetRoot() Inode
	CreateFile(name string, p Inode, mode int) (n Inode)
}
type Inode interface {
	Create()
	Link()       // create a hard link
	FollowLink() //follow a symbolic link to the real path
	LookUp(name string) int
	SeAttr(data InodeAttr)
	GetAttr() InodeAttr
	List()([]string,bool)
	ReadAll() []byte
	WriteAt(offset int, data []byte) int
	Append(data []byte) int
}
type InodeAttr struct {
	InodeNumber uint16
	Mode        uint16
	LinkCount   uint16
	Uid         uint16
	Gid         uint16
	Ctime       uint32 // inode last changed
	Mtime       uint32 // file last time modified
	Atime       uint32 // access time
	Size        uint32
	BlockCount  uint16
	FileType    uint16
	StartAddr   uint16
}
type vfsInode struct {
	data  InodeAttr
	sb    SuperBlock
	inode Inode
}
type vfsMount struct {
	mountPoint Path
	sb         SuperBlock
	root       Path
	order      int
}
type file struct {
	count int // open count
	fPath Path
}

type Path struct {
	pathString   string
	currentIndex int
	pathSlice    []string
	depth        int
}
type inodeHash struct {
	fsMagic int
	num     int
}

type Vfs struct {
	rootVnode vfsInode
	//rootDentry Dentry
	inodeCache     cache.Cache
	rootSb         SuperBlock
	mountPointList []Path
	mount          []vfsMount
	curDir         Path
}
