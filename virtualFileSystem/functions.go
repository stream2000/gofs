package virtualFileSystem

import (
	cache "../LruCache"
	u "../utilities"
	"fmt"
	"github.com/fatih/color"
	"strconv"
	"strings"
	"time"
)

func Hash(fsMagic int, inodeNum int) string {
	return strconv.Itoa(fsMagic) + "|" + strconv.Itoa(inodeNum)
}

func (v Vfs) initPath(path string) (p Path) {

	path = v.parseRelativePath(path)

	p.pathSlice = strings.Split(path, "/")[1:]
	p.depth = len(p.pathSlice)
	p.currentIndex = 0
	p.pathString = path
	return
}

func (v Vfs) isMountPoint(p string) bool {
	//for _, x := range v.mountPointList {
	//	if p == x.pathString {
	//		return true
	//	}
	//}
	return false
}

// 查询过程中，很重要的一个点是判断当前的目录是不是一个挂载点
// 如果是的话，通过vfsmount结构可以得到当前目录的超级块
// 一个目录项若要成为挂载点，那么它首先应该存在，并且为空目录
func (v *Vfs) Init(sb SuperBlock) {
	sb.Init()
	v.rootSb = sb
	v.rootVnode.inode = sb.GetRoot()
	v.rootVnode.sb = sb
	v.curDir = v.initPath("/")
	v.mountPointList = append(v.mountPointList, v.initPath("/"))
	v.mount = append(v.mount, vfsMount{mountPoint: v.mountPointList[0], sb: sb, order: 0})
	v.inodeCache = cache.NewMemCache(30)
}

func (v Vfs) Pwd() {
	fmt.Println(v.curDir.pathString)
}
func (v Vfs) GetCur() string {
	return v.curDir.pathString
}
func (v Vfs) getInodeByPath(path string) (Inode, bool) {
	if path == "/" {
		v.curDir = v.initPath(path)
		return v.rootVnode.inode, true
	}

	p := v.initPath(path)
	path = p.pathString

	curInode := v.rootVnode.inode
	curMnt := v.mount[0]
	curDir := "/"
	curMnt.sb = v.rootSb
	for _, x := range p.pathSlice {
		curDir += x
		// FIXME mount
		if v.isMountPoint(curDir) {

		} else {
			newInodeNum := curInode.LookUp(x)

			if newInodeNum > 0 {
				hashValueOfCurInode := Hash(0, newInodeNum)
				// 从缓存中搜索inode
				cachedInode, flag := v.inodeCache.Get(hashValueOfCurInode)
				if flag {
					curInode = cachedInode.(Inode)
				} else {
					nInode := curMnt.sb.ReadInode(newInodeNum)
					// 将新读取的inode加入缓存
					nInode.SetSb(v.rootSb)
					v.inodeCache.Set(Hash(curMnt.order, newInodeNum), nInode)
					curInode = nInode
				}
			} else { // 并没有这样的目录项
				return curInode, false
			}
		}
	}
	curInode.SetSb(v.rootSb)
	return curInode, true
}

// 工作目录必须是有效的
func (v *Vfs) ChangeDir(path string) {

	p := v.initPath(path)
	path = p.pathString
	ino, ok := v.getInodeByPath(path)
	if ok {
		if ino.GetAttr().FileType != u.Directory {
			fmt.Println("这不是一个目录！")
		} else {
			v.curDir = v.initPath(path)
		}
	} else {
		fmt.Println("不存在这样的目录")
	}
}
func (v Vfs) GetFileListInCurrentDir() (list []string, ok bool) {
	ino, ok := v.getInodeByPath(v.curDir.pathString)
	if ok {
		if ino.GetAttr().FileType != u.Directory {
			fmt.Println("错误！当前项不是一个目录！")
		} else {
			list, ok := ino.List()
			if ok {
				return list, true
			}
		}
	} else {
		fmt.Println("当前目录不存在")
	}
	return
}

