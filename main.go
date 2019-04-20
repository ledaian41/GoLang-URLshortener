package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const CONFIGURE_COMMAND = "configure"
const RUN_COMMAND = "run"
const FILE_NAME = "config.yaml"
const DEFAULT_PORT = 8080
const LETTER_BYTES = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Config struct {
	Redirects map[string]UrlInfo `yaml:"redirects"`
}

type UrlInfo struct {
	Url  string `yaml:"url"`
	Used int64  `yaml:"used"`
}

func main() {
	fmt.Println("*** Assignment: URL Shortener by An Le ***")

	// Define flag set and input pointers
	configure := flag.NewFlagSet(CONFIGURE_COMMAND, flag.ExitOnError)
	appendValue := configure.String("a", "", "Append to redirection list")
	url := configure.String("u", "", "URL")
	run := flag.NewFlagSet(RUN_COMMAND, flag.ExitOnError)
	portNumber := run.Int64("p", DEFAULT_PORT, "Port number")
	deleteValue := flag.String("d", "", "Delete from redirection list")
	list := flag.Bool("l", false, "List redirection list")
	help := flag.Bool("h", false, "Print usage info")

	if isNoCommand() {
		fmt.Println("Exit")
		os.Exit(1)
	}
	config := readYamlFile()
	subCommand := os.Args[1]
	subCommandArg := os.Args[2:]
	switch subCommand {
	case CONFIGURE_COMMAND:
		err := configure.Parse(subCommandArg)
		checkError(err)
		err = config.appendRedirect(*appendValue, *url)
		checkError(err)
	case RUN_COMMAND:
		err := run.Parse(subCommandArg)
		checkError(err)
		setRedirectHandler(config)
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

func readYamlFile() Config {
	config := Config{}
	yamlFile, err := ioutil.ReadFile(FILE_NAME)
	checkError(err)
	err = yaml.Unmarshal(yamlFile, &config)
	checkError(err)
	return config
}

func setRedirectHandler(config Config) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[1:]
		if urlInfo, ok := config.Redirects[key]; ok {
			config.increaseUsedTimes(key)
			http.Redirect(w, r, urlInfo.Url, http.StatusMovedPermanently)
		}
	}
	http.HandleFunc("/", handler)
}

func startServer(portNumber int64) {
	fmt.Printf("Server is running on port: %d", portNumber)
	port := fmt.Sprintf(":%d", portNumber)
	log.Fatal(http.ListenAndServe(port, nil))
}

func (config Config) appendRedirect(value string, url string) error {
	var err error = nil
	if url != "" {
		key := value
		if value == "" {
			key = randString(8)
		}
		if config.Redirects == nil {
			config.Redirects = map[string]UrlInfo{}
		}
		config.Redirects[key] = UrlInfo{url, 0}
		config.saveToYamlFile()
		fmt.Printf("Value: %s - URL: %s is appended to config file!", key, url)
	} else {
		err = fmt.Errorf("missing url")
	}
	return err
}

func (config Config) increaseUsedTimes(key string) {
	urlInfo := config.Redirects[key]
	urlInfo.Used += 1
	config.Redirects[key] = urlInfo
	config.saveToYamlFile()
}

func (config Config) printAllRedirects() {
	fmt.Println("Redirection list:")
	for value, urlInfo := range config.Redirects {
		fmt.Printf("Value: %s - URL: %s - Used: %d \n", value, urlInfo.Url, urlInfo.Used)
	}
}

func (config Config) deleteRedirect(deleteValue string) {
	if _, ok := config.Redirects[deleteValue]; ok {
		delete(config.Redirects, deleteValue)
		config.saveToYamlFile()
		fmt.Printf("%s is deleted from config file!", deleteValue)
	} else {
		fmt.Printf("%s is not found in config file!", deleteValue)
	}
}

func (config Config) saveToYamlFile() {
	changed, err := yaml.Marshal(&config)
	checkError(err)
	err = ioutil.WriteFile(FILE_NAME, changed, 0644)
	checkError(err)
}

func printUsageInfo() {
	usageInfo := "Usage:\n" +
		"        ./urlshorten <command> [option]\n" +
		"The commands are:\n" +
		"        configure  configure yaml config file\n" +
		"        run        run server\n" +
		"The options are:\n" +
		"        -a         append value to redirection list\n" +
		"        -u         url is appended to direction list\n" +
		"        -d         delete value from redirection list\n" +
		"        -p         port number\n" +
		"        -l         list all redirection\n" +
		"        -h         usage info\n"
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, usageInfo)
		checkError(err)
	}
	flag.Usage()
}

func randString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	l := len(LETTER_BYTES)
	for i := range b {
		b[i] = LETTER_BYTES[rand.Intn(l)]
	}
	return string(b)
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
