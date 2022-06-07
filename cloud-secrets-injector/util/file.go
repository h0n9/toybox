package util

import "os"

func SaveStringToFile(data, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.WriteString(data)
	if err != nil {
		return err
	}
	return nil
}
