/* *********************************************************************
 * This Original Work is copyright of 51 Degrees Mobile Experts Limited.
 * Copyright 2019 51 Degrees Mobile Experts Limited, 5 Charlotte Close,
 * Caversham, Reading, Berkshire, United Kingdom RG4 7BY.
 *
 * This Original Work is licensed under the European Union Public Licence (EUPL)
 * v.1.2 and is subject to its terms as set out below.
 *
 * If a copy of the EUPL was not distributed with this file, You can obtain
 * one at https://opensource.org/licenses/EUPL-1.2.
 *
 * The 'Compatible Licences' set out in the Appendix to the EUPL (as may be
 * amended by the European Commission) shall be deemed incompatible for
 * the purposes of the Work and the provisions of the compatibility
 * clause in Article 5 of the EUPL shall not apply.
 *
 * If using the Work as, or as part of, a network application, by
 * including the attribution notice(s) required under Article 5 of the EUPL
 * in the end user terms of the application under an appropriate heading,
 * such notice(s) shall fulfill the requirements of that article.
 * ********************************************************************* */

package main

/*
This example illustrates the performance of 51Degrees device detection solution.

Expected output is as described at the "// Output:..." section locate at the
bottom of this example.

To run this example, perform the following command:
```
go test -run Example_performance
```

This example will output a report to ./performance_report.log. The report
content is in the below format:
```
Average 0.00456 ms per Evidence Record
Total Evidence Records: 80000
IsMobile Evidence Records: 58076
Processed Evidence Records: 80000
Number of CPUs: 2
```
*/

import ( //	"runtime"
	"bufio"
	"fmt"
	"io"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	dd_example "github.com/51Degrees/device-detection-examples-go/v4/dd"
	"github.com/51Degrees/device-detection-examples-go/v4/onpremise/common"
	"gopkg.in/yaml.v3"

	"github.com/51Degrees/device-detection-go/v4/dd"
	"github.com/51Degrees/device-detection-go/v4/onpremise"
)

// File to output the performance report
const reportFile = "performance_report.log"
const fIterationCount = 4

// Report struct for each performance run
type report struct {
	evidenceCount     uint64
	evidenceIsMobile  uint64
	evidenceProcessed uint64
	processingTime    int64
}

// Perform device detection on a Evidence Record
func matchEvidenceRecord(
	pl *onpremise.Engine,
	wg *sync.WaitGroup,
	evidence []onpremise.Evidence,
	rep *report) {
	// Increase the number of Evidence Record being processed
	atomic.AddUint64(&rep.evidenceProcessed, 1)

	results, err := pl.Process(evidence)
	if err != nil {
		log.Fatalln(err)
	}
	defer results.Free()

	// Get the value in string
	res, err := results.ValuesString(
		"IsMobile",
		",")
	if err != nil {
		log.Fatalln(err)
	}

	// Update report
	if strings.Compare("True", res) == 0 {
		atomic.AddUint64(&rep.evidenceIsMobile, 1)
	}

	// Complete and mark as done
	defer wg.Done()
}

// Run the performance test. Determine the number of records in a Evidence
// file. Iterate through the Evidence file and perform detection on each
// Evidence. Record the processing time and update a report statistic.
func performDetections(
	pl *onpremise.Engine,
	params common.ExampleParams,
	rep *report) {
	// Create a wait group
	var wg sync.WaitGroup
	evidenceFilePath := dd_example.GetFilePathByPath(params.EvidenceYaml)

	// Read and extract Evidence for the performance check
	evidenceSlice := readYAMLFile(evidenceFilePath)

	start := time.Now()
	for i := 0; i < int(fIterationCount); i++ {
		// Actual processing
		for _, evidence := range evidenceSlice {
			// Increase wait group
			wg.Add(1)
			rep.evidenceCount += 1

			go matchEvidenceRecord(
				pl,
				&wg,
				evidence,
				rep)
		}
	}
	// Wait until all goroutines finish
	wg.Wait()
	rep.processingTime = time.Since(start).Milliseconds()
}

