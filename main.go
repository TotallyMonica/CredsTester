package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type testResults struct {
	expectedResult bool
	actualResult   bool
	testedPassword string
}

var runShouldPass bool
var runShouldFailLength bool
var runShouldFailUpper bool
var runShouldFailLower bool
var runShouldFailNumber bool
var runShouldFailSpecialChars bool
var runShouldFailIllegalChars bool
var doTests bool
var doEvals bool
var exitOnFail bool
var testsToRun int
var showProgress bool

var validPasswordRegex *regexp.Regexp
var passwordUpperLettersRegex *regexp.Regexp
var passwordLowerLettersRegex *regexp.Regexp
var passwordNumbersRegex *regexp.Regexp
var passwordLegalSpecialRegex *regexp.Regexp
var legalChars = []string{"-", "_", ".", "!", "$", "|", "@", "%", "^", "&", "*"}
var illegalChars = []string{"+", "=", "(", ")", "#", "~", "}", "{", "[", "]", "\\", "<", ">", "/", "?", " ", "\"", "'", "`", ","}

const updateFrequency = 1000 * 100 // Change right number to change decimal precision, 1 means ever 0.01% increase

func runRegexp(passwd string) bool {
	upperCharCheck := passwordUpperLettersRegex.FindString(passwd) != ""
	lowerCharCheck := passwordLowerLettersRegex.FindString(passwd) != ""
	numberCheck := passwordNumbersRegex.FindString(passwd) != ""
	legalSpecialCharCheck := passwordLegalSpecialRegex.FindString(passwd) != ""
	validPasswordCheck := validPasswordRegex.FindString(passwd) != ""
	emptyStringCheck := passwd == ""

	return upperCharCheck && lowerCharCheck && numberCheck && legalSpecialCharCheck && validPasswordCheck && !emptyStringCheck
}

