package ext0

import (
	vfs "../virtualFileSystem"
	"encoding/binary"
)

/* ***********************************SuperBlock***********************************  */
func (sb *Ext0SuperBlock) writeSuperBlock() {
	sb.disk.SetBlockSize(BlockSize)
	sbSlice := sb.disk.ReadBlock(0)

	binary.BigEndian.PutUint64(sbSlice[:8], uint64(sb.BlockNumber))
	binary.BigEndian.PutUint64(sbSlice[8:16], uint64(sb.InodeNumber))
	binary.BigEndian.PutUint64(sbSlice[16:24], uint64(sb.FreeBlockNumber))
	binary.BigEndian.PutUint64(sbSlice[24:32], uint64(sb.FreeInodeNumber))
	binary.BigEndian.PutUint64(sbSlice[32:40], sb.sysType)

}

// TODO recover
func (sb *Ext0SuperBlock) RecoverFromDisk() {
	sbSlice := sb.disk.UnsaveRead(SuperBlockStartAddr, FatStartAddr)
	sb.InodeNumber = binary.BigEndian.Uint64(sbSlice[8:16])
	sb.BlockNumber = binary.BigEndian.Uint64(sbSlice[:8])
	sb.FreeBlockNumber = binary.BigEndian.Uint64(sbSlice[16:24])
	sb.FreeInodeNumber = binary.BigEndian.Uint64(sbSlice[24:32])
	sb.sysType = binary.BigEndian.Uint64(sbSlice[32:40])
}

/* ***********************************FAT***********************************  */
func (sb *Ext0SuperBlock) getFat(num uint16) (value uint16) {
	fatSlice := sb.disk.UnsaveRead(FatStartAddr, BlockBitmapStartAddr)
	if num <= 0 {
		return
	} else {
		// 读取当前Fat中存储的地址
		value = binary.BigEndian.Uint16(fatSlice[num*2 : (num+1)*2])
		return
	}
}
func (sb *Ext0SuperBlock) setFat(num int, addr uint16) {
	fatSlice := sb.disk.UnsaveRead(FatStartAddr, BlockBitmapStartAddr)
	binary.BigEndian.PutUint16(fatSlice[num*2:(num+1)*2], addr)
}

func (sb *Ext0SuperBlock) freeFat(num uint16) (next uint16) {
	if !(num > 0) {
		return
	} else {
		// 读取当前Fat中存储的地址
		next = sb.getFat(num)
		sb.setFat(int(num), 0)
		return
	}
}

/* ***********************************BlockBitmap***********************************  */
func (sb *Ext0SuperBlock) setBlockBitmap(num int, value bool) {
	blockBitmapSlice := sb.disk.UnsaveRead(BlockBitmapStartAddr, InodeBitmapStartAddr)
	if value {
		binary.BigEndian.PutUint16(blockBitmapSlice[num*2:(num+1)*2], 1)
	} else {
		binary.BigEndian.PutUint16(blockBitmapSlice[num*2:(num+1)*2], 0)
	}
}

func (sb *Ext0SuperBlock) getBlockBitmap(num int) bool {
	blockBitmapSlice := sb.disk.UnsaveRead(BlockBitmapStartAddr, InodeBitmapStartAddr)
	value := binary.BigEndian.Uint16(blockBitmapSlice[num*2 : (num+1)*2])
	return value > 0
}

/* ***********************************InodeBitmap***********************************  */
func (sb *Ext0SuperBlock) setInodeBitmap(num int, value bool) {
	inodeBitmapSlice := sb.disk.UnsaveRead(InodeBitmapStartAddr+2*num, InodeBitmapStartAddr+(num+1)*2)
	if value {
		binary.BigEndian.PutUint16(inodeBitmapSlice[num*2:(num+1)*2], 1)
	} else {
		binary.BigEndian.PutUint16(inodeBitmapSlice[num*2:(num+1)*2], 0)
	}
}
func (sb *Ext0SuperBlock) getInodeBitmap(num int) bool {
	inodeBitmapSlice := sb.disk.UnsaveRead(InodeBitmapStartAddr+2*num, InodeBitmapStartAddr+(num+1)*2)
	value := binary.BigEndian.Uint16(inodeBitmapSlice[num*2 : (num+1)*2])
	return value > 0
}

