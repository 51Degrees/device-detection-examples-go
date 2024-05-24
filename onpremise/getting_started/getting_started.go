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
This example illustrates how to perform simple device detections on given
User-Agent strings.
*/

import (
	"fmt"
	"log"

	"github.com/51Degrees/device-detection-examples-go/v4/onpremise/common"

	"github.com/51Degrees/device-detection-go/v4/dd"
	"github.com/51Degrees/device-detection-go/v4/onpremise"
)

// function match performs a match on an input User-Agent string and determine
// if the device is a mobile device. Returns output string.
func match(
	engine *onpremise.Engine,
	evidence []onpremise.Evidence) string {
	// Perform detection
	results, _ := engine.Process(evidence)

	// Make sure results object is freed after function execution.
	defer results.Free()

	propertyName := "IsMobile"

	// If results has values for required property
	hasValues, err := results.HasValues(propertyName)
	if err != nil {
		log.Fatalln(err)
	}

	returnStr := ""
	if !hasValues {
		returnStr = fmt.Sprintf("Property %s does not have a matched value.\n", propertyName)
	} else {
		// Get the values in string
		value, err := results.ValuesString(
			propertyName,
			",")
		if err != nil {
			log.Fatalln(err)
		}
		returnStr = fmt.Sprintf("\tIsMobile: %s\n", value)
	}

	return returnStr
}

func getEvidenceUserAgent(evidence []onpremise.Evidence) string {
	for i := range evidence {
		if evidence[i].Key == "User-Agent" {
			return evidence[i].Value
		}
	}
	return ""
}

func runGettingStarted(engine *onpremise.Engine) {
	// Perform detection on mobile Evidence
	actual := fmt.Sprintf("Mobile User-Agent: %s\n", getEvidenceUserAgent(common.ExampleEvidenceMobile))
	actual += match(engine, common.ExampleEvidenceMobile)

	// Perform detection on desktop Evidence
	actual += fmt.Sprintf("\nDesktop User-Agent: %v\n", getEvidenceUserAgent(common.ExampleEvidenceDesktop))
	actual += match(engine, common.ExampleEvidenceDesktop)

	// Perform detection on MediaHub Evidence
	actual += fmt.Sprintf("\nMediaHub User-Agent: %v\n", getEvidenceUserAgent(common.ExampleEvidenceMediaHub))
	actual += match(engine, common.ExampleEvidenceMediaHub)

	// Expected output
	expected := "Mobile User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1.2 Mobile/15E148 Safari/604.1\n"
	expected += "\tIsMobile: True\n"
	expected += "\n"
	expected += "Desktop User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36\n"
	expected += "\tIsMobile: False\n"
	expected += "\n"
	expected += "MediaHub User-Agent: Mozilla/5.0 (Linux; U; Android 4.4.2; en-us; A464BG Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.36\n"
	expected += "\tIsMobile: True\n"
	if actual != expected {
		log.Println("Expected:")
		log.Println(expected)
		log.Println("")
		log.Println("Actual:")
		log.Println(actual)
		log.Fatalln("Output does not match expected.")
	}

	log.Println(actual)
}

func main() {
	common.RunExample(
		func(params common.ExampleParams) error {
			//... Example code
			//Create config
			config := dd.NewConfigHash(dd.Default)

			//Create on-premise engine
			engine, err := onpremise.New(
				// Optimized config provided
				onpremise.WithConfigHash(config),
				// Path to your data file
				onpremise.WithDataFile(params.DataFile),
				// Enable automatic updates.
				onpremise.WithAutoUpdate(false),
			)

			if err != nil {
				log.Fatalf("Failed to create engine: %v", err)
			}

			// Run example
			runGettingStarted(engine)

			engine.Stop()

			return nil
		},
	)
}
