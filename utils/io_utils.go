package utils

import (
	"bufio"
	"os"
)

func ReadFileLinesIntoBytes(filepath string) ([][]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var byteStrings [][]byte

	for scanner.Scan() {
		byteStrings = append(byteStrings, scanner.Bytes())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return byteStrings, nil
}

func WriteBytesInfoFileLines(data [][]byte, filepath string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		os.Create(filepath)
	}

	file, err := os.OpenFile(filepath, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	for i := 0; i < len(data); i++ {
		_, err := file.Write([]byte(data[i]))
		if err != nil {
			return err
		}

		_, err = file.WriteString("\n")
		if err != nil {
			return err
		}
	}

	return nil
}