/* ***********************************Inode***********************************  */
func (sb *Ext0SuperBlock) ReadInode(number int) (ino vfs.Inode) {
	rawData := sb.disk.UnsaveRead(InodeStartAddr+InodeSize*number, InodeStartAddr+InodeSize*(number+1))
	inoAttr := vfs.InodeAttr{
		InodeNumber: binary.BigEndian.Uint16(rawData[0:2]),
		Mode:        binary.BigEndian.Uint16(rawData[2:4]),
		LinkCount:   binary.BigEndian.Uint16(rawData[4:6]),
		Uid:         binary.BigEndian.Uint16(rawData[6:8]),
		Gid:         binary.BigEndian.Uint16(rawData[8:10]),
		Ctime:       binary.BigEndian.Uint32(rawData[10:14]),
		Mtime:       binary.BigEndian.Uint32(rawData[14:18]),
		Atime:       binary.BigEndian.Uint32(rawData[18:22]),
		Size:        binary.BigEndian.Uint32(rawData[22:26]),
		BlockCount:  binary.BigEndian.Uint16(rawData[26:28]),
		FileType:    binary.BigEndian.Uint16(rawData[28:30]),
		StartAddr:   binary.BigEndian.Uint16(rawData[30:32]),
	}
	ino = &Ext0Inode{attr: inoAttr, sb: sb}
	return
}

func (sb *Ext0SuperBlock) WriteInode(number int, attr vfs.InodeAttr) {
	inodeSlice := sb.disk.UnsaveRead(InodeStartAddr+InodeSize*number, InodeStartAddr+InodeSize*(number+1))

	binary.BigEndian.PutUint16(inodeSlice[0:2], attr.InodeNumber)
	binary.BigEndian.PutUint16(inodeSlice[2:4], attr.Mode)
	binary.BigEndian.PutUint16(inodeSlice[4:6], attr.LinkCount)
	binary.BigEndian.PutUint16(inodeSlice[6:8], attr.Uid)
	binary.BigEndian.PutUint16(inodeSlice[8:10], attr.Gid)
	binary.BigEndian.PutUint32(inodeSlice[10:14], attr.Ctime)
	binary.BigEndian.PutUint32(inodeSlice[14:18], attr.Mtime)
	binary.BigEndian.PutUint32(inodeSlice[18:22], attr.Atime)
	binary.BigEndian.PutUint32(inodeSlice[22:26], attr.Size)
	binary.BigEndian.PutUint16(inodeSlice[26:28], attr.BlockCount)
	binary.BigEndian.PutUint16(inodeSlice[28:30], attr.FileType)
	binary.BigEndian.PutUint16(inodeSlice[30:32], attr.StartAddr)
}

/* ***********************************DataBlock***********************************  */

func (sb *Ext0SuperBlock) freeBlock(num int) {
	dataSlice := sb.disk.UnsaveRead(DataStartAddr, DataEndAddr)
	var emptyByte byte
	// 将块位图中对应部分清0
	sb.setBlockBitmap(num, false)
	// 将数据块清0
	memsetRepeat(dataSlice[num*BlockSize:(num+1)*BlockSize], emptyByte)
}
func (sb *Ext0SuperBlock) getData(num int) []byte {
	dataSlice := sb.disk.UnsaveRead(DataStartAddr, DataEndAddr)
	return dataSlice[num*BlockSize : (num+1)*BlockSize]

}
func (sb *Ext0SuperBlock) setData(num int, data []byte) {
	dataSlice := sb.disk.UnsaveRead(DataStartAddr, DataEndAddr)
	//sb.dataSlice[num*BlockSize : (num+1)*BlockSize] = data
	for i, x := range data {
		dataSlice[num*BlockSize+i] = x
	}
}
