package disk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path"
)

// 随机访问的磁盘
type Harddisk struct {
	storage   [2 << 20]byte // 20 Mib
	blockSize int
}

func (d Harddisk) ReadBlock(blockNumber int) []byte {
	return d.storage[blockNumber*d.blockSize : (blockNumber+1)*d.blockSize]
}
func (d *Harddisk) SetBlockSize(size int) {
	if size%2 != 0 || size < 0 {
		return
	}
	d.blockSize = size
}
func (d *Harddisk) SetBlock(newSlice []byte, order int) {
	base := d.blockSize * order
	if len(newSlice) < int(d.blockSize) {
		for i, x := range newSlice {
			d.storage[base+i] = x
		}
	}
}
func (d *Harddisk) UnsaveRead(begin int, end int) []byte {
	if end > len(d.storage) {
		fmt.Println(end)
		fmt.Println(len(d.storage))
	}
	return d.storage[begin:end]
}
func (d Harddisk) Dump() {
	fp, _ := os.Create(path.Join("dev", "zero.ext0fs"))
	defer fp.Close()
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, d.storage)
	if err != nil {
		fmt.Println(err)
	}
}
func (d *Harddisk) InitFromDiskFile(name string) {
	fp, _ := os.Open(name)
	defer fp.Close()
	dataSlice := d.storage[:]
	n, _ := fp.Read(dataSlice)
	if n != 20<<20 {
		fmt.Errorf("fatal error")
	}

}
