package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

//go:embed qleverlfile_template
var qlevelTemplate string

const cloneQuery string = `CONSTRUCT { ?s  ?p ?o }  WHERE  {
      ?s  ?p ?o.
     FILTER (?p != <http://vocab.sindice.net/analytics#cardinality>)
}`

func main() {
	inputGlob := flag.String("g", "", "input directory (required)")
	output := flag.String("o", "", "output directory (required)")
	flag.Parse()

	if *inputGlob == "" {
		fmt.Println("the input glob (-g) should be defined")
		os.Exit(1)
	}
	if *output == "" {
		fmt.Println("the output directory (-o) should be defined")
		os.Exit(1)
	}
	runParams, err := generateRuns(*inputGlob, *output)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	runErrors := []RunError{}
	for _, param := range runParams {
		if err := run(param); err != nil {
			fmt.Println(err)
			runErrors = append(runErrors, newRunError(param, err))
		}
	}

	if err := writeErrorReport(runErrors); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateRuns(filesGlob, outputFolder string) ([]CleaningRun, error) {
	resp := []CleaningRun{}
	matches, err := filepath.Glob(filesGlob)
	if err != nil {
		return resp, err
	}
	for _, match := range matches {
		inputFolder, file := filepath.Split(match)
		outputFilePath := filepath.Join(outputFolder, file)

		run := CleaningRun{
			InitialFilePath: match,
			OutputFilePath:  outputFilePath,
			QleverFileFolderPath:  inputFolder,
		}
		resp = append(resp, run)
	}
	return resp, nil
}

func run(config CleaningRun) error {
	qleverFileContent := fillQleverFileTemplate(config.InitialFilePath)
	if err := generateQleverFile(qleverFileContent, config.QleverFileFolderPath); err != nil {
		return err
	}

	if err := queryQlever(config.QleverFileFolderPath, config.OutputFilePath); err != nil {
		return err
	}

	return nil
}

func fillQleverFileTemplate(inputFilePath string) string {
	fileName := path.Base(inputFilePath)
	datasetName := strings.TrimSuffix(fileName, filepath.Ext(inputFilePath))

	qleverFile := strings.Replace(qlevelTemplate, "!{dataset-name}", datasetName, 1)
	qleverFile = strings.Replace(qleverFile, "!{input-file}", fileName, 1)
	return qleverFile
}

func generateQleverFile(qleverFileContent, generationPath string) error {
	qleverFilePath := filepath.Join(generationPath, "Qleverfile")
	file, err := os.Create(qleverFilePath)
	if err != nil {
		return err
	}

	_, err = file.WriteString(qleverFileContent)
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return closeErr
		}
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func queryQlever(dir string, outputFile string) error {
	if _, err := runQlever(dir, "index", "--overwrite-existing"); err != nil {
		return err
	}

	if _, err := runQlever(dir, "start", "--kill-existing-with-same-port"); err != nil {
		return err
	}

	cleanedKb, err := runQlever(dir, "query", cloneQuery)
	if err != nil {
		return err
	}
	// Transpose the KG to the new file
	cleanedKb = strings.ReplaceAll(cleanedKb, "\n", " .\n")

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	_, err = file.WriteString(cleanedKb)
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return closeErr
		}
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func runQlever(dir string, args ...string) (string, error) {
	cmd := exec.Command("qlever", args...)
	cmd.Dir = dir

	var outb bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg != "" {
			return "", fmt.Errorf("qlever %q failed: %w: %s", strings.Join(args, " "), err, msg)
		}
		return "", fmt.Errorf("qlever %q failed: %w", strings.Join(args, " "), err)
	}

	return outb.String(), nil
}

func writeErrorReport(runErrors []RunError) error {
	file, err := os.Create("error_report.txt")
	if err != nil {
		return err
	}

	var report strings.Builder
	report.WriteString("Dataset Cleaning Error Report\n")
	report.WriteString("============================\n\n")
	fmt.Fprintf(&report, "Runs with errors: %d\n\n", len(runErrors))

	if len(runErrors) == 0 {
		report.WriteString("No run errors were recorded.\n")
	} else {
		for i, runErr := range runErrors {
			fmt.Fprintf(&report, "Error #%d\n", i+1)
			fmt.Fprintf(&report, "Input file: %s\n", runErr.Run.InitialFilePath)
			fmt.Fprintf(&report, "Output file: %s\n", runErr.Run.OutputFilePath)
			fmt.Fprintf(&report, "Qlever directory: %s\n", runErr.Run.QleverFileFolderPath)
			fmt.Fprintf(&report, "Message: %s\n\n", runErr.Err)
		}
	}

	_, err = file.WriteString(report.String())
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return closeErr
		}
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func newRunError(run CleaningRun, err error) RunError {
	return RunError{
		Run: run,
		Err: err,
	}
}

type CleaningRun struct {
	// the path of the initial kg
	InitialFilePath string
	// the path where the clean data will live
	OutputFilePath string
	// the path for the qleverfile
	QleverFileFolderPath string
}

type RunError struct {
	Run CleaningRun
	Err error
}
