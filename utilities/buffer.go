package utilities

// 设计这样结构的动机是什么呢？
// 当你要读取一个文件，将它的全部块读取到内存中（有点暴力？）
// 因为文件的物理存储是离散的啊
// 果然还是要用LRU吗？
// 太过复杂，算了

// 事实证明，对内部的slice的操作可以直接影响

// 这个缓冲区就是
type UnifiedBuffer struct {
	Data        [][]byte
	BlockSize   int // 1kb
	CurrentSize int
}

func (u *UnifiedBuffer) Append(a []byte) {
	u.Data = append(u.Data, a)
}

// 向后写
func (u *UnifiedBuffer) Write(data []byte) int {
	blockAddr := u.CurrentSize / u.BlockSize
	if len(data) == 0 {
		return 0
	}
	currentOffset := 0
	offset := u.CurrentSize % u.BlockSize
	if offset == 0 {
		u.Data = append(u.Data, make([]byte, u.BlockSize))

	}
	if len(data)+offset <= u.BlockSize {
		u.Data[blockAddr] = append(u.Data[blockAddr][:offset], data...)
		u.CurrentSize += len(data)
		currentOffset += len(data)
	} else {
		u.Data[blockAddr] = append(u.Data[blockAddr][:offset], data[:u.BlockSize-offset]...)
		currentOffset += u.BlockSize - offset
		u.CurrentSize += u.BlockSize - offset
	}
	for len(data)-currentOffset > 0 {
		var next int
		if (len(data) - currentOffset) > u.BlockSize {
			next = u.BlockSize
		} else {
			next = len(data) - currentOffset
		}
		blockAddr += 1
		u.Data = append(u.Data, data[currentOffset:currentOffset+next])

		currentOffset += next
		u.CurrentSize += next
	}
	return len(data)
}
func (u *UnifiedBuffer) modify(offset int, data []byte) {
	if offset+len(data)-u.CurrentSize > 0 || len(data) == 0 {
		return
	} else {
		blockAddr := offset / u.BlockSize

		currentOffset := 0
		u.setDataInsideBlock(blockAddr, offset%u.BlockSize, data[:u.BlockSize-offset%u.BlockSize])
		currentOffset += u.BlockSize - offset%u.BlockSize
		for len(data)-currentOffset > 0 {
			var next int
			if (len(data) - currentOffset) > u.BlockSize {
				next = u.BlockSize
			} else {
				next = len(data) - currentOffset
			}
			blockAddr += 1
			u.setDataInsideBlock(blockAddr, next, data[currentOffset:currentOffset+next])
			currentOffset += next
		}
	}
}
func (u *UnifiedBuffer) setDataInsideBlock(num int, offsetInBlock int, data []byte) int {
	if len(data) > u.BlockSize-offsetInBlock {
		return -1
	} else {
		var i int
		for i, x := range data {
			u.Data[num][offsetInBlock+i] = x
		}
		return i
	}
}
func (u *UnifiedBuffer) Init(size int) {
	u.BlockSize = size
}

// 如果offset大于currentSize 怎么办？
// 答案很简单，就用和dd类似的办法，中间的全部填写为0
// 要懂得复用代码
// currentSize内包含的部分就用modify，不包含的部分就用Write，注意切割
func (u *UnifiedBuffer) WriteAt(offset int, data []byte) {
	// 假设 offset = 1025 则 start = 1
	endAddr := offset + len(data)
	if endAddr < u.CurrentSize {
		u.modify(offset, data)
	} else {
		if offset > u.CurrentSize {
			empty := make([]byte, offset-u.CurrentSize)
			u.Write(empty)
			offset = u.CurrentSize
		}
		u.modify(offset, data[:u.CurrentSize-offset])
		u.Write(data[u.CurrentSize-offset:])
	}
}
func (u UnifiedBuffer) ReadAll() []byte {
	result := make([]byte, u.CurrentSize)
	for _, x := range u.Data {
		result = append(result, x...)
	}
	return result
}