// Open, read, decode and extract Evidence to be used in the performance test.
// Data can be reused for multiple iterations.
func readYAMLFile(evidenceFilePath string) [][]onpremise.Evidence {
	// Open YAML file
	file, err := os.OpenFile(evidenceFilePath, os.O_RDONLY, 0444)
	if err != nil {
		log.Fatalf("ERROR: Failed to open file \"%s\".\n", evidenceFilePath)
	}
	defer func() {
		// Make sure the file is closed properly
		if err := file.Close(); err != nil {
			log.Fatalf("ERROR: Failed to close file \"%s\".\n", evidenceFilePath)
		}
	}()

	// Decode YAML file
	var res [][]onpremise.Evidence
	dec := yaml.NewDecoder(file)
	for {
		// Decode Evidence file by line
		var doc map[string]string
		if err := dec.Decode(&doc); err == io.EOF {
			break
		} else if err != nil {
			// Make sure there is no decoder error
			log.Fatalf("ERROR: Failed during decoding file \"%s\". %v\n", evidenceFilePath, err)
		}
		// Prepare evidence for usage
		evidence := common.ConvertToEvidence(doc)

		res = append(res, evidence)
	}
	return res
}

// Check a error returned from writing to a buffer
func checkWriteError(err error) {
	if err != nil {
		log.Fatalln("ERROR: Failed to write to buffer.")
	}
}

// Print report to a report file and return output message.
func printReport(actR *report, logOutputPath string) {
	// Get relative output path for testing
	var path string
	if filepath.IsAbs(logOutputPath) {
		path = logOutputPath
	} else {
		rootDir, e := os.Getwd()
		if e != nil {
			log.Fatalln("Failed to get current directory.")
		}
		path = filepath.Join(rootDir, logOutputPath)
	}
	path = filepath.Join(path, reportFile)

	// Create a report file
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("ERROR: Failed to create report file \"%s\".", path)
	}
	defer f.Close()

	// Create a writer
	w := bufio.NewWriter(f)

	msPerRecord := float64(actR.processingTime) / float64(actR.evidenceCount)
	_, err = fmt.Fprintf(w, "Average %.5f ms per Evidence Record\n", msPerRecord)
	checkWriteError(err)
	detectionsPerSecond := float64(actR.evidenceCount) / float64(actR.processingTime)
	_, err = fmt.Fprintf(w, "Average %.2f detections per second\n", detectionsPerSecond)
	checkWriteError(err)
	_, err = fmt.Fprintf(w, "Total Evidence Records: %d\n", actR.evidenceCount)
	checkWriteError(err)
	_, err = fmt.Fprintf(w, "IsMobile Evidence Records: %d\n", actR.evidenceIsMobile)
	checkWriteError(err)
	_, err = fmt.Fprintf(w, "Processed Evidence Records: %d\n", actR.evidenceProcessed)
	checkWriteError(err)
	_, err = fmt.Fprintf(w, "Number of CPUs: %d\n", runtime.NumCPU())
	checkWriteError(err)
	w.Flush()
	fmt.Printf("Output report to file \"%s\".\n", path)
}

func main() {
	common.RunExample(
		func(params common.ExampleParams) error {
			//... Example code
			//Create config
			config := dd.NewConfigHash(dd.InMemory)
			config.SetConcurrency(uint16(runtime.NumCPU()))
			config.SetUsePredictiveGraph(false)
			config.SetUsePerformanceGraph(true)
			config.SetUseUpperPrefixHeaders(false)
			config.SetUpdateMatchedUserAgent(false)

			//Create on-premise engine
			pl, err := onpremise.New(
				// A single property detection
				onpremise.WithProperties([]string{"IsMobile"}),
				// Optimized config provided
				onpremise.WithConfigHash(config),
				// Path to your data file
				onpremise.WithDataFile(params.DataFile),
				// Disable automatic updates.
				onpremise.WithAutoUpdate(false),
			)

			if err != nil {
				log.Fatalf("Failed to create engine: %v", err)
			}

			// Action
			actReport := report{0, 0, 0, 0}
			performDetections(pl, params, &actReport)
			// Validation to make sure same number of Evidences have been read and processed
			if actReport.evidenceCount != actReport.evidenceProcessed {
				log.Fatalln("ERROR: Not all Evidence Records have been processed.")
			}

			// Print the final performance report
			printReport(&actReport, "")

			pl.Stop()

			return nil
		},
	)
}
