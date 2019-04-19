package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const CONFIGURE_COMMAND = "configure"
const RUN_COMMAND = "run"
const FILE_NAME = "config.yaml"

type PortNumber int64

type Config struct {
	Redirects map[string]string `yaml:"redirects"`
}

type Redirect struct {
	Value string `yaml:"value"`
	URL   string `yaml:"url"`
}

func main() {
	fmt.Println("*** Assignment: URL Shortener by An Le ***")

	// Define flag set and input pointers
	configure := flag.NewFlagSet(CONFIGURE_COMMAND, flag.ExitOnError)
	appendValue := configure.String("a", "", "Append to redirection list")
	url := configure.String("u", "", "URL")

	run := flag.NewFlagSet(RUN_COMMAND, flag.ExitOnError)
	portNumber := run.Int64("p", 8080, "Port number")

	deleteValue := flag.String("d", "", "Delete from redirection list")
	list := flag.Bool("l", false, "List redirection list")
	help := flag.Bool("h", false, "Print usage info")

	if isNoCommand() {
		printUsageInfo()
		os.Exit(-1)
	}

	config := readYamlFile()

	subCommand := os.Args[1]
	subCommandArg := os.Args[2:]
	switch subCommand {
	case CONFIGURE_COMMAND:
		err := configure.Parse(subCommandArg)
		checkError(err)
		newRedirect := Redirect{*appendValue, *url}
		config.appendRedirect(newRedirect)
	case RUN_COMMAND:
		err := run.Parse(subCommandArg)
		checkError(err)
		startServer(*portNumber)
	default:
		flag.Parse()
		delValue := *deleteValue
		if *list {
			config.printAllRedirects()
		} else if delValue != "" {
			config.deleteRedirect(delValue)
		} else if *help {
			printUsageInfo()
		} else {
			fmt.Println("Exit")
			os.Exit(1)
		}
	}
}

func printUsageInfo() {
	usage := "Usage:\n" +
		"        ./urlshorten <command> [option]\n" +
		"The commands are:\n" +
		"        configure  configure yaml config file\n" +
		"        run        run server\n" +
		"The options are:\n" +
		"        -a         append to redirection list\n" +
		"        -u         URL\n" +
		"        -d         delete from redirection list\n" +
		"        -p         port number\n" +
		"        -l         usage info\n"
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, usage)
		checkError(err)
	}
	flag.Usage()
}

func startServer(portNumber int64) {
	port := PortNumber(portNumber).toString()
	fmt.Printf("Server is running on port: %d", portNumber)
	log.Fatal(http.ListenAndServe(port, nil))
}

func (p PortNumber) toString() string {
	return fmt.Sprintf(":%d", p)
}

func readYamlFile() Config {
	config := Config{}
	yamlFile, err := ioutil.ReadFile(FILE_NAME)
	checkError(err)
	err = yaml.Unmarshal(yamlFile, &config)
	checkError(err)
	return config
}

func (config Config) appendRedirect(newRedirect Redirect) {
	if newRedirect.valid() {
		config.Redirects[newRedirect.Value] = newRedirect.URL
		config.saveToYamlFile()
		fmt.Printf("%s is appended to config file!", newRedirect.displayValue())
	} else {
		printUsageInfo()
		os.Exit(-1)
	}
}

func (redirect Redirect) valid() bool {
	return redirect.Value != "" && redirect.URL != ""
}

func (config Config) printAllRedirects() {
	fmt.Println("Redirection list:")
	for value, url := range config.Redirects {
		redirect := Redirect{value, url}
		fmt.Println(redirect.displayValue())
	}
}

func (redirect Redirect) displayValue() string {
	format := "Value: %s - URL: %s"
	return fmt.Sprintf(format, redirect.Value, redirect.URL)
}

func (config Config) deleteRedirect(deleteValue string) {
	delete(config.Redirects, deleteValue)
	config.saveToYamlFile()
	fmt.Printf("%s is deleted from config file!", deleteValue)
}

func (config Config) saveToYamlFile() {
	changed, err := yaml.Marshal(&config)
	checkError(err)
	err = ioutil.WriteFile(FILE_NAME, changed, 0644)
	checkError(err)
}

func isNoCommand() bool {
	commandLength := len(os.Args)
	return commandLength < 2
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
