package ext0

// 设计这样结构的动机是什么呢？
// 当你要读取一个文件，可能需要将它的全部块读取到内存中，并且对这些块内的内容进行写入
// 但是文件的物理存储是离散的
// 经常会发生不在同一个块内进行写入的情况
// 这时候就需要一个统一的buffer进行文件的读写啦
// 这个buffer必须表现得和一个连续的文件一模一样

// 事实证明，对内部的slice的操作可以直接影响

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

func (buf *unifiedBuffer) Init(size int, ino *Ext0Inode) {
	// 1. assigning  buffer attributes
	buf.BlockSize = size
	buf.ino = ino
	buf.sb = ino.sb
	buf.CurrentSize = int(ino.attr.Size)
	attr := ino.attr
	addr := attr.StartAddr
	if addr == 0 {
		return
	} else {
		for addr > 0 {
			buf.blockNumbers = append(buf.blockNumbers, int(addr))
			buf.Data = append(buf.Data, ino.sb.getData(int(addr)))
			addr = ino.sb.getFat(addr)
		}
	}
}
// write back all the data to the disk
func (buf *unifiedBuffer) writeBackToDisk() {

}
//alloc new blocks, it won't change the size of file
// 这个函数修改了什么？有未同步的项吗？
// 在alloc中自动修改了很多属性，都是直接写入的
// 对于在inode中需要存储的信息，除却时间外几乎没什么变动？
func (buf *unifiedBuffer)allocNewBlock(num int){
	for i := 0; i < num; i++ {
		addr := buf.ino.allocBlock()
		buf.blockNumbers  = append(buf.blockNumbers, addr)
		buf.Data = append(buf.Data, buf.ino.sb.getData(addr))
	}
}
// 向后写
func (buf *unifiedBuffer) Write(data []byte) int {
	if len(data) == 0 {
		// nothing to write
		return 0
	}
	defer func() {
		// keep the consistency
		buf.ino.attr.Size = uint32( buf.CurrentSize)
		buf.sb.WriteInode(int(buf.ino.attr.InodeNumber), buf.ino.attr)
	}()
	// offset of the new data
	currentOffset := 0
	// offset inside current block
	offset := buf.CurrentSize % buf.BlockSize

	// get max block number,if it is 0, alloc one new block
	blockNum := len(buf.blockNumbers) -1
	if offset == 0 {
		// we don't alloc blocks unless needed, now, it is time to alloc at least one  new block
		buf.allocNewBlock(1)
		blockNum += 1
	}
	if len(data)+buf.CurrentSize <= buf.BlockSize*len(buf.blockNumbers) {
		//the capacity of current file is enough, don't need to allocate new block
		buf.Data[blockNum] = append(buf.Data[blockNum][:offset], data...)
		buf.CurrentSize += len(data)
		return len(data)
	} else {
		//the capacity of current file is not  enough, we need to allocate new blocks from disk

		// calculate the number of blocks that we need to  allocate, then call the allocation method
		finalSize := len(data) + buf.CurrentSize
		numberOfNewBlocks := (finalSize % buf.BlockSize) - len(buf.blockNumbers)
		buf.allocNewBlock(numberOfNewBlocks)

		// firstly, fill the current block
		buf.Data[blockNum] = append(buf.Data[blockNum][:offset], data[:buf.BlockSize-offset]...)
		currentOffset += buf.BlockSize - offset
		buf.CurrentSize += buf.BlockSize - offset
	}
	for len(data)-currentOffset > 0 {
		var next int
		if (len(data) - currentOffset) > buf.BlockSize {
			next = buf.BlockSize
		} else {
			next = len(data) - currentOffset
		}
		blockNum += 1
		buf.setDataInsideBlock(blockNum, next, data[currentOffset:currentOffset+next])
		currentOffset += next
		buf.CurrentSize += next
	}

	// modify the attr "blockCount"
	return len(data)
}
func (buf *unifiedBuffer) modify(offset int, data []byte) {
	if offset+len(data)-buf.CurrentSize > 0 || len(data) == 0 {
		return
	} else {
		blockAddr := offset / buf.BlockSize

		currentOffset := 0
		buf.setDataInsideBlock(blockAddr, offset%buf.BlockSize, data[:buf.BlockSize-offset%buf.BlockSize])
		currentOffset += buf.BlockSize - offset%buf.BlockSize
		for len(data)-currentOffset > 0 {
			var next int
			if (len(data) - currentOffset) > buf.BlockSize {
				next = buf.BlockSize
			} else {
				next = len(data) - currentOffset
			}
			blockAddr += 1
			buf.setDataInsideBlock(blockAddr, next, data[currentOffset:currentOffset+next])
			currentOffset += next
		}
	}
}
func (buf *unifiedBuffer) setDataInsideBlock(num int, offsetInBlock int, data []byte) int {
	if len(data) > buf.BlockSize-offsetInBlock {
		return -1
	} else {
		var i int
		for i, x := range data {
			buf.Data[num][offsetInBlock+i] = x
		}
		return i
	}
}

// 如果offset大于currentSize 怎么办？
// 答案很简单，就用和dd类似的办法，中间的全部填写为0
// 要懂得复用代码
// currentSize内包含的部分就用modify，不包含的部分就用Write，注意切割
func (buf *unifiedBuffer) WriteAt(offset int, data []byte) int{
	// 假设 offset = 1025 则 start = 1
	endAddr := offset + len(data)
	if endAddr < buf.CurrentSize {
		buf.modify(offset, data)
	} else {
		if offset > buf.CurrentSize {
			empty := make([]byte, offset-buf.CurrentSize)
			buf.Write(empty)
			offset = buf.CurrentSize
		}
		buf.modify(offset, data[:buf.CurrentSize-offset])
		buf.Write(data[buf.CurrentSize-offset:])
	}
	return len(data)
}
func (buf unifiedBuffer) ReadAll()(result []byte ) {
	for _, x := range buf.Data {
		result = append(result, x...)
	}
	return result
}
