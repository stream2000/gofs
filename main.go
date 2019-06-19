package main

import (
	"./ext0"
	"./virtualFileSystem"
	"fmt"
	"github.com/abiosoft/ishell"

	"github.com/fatih/color"
	"strconv"
)

func main() {
	//var sbI virtualFileSystem.SuperBlock = new (ext0.Ext0SuperBlock)

	var sb virtualFileSystem.SuperBlock
	sb = &ext0.Ext0SuperBlock{}
	var v virtualFileSystem.Vfs
	v.Init(sb)
	var words  = []string{"fuqijun@My-Arch-Linux in /usr/bin","qwqw"}
	var count int
	blue := color.New(color.FgHiRed).SprintFunc()
	var msg =[3]string{"fuqijun", "My-Arch-Linux","/usr/bin"}

	//v.Touch("/mnt/win/go.test")
	//v.Touch("/bin")
	//v.Touch("/var")
	//v.ChangeDir("/mnt/win")
	//v.Pwd()
	//v.Ls()
	//v.ChangeDir("/")
	//v.Ls()
	shell := ishell.New()
	shell.SetPrompt(blue("$ "))
	// display welcome info.
	printMessage(msg)
	// register a function for "greet" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "ls",
		Help: "list",
		Func: func(c *ishell.Context) {
			v.Ls()
			printMessage(msg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "cd",
		Help: "list",

		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				v.ChangeDir("/")
			} else {
				v.ChangeDir(c.Args[0])
			}
			printMessage(msg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "pwd",
		Help: "list",
		Func: func(c *ishell.Context) {
			v.Pwd()
			printMessage(msg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "touch",
		Help: "list",
		Func: func(c *ishell.Context) {
			v.Touch(c.Args[0])
			printMessage(msg)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "test",
		Help: "list",
		Completer: func([]string) []string {
			return words
		},
		Func: func(c *ishell.Context) {
			count += 1
			var s string
			s += "test" + strconv.Itoa(count)
			words = append(words, s)
			printMessage(msg)
		},
	})
	// run shell
	shell.Run()
}
func printMessage(s [3]string){
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgHiYellow).SprintFunc()
	blue := color.New(color.FgHiCyan).SprintFunc()
	nBlule := color.New(color.FgHiBlue).SprintFunc()
	fmt.Printf("%s %s @ %s in %s\n",nBlule("#"),blue(s[0]),green(s[1]),yellow(s[2]))
}