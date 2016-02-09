package main

import (
	// "bytes"
	"gopkg.in/fsnotify.v1"
	io "io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	// "strings"
)

func create_home() (in string, out string) {
	home := os.Getenv("DROPBOX_HOME")
	if len(home) > 0 {
		home = path.Join(home, ".dropbox-ssh")
	} else {
		log.Fatal("DROPBOX_HOME environment variable not set!")
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not get hostname!")
	} else {
		home = path.Join(home, hostname)
	}

	if _, err := os.Stat(home); os.IsNotExist(err) {
		err := os.MkdirAll(home, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	in = path.Join(home, "in")
	if _, err := os.Stat(in); os.IsNotExist(err) {
		err := io.WriteFile(in, []byte{}, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	out = path.Join(home, "out")
	if _, err := os.Stat(out); os.IsNotExist(err) {
		err := io.WriteFile(out, []byte{}, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	in, out := create_home()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	cmd := exec.Command("bash")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Open(out)
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdout = outFile
	cmd.Stderr = outFile

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					contents, _ := io.ReadFile(event.Name)
					if len(contents) > 0 {
						outFile.Sync()
						log.Print(string(contents))
						stdin.Write(contents)
						stdin.Write([]byte{'\n'})
						inFile, err := os.Open(in)
						if err != nil {
							log.Fatal(err)
						}
						inFile.WriteString("")
						inFile.Close()
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(in)
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	<-done
}
