package main

import (
	"os"
	"fmt"
	"bufio"
	"os/exec"
	"regexp"
	"flag"
	"path/filepath"
)


const CONFIG_FILE = ".vim/vim-modules.conf"

type cmdFunc func() error
type cmd map[string] cmdFunc

type config struct {
	lines []string
}

var commands = cmd{
	"install": cmdInstall,
	"clean": cmdClean,
}

var dryRun bool
func main() {

	help := flag.Bool("h", false, "This help screen")
	flag.BoolVar(&dryRun, "dry-run", false, "Make no modifications")
	flag.Parse()

	if *help {
		showHelp()
		os.Exit(1)
	}

	input := flag.Arg(0)
	c := commands[input]
	if c == nil {
		fmt.Println("Invalid command:", input)
		showHelp()
		os.Exit(1)
	}

	err := c()
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Printf("%s [options] install | clear\n\nOptions:\n", os.Args[0])
	flag.VisitAll(func(f *flag.Flag){
		fmt.Printf(" - %s (%s) %s\n", f.Name, f.DefValue, f.Usage)
	})
	fmt.Println("")
}
func cdBundleDir() error {
	return os.Chdir(os.Getenv("HOME") + "/.vim/bundle")
}

func cmdClean() error {
	fmt.Printf("Cleaning modules ...")
	fmt.Printf("This is not implemented")
	return nil
}

func cmdInstall() error {

	err := os.Chdir(os.Getenv("HOME") + "/.vim/bundle")
	if err != nil {
		return fmt.Errorf("failed to cd into ./vim: %s", err)
	}

	conf, err := getConfig()
	if err != nil {
		return err
	}

	for _, module := range conf.lines {
		name := gitParseRepoName(module)
		if _, err := os.Stat(name); err == nil {
			fmt.Printf("Module '%s' already exists. Skipping.\n", name)
			continue
		}
		err := gitClone(module)
		if err != nil {
			return err
		}
	}
	return nil
}

func getConfig() (*config, error) {

	home := os.Getenv("HOME")
	path := home + "/" + CONFIG_FILE

	fd, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file '%s': %s", path, err)
	}

	defer fd.Close()

	scanner := bufio.NewScanner(fd)

	reComments := regexp.MustCompile(`#.*$`)
	reLeadingWs := regexp.MustCompile(`^\s*`)
	conf := &config{}
	for scanner.Scan() {
		line := scanner.Text()
		line = reComments.ReplaceAllString(line, "")
		line = reLeadingWs.ReplaceAllString(line, "")
		if line != "" {
			conf.lines = append(conf.lines, scanner.Text())
		}
	}


	return conf, nil
}

func gitParseRepoName(repo string) string {

	result := filepath.Base(repo)
	return regexp.MustCompile(`\.git`).ReplaceAllString(result, "")
}

func gitClone(repo string) error {
	if dryRun {
		fmt.Printf("git clone %s\n", repo)
		return nil
	}

	return runIt("git", "clone", repo)
}

func runIt(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
