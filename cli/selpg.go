package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"

	"github.com/spf13/pflag"
)

type selpg_args struct {
	startPage      int
	endPage        int
	inFile         string
	pageLen        int
	pageType       bool // true for -f, false for -lNumber
	outDestination string
}

func main() {

	var args selpg_args

	// 解析获取参数
	getArgs(&args)

	// 检查参数合法性
	checkArgs(&args)

	// 执行命令
	processInput(&args)
}

// 解析获取参数
func getArgs(args *selpg_args) {
	pflag.IntVarP(&(args.startPage), "startPage", "s", -1, "start page")
	pflag.IntVarP(&(args.endPage), "endPage", "e", -1, "end page")
	pflag.IntVarP(&(args.pageLen), "pageLen", "l", 72, "the length of page")
	pflag.BoolVarP(&(args.pageType), "pageType", "f", false, "page type")
	pflag.StringVarP(&(args.outDestination), "outDestination", "d", "", "print destination")
	pflag.Parse()

	other := pflag.Args() // 其余参数
	if len(other) > 0 {
		args.inFile = other[0]
	} else {
		args.inFile = ""
	}
}

// 检查参数合法性
func checkArgs(args *selpg_args) {
	if args.startPage == -1 || args.endPage == -1 {
		os.Stderr.Write([]byte("You shouid input like selpg -sNumber -eNumber ... \n"))
		os.Exit(0)
	}

	if args.startPage < 1 || args.startPage > math.MaxInt32 {
		os.Stderr.Write([]byte("You should input valid start page\n"))
		os.Exit(0)
	}

	if args.endPage < 1 || args.endPage > math.MaxInt32 || args.endPage < args.startPage {
		os.Stderr.Write([]byte("You should input valid end page\n"))
		os.Exit(0)
	}

	if (!args.pageType) && (args.pageLen < 1 || args.pageLen > math.MaxInt32) {
		os.Stderr.Write([]byte("You should input valid page length\n"))
		os.Exit(0)
	}
}

// 执行命令
func processInput(args *selpg_args) {

	// read the file
	var reader *bufio.Reader

	if args.inFile == "" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		fileIn, err := os.Open(args.inFile)
		defer fileIn.Close()
		if err != nil {
			os.Stderr.Write([]byte("Open file error\n"))
			os.Exit(0)
		}
		reader = bufio.NewReader(fileIn)
	}

	// output the file
	if args.outDestination == "" {
		// 输出到当前命令行
		outputCurrent(reader, args)
	} else {
		// 输出到目的地
		outputToDest(reader, args)
	}
}

// 输出到当前命令行
func outputCurrent(reader *bufio.Reader, args *selpg_args) {
	writer := bufio.NewWriter(os.Stdout)
	lineCtr := 0
	pageCtr := 1

	endSign := '\n'
	if args.pageType == true {
		endSign = '\f'
	}

	for {
		strLine, errR := reader.ReadBytes(byte(endSign))
		if errR != nil {
			if errR == io.EOF {
				writer.Flush()
				break
			} else {
				os.Stderr.Write([]byte("Read bytes from reader fail\n"))
				os.Exit(0)
			}
		}

		if pageCtr >= args.startPage && pageCtr <= args.endPage {
			_, errW := writer.Write(strLine)
			if errW != nil {
				os.Stderr.Write([]byte("Write bytes to out fail\n"))
				os.Exit(0)
			}
		}

		if args.pageType == true {
			pageCtr++
		} else {
			lineCtr++
		}

		if args.pageType != true && lineCtr == args.pageLen {
			lineCtr = 0
			pageCtr++
			if pageCtr > args.endPage {
				writer.Flush()
				break
			}
		}
	}
	checkPageNum(args, pageCtr)
}

// 输出到指定目的地
func outputToDest(reader *bufio.Reader, args *selpg_args) {
	cmd := exec.Command("./" + args.outDestination)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	startErr := cmd.Start()
	if startErr != nil {
		fmt.Println(startErr)
		os.Exit(0)
	}

	lineCtr := 0
	pageCtr := 1

	endSign := '\n'
	if args.pageType == true {
		endSign = '\f'
	}

	for {
		strLine, errR := reader.ReadBytes(byte(endSign))
		if errR != nil {
			if errR == io.EOF {
				break
			} else {
				os.Stderr.Write([]byte("Read bytes from reader fail\n"))
				os.Exit(0)
			}
		}

		if args.pageType == true {
			pageCtr++
		} else {
			lineCtr++
		}

		if pageCtr >= args.startPage && pageCtr <= args.endPage {
			_, errW := stdin.Write(strLine)
			if errW != nil {
				fmt.Println(errW)
				os.Stderr.Write([]byte("Write bytes to out fail\n"))
				os.Exit(0)
			}
		}
		if args.pageType != true && lineCtr == args.pageLen {
			lineCtr = 0
			pageCtr++
			stdin.Write([]byte("\f"))
			if pageCtr > args.endPage {
				break
			}
		}
	}

	stdin.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	checkPageNum(args, pageCtr)
}

// 检查开始页号与结束页号的实际合理性
func checkPageNum(args *selpg_args, pageCtr int) {

	if pageCtr < args.startPage {
		os.Stderr.Write([]byte("Start page is bigger than the total page num\n"))
		os.Exit(0)
	}

	if pageCtr < args.endPage {
		os.Stderr.Write([]byte("End page is bigger than the total page num\n"))
		os.Exit(0)
	}
}