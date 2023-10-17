package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

type pattern struct {
	Flags    string   `json:"flags,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	Patterns []string `json:"patterns,omitempty"`
	Engine   string   `json:"engine,omitempty"`
}

// main is the entry point of the gf (grep-friendly) tool.
func main() {
    // Define command-line flags to control the tool's behavior.
    var saveMode bool
    flag.BoolVar(&saveMode, "save", false, "Save a pattern (e.g: gf -save pat-name -Hnri 'search-pattern')")

    var listMode bool
    flag.BoolVar(&listMode, "list", false, "List available patterns")

    var dumpMode bool
    flag.BoolVar(&dumpMode, "dump", false, "Prints the grep command rather than executing it")

    // Parse the command-line arguments.
    flag.Parse()

    // If listMode is enabled, list available patterns.
    if listMode {
        patternNames, err := getPatterns()
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n", err)
            return
        }
        
        output := strings.Join(patternNames, "\n")
        fmt.Println(output)
        return
    }

    // If saveMode is enabled, save a new pattern.
    if saveMode {
        name := flag.Arg(0)
        flags := flag.Arg(1)
        searchPattern := flag.Arg(2)

        err := savePattern(name, flags, searchPattern)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n", err)
        }
        return
    }

    // Extract arguments for executing a pattern.
    patternName := flag.Arg(0)
    files := flag.Arg(1)
    if files == "" {
        files = "."
    }

    // Get the pattern directory.
    patternDir, err := getPatternDir()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Unable to open user's pattern directory")
        return
    }

    // Construct the filename for the pattern.
    filename := filepath.Join(patternDir, patternName+".json")
    file, err := os.Open(filename)
    if err != nil {
        fmt.Fprintln(os.Stderr, "No such pattern")
        return
    }
    defer file.Close()

    // Decode the pattern from the JSON file.
    pattern := pattern{}
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&pattern)

    if err != nil {
        fmt.Fprintf(os.Stderr, "Pattern file '%s' is malformed: %s\n", filename, err)
        return
    }

    if pattern.Pattern == "" {
        // Check for multiple patterns and construct a regex.
        if len(pattern.Patterns) == 0 {
            fmt.Fprintf(os.Stderr, "Pattern file '%s' contains no pattern(s)\n", filename)
            return
        }

        pattern.Pattern = "(" + strings.Join(pattern.Patterns, "|") + ")"
    }

    // If dumpMode is enabled, print the grep command.
    if dumpMode {
        fmt.Printf("grep %v %q %v\n", pattern.Flags, pattern.Pattern, files)
    } else {
        var cmd *exec.Cmd
        operator := "grep"
        if pattern.Engine != "" {
            operator = pattern.Engine
        }

        if isStdinPiped() {
            // Prepare the command to execute when input is piped.
            cmdArgs := append(strings.Fields(pattern.Flags), pattern.Pattern)
            cmd = exec.Command(operator, cmdArgs...)
        } else {
            // Prepare the command to execute when there's no piped input.
            cmdArgs := append(strings.Fields(pattern.Flags), pattern.Pattern, files)
            cmd = exec.Command(operator, cmdArgs...)
        }
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        cmd.Run()
    }
}

// getPatternDir returns the directory where pattern JSON files are stored.
func getPatternDir() (string, error) {
    // Get the current user's information.
    currentUser, err := user.Current()
    if err != nil {
        return "", err
    }

    // Define the path for the .config/gf directory.
    configDirPath := filepath.Join(currentUser.HomeDir, ".config/gf")

    // Check if the .config/gf directory exists.
    if _, err := os.Stat(configDirPath); !os.IsNotExist(err) {
        // .config/gf exists, return its path.
        return configDirPath, nil
    }

    // .config/gf doesn't exist, return the .gf directory path.
    return filepath.Join(currentUser.HomeDir, ".gf"), nil
}

// savePattern saves a pattern to a JSON file with the given name.
func savePattern(patternName, flags, patternValue string) error {
    // Check for empty name and pattern.
    if patternName == "" {
        return errors.New("pattern name cannot be empty")
    }
    if patternValue == "" {
        return errors.New("pattern cannot be empty")
    }
   // Create a pattern struct.
    p := &pattern{
        Flags:   flags,
        Pattern: patternValue,
    }

    // Get the pattern directory path.
    patternDir, err := getPatternDir()
    if err != nil {
        return fmt.Errorf("failed to determine pattern directory: %s", err)
    }

    // Construct the file path for the pattern.
    patternFilePath := filepath.Join(patternDir, patternName+".json")

    // Open the file for writing, creating it if it doesn't exist.
    file, err := os.OpenFile(patternFilePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
    if err != nil {
        return fmt.Errorf("failed to create pattern file: %s", err)
    }
    defer file.Close()

    // Create a JSON encoder with an indentation for formatting.
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "    ")

    // Encode the pattern and write it to the file.
    err = encoder.Encode(p)
    if err != nil {
        return fmt.Errorf("failed to write pattern file: %s", err)
    }

    return nil
}

// getPatterns retrieves a list of pattern names from JSON files located in a pattern directory.
// It returns a slice of strings containing the pattern names and an error, if any.
func getPatterns() ([]string, error) {
    // Initialize an empty slice to store the pattern names.
    var patterns []string

    // Get the pattern directory path.
    patternDirectory, err := getPatternDir()
    if err != nil {
        return patterns, fmt.Errorf("failed to determine pattern directory: %s", err)
    }

    // List JSON files in the pattern directory.
    jsonFiles, err := filepath.Glob(patternDirectory + "/*.json")
    if err != nil {
        return patterns, err
    }

    // Iterate over the list of JSON files and extract pattern names.
    for _, jsonFile := range jsonFiles {
        // Remove the pattern directory path and ".json" extension to get the pattern name.
        patternName := jsonFile[len(patternDirectory)+1 : len(jsonFile)-5]
        patterns = append(patterns, patternName)
    }

    // Return the list of pattern names and a nil error to indicate success.
    return patterns, nil
}

// isStdinPiped checks if the standard input (stdin) is a pipe.
func isStdinPiped() bool {
    // Get the file information of the standard input.
    stdinFileInfo, _ := os.Stdin.Stat()

    // Check if the standard input is a pipe (not a character device).
    return (stdinFileInfo.Mode() & os.ModeCharDevice) == 0
}
