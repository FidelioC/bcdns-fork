package go_modules

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
)

func stringToBase256Int(s string) *big.Int {
    result := big.NewInt(0)
    base := big.NewInt(256)

    for i := 0; i < len(s); i++ {
        result.Mul(result, base)
        result.Add(result, big.NewInt(int64(s[i])))
    }
    return result
}

func FieldToInt(inputPath string) (map[string] string, error){
	file, err := os.ReadFile(inputPath)
    if err != nil {
        return nil, err
    }

    var original map[string]interface{}
    if err := json.Unmarshal(file, &original); err != nil {
        return nil, err
    }

    result := make(map[string]string)
    keys := []string{"name", "id"}

    for _, key := range keys {
        if val, ok := original[key].(string); ok {
            num := stringToBase256Int(val)
            result[key] = num.String()
        }
    }

    return result, nil
}

func WriteJSONToFile(data map[string]string, outputPath string) error {
    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(data)
}

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

func InitSpecInputPy(domain string, tldOutput string, targetOutput string) error{
	cmd := exec.Command("python3", "initSpecInput.py", "--domain", domain, "--tld-output", tldOutput, "--target-output", targetOutput)

	output, err := cmd.CombinedOutput()
	if err != nil{
		printErrMessage(output, err)
		return err
	}

	fmt.Println(string(output))
	return nil
}

func CheckFormatSpecPy(formatInput string, formatResultOutput string) error{
	cmd := exec.Command("python3", "checkSpecFormat.py", "--input-file", formatInput, "--output-file", formatResultOutput)

	output, err := cmd.CombinedOutput()
	if err != nil{
		printErrMessage(output, err)
		return err
	}

	fmt.Println(string(output))
	return nil
}

func CombineJSONFile(file1 string, file2 string, combineFileOutput string) error{
	cmd := exec.Command("python3", "combineJson.py", "--file1", file1, "--file2", file2, "--output", combineFileOutput)

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