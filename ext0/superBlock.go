package ext0

import (
	"../disk"
	u "../utilities"
	vfs "../virtualFileSystem"
	"encoding/binary"
	"time"
)

//在 go语言中不需要手动管理内存，所以alloc_inode 和destroy是不是不需要的

/* TODO  2. 从磁盘文件恢复系统

 */
// 以下基本以byte为单位
const (
	InodeNumber          = 1 << 7
	BlockSize            = 1 << 10 //  1 Kilobytes
	InodeSize            = 1 << 6  // 64 bits | used 36 bits 8 bytes
	DataBlockNumber      = 1 << 10 // 1024 个 总大小 4*1024 * 1kb = 4mb
	SuperBlockSize       = BlockSize
	DirStorageSIze       = BlockSize / 16
	SuperBlockStartAddr  = 0
	FatStartAddr         = SuperBlockSize               // fat uint16
	BlockBitmapStartAddr = FatStartAddr + 2*InodeNumber // bitmap uint16
	InodeBitmapStartAddr = BlockBitmapStartAddr + 2*InodeNumber
	InodeStartAddr       = InodeBitmapStartAddr + 2*InodeNumber
	DataStartAddr        = InodeStartAddr + InodeSize*InodeNumber
	DataEndAddr          = DataStartAddr + DataBlockNumber*BlockSize
)

type Ext0SuperBlock struct {
	BlockNumber     uint64
	InodeNumber     uint64
	FreeInodeNumber uint64
	FreeBlockNumber uint64
	disk            disk.Harddisk
	sysType         uint64
}

// 删除一个inode 分两步 先删除文件数据再删除inode数据 首先找到i节点对应的数据区域块，将他们对应的位图全部清0，修改空闲区域
// startAddr 使用在fat和block表里的数据

func (sb Ext0SuperBlock) GetFileSystemType() int {
	return u.Ext0
}
func (sb *Ext0SuperBlock) NewSuperBlock() {
	sb.disk.SetBlockSize(BlockSize)
	sb.FreeInodeNumber = InodeNumber
	sb.FreeBlockNumber = DataBlockNumber
	sb.InodeNumber = InodeNumber
	sb.BlockNumber = DataBlockNumber
	sb.sysType = u.Ext0
}
func (sb *Ext0SuperBlock) formatDisk() {
	var emptyByte byte
	fatSlice := sb.disk.UnsaveRead(FatStartAddr, BlockBitmapStartAddr)
	blockBitmapSlice := sb.disk.UnsaveRead(BlockBitmapStartAddr, InodeBitmapStartAddr)
	inodeBitmapSlice := sb.disk.UnsaveRead(InodeBitmapStartAddr, InodeStartAddr)
	inodeSlice := sb.disk.UnsaveRead(InodeStartAddr, DataStartAddr)
	dataSlice := sb.disk.UnsaveRead(DataStartAddr, DataEndAddr)
	memsetRepeat(fatSlice, emptyByte)
	memsetRepeat(blockBitmapSlice, emptyByte)
	memsetRepeat(inodeBitmapSlice, emptyByte)
	memsetRepeat(inodeSlice, emptyByte)
	memsetRepeat(dataSlice, emptyByte)
	sb.NewSuperBlock()

	sb.writeSuperBlock()
}

// TODO initFromDisk
func (sb *Ext0SuperBlock) Init() {
	sb.formatDisk()
	sb.initRootInode()
}
func (sb *Ext0SuperBlock) CreateFile(name string, p vfs.Inode, fileType int) (n vfs.Inode) {
	var ino Ext0Inode
	if fileType == int(u.OrdinaryFile) {
		num := ino.initAsOrdinaryFile(sb)
		p.(*Ext0Inode).createChild(name, num)
	} else if fileType == int(u.Directory) {
		// 这会为ino分配磁盘空间并初始化一些属性最后写入磁盘
		num := ino.initAsDir(p.GetAttr().InodeNumber, sb)
		// 将新ino的信息写到父目录的信息区中
		p.(*Ext0Inode).createChild(name, num)
	}
	return
}
func (sb Ext0SuperBlock) initAttr() (attr vfs.InodeAttr) {
	attr = vfs.InodeAttr{
		LinkCount: 1,
		Ctime:     uint32(time.Now().Unix()),
		Mtime:     uint32(time.Now().Unix()),
		Atime:     uint32(time.Now().Unix()),
		Size:      0,
	}
	return attr
}
func memsetRepeat(a []byte, v byte) {
	if len(a) == 0 {
		return
	}
	a[0] = v
	for bp := 1; bp < len(a); bp *= 2 {
		copy(a[bp:], a[:bp])
	}
}

