package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func printErrMessage(output []byte, err error){
	fmt.Println("Error:", err)
	fmt.Println("Output:\n", string(output))
}

func CreateFolder(folderName string) error{
	err := os.Mkdir(folderName, 0755) // 0755 = rwxr-xr-x
	if err != nil {
		fmt.Printf("Error creating folder %s: %v\n", folderName, err)
		return err
	}
	fmt.Printf("Folder: %s created successfully\n\n", folderName)
	return nil
}

func InitSpecInputPy(pythonScript string, domain string, tldOutput string, targetOutput string) error{
	cmd := exec.Command("python3", pythonScript, "--domain", domain, "--tld-output", tldOutput, "--target-output", targetOutput)

	return getResultCmd(*cmd)
}

func CheckFormatSpecPy(pythonScript string, formatInput string, formatResultOutput string) error{
	cmd := exec.Command("python3", pythonScript, "--input-file", formatInput, "--output-file", formatResultOutput)

	return getResultCmd(*cmd)
}

func CombineJSONFile(pythonScript string, file1 string, file2 string, combineFileOutput string) error{
	cmd := exec.Command("python3", pythonScript, "--file1", file1, "--file2", file2, "--output", combineFileOutput)

	return getResultCmd(*cmd)
}

func NameIdConvertPy(pythonScript string, input string, outputFile string) error{
	cmd := exec.Command("python3", pythonScript, "--input", input, "--output", outputFile)

	return getResultCmd(*cmd)
}

func getResultCmd(cmd exec.Cmd) error{
	output, err := cmd.CombinedOutput()

	if err != nil{
		printErrMessage(output, err)
		return err
	}

	fmt.Println(string(output))

	return nil
}

func CallBashFiles(scriptPath string, args ...string) error{
	cmd := exec.Command("bash", append([]string{scriptPath}, args...)...)

	// connect scripts stdin, stdout and stderr, to Go's terminal
	// in order for it to be able to receive user's input
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing script %s: %v\n", scriptPath, err)
		return err
	}
	return nil
}

func GetFileNameWithoutExt(path string) string {
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, ext)
}

func RemovePrefixFlag(flag string) string {
	return strings.TrimPrefix(flag, "--")
}