package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func FormatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}

func PasswordToHash(password, secretKey string) string {
	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password))
	hashedPassword := hex.EncodeToString(h.Sum(nil))
	return hashedPassword
}

// ComparePasswordHash compares a given password and secret key with a provided hash.
// It returns true if the computed hash matches the provided hash, otherwise false.
func ComparePasswordHash(password, secretKey, hash string) bool {
	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password))
	computedHash := hex.EncodeToString(h.Sum(nil))
	return computedHash == hash
}

func HasNil(slice ...interface{}) bool {
	for _, v := range slice {
		for _, j := range v.([]interface{}) {
			if j == nil {
				return true
			}
		}
	}
	return false
}

func CheckEmptyFields(s interface{}) []string {
	detectedFields := []string{}
	v := reflect.ValueOf(s)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		if field.Interface() != reflect.Zero(fieldType.Type).Interface() {
			detectedFields = append(detectedFields, fieldType.Name)
		}
	}
	return detectedFields
}

func ContainsStringInArray(substr string, arr []string) bool {
	for _, field := range arr {
		if strings.ToLower(substr) == strings.ToLower(field) {
			return true
		}
	}
	return false
}

func GetIdField(id any) string {
	return fmt.Sprintf("user%v", id)
}

func GetRootDirectory(file string) (string, error) {
	// Находим путь до корневого каталога, где и находится config.yaml
	projectDirPath, err := filepath.Abs("")
	if err != nil {
		return "", err
	}

	// Проверяем есть ли файл в данном каталоге, если нет, то поднимаемся на каталог вверх
	for i := 0; i <= 10; i++ {
		if _, err = os.Stat(projectDirPath + "\\" + file); os.IsNotExist(err) {
			projectDirPath = filepath.Dir(projectDirPath)
		} else if err == nil {
			break
		} else {
			return "", fmt.Errorf("not found root directory")
		}
	}

	return projectDirPath, nil
}
