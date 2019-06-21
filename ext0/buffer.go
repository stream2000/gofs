package ext0

// 设计这样结构的动机是什么呢？
// 当你要读取一个文件，可能需要将它的全部块读取到内存中，并且对这些块内的内容进行写入
// 但是文件的物理存储是离散的
// 经常会发生不在同一个块内进行写入的情况
// 这时候就需要一个统一的buffer进行文件的读写啦
// 这个buffer必须表现得和一个连续的文件一模一样

// All operations about data blocks in memory disk  can be achieved by this buffer
/*
Features :
1. automatically allocate and free data blocks of current file
2. providing api that can help users treat discrete blocks in disk as one unified block
*/

/*
1. the function write and writeAt can only lengthen the size of current file
2. the function free, delete , trunc can shorten the file
*/
type unifiedBuffer struct {
	// 切片的切片，其中的每一个元素代表一个真实的内存磁盘块
	Data         [][]byte
	BlockSize    int // 1kb
	CurrentSize  int
	blockNumbers []int
	// 一个inode的指针
	ino *Ext0Inode
	// 超级块指针
	sb *Ext0SuperBlock
}

func (b *unifiedBuffer) Init(size int, ino *Ext0Inode) {
	// 1. assigning  buffer attributes
	b.BlockSize = size
	b.ino = ino
	b.sb = ino.sb
	b.CurrentSize = int(ino.attr.Size)
	attr := ino.attr
	addr := attr.StartAddr
	if addr == 0 {
		return
	} else {
		for addr > 0 {
			b.blockNumbers = append(b.blockNumbers, int(addr))
			b.Data = append(b.Data, ino.sb.getData(int(addr)))
			addr = ino.sb.getFat(addr)
		}
	}
}

// write back all the data to the disk
func (b *unifiedBuffer) writeBackToDisk() {

}

//alloc new blocks, it won't change the size of file
// 这个函数修改了什么？有未同步的项吗？
// 在alloc中自动修改了很多属性，都是直接写入的
// 对于在inode中需要存储的信息，除却时间外几乎没什么变动？
func (b *unifiedBuffer) allocNewBlock(num int) {
	for i := 0; i < num; i++ {
		addr := b.ino.allocBlock()
		b.blockNumbers = append(b.blockNumbers, addr)
		b.Data = append(b.Data, b.ino.sb.getData(addr))
	}
}

