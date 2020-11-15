package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func usage(program string) {
	fmt.Printf(`
	arguments of %s:

		--cmd=[command that you want to run, need full path, "--cmd=xx xx" if there are spaces inside command]
		--interval=[interval seconds between execution of command]
		--time=[how many times to run command]
		--seconds=[how many seconds to run the test, if both --time and --seconds were given, use --seconds]
		--mode=[seq | con, seq means sequence execute, con means concurrent]
		--quiet  ignore output.
	hanfei@g-cloud.com.cn
	version 0.2, 2020.10.14
`, program)
}

type argument struct {
	command  string
	interval float64
	time     int
	seconds  int
	mode     string
	quiet    bool
	procs    chan *exec.Cmd  // passing *Cmd object.
	stopSig  chan int        // passing stop signal.
}

func main() {
	proName := filepath.Base(os.Args[0])
	argu := argument{}
	argu.procs = make(chan *exec.Cmd)
	argu.stopSig = make(chan int)
	var err error

	for _, arg := range os.Args[1:] {

		if arg == "--quiet" {
			argu.quiet = true
			continue
		}
		position := strings.IndexAny(arg, "=")
		switch arg[0:position] {
		case "--cmd":
			argu.command = arg[position+1:]
		case "--interval":
			argu.interval, err = strconv.ParseFloat(arg[position+1:], 64)
			if err != nil {
				fmt.Println(err.Error())
				usage(proName)
				os.Exit(-1)
			}
		case "--time":
			argu.time, err = strconv.Atoi(arg[position+1:])
			if err != nil {
				usage(proName)
				os.Exit(-1)
			}
		case "--seconds":
			argu.seconds, err = strconv.Atoi(arg[position+1:])
			if err != nil {
				usage(proName)
				os.Exit(-1)
			}
		case "--mode":
			argu.mode = arg[position+1:]
			if argu.mode != "seq" && argu.mode != "con" {
				usage(proName)
				os.Exit(-1)
			}
		default:
			usage(proName)
			os.Exit(-1)
		}
	}

	if len(argu.command) == 0 || len(argu.mode) == 0 {
		usage(proName)
		os.Exit(-1)
	}

	go func() {
		argu.output("command: %s, interval: %v, run seconds: %d, times:%d, mode=%s\n",
					argu.command, argu.interval, argu.seconds, argu.time, argu.mode)
		if err := argu.exec(); err != nil {
			os.Exit(-1)
		}
	}()

	// waiting stop signal and exit.
	for {
		cmd := <- argu.procs
		if cmd == nil {
			argu.output("all command processes finished.")
			os.Exit(0)
		}
		argu.output("one command running finished.")
		_ = cmd.Wait()
	}
}

func (arg *argument) _exec(path string, args []string) {
	var err error
	cmds := exec.Command(path, args...) //uments[1:]...)
	if !arg.quiet {
		cmds.Stderr = os.Stderr
		cmds.Stdout = os.Stdout
	}
	if arg.mode == "con" {
		// block by command
		err = cmds.Start()
	} else {
		err = cmds.Run()
	}
	if err != nil {
		arg.output("exec command return failed: %s\n", err.Error())
	} else {
		arg.procs <- cmds
	}
}

func (arg *argument) exec() error {
	arguments := strings.Split(arg.command, " ")
	path, _ := exec.LookPath(arguments[0])
	duration := time.Duration(arg.interval * 1000) * time.Millisecond

	go func() {
		// exec for a specific time
		if arg.seconds > 0 {
			for {
				arg._exec(path, arguments[1:])

				if arg.interval > 0 {
					time.Sleep(duration)
				}

				select {
				case <- arg.stopSig:
					arg.output("exec time is out.")
					break
				default:
					// next round.
				}
			}
		} else {
			// exec for counting times.
			for count:= 0; count < arg.time; count++ {
				arg.output("the %d time running command: [%s]", count+1, arg.command)
				arg._exec(path, arguments[1:])

				if arg.interval > 0 {
					time.Sleep(duration)
				}
			}
		}
	}()

	if arg.seconds > 0 {
		time.Sleep(time.Duration(arg.seconds) * time.Second)
		arg.stopSig <- 1
	}

	arg.output("Done.")
	arg.procs <- nil

	return nil
}

func (arg *argument) output(format string, a ...interface{}) {
	if !arg.quiet {
		fmt.Printf(format, a...)
		fmt.Println()
	}
}
