设计
我的文件系统取名叫ext0,原因在于他是对linux文件系统ext2虚拟文件系统的一次拙劣模仿。ext0从上至下分为4层，每个层都用一个go语言的包表示（可以理解为子系统）。

最底层是物理存储层disk，我将它设置为提供随机访问的磁盘。上层可以通过接口一次性读取一整个块。

第二层是exto文件系统层。这个层实现了对底层裸数据的管理。通过五个控制信息块，即超级块，inode表，inode位图，Fat表，块位图实现了对数据区的控制管理。

第三层是虚拟文件系统层。这个层首先定义两个接口，分别是inode和超级块，实现了inode和超级块接口的文件系统可以完美的接入文件系统进行管理。该层通过调用下层接口，将创建删除文件，读写文件等封装成以路径名为参数的接口。

第四层是应用层，实现了用户友好的命令行交互界面。

## 难点和实现
我将列出一些我在编程中遇到的难点以及他们的解决办法

### 1. 整体架构
设计之初最重要的是整体的架构，我一开始就看上了linux的虚拟文件系统，因为我自己在使用linux系统的时候可以感受到虚拟文件系统带来的统一访问方式的便利。

为了解决整个架构的问题，我画了两天天的时间尝试理解linux的ext2文件系统的源码，了解了linux文件系统在磁盘内的存储和内存中的表示。但是仍然有许多的细节不是十分理解。

### 2. 磁盘中数据结构的设计
由于索引访问的控制变量较多，出于简便我选择了FAT+Bitmap的实现方式，再加上超级块，inode表，inode bitmap 和数据区，磁盘中一共存储了六种数据。

### 3. 对磁盘区数据Rawdata的封装
由于文件系统的底层就是一个没有任何装饰的数组，需要对这个数组的数据进行一些封装，以简化读取数据，修改和读取Fat，修改和读取位图等操作。

具体解决办法就是为各个数据区设置读写函数。由此在极度仿真的情况下，也取得了极其便利的操作方式。

### 3. 目录文件的存储
这里的目录文件不是指的inode类型的索引目录，而是在数据区存储的目录。我一开始十分不明白目录区要存储什么样的数据，后来选择了最简单的，将每一个目录项设置为定长文件名+inode编号的方式。

### 4. 对磁盘空间的管理
在磁盘空间的管理上我采用的是Fat和Bitmap方案，这是一种链接存储的方案，所以给空间的管理增加了一点麻烦。最核心的一点是，如何做到当文件的size变大的时候分配空间，当文件的size变小的时候自动释放并回收空间。

我的解决方案是设计一个缓冲区unifiedBuffer,这个缓冲区有几种方法，1、初始化。给定一个inode节点，根据他fat表的起始地址初始化存储它的数据区，会自动的读取所有该inode的数据。2、自动管理空间的读写Api。如写文件的基本api，WriteAt顾名思义会给定一个文件内偏移并写入数据，当修改后文件尺寸增加以至于现有的块不能容纳，会自动的向超级块申请空间并修改块位图和Fat表。再比如Trunc函数，会自动截取文件（在我的代码里该函数叫Resize），如果截取后的文件所需要的块空间变少，也会自动的向超级块申请释放空间。Resize（0）就是完全释放文件的数据区。

缓冲区名unifiedBuffer正如其名，使用者完全不用担心数据是离散存放的还是连续存放的，在用户看来这些数据都统一为了unifiedBuffer，用户只需要对unifiedBuffer进行读写，底层会自动完成空间的分配与释放。这个buffer的实现是我本次代码的亮点之一。

### 5. superBlock函数和inode函数的实现
superBlock接口如下

```go
type SuperBlock interface {
   Format()
   ReadInode(number int) Inode          // when create an vInode,read it from disk
   WriteInode(number int, data InodeAttr) // write back inode to disk
   RecoverFromDisk()
   Init(format bool)
   GetRoot() Inode
   CreateFile(name string, p Inode, mode int)
}
```