// 向后写
func (b *unifiedBuffer) Write(data []byte) int {
	if len(data) == 0 {
		// nothing to write
		return 0
	}
	defer func() {
		// keep the consistency
		b.ino.attr.Size = uint32(b.CurrentSize)
		b.sb.WriteInode(int(b.ino.attr.InodeNumber), b.ino.attr)
	}()
	// offset of the new data
	currentOffset := 0
	// offset inside current block
	offset := b.CurrentSize % b.BlockSize

	// get max block number,if it is 0, alloc one new block
	blockNum := len(b.blockNumbers) - 1
	if offset == 0 {
		// we don't alloc blocks unless needed, now, it is time to alloc at least one  new block
		b.allocNewBlock(1)
		blockNum += 1
	}
	if len(data)+b.CurrentSize <= b.BlockSize*len(b.blockNumbers) {
		//the capacity of current file is enough, don't need to allocate new block
		b.Data[blockNum] = append(b.Data[blockNum][:offset], data...)
		b.CurrentSize += len(data)
		return len(data)
	} else {
		//the capacity of current file is not  enough, we need to allocate new blocks from disk

		// calculate the number of blocks that we need to  allocate, then call the allocation method
		finalSize := len(data) + b.CurrentSize
		numberOfNewBlocks := (finalSize % b.BlockSize) - len(b.blockNumbers)
		b.allocNewBlock(numberOfNewBlocks)

		// firstly, fill the current block
		b.Data[blockNum] = append(b.Data[blockNum][:offset], data[:b.BlockSize-offset]...)
		currentOffset += b.BlockSize - offset
		b.CurrentSize += b.BlockSize - offset
	}
	for len(data)-currentOffset > 0 {
		var next int
		if (len(data) - currentOffset) > b.BlockSize {
			next = b.BlockSize
		} else {
			next = len(data) - currentOffset
		}
		blockNum += 1
		b.setDataInsideBlock(blockNum, next, data[currentOffset:currentOffset+next])
		currentOffset += next
		b.CurrentSize += next
	}

	// modify the attr "blockCount"
	return len(data)
}
func (b *unifiedBuffer) modify(offset int, data []byte) {
	if offset+len(data)-b.CurrentSize > 0 || len(data) == 0 {
		return
	} else {
		blockAddr := offset / b.BlockSize

		currentOffset := 0

		offSetInsideBlock := offset % b.BlockSize
		if offSetInsideBlock+len(data) <= b.BlockSize {
			b.setDataInsideBlock(blockAddr, offset%b.BlockSize, data)
			return
		} else {
			b.setDataInsideBlock(blockAddr, offset%b.BlockSize, data[:b.BlockSize-offset%b.BlockSize])
			currentOffset += b.BlockSize - offset%b.BlockSize
		}
		for len(data)-currentOffset > 0 {
			var next int
			if (len(data) - currentOffset) > b.BlockSize {
				next = b.BlockSize
			} else {
				next = len(data) - currentOffset
			}
			blockAddr += 1
			b.setDataInsideBlock(blockAddr, next, data[currentOffset:currentOffset+next])
			currentOffset += next
		}
	}
}
func (b *unifiedBuffer) setDataInsideBlock(num int, offsetInBlock int, data []byte) int {
	if len(data) > b.BlockSize-offsetInBlock {
		return -1
	} else {
		var i int
		for i, x := range data {
			b.Data[num][offsetInBlock+i] = x
		}
		return i
	}
}

// 如果offset大于currentSize 怎么办？
// 答案很简单，就用和dd类似的办法，中间的全部填写为0
// 要懂得复用代码
// currentSize内包含的部分就用modify，不包含的部分就用Write，注意切割
func (b *unifiedBuffer) WriteAt(offset int, data []byte) int {
	// 假设 offset = 1025 则 start = 1
	endAddr := offset + len(data)
	if endAddr < b.CurrentSize {
		b.modify(offset, data)
	} else {
		if offset > b.CurrentSize {
			empty := make([]byte, offset-b.CurrentSize)
			b.Write(empty)
			offset = b.CurrentSize
		}
		b.modify(offset, data[:b.CurrentSize-offset])
		b.Write(data[b.CurrentSize-offset:])
	}
	return len(data)
}
func (b unifiedBuffer) ReadAll() (result []byte) {
	for _, x := range b.Data {
		result = append(result, x...)
	}
	return result[:b.CurrentSize]
}

// trunc
func (b *unifiedBuffer) resize(newSize int) bool {
	blocksNeeded := newSize/b.BlockSize + 1
	defer b.sb.WriteInode(int(b.ino.attr.InodeNumber), b.ino.attr)
	if newSize%b.BlockSize == 0 {
		blocksNeeded -= 1
	}
	b.sb.FreeBlockNumber -= uint64(len(b.blockNumbers) - blocksNeeded)
	if blocksNeeded == 0 {
		b.ino.attr.StartAddr = 0
		for _, x := range b.blockNumbers {
			b.ino.attr.BlockCount -= 1
			b.sb.setFat(x, 0)
			b.sb.freeBlock(x)
		}
	} else {
		n := b.blockNumbers[blocksNeeded-1]
		b.sb.setFat(n, 0)
		for _, x := range b.blockNumbers[blocksNeeded:] {
			b.ino.attr.BlockCount -= 1
			b.sb.setFat(x, 0)
			b.sb.freeBlock(x)
		}
	}
	b.blockNumbers = b.blockNumbers[:blocksNeeded]
	b.Data = b.Data[:blocksNeeded]
	b.CurrentSize = newSize
	b.ino.attr.Size = uint32(newSize)

	return true
}