// 仅仅返回数字，不多做操作
func (sb *Ext0SuperBlock) GetNextFreeInodeNumber() uint16 {
	if sb.FreeInodeNumber > 0 {
		// 遍历inode位图，查找第一个为free的inode
		for i := 1; i < int(sb.InodeNumber); i++ {
			if !sb.getInodeBitmap(i) {
				sb.FreeInodeNumber -= 1
				return uint16(i)
			}
		}
		return 0
	} else {
		return 0
	}
}

// 仅仅返回数字，不多做操作
func (sb *Ext0SuperBlock) GetNextFreeBlockeNumber() uint16 {
	if sb.FreeBlockNumber > 0 {
		// 遍历inode位图，查找第一个为free的inode
		for i := 1; i < int(sb.BlockNumber); i++ {
			if !sb.getBlockBitmap(i) {
				sb.FreeBlockNumber -= 1
				return uint16(i)
			}
		}
		return 0
	} else {
		return 0
	}
}
func (sb *Ext0SuperBlock) ReadDir(attr vfs.InodeAttr) (dir []Exto0DirectoryStorage) {
	if attr.FileType != u.Directory {
		return
	}
	var buf ExtendedBuffer
	buf.Init(BlockSize,sb.ReadInode(int(attr.InodeNumber)).(*Ext0Inode))
	size := attr.Size
	if size%DirStorageSIze != 0 {
		return
	}
	var dentryNumber = size / DirStorageSIze

	block := buf.ReadAll()

	for i := 0; i < int(dentryNumber); i++ {
		var dirName [DirStorageSIze - 2]byte
		var dirInodeNumber uint16
		for j := 0; j < DirStorageSIze-2; j++ {
			dirName[j] = block[DirStorageSIze*i+j]
		}
		dirInodeNumber = binary.BigEndian.Uint16(block[DirStorageSIze*i+DirStorageSIze-2 : DirStorageSIze*i+DirStorageSIze])

		d := Exto0DirectoryStorage{
			name:        dirName,
			inodeNumber: dirInodeNumber,
		}
		dir = append(dir, d)
	}
	return
}
func (sb Ext0SuperBlock) GetRoot() vfs.Inode {
	return sb.ReadInode(0)
}
func (sb *Ext0SuperBlock) DestroyInode(num int) (ok bool) {
	ino := sb.ReadInode(num)
	data := ino.GetAttr()
	addr := data.StartAddr
	//var emptyByte byte
	// 当硬链接数大于1时无法删除
	if addr == 0 || ino.GetAttr().LinkCount > 1 {
		return false
	} else {
		for addr > 0 {
			// 先释放数据区
			sb.freeBlock(int(addr))
			// 再释放Fat，取得下一个数据块
			addr = sb.freeFat(addr)
			sb.FreeBlockNumber++
		}
	}
	//	最后释放inode以及他的位图
	return true
}
func (sb *Ext0SuperBlock) initRootInode() {
	var ino = &Ext0Inode{}
	ino.attr = sb.initAttr()
	ino.attr.BlockCount = 0
	ino.attr.FileType = u.Directory
	ino.sb = sb
	buf := ExtendedBuffer{}
	num := uint16(0)
	ino.attr.InodeNumber = num
	buf.Init(BlockSize,ino)
	buf.WriteAt(0, makeDirData(".", 0))
	buf.WriteAt(DirStorageSIze, makeDirData("..", 0))
	sb.setInodeBitmap(int(num), true)
	sb.WriteInode(int(num), ino.attr)
}
