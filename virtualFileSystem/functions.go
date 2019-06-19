package virtualFileSystem

import (
	u "../utilities"
	"fmt"
	"strings"
)

//func (fs *fileSystem)ReadSuperBlock(){
//
//}
//
//func (fs *fileSystem)registerFileSystem(){
//
//
//}
// a simplest hash function
func Hash(fsMagic int, inodeNum int) string {
	return string(fsMagic*1<<20 + inodeNum)
}

func (v Vfs) parsePathName(name string) (p Path) {
	p.pathSlice = strings.Split(name, "/")[1:]
	p.depth = len(p.pathSlice)
	p.currentIndex = 0
	p.pathString = name
	return
}

func (v Vfs) isMountPoint(p string) bool {
	for _, x := range v.mountPointList {
		if p == x.pathString {
			return true
		}
	}
	return false
}

// 查询过程中，很重要的一个点是判断当前的目录是不是一个挂载点
// 如果是的话，通过vfsmount结构可以得到当前目录的超级块
// 一个目录项若要成为挂载点，那么它首先应该存在，并且为空目录
func (v *Vfs) Init(sb SuperBlock) {
	v.rootSb = sb

	sb.Init()
	//v.rootVnode.inode = sb.ReadInode(0)
	//v.rootVnode.sb = sb
	//v.rootVnode.data = v.rootVnode.inode.GetAttr()
	v.rootVnode.inode = sb.GetRoot()
	v.rootVnode.sb = sb
	v.curDir = v.parsePathName("/")
	v.mountPointList = append(v.mountPointList, v.parsePathName("/"))
}

func (v Vfs) Pwd() {
	fmt.Println(v.curDir.pathString)
}

func (v Vfs) GetInodeByPath(path string) (ino Inode, ok bool) {
	root := v.rootVnode
	if path == "/" {
		return root.inode, true
	}
	p := v.parsePathName(path)
	if p.depth < 1 {
		return ino, false
	}
	ino = root.inode
	curSb := v.rootSb
	for _, x := range p.pathSlice {
		num := ino.LookUp(x)
		if num == 0 {
			fmt.Println("错误，没有找到 ", x)
			return ino, false
		} else {
			ino = curSb.ReadInode(num)
		}

	}
	return ino, true
}

// 工作目录必须是有效的
func (v *Vfs) ChangeDir(path string) {
	ino, ok := v.GetInodeByPath(path)
	if ok {
		if ino.GetAttr().FileType != u.Directory {
			fmt.Println("这不是一个目录！")
		} else {
			v.curDir = v.parsePathName(path)
		}
	} else {
		fmt.Println("不存在这样的目录")
	}
}
func (v Vfs) Ls() {
	ino, ok := v.GetInodeByPath(v.curDir.pathString)
	if ok {
		if ino.GetAttr().FileType != u.Directory {
			fmt.Println("错误！当前项不是一个目录！")
		} else {
			ino.List()
		}
	} else {
		fmt.Println("当前目录不存在")
	}
}
func (v *Vfs) Touch(name string) {
	root := v.rootVnode
	p := v.parsePathName(name)
	if p.depth < 1 {
		return
	}
	curInode := root.inode
	curSb := v.rootSb
	for _, x := range p.pathSlice[:p.depth-1] {
		num := curInode.LookUp(x)
		if num == 0 {
			curSb.CreateFile(x, curInode, 2)
			num := curInode.LookUp(x)
			if num > 0 {
				curInode = curSb.ReadInode(num)
			}
		} else {

		}
	}
	curSb.CreateFile(p.pathSlice[p.depth-1], curInode, 1)
}
func (p Path) GetPart(depth int) (name string) {
	name += "/"
	for i, x := range p.pathSlice {
		if i < depth {
			name += x
		} else {
			return
		}
	}
	return
}

func (v Vfs) Open(name string) {
	//
	//d,flag := v.inodeCache.Get("123")
	//if flag{
	//	d.(Inode).LookUp("mnt")
	//}
	p := v.parsePathName(name)

	curInode := v.rootVnode
	curMnt := v.mount[0]
	curDir := "/"

	for _, x := range p.pathSlice {
		curDir += x
		if v.isMountPoint(curDir) {

		} else {
			newInodeNum := curInode.inode.LookUp(x)

			if newInodeNum > 0 {
				hashValueOfCurInode := Hash(curMnt.order, newInodeNum)
				// 从缓存中搜索inode
				cachedInode, flag := v.inodeCache.Get(hashValueOfCurInode)
				if flag {
					curInode = vfsInode{
						inode: cachedInode.(Inode),
						sb:    curMnt.sb,
					}
				} else {
					nInode := curMnt.sb.ReadInode(newInodeNum)
					// 将新读取的inode加入缓存
					v.inodeCache.Set(Hash(curMnt.order, newInodeNum), nInode)
					curInode = vfsInode{
						inode: nInode,
						sb:    curMnt.sb,
					}
				}
			} else { // 并没有这样的目录项
				return
			}
		}
	}

}