func (v *Vfs) ListCurrentDir() {
	ino, ok := v.getInodeByPath(v.curDir.pathString)
	if ok {
		if ino.GetAttr().FileType != u.Directory {
			fmt.Println("错误！当前项不是一个目录！")
		} else {
			list, ok := ino.List()
			if ok {
				for _, x := range list {
					in,_ := v.getInodeByPath(x)
					fmt.Println(beutifyString(in.GetAttr(),x))
				}

			}
		}
	} else {
		fmt.Println("当前目录不存在")
	}
}
func (v *Vfs) ListDir(path string) {
	ino, ok := v.getInodeByPath(path)

	if ok {
		fmt.Println(ino.GetAttr())
	} else {
		_ = fmt.Errorf("stat error: path %s not found", path)
	}
}
func (v Vfs) Stat(path string) {
	ino, ok := v.getInodeByPath(path)

	if ok {
		fmt.Println(ino.GetAttr())
	} else {
		_ = fmt.Errorf("stat error: path %s not found", path)
	}
}
func (v *Vfs) Touch(path string) {
	p := v.initPath(path)
	path = p.pathString
	if p.depth < 1 {
		return
	}
	parentPath, childName := p.splitParentAndChild()
flag:
	parentInode, ok := v.getInodeByPath(parentPath)
	if ok {
		v.rootSb.CreateFile(childName, parentInode, 1)
	} else {
		v.createParentDir(parentPath)
		goto flag
	}

}
func (v *Vfs) MakeDir(path string) {
	p := v.initPath(path)
	path = p.pathString
	if p.depth < 1 {
		return
	}
	parentPath, childName := p.splitParentAndChild()

flag:
	parentInode, ok := v.getInodeByPath(parentPath)
	if ok {
		if parentInode.GetAttr().FileType != u.Directory {
			fmt.Println("mkdir error: ", "path: ", parentPath, " is not a directory")
			return
		}
		v.rootSb.CreateFile(childName, parentInode, int(u.Directory))
	} else {
		v.createParentDir(parentPath)
		fmt.Println("flag!")
		goto flag
	}
}
func (p Path) splitParentAndChild() (parent string, child string) {

	if p.depth == 1 {
		return "/", p.pathSlice[0]
	}
	parent = "/"
	for i, x := range p.pathSlice[:p.depth-1] {
		if i != p.depth-2 {
			parent += x + "/"
		} else {
			parent += x
		}
	}
	child = p.pathSlice[p.depth-1]
	return
}

func (v Vfs) parseRelativePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		// 说明要解析的是相对路径
		if v.curDir.pathString == "/" {
			path = "/" + path
		} else {
			path = v.curDir.pathString + "/" + path
		}
	}
	return path
}
func (v *Vfs) createParentDir(path string) {
	root := v.rootVnode
	p := v.initPath(path)
	path = p.pathString
	if p.depth < 1 {
		return
	}
	curInode := root.inode
	curSb := v.rootSb
	for _, x := range p.pathSlice {
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
}
func (v *Vfs) Remove(path string) {
	p := v.initPath(path)
	path = p.pathString
	if p.depth < 1 {
		return
	}
	parentPath, childName := p.splitParentAndChild()

	parentInode, ok := v.getInodeByPath(parentPath)

	if ok && parentInode.GetAttr().FileType == u.Directory {
		parentInode.SetSb(v.rootSb)
		ok = parentInode.Remove(childName)
		if !ok {
			fmt.Errorf("fatal error, can't delete file ", path)
		}
	} else {
		_ = fmt.Errorf("Not such file ", path)
	}
}
func beutifyString(attr InodeAttr,name string)string{
	tm := time.Unix(int64(attr.Ctime), 0)
	blue := color.New(color.FgHiCyan).SprintFunc()
	time := fmt.Sprintf(tm.Format("2006-01-02 03:04:05 PM"))
	return fmt.Sprintf("drwxr-xr-x %db %-5s %-10s",attr.Size,time,blue(name))
}
func (v *Vfs)Append(path string,data string){
	p,ok := v.getInodeByPath(path)
	if ok{
		p.Append(data)
	}else {
		fmt.Println("not fount")
	}
}
func (v *Vfs)Cat(path string){
	ino, ok := v.getInodeByPath(path)

	if ok {
		data := ino.ReadAll()
		fmt.Println(string(data))
	} else {
		_ = fmt.Errorf("stat error: path %s not found", path)
	}
}