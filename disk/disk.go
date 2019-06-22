package disk

import (
	"fmt"
	"io/ioutil"
	"os"
)

// 随机访问的磁盘
type HardDisk struct {
	storage   []byte // 20 Mib
	blockSize int
}

func (d HardDisk) ReadBlock(blockNumber int) []byte {
	return d.storage[blockNumber*d.blockSize : (blockNumber+1)*d.blockSize]
}
func (d *HardDisk) setBlockSize(size int) {
	if size%2 != 0 || size < 0 {
		return
	}
	d.blockSize = size
}
func (d *HardDisk) UnsaveRead(begin int, end int) []byte {
	if end > len(d.storage) {
		fmt.Println(end)
		fmt.Println(len(d.storage))
	}
	return d.storage[begin:end]
}
func (d HardDisk) Dump() {

	err := ioutil.WriteFile("ext0fs.bk", d.storage, 0777)
	if err != nil {
		_ = fmt.Errorf("error when dump")
	}
}
func NewDisk(path string, format bool, BlockSize int) (d HardDisk, ok bool) {
	if BlockSize%2 != 0 || BlockSize < 0 {
		fmt.Println("wrong block size")
		return
	}
	if format {
		d = HardDisk{
			storage:   make([]byte, 2<<20),
			blockSize: BlockSize,
		}
		return d, true
	} else {
		d.storage = make([]byte, 2<<20)
		d.blockSize = BlockSize

		fp, err := os.Open(path)
		defer fp.Close()
		if err != nil {
			fmt.Errorf("Read file error")
			return d, false
		} else {
			dataSlice := d.storage
			n, _ := fp.Read(dataSlice)
			if n < 2<<20 {
				_ = fmt.Errorf("fatal error,size is not fit,expect to read %d, in fact %d", 2<<20, n)
				return d, false
			}
		}
	}
	return d, true
}
