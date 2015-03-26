package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

const configFile = ".vim/vim-modules.conf"

type cmdFunc func() error
type cmd map[string]cmdFunc

type config struct {
	lines []string
}

var commands = cmd{
	"install": cmdInstall,
	"clean":   cmdClean,
}

var dryRun bool
var saveModule bool

func main() {

	help := flag.Bool("h", false, "This help screen")
	flag.BoolVar(&dryRun, "dry-run", false, "Make no modifications")
	flag.BoolVar(&saveModule, "s", false, "Save the module to the config file")
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
	flag.VisitAll(func(f *flag.Flag) {
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

	// No module to install? use the config
	newModule := flag.Arg(1)
	if newModule == "" {
		conf, err := getConfig()
		if err != nil {
			return err
		}

		for _, module := range conf.lines {
			err = installOne(module)
			if err != nil {
				return err
			}
		}
	} else {
		err = installOne(newModule)
		if err != nil {
			return err
		}
		if saveModule {
			conf, err := getConfig()
			conf.lines = append(conf.lines, newModule)
			err = saveConfig(conf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func installOne(module string) error {

	name := gitParseRepoName(module)
	if _, err := os.Stat(name); err == nil {
		fmt.Printf("Module '%s' already exists. Skipping.\n", name)
		return nil
	}
	return gitClone(module)
}

func getConfig() (*config, error) {

	home := os.Getenv("HOME")
	path := home + "/" + configFile

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

func saveConfig(conf *config) error {
	home := os.Getenv("HOME")
	path := home + "/" + configFile

	fd, err := ioutil.TempFile(filepath.Dir(path), "vim-modules.conf")
	if err != nil {
		return err
	}

	buff := bufio.NewWriter(fd)
	for _, line := range conf.lines {
		buff.Write([]byte(line))
		buff.Write([]byte("\n"))
	}

	err = buff.Flush()
	if err != nil {
		return err
	}

	err = fd.Close()
	if err != nil {
		return err
	}
	return os.Rename(fd.Name(), path)
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