func printUpdate(prefix string, current int, max int, elapsed time.Duration) {
	fmt.Printf("--- %s --- Ran %d out of %d tests (%.2f%%) in ", prefix, current, max, float32(current)/float32(max)*100)
	if elapsed.Nanoseconds() < 1000 {
		fmt.Printf("%d nanoseconds\n", elapsed.Nanoseconds())
	} else if elapsed.Microseconds() < 1000 {
		fmt.Printf("%d microseconds\n", elapsed.Microseconds())
	} else if elapsed.Milliseconds() < 1000 {
		fmt.Printf("%d milliseconds\n", elapsed.Milliseconds())
	} else if elapsed.Seconds() < 60 {
		fmt.Printf("%.3g seconds\n", elapsed.Seconds())
	} else if elapsed.Minutes() < 60 {
		fmt.Printf("%d minutes, %d seconds\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
	} else {
		fmt.Printf("%d hours, %d minutes, %d seconds\n", int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60)
	}
}

func shouldFailIllegalChars(c chan testResults) {
	fmt.Printf("Running %d tests that should fail on illegal characters\n", testsToRun)
	start := time.Now()
	testCount := 0
	// Run 50 iterations of the test
	for i := 0; i < testsToRun; i++ {
		// Generate random number of numbers
		numberCount := rand.Intn(25)
		nums := make([]string, numberCount)
		for i := 0; i < numberCount; i++ {
			nums = append(nums, strconv.Itoa(rand.Intn(10)))
		}

		// Generate random number of lower characters
		lowerCharCount := rand.Intn(25)
		lowerChars := make([]string, lowerCharCount)
		for i := 0; i < lowerCharCount; i++ {
			lowerChars = append(lowerChars, string(rune(rand.Intn(26)+97)))
		}

		// Generate random number of lower characters
		upperCharsCount := rand.Intn(25)
		upperChars := make([]string, upperCharsCount)
		for i := 0; i < upperCharsCount; i++ {
			upperChars = append(upperChars, string(rune(rand.Intn(26)+65)))
		}

		// Generate random number of _legal_ special characters
		legalSpecialCharsCount := rand.Intn(25)
		legalSpecialChars := make([]string, legalSpecialCharsCount)
		for i := 0; i < legalSpecialCharsCount; i++ {
			legalSpecialChars = append(legalSpecialChars, legalChars[rand.Intn(len(legalChars))])
		}

		// Generate random number of _illegal_ special characters
		illegalSpecialCharsCount := rand.Intn(25) + 1
		illegalSpecialChars := make([]string, illegalSpecialCharsCount)
		for i := 0; i < illegalSpecialCharsCount; i++ {
			illegalSpecialChars = append(illegalSpecialChars, illegalChars[rand.Intn(len(illegalChars))])
		}

		generatedPassword := strings.Join(legalSpecialChars[:], "") + strings.Join(illegalSpecialChars[:], "") + strings.Join(lowerChars[:], "") + strings.Join(upperChars[:], "") + strings.Join(nums[:], "")

		shuff := []rune(generatedPassword)
		rand.Shuffle(len(shuff), func(i, j int) {
			shuff[i], shuff[j] = shuff[j], shuff[i]
		})
		generatedPassword = string(shuff)

		c <- testResults{
			expectedResult: false,
			actualResult:   runRegexp(generatedPassword),
			testedPassword: generatedPassword,
		}

		testCount += 1
		t := time.Now()
		elapsed := t.Sub(start)
		if showProgress && testCount%updateFrequency == 0 {
			printUpdate("SHOULD FAIL ILLEGAL CHARACTERS", testCount, testsToRun, elapsed)
		}
	}

	t := time.Now()
	elapsed := t.Sub(start)
	printUpdate("SHOULD FAIL ILLEGAL CHARACTERS", testCount, testsToRun, elapsed)
}

func shouldPass(c chan testResults) {
	fmt.Printf("Running %d tests that should pass successfully\n", testsToRun)
	start := time.Now()
	testCount := 0
	// Run 50 iterations of the test
	for i := 0; i < testsToRun; i++ {
		// Generate random number of numbers
		numberCount := rand.Intn(25) + 2
		nums := make([]string, numberCount)
		for i := 0; i < numberCount; i++ {
			nums = append(nums, strconv.Itoa(rand.Intn(10)))
		}

		// Generate random number of lower characters
		lowerCharCount := rand.Intn(23) + 2
		lowerChars := make([]string, lowerCharCount)
		for i := 0; i < lowerCharCount; i++ {
			lowerChars = append(lowerChars, string(rune(rand.Intn(26)+97)))
		}

		// Generate random number of lower characters
		upperCharsCount := rand.Intn(23) + 2
		upperChars := make([]string, upperCharsCount)
		for i := 0; i < upperCharsCount; i++ {
			upperChars = append(upperChars, string(rune(rand.Intn(26)+65)))
		}

		// Generate random number of _legal_ special characters
		legalSpecialCharsCount := rand.Intn(23) + 2
		legalSpecialChars := make([]string, legalSpecialCharsCount)
		for i := 0; i < legalSpecialCharsCount; i++ {
			legalSpecialChars = append(legalSpecialChars, legalChars[rand.Intn(len(legalChars))])
		}

		generatedPassword := strings.Join(legalSpecialChars[:], "") + strings.Join(lowerChars[:], "") + strings.Join(upperChars[:], "") + strings.Join(nums[:], "")

		shuff := []rune(generatedPassword)
		rand.Shuffle(len(shuff), func(i, j int) {
			shuff[i], shuff[j] = shuff[j], shuff[i]
		})
		generatedPassword = string(shuff)
		pass := runRegexp(generatedPassword)

		results := testResults{
			expectedResult: true,
			actualResult:   pass,
			testedPassword: generatedPassword,
		}

		c <- results
		testCount += 1
		if showProgress && testCount%updateFrequency == 0 {
			t := time.Now()
			elapsed := t.Sub(start)
			printUpdate("SHOULD PASS", testCount, testsToRun, elapsed)
		}
	}

	t := time.Now()
	elapsed := t.Sub(start)
	printUpdate("SHOULD PASS", testCount, testsToRun, elapsed)
}

func shouldFailSpecialChars(c chan testResults) {
	fmt.Printf("Running %d tests that should fail on missing special characters\n", testsToRun)
	start := time.Now()
	testCount := 0
	// Run 50 iterations of the test
	for i := 0; i < testsToRun; i++ {
		// Generate random number of numbers
		numberCount := rand.Intn(25) + 2
		nums := make([]string, numberCount)
		for i := 0; i < numberCount; i++ {
			nums = append(nums, strconv.Itoa(rand.Intn(10)))
		}

		// Generate random number of lower characters
		lowerCharCount := rand.Intn(23) + 2
		lowerChars := make([]string, lowerCharCount)
		for i := 0; i < lowerCharCount; i++ {
			lowerChars = append(lowerChars, string(rune(rand.Intn(26)+97)))
		}

		// Generate random number of lower characters
		upperCharsCount := rand.Intn(23) + 2
		upperChars := make([]string, upperCharsCount)
		for i := 0; i < upperCharsCount; i++ {
			upperChars = append(upperChars, string(rune(rand.Intn(26)+65)))
		}

		// Generate random number of _legal_ special characters
		legalSpecialCharsCount := rand.Intn(23) + 2
		legalSpecialChars := make([]string, legalSpecialCharsCount)
		for i := 0; i < legalSpecialCharsCount; i++ {
			legalSpecialChars = append(legalSpecialChars, "")
		}

		generatedPassword := strings.Join(legalSpecialChars[:], "") + strings.Join(lowerChars[:], "") + strings.Join(upperChars[:], "") + strings.Join(nums[:], "")

		shuff := []rune(generatedPassword)
		rand.Shuffle(len(shuff), func(i, j int) {
			shuff[i], shuff[j] = shuff[j], shuff[i]
		})
		generatedPassword = string(shuff)

		c <- testResults{
			expectedResult: false,
			actualResult:   runRegexp(generatedPassword),
			testedPassword: generatedPassword,
		}
		testCount += 1
		if showProgress && testCount%updateFrequency == 0 {
			t := time.Now()
			elapsed := t.Sub(start)
			printUpdate("SHOULD FAIL SPECIAL CHARS", testCount, testsToRun, elapsed)
		}
	}

	t := time.Now()
	elapsed := t.Sub(start)
	printUpdate("SHOULD FAIL SPECIAL CHARS", testCount, testsToRun, elapsed)
}

func shouldFailNumber(c chan testResults) {
	fmt.Printf("Running %d tests that should fail on missing numbers\n", testsToRun)
	start := time.Now()
	testCount := 0
	// Run 50 iterations of the test
	for i := 0; i < testsToRun; i++ {
		// Generate random number of numbers
		numberCount := rand.Intn(25) + 2
		nums := make([]string, numberCount)
		for i := 0; i < numberCount; i++ {
			nums = append(nums, "")
		}

		// Generate random number of lower characters
		lowerCharCount := rand.Intn(23) + 2
		lowerChars := make([]string, lowerCharCount)
		for i := 0; i < lowerCharCount; i++ {
			lowerChars = append(lowerChars, string(rune(rand.Intn(26)+97)))
		}

		// Generate random number of lower characters
		upperCharsCount := rand.Intn(23) + 2
		upperChars := make([]string, upperCharsCount)
		for i := 0; i < upperCharsCount; i++ {
			upperChars = append(upperChars, string(rune(rand.Intn(26)+65)))
		}

		// Generate random number of _legal_ special characters
		legalSpecialCharsCount := rand.Intn(23) + 2
		legalSpecialChars := make([]string, legalSpecialCharsCount)
		for i := 0; i < legalSpecialCharsCount; i++ {
			legalSpecialChars = append(legalSpecialChars, legalChars[rand.Intn(len(legalChars))])
		}

		generatedPassword := strings.Join(legalSpecialChars[:], "") + strings.Join(lowerChars[:], "") + strings.Join(upperChars[:], "") + strings.Join(nums[:], "")

		shuff := []rune(generatedPassword)
		rand.Shuffle(len(shuff), func(i, j int) {
			shuff[i], shuff[j] = shuff[j], shuff[i]
		})
		generatedPassword = string(shuff)

		c <- testResults{
			expectedResult: false,
			actualResult:   runRegexp(generatedPassword),
			testedPassword: generatedPassword,
		}
		testCount += 1
		if showProgress && testCount%updateFrequency == 0 {
			t := time.Now()
			elapsed := t.Sub(start)
			printUpdate("SHOULD FAIL NUMBER", testCount, testsToRun, elapsed)
		}
	}
	t := time.Now()
	elapsed := t.Sub(start)
	printUpdate("SHOULD FAIL NUMBER", testCount, testsToRun, elapsed)
}

func shouldFailLower(c chan testResults) {
	fmt.Printf("Running %d tests that should fail on missing lowercase letters\n", testsToRun)
	start := time.Now()
	testCount := 0
	// Run 50 iterations of the test
	for i := 0; i < testsToRun; i++ {
		// Generate random number of numbers
		numberCount := rand.Intn(25) + 2
		nums := make([]string, numberCount)
		for i := 0; i < numberCount; i++ {
			nums = append(nums, strconv.Itoa(rand.Intn(10)))
		}

		// Generate random number of lower characters
		lowerCharCount := rand.Intn(23) + 2
		lowerChars := make([]string, lowerCharCount)
		for i := 0; i < lowerCharCount; i++ {
			lowerChars = append(lowerChars, "")
		}

		// Generate random number of lower characters
		upperCharsCount := rand.Intn(23) + 2
		upperChars := make([]string, upperCharsCount)
		for i := 0; i < upperCharsCount; i++ {
			upperChars = append(upperChars, string(rune(rand.Intn(26)+65)))
		}

		// Generate random number of _legal_ special characters
		legalSpecialCharsCount := rand.Intn(23) + 2
		legalSpecialChars := make([]string, legalSpecialCharsCount)
		for i := 0; i < legalSpecialCharsCount; i++ {
			legalSpecialChars = append(legalSpecialChars, legalChars[rand.Intn(len(legalChars))])
		}

		generatedPassword := strings.Join(legalSpecialChars[:], "") + strings.Join(lowerChars[:], "") + strings.Join(upperChars[:], "") + strings.Join(nums[:], "")

		shuff := []rune(generatedPassword)
		rand.Shuffle(len(shuff), func(i, j int) {
			shuff[i], shuff[j] = shuff[j], shuff[i]
		})
		generatedPassword = string(shuff)

		c <- testResults{
			expectedResult: false,
			actualResult:   runRegexp(generatedPassword),
			testedPassword: generatedPassword,
		}
		testCount += 1
		if showProgress && testCount%updateFrequency == 0 {
			t := time.Now()
			elapsed := t.Sub(start)
			printUpdate("SHOULD FAIL LOWER", testCount, testsToRun, elapsed)
		}
	}

	t := time.Now()
	elapsed := t.Sub(start)
	printUpdate("SHOULD FAIL LOWER", testCount, testsToRun, elapsed)
}

func shouldFailUpper(c chan testResults) {
	fmt.Printf("Running %d tests that should fail on missing uppercase letters\n", testsToRun)
	start := time.Now()
	testCount := 0
	// Run 50 iterations of the test
	for i := 0; i < testsToRun; i++ {
		// Generate random number of numbers
		numberCount := rand.Intn(25) + 2
		nums := make([]string, numberCount)
		for i := 0; i < numberCount; i++ {
			nums = append(nums, strconv.Itoa(rand.Intn(10)))
		}

		// Generate random number of lower characters
		lowerCharCount := rand.Intn(23) + 2
		lowerChars := make([]string, lowerCharCount)
		for i := 0; i < lowerCharCount; i++ {
			lowerChars = append(lowerChars, string(rune(rand.Intn(26)+97)))
		}

		// Generate random number of lower characters
		upperCharsCount := rand.Intn(23) + 2
		upperChars := make([]string, upperCharsCount)
		for i := 0; i < upperCharsCount; i++ {
			upperChars = append(upperChars, "")
		}

		// Generate random number of _legal_ special characters
		legalSpecialCharsCount := rand.Intn(23) + 2
		legalSpecialChars := make([]string, legalSpecialCharsCount)
		for i := 0; i < legalSpecialCharsCount; i++ {
			legalSpecialChars = append(legalSpecialChars, legalChars[rand.Intn(len(legalChars))])
		}

		generatedPassword := strings.Join(legalSpecialChars[:], "") + strings.Join(lowerChars[:], "") + strings.Join(upperChars[:], "") + strings.Join(nums[:], "")

		shuff := []rune(generatedPassword)
		rand.Shuffle(len(shuff), func(i, j int) {
			shuff[i], shuff[j] = shuff[j], shuff[i]
		})
		generatedPassword = string(shuff)

		c <- testResults{
			expectedResult: false,
			actualResult:   runRegexp(generatedPassword),
			testedPassword: generatedPassword,
		}
		testCount += 1
		if showProgress && testCount%updateFrequency == 0 {
			t := time.Now()
			elapsed := t.Sub(start)
			printUpdate("SHOULD FAIL UPPER", testCount, testsToRun, elapsed)
		}
	}
	t := time.Now()
	elapsed := t.Sub(start)
	printUpdate("SHOULD FAIL UPPER", testCount, testsToRun, elapsed)
}

func shouldFailLength(c chan testResults) {
	fmt.Printf("Running %d tests that should fail on too short of a password\n", testsToRun)
	start := time.Now()
	testCount := 0
	// Run 50 iterations of the test
	for i := 0; i < testsToRun; i++ {
		// Generate random number of numbers
		numberCount := rand.Intn(25) + 2
		nums := make([]string, numberCount)
		for i := 0; i < numberCount; i++ {
			nums = append(nums, strconv.Itoa(rand.Intn(10)))
		}

		// Generate random number of lower characters
		lowerCharCount := rand.Intn(23) + 2
		lowerChars := make([]string, lowerCharCount)
		for i := 0; i < lowerCharCount; i++ {
			lowerChars = append(lowerChars, string(rune(rand.Intn(26)+97)))
		}

		// Generate random number of lower characters
		upperCharsCount := rand.Intn(23) + 2
		upperChars := make([]string, upperCharsCount)
		for i := 0; i < upperCharsCount; i++ {
			upperChars = append(upperChars, string(rune(rand.Intn(26)+65)))
		}

		// Generate random number of _legal_ special characters
		legalSpecialCharsCount := rand.Intn(23) + 2
		legalSpecialChars := make([]string, legalSpecialCharsCount)
		for i := 0; i < legalSpecialCharsCount; i++ {
			legalSpecialChars = append(legalSpecialChars, legalChars[rand.Intn(len(legalChars))])
		}

		generatedPassword := strings.Join(legalSpecialChars[:], "") + strings.Join(lowerChars[:], "") + strings.Join(upperChars[:], "") + strings.Join(nums[:], "")

		shuff := []rune(generatedPassword)
		rand.Shuffle(len(shuff), func(i, j int) {
			shuff[i], shuff[j] = shuff[j], shuff[i]
		})
		generatedPassword = string(shuff)

		generatedPassword = generatedPassword[:rand.Intn(7)+1]

		c <- testResults{
			expectedResult: false,
			actualResult:   runRegexp(generatedPassword),
			testedPassword: generatedPassword,
		}
		testCount += 1
		if showProgress && testCount%updateFrequency == 0 {
			t := time.Now()
			elapsed := t.Sub(start)
			printUpdate("SHOULD FAIL LENGTH", testCount, testsToRun, elapsed)
		}
	}
	t := time.Now()
	elapsed := t.Sub(start)
	printUpdate("SHOULD FAIL LENGTH", testCount, testsToRun, elapsed)
}

func main() {
	// Parse flags
	flag.BoolVar(&doTests, "run-tests", false, "Run the tests. Omission takes precedence over -all and specifying individual tests")
	flag.BoolVar(&doEvals, "run-evals", false, "Evaluate results.csv")
	flag.BoolVar(&showProgress, "show-progress", false, "Print progress to stdout")
	flag.BoolVar(&runShouldFailSpecialChars, "run-special-char-test", false, "Test to make sure special characters are required")
	flag.BoolVar(&runShouldFailIllegalChars, "run-illegal-char-test", false, "Test to make sure illegal characters aren't allowed")
	flag.BoolVar(&runShouldFailLength, "run-length-test", false, "Test to make sure passwords need to be sufficiently long enough")
	flag.BoolVar(&runShouldFailLower, "run-lowercase-test", false, "Test to make sure lowercase letters are required")
	flag.BoolVar(&runShouldFailUpper, "run-uppercase-test", false, "Test to make sure uppercase letters are required")
	flag.BoolVar(&runShouldFailNumber, "run-numbers-test", false, "Test to make sure numbers are required")
	flag.BoolVar(&exitOnFail, "exit-on-fail", false, "Exit immediately on fail")
	flag.IntVar(&testsToRun, "run", 100, "Specify how many times a test should be run")
	runAllTests := flag.Bool("run-all-tests", false, "Runs all tests. Takes precedence of running specific tests")

	flag.Parse()

	// Compile regexes
	var err error
	validPasswordRegex, err = regexp.Compile(`^([A-Z]|[a-z]|[0-9]|-|_|\.|!|\$|\||@|%|\^|&|\*){8,}$`)
	if err != nil {
		log.Fatal("Error while compiling regex\n", err)
	}
	passwordUpperLettersRegex, err = regexp.Compile(`[A-Z]`)
	if err != nil {
		log.Fatal("Error while compiling regex\n", err)
	}
	passwordLowerLettersRegex, err = regexp.Compile(`[a-z]`)
	if err != nil {
		log.Fatal("Error while compiling regex\n", err)
	}
	passwordNumbersRegex, err = regexp.Compile(`[0-9]`)
	if err != nil {
		log.Fatal("Error while compiling regex\n", err)
	}
	passwordLegalSpecialRegex, err = regexp.Compile(`([-_.!$|@%^&*])`)
	if err != nil {
		log.Fatal("Error while compiling regex\n", err)
	}

	fmt.Printf("Do tests:                %t\n", doTests)
	fmt.Printf("Do evals:                %t\n", doEvals)
	fmt.Printf("Run all tests:           %t\n", doTests && *runAllTests)
	fmt.Printf("Test special characters: %t\n", doTests && (runShouldFailSpecialChars || *runAllTests))
	fmt.Printf("Test illegal characters: %t\n", doTests && (runShouldFailIllegalChars || *runAllTests))
	fmt.Printf("Test uppercase letters:  %t\n", doTests && (runShouldFailUpper || *runAllTests))
	fmt.Printf("Test lowercase letters:  %t\n", doTests && (runShouldFailLower || *runAllTests))
	fmt.Printf("Test numbers:            %t\n", doTests && (runShouldFailNumber || *runAllTests))
	fmt.Printf("Test length:             %t\n", doTests && (runShouldFailLength || *runAllTests))
	fmt.Printf("Show progress:           %t\n", showProgress)
	fmt.Printf("Exit on fail:            %t\n", exitOnFail)
	fmt.Printf("Test repeat count:       %d\n", testsToRun)

	start := time.Now()
	if doTests {
		c := make(chan testResults)
		tests := 0

		// Create CSV for writing results
		file, err := os.Create("results.csv")
		if err != nil {
			log.Fatal(err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatalf("Error while closing file %s\n%s\n", file.Name(), err)
			}
		}(file)

		// Create writer to handle the writing
		writer := csv.NewWriter(file)

		// Add the headers
		err = writer.Write([]string{"password", "expected", "actual"})
		if err != nil {
			log.Fatalf("Error while writing headers to file %s\n%s\n", file.Name(), err)
		}

		// Start the various tests
		if runShouldPass || *runAllTests {
			go shouldPass(c)
			tests += 1
		}
		if runShouldFailSpecialChars || *runAllTests {
			go shouldFailSpecialChars(c)
			tests += 1
		}
		if runShouldFailIllegalChars || *runAllTests {
			go shouldFailIllegalChars(c)
			tests += 1
		}
		if runShouldFailNumber || *runAllTests {
			go shouldFailNumber(c)
			tests += 1
		}
		if runShouldFailUpper || *runAllTests {
			go shouldFailUpper(c)
			tests += 1
		}
		if runShouldFailLower || *runAllTests {
			go shouldFailLower(c)
			tests += 1
		}
		if runShouldFailLength || *runAllTests {
			go shouldFailLength(c)
			tests += 1
		}

		// Write the results of each test to the CSV
		for i := 0; i < testsToRun*tests; i++ {
			result := <-c
			row := []string{result.testedPassword, fmt.Sprintf("%t", result.expectedResult), fmt.Sprintf("%t", result.actualResult)}
			err := writer.Write(row)
			if err != nil {
				log.Panicf("Issue while writing to file %s\n%s\n", file.Name(), err)
			}
			t := time.Now()
			elapsed := t.Sub(start)
			if i%updateFrequency == 0 {
				writer.Flush()
				if showProgress {
					printUpdate("OVERALL", i, testsToRun*tests, elapsed)
				}
			}
			if exitOnFail && result.expectedResult != result.actualResult {
				printUpdate("OVERALL", i, testsToRun*tests, elapsed)
				fmt.Printf("Password %s failed (Expected %t, got %t)\n", result.testedPassword, result.expectedResult, result.actualResult)
				os.Exit(1)
			}
		}
		writer.Flush()
		t := time.Now()
		elapsed := t.Sub(start)

		fmt.Printf("Total time to run tests: ")
		if elapsed.Nanoseconds() < 1000 {
			fmt.Printf("%d nanoseconds\n", elapsed.Nanoseconds())
		} else if elapsed.Microseconds() < 1000 {
			fmt.Printf("%d microseconds\n", elapsed.Microseconds())
		} else if elapsed.Milliseconds() < 1000 {
			fmt.Printf("%d milliseconds\n", elapsed.Milliseconds())
		} else if elapsed.Seconds() < 60 {
			fmt.Printf("%.3g seconds\n", elapsed.Seconds())
		} else if elapsed.Minutes() < 60 {
			fmt.Printf("%d minutes, %d seconds\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
		} else {
			fmt.Printf("%d hours, %d minutes, %d seconds\n", int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60)
		}
	}

	if doEvals {
		evalStart := time.Now()
		// Create CSV Reader
		file, err := os.Open("results.csv")
		if err != nil {
			log.Fatalf("Error while opening file %s\n", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatalf("Error while closing file %s\n%s\n", file.Name(), err)
			}
		}(file)
		reader := csv.NewReader(file)

		// We want to ignore the header
		line, err := reader.Read()
		if err != nil {
			log.Fatal("Error while parsing CSV: ", err)
		}
		line, err = reader.Read()
		if err != nil && err != io.EOF {
			log.Fatal("Error while parsing CSV: ", err)
		}

		// Get all the tests
		failedTests := 0
		totalOfTests := 0
		for err != io.EOF {
			password := strings.Join(line[:len(line)-2], ",")
			expected := strings.ToLower(line[len(line)-2])
			actual := strings.ToLower(line[len(line)-1])

			// Report if a password had different results from what was expected
			if expected != actual {
				fmt.Printf("%s did not meet expectations (Expected result of %s, got %s)\n", password, expected, actual)
				failedTests += 1
			}
			totalOfTests += 1
			line, err = reader.Read()
			if err != nil && err != io.EOF {
				log.Fatal("Error while parsing CSV: ", err)
			}
		}
		t := time.Now()
		elapsed := t.Sub(evalStart)

		fmt.Printf("Total number of tests ran: %d\n", totalOfTests)
		fmt.Printf("Number of passing tests: %d (%.3g%%)\n", totalOfTests-failedTests, float32((totalOfTests-failedTests)/totalOfTests)*100)

		fmt.Printf("Total time to evaluate test results: ")
		if elapsed.Nanoseconds() < 1000 {
			fmt.Printf("%d nanoseconds\n", elapsed.Nanoseconds())
		} else if elapsed.Microseconds() < 1000 {
			fmt.Printf("%d microseconds\n", elapsed.Microseconds())
		} else if elapsed.Milliseconds() < 1000 {
			fmt.Printf("%d milliseconds\n", elapsed.Milliseconds())
		} else if elapsed.Seconds() < 60 {
			fmt.Printf("%.3g seconds\n", elapsed.Seconds())
		} else if elapsed.Minutes() < 60 {
			fmt.Printf("%d minutes, %d seconds\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
		} else {
			fmt.Printf("%d hours, %d minutes, %d seconds\n", int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60)
		}

		t = time.Now()
		elapsed = t.Sub(start)
		fmt.Printf("Overall time to evaluate test results: ")
		if elapsed.Nanoseconds() < 1000 {
			fmt.Printf("%d nanoseconds\n", elapsed.Nanoseconds())
		} else if elapsed.Microseconds() < 1000 {
			fmt.Printf("%d microseconds\n", elapsed.Microseconds())
		} else if elapsed.Milliseconds() < 1000 {
			fmt.Printf("%d milliseconds\n", elapsed.Milliseconds())
		} else if elapsed.Seconds() < 60 {
			fmt.Printf("%.3g seconds	\n", elapsed.Seconds())
		} else if elapsed.Minutes() < 60 {
			fmt.Printf("%d minutes, %d seconds\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
		} else {
			fmt.Printf("%d hours, %d minutes, %d seconds\n", int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60)
		}

		os.Exit(failedTests)
	}
}
