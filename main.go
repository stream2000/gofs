package main

import (
	"./ext0"
	"./virtualFileSystem"
	"fmt"
	"github.com/abiosoft/ishell"
	"time"

	"github.com/fatih/color"
)

func main() {
	shell()
}

func testInstructions(){
	//var sbI virtualFileSystem.SuperBlock = new (ext0.Ext0SuperBlock)
	var sb virtualFileSystem.SuperBlock
	sb = &ext0.Ext0SuperBlock{}
	sb.Init(true)
	//	defer sb.(*ext0.Ext0SuperBlock).Dump()
	var v virtualFileSystem.Vfs
	v.Init(sb)
	v.Touch("new")
	v.ListCurrentDir()
	v.MakeDir("mnt/kk/kk")
	v.ListCurrentDir()
	v.Remove("mnt")
	v.ListCurrentDir()
	v.Stat("new")
	v.Touch("new2")
	v.Stat("new2")
	v.MakeDir("mnt/new/mkmkk/lo")
	v.ChangeDir("mnt")
	v.Remove("new")
	v.Touch("kk")
	v.Stat("kk")
}
func printMessage(s [3]string) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgHiYellow).SprintFunc()
	blue := color.New(color.FgHiCyan).SprintFunc()
	nBlule := color.New(color.FgHiBlue).SprintFunc()
	t := time.Now()
	time := fmt.Sprintf(t.Format("03:04:05"))
	fmt.Printf("%s %s @ %s in %s [%s]\n", nBlule("#"), blue(s[0]), green(s[1]), yellow(s[2]), time)
}

func shell(){
	var sb virtualFileSystem.SuperBlock
	sb = &ext0.Ext0SuperBlock{}
	sb.Init(true)
	defer sb.(*ext0.Ext0SuperBlock).Dump()
	var v virtualFileSystem.Vfs
	v.Init(sb)
	shell := ishell.New()
	red := color.New(color.FgHiRed).SprintFunc()
	shell.SetPrompt(red("$ "))
	// display welcome info.
	var promptMsg = [3]string{"fuqijun", "My-Arch-Linux", "/"}
	printMessage(promptMsg)

	shell.AddCmd(&ishell.Cmd{
		Name: "ls",
		Help: "list",
		Func: func(c *ishell.Context) {
			v.ListCurrentDir()
			printMessage(promptMsg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "cd",
		Help: "change work directory",
		Completer: func([]string) []string {
			dir, _ := v.GetDiristInCurrentDir()
			return dir
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				v.ChangeDir("/")
			} else {
				v.ChangeDir(c.Args[0])
			}
			promptMsg[2] = v.GetCur()
			printMessage(promptMsg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "pwd",
		Help: "print work directory",
		Func: func(c *ishell.Context) {
			v.Pwd()
			printMessage(promptMsg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "touch",
		Help: "list",
		Func: func(c *ishell.Context) {

			if len(c.Args) == 0 {
				_ = fmt.Errorf("Touch error: you must input the name")
			} else {
				v.Touch(c.Args[0])
			}
			printMessage(promptMsg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "mkdir",
		Help: "list",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				_ = fmt.Errorf("mkdir error: you must input the name")
			} else {
				v.MakeDir(c.Args[0])
			}
			printMessage(promptMsg)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "stat",
		Help: "view the information of file",
		Completer: func([]string) []string {
			dir, _ := v.GetFileListInCurrentDir()
			return dir
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				_ = fmt.Errorf("stat error: you must input the name")
			} else {
				v.Stat(c.Args[0])
			}
			printMessage(promptMsg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "rm",
		Help: "delete file or dir,will delete all its children at the same time",
		Completer: func([]string) []string {
			dir, _ := v.GetFileListInCurrentDir()
			return dir
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				_ = fmt.Errorf("stat error: you must input the name")
			} else {
				v.Remove(c.Args[0])
			}
			printMessage(promptMsg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "append",
		Help: "append some text to the file",
		Completer: func([]string) []string {
			dir, _ := v.GetFileListInCurrentDir()
			return dir
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				_ = fmt.Errorf("append error: params error")
			} else {
				yellow := color.New(color.FgHiYellow).SprintFunc()
				c.Printf("%s", yellow("Input multiple lines and end with semicolon ';'.\n"))
				// 设置结束符
				lines := c.ReadMultiLines(";")
				c.Printf("%s", yellow("Input finished\n"))
				v.Append(c.Args[0], lines)
			}
			printMessage(promptMsg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "cat",
		Help: "read file to stdin",
		Completer: func([]string) []string {
			dir, _ := v.GetFileListInCurrentDir()
			return dir
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				_ = fmt.Errorf("cat error: you must input the name")
			} else {
				v.Cat(c.Args[0])
			}
			printMessage(promptMsg)
		},
	})
	// run shell
	shell.Run()
}