其中经常被上层调用的是1. readInode，将一个inode从数据区读到内存中 2. CreateFile，给定一个父目录，创建它的孩子

Inode接口如下

```go
type Inode interface {
   Create()
   Link()       // create a hard link
   FollowLink() //follow a symbolic link to the real path
   LookUp(name string) int
   SeAttr(data InodeAttr)
   GetAttr() InodeAttr
   List() ([]string, bool)
   ReadAll() []byte
   WriteAt(offset int, data []byte) int
   Append(data string) int
   Remove(name string) bool
   GetSb() SuperBlock
   SetSb(block SuperBlock)
}
```

inode的函数运用的都比较频繁。

### CLI的实现
cli使用go语言的一个cli交互库ishell，通过这个库，我实现了命令解析，自动补全的功能。

##总结

先说一说我的这个项目中和虚拟文件系统相关的部分。我用了大量的时间和精力将虚拟文件层和逻辑文件系统层分离，达成了vfs包完全不用考虑底层实现而运作的。但是有一点非常遗憾，就是时间的不充足，导致我没有办法继续雕琢我的文件系统。本来作为虚拟文件系统，理所当然的应该拥有挂载多文件系统的能力，以及软硬链接都应该能基本实现了，但是我只来得及达成基本功能，就草草收手了。

我的模拟虚拟文件系统分层清晰，实现了文件系统，下层都向上层提供了强有力的api供调用。个人认为一定程度上做到了高内聚低耦合的设计。但是时间关系不能做到完美，只能将一些开始时的设计做一个说明。

1. 多系统挂载。 所有的文件操作都离不开解析目录这个过程。而多系统挂载的玄机就在解析目录这里。我的原计划是对现有的目录解析做一个小修改，当系统解析到当前的目录是一个挂载点（当一个文件系统被挂载，它的超级块信息就会被注册，所以挂载信息可以从注册表中查到），就会将当前执行职能的超级块切换为被挂载系统的超级块，继续进行目录解析。注意到多文件系统都实现了虚拟文件系统的接口，所以操作起来并没有太大的差别。

2. 缓存机制。在linux文件系统一个很大的区别就是物理数据结构和内存数据结构。我在高度仿真了物理数据结构以后并没有因为性能问题设置内存中数据的缓存。导致的一个问题很多频繁的操作会进行无意义重复。 其实我一开始也实现了一个inode在内存中的lru缓存，但是因为在层叠删除子文件这个操作下，来不及考虑内存和物理存储的一致性，可能会导致读取错误的节点，所以将缓存功能暂时删去。

3. 链接功能。老实说，按照我的架构，实现软硬链接都是十分简单的。采用了inode架构后，文件和文件名字相互分离。 对于硬链接，只需要在多个directory下不同的文件名指向同一个文件就好了。但是硬链接有一个特点，那就是硬链接数大于1的文件无法删除。 由此衍生出很多管理成本。其实，我已经部分实现了硬链接，只是没有向cli提供接口。我在每个文件目录（directory）创建的时候，会自动增加两个项”.“ 和 ”..“ ，分别是本身和父目录的硬链接。对于软连接，还是目录解析的问题，我在程序实现了通过路径获得inode的功能，实现软连接需要在额外的目录解析（也就是取得软链接后还要读取其数据解析它指向的地址），我并没有时间去完成这个。

## 测试
我给出的源码版本是会自动读取历史文件并写入历史文件的，所以一开始就初始化了几个一些文件和目录。

操作说明：
append      append some text to the file
cat         read file to stdin
cd          change work directory
clear       clear the screen
exit        exit the program
format      format the disk
help        display help
ls          list
mkdir       make dir
pwd         print work directory
rm          delete file or dir, will delete all its children at the same time
stat        view the information of file
touch       create new empty file

基本和linux文件系统一样，只是对修改文件支持十分残缺，苦于找不到合适的命令行编辑工具，只能做了一个append向文件尾部添加一些数据。

**提供一个mac可执行文件和一个linux可执行文件，在dev文件夹中存放了一个文件系统文件的备份。**

