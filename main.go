package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	tailer "github.com/nxadm/tail"
	"github.com/radovskyb/watcher"
)

const (
	fileMode = iota
	directoryMode
	start
	stop
)

var (
	tails    sync.Map
	poll     bool
	exclude  []string
	json     bool
	minLevel int
)

type tailEvent struct {
	Op   int
	File string
}

func init() {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.BoolVar(&poll, "poll", false, "Use polling instead of inotify")
	fs.BoolVar(&json, "json", false, "Keep only JSON logs")
	fs.IntVar(&minLevel, "minlevel", 0, "Minimum log level")
	e := fs.String("exclude", "sasl|config|lock|@|gzip|tgz|gz", "blacklist (sep. by |): excludes any file containing any of those words in their fullpath.")
	fs.SetOutput(ioutil.Discard)
	fs.Parse(os.Args[1:])
	exclude = strings.Split(*e, "|")
}

func main() {
	c := make(chan *tailEvent)
	ff := filesFromArgs()
	go tail(c)
	if v, ok := ff[directoryMode]; ok {
		go watch(v, c)
	}
	for _, f := range ff[fileMode] {
		c <- &tailEvent{Op: start, File: f}
	}
	runtime.Goexit()
}

func watchFilter(info os.FileInfo, fullPath string) error {
	if info.IsDir() {
		return watcher.ErrSkip
	}
	for _, w := range exclude {
		if strings.Contains(fullPath, w) {
			return watcher.ErrSkip
		}
	}
	return nil
}

func watch(dirs []string, c chan *tailEvent) {
	w := watcher.New()
	w.AddFilterHook(watchFilter)
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove)
	go func() {
		for {
			select {
			case event := <-w.Event:
				switch event.Op {
				case watcher.Create, watcher.Write:
					_, ok := tails.Load(event.Path)
					if !ok {
						c <- &tailEvent{Op: start, File: event.Path}
					}
				case watcher.Remove:
					c <- &tailEvent{Op: stop, File: event.Path}
				default:
					panic(nil)
				}
			case err := <-w.Error:
				panic(err)
			case <-w.Closed:
				return
			}
		}
	}()
	for _, d := range dirs {
		if err := w.AddRecursive(d); err != nil {
			panic(err)
		}
	}
	if err := w.Start(time.Millisecond * 100); err != nil {
		panic(err)
	}
}

func tail(c chan *tailEvent) {
	for evt := range c {
		switch evt.Op {
		case start:
			go func() {
				t, err := tailer.TailFile(evt.File, tailer.Config{Poll: poll, Follow: true, ReOpen: true, Location: &tailer.SeekInfo{Whence: io.SeekEnd}, Logger: tailer.DiscardingLogger})
				if err != nil {
					panic(err)
				}
				tails.Store(evt.File, t)
				for line := range t.Lines {
					log(t.Filename, line.Text)
				}
			}()
		case stop:
			v, ok := tails.Load(evt.File)
			if ok {
				v.(*tailer.Tail).Stop()
				tails.Delete(evt.File)
			}
		default:
			panic(nil)
		}
	}
}

func filesFromArgs() map[int][]string {
	var filePerMode map[int][]string = make(map[int][]string)
	for _, f := range os.Args[1:] {
		if strings.HasPrefix(f, "/") {
			mode := fileMode
			if isDir(f) {
				mode = directoryMode
			}
			filePerMode[mode] = append(filePerMode[mode], f)
		}
	}
	return filePerMode
}

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func log(file, message string) {
	if len(message) == 0 {
		return
	}
	j := isJson(message)
	if json && !j {
		return
	}
	level, label := detectLevel(message)
	if level < minLevel {
		return
	}
	if !j {
		message = strconv.Quote(message)
	}
	date := time.Now().Format(time.RFC3339)
	component, subcomponent := components(file)
	fmt.Printf("{\"date\":\"%s\",\"component\": \"%s\",\"subcomponent\":\"%s\",\"level\":\"%s\",\"file\":%q,\"message\":%s}\n", date, component, subcomponent, label, file, message)
}

func detectLevel(message string) (int, string) {
	level := 100
	label := "unknown"
	message = strings.ToLower(message)
	debug := []string{"debug"}
	info := []string{"info", "log", "get", "post", "processing", "starting", "started", "completed", "success", "saving", "saved", "creating", "created"}
	notice := []string{"notice"}
	warn := []string{"warn"}
	err := []string{"error", "failed"}
	fatal := []string{"fatal", "emerg"}
	for _, w := range debug {
		if strings.Contains(message, w) {
			level = 1
			label = "debug"
			break
		}
	}
	for _, w := range info {
		if strings.Contains(message, w) {
			level = 2
			label = "info"
			break
		}
	}
	for _, w := range notice {
		if strings.Contains(message, w) {
			level = 3
			label = "notice"
			break
		}
	}
	for _, w := range warn {
		if strings.Contains(message, w) {
			level = 4
			label = "warning"
			break
		}
	}
	for _, w := range err {
		if strings.Contains(message, w) {
			level = 5
			label = "error"
			break
		}
	}
	for _, w := range fatal {
		if strings.Contains(message, w) {
			level = 6
			label = "fatal"
			break
		}
	}
	return level, label
}

func components(file string) (string, string) {
	parts := strings.Split(file, "/")
	component := parts[len(parts)-2:][0]
	subcomponent := parts[len(parts)-1:][0]
	subcomponent = strings.TrimSuffix(subcomponent, filepath.Ext(subcomponent))
	return component, subcomponent
}

func isJson(s string) bool {
	//cheap detection; must be good / fastest for such logs parsing
	return s[:1] == "{" && s[len(s)-1:] == "}"
}
