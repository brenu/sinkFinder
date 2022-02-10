package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

type sink struct {
	name        string
	regexString string
	regex       *regexp.Regexp
}

type occurrence struct {
	keyword string
	line    int
	column  int
}

func appendOccurrences(occurrences []occurrence, newOccurrences [][]int, line int, keyword string) []occurrence {
	for i := 0; i < len(newOccurrences); i++ {
		newOcurrence := occurrence{
			keyword: keyword,
			line:    line,
			column:  newOccurrences[i][0] + 1,
		}

		occurrences = append(occurrences, newOcurrence)
	}

	return occurrences
}

func initialize(outputPath string) ([]sink, *os.File, *bufio.Scanner, error) {
	var sinks = []sink{
		{name: "document.write", regexString: "document\\.write\\(.\\)"},
		{name: "document.writeln", regexString: "document\\.writeln\\(.\\)"},
		{name: "innerHTML", regexString: "innerHTML *="},
		{name: "outerHTML", regexString: "outerHTML *="},
		{name: "insertAdjacentHTML", regexString: "insertAdjacentHTML\\(.\\)"},
		{name: "eval", regexString: "eval\\(.\\)"},
		{name: "new Function", regexString: "new Function\\(.\\)"},
		{name: "onevent", regexString: "onevent\\(.\\)"},
	}

	for index, value := range sinks {
		sinks[index].regex = regexp.MustCompile(value.regexString)
	}

	var file *os.File
	var err error

	if outputPath != "" {
		file, err = os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

		if err != nil {
			return nil, nil, nil, err
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	return sinks, file, scanner, nil
}

func fileLookup(URL string, keywords *[]sink) ([]occurrence, error) {

	var occurrences []occurrence

	response, err := http.Get(URL)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	lines := bytes.Split(body, []byte{10}) // Splitting on each \n character
	for i := 0; i < len(lines); i++ {
		for j := 0; j < len(*keywords); j++ {
			lineToSuffixArray := suffixarray.New([]byte(lines[i]))
			results := lineToSuffixArray.FindAllIndex((*keywords)[j].regex, -1)

			occurrences = appendOccurrences(occurrences, results, i+1, (*keywords)[j].name)
		}
	}

	return occurrences, nil
}

func printResult(waitGroup *sync.WaitGroup, file *os.File, URL string, occurrences []occurrence) {

	if len(occurrences) == 0 {
		defer waitGroup.Done()
		return
	}

	fmt.Println("[*]", URL)
	for i := 0; i < len(occurrences); i++ {
		fmt.Printf("    • %s at line %d:%d\n", occurrences[i].keyword, occurrences[i].line, occurrences[i].column)
	}

	if file != nil {
		file.WriteString(fmt.Sprint("[*] ", URL, "\n"))
		for i := 0; i < len(occurrences); i++ {
			file.WriteString(fmt.Sprintf("    • %s at line %d:%d\n", occurrences[i].keyword, occurrences[i].line, occurrences[i].column))
		}
	}

	defer waitGroup.Done()
}

func main() {

	var outputPath string
	var maxChans int
	var rateLimit int

	flag.StringVar(&outputPath, "o", "", "Path to the output file. Optional, and it uses an append approach, so that whenever you choose a file with content inside, it will not erase it.")
	flag.IntVar(&maxChans, "t", 1, "Number of concurrent 'threads'.")
	flag.IntVar(&rateLimit, "r", 1, "Number of requests per second (rate limiting).")
	flag.Parse()

	sinkKeywords, file, scanner, err := initialize(outputPath)

	if err != nil {
		log.Fatal(err)
	}

	var waitGroup sync.WaitGroup

	sem := make(chan int, maxChans)
	completed := false

	for !completed {
		select {
		case <-time.After(time.Second / time.Duration(rateLimit)):
			for scanner.Scan() {
				sem <- 1
				waitGroup.Add(1)

				go func(URL string, file *os.File) {
					occurrences, err := fileLookup(URL, &sinkKeywords)

					if err != nil {
						defer waitGroup.Done()
						return
					}

					printResult(&waitGroup, file, URL, occurrences)
					<-sem
				}(scanner.Text(), file)
			}
			completed = true
		}
	}

	waitGroup.Wait()

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

}
