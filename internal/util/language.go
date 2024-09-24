package util

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// getLanguageViaEnv retrieves the language setting from the environment variable "LANG".
// It uses a fallback mechanism provided by GetEnvsWithFallback to ensure a valid language
// is returned. If the "LANG" environment variable is set and not empty, a pointer to the
// language string is returned. If "LANG" is not set or is empty, the function returns nil.
func getLanguageViaEnv() *string {
	lang := GetEnvsWithFallback(os.Getenv("LANG"), "LANG", "CL_ALL")

	if lang != "" {
		return &lang
	}

	return nil
}

// getLanguageViaApi retrieves the preferred language of the user based on the operating system.
// It checks the OS type (Windows, macOS, or Linux) and executes the appropriate command to obtain
// the language settings.
//
// For Windows and macOS, it uses the `defaults` command to read the AppleLanguages setting,
// processes the output to convert it from a tuple format to JSON, and then unmarshals it into
// a slice of strings. The first language in the list is returned.
//
// For Linux, it reads the `/etc/locale.conf` file to find the LANG setting and returns it.
//
// Returns:
// - A pointer to a string containing the preferred language, or nil if the language could not be determined.
func getLanguageViaApi() *string {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("defaults", "read", "-g", "AppleLanguages")

		output, err := cmd.Output()

		if err != nil {
			return nil
		}

		// (
		// 		"en-US",
		// 		"zh-Hans"
		// )
		jsonString := strings.ReplaceAll(strings.ReplaceAll(string(output), "(", "["), ")", "]")

		var languages []string

		// Decode JSON
		if err := json.Unmarshal([]byte(jsonString), &languages); err != nil {
			return nil
		}

		if len(languages) == 0 {
			return nil
		}

		lang := strings.TrimSpace(languages[0])

		return &lang
	case "darwin":
		cmd := exec.Command("defaults", "read", "-g", "AppleLanguages")

		output, err := cmd.Output()

		if err != nil {
			return nil
		}

		// (
		// 		"en-US",
		// 		"zh-Hans"
		// )
		jsonString := strings.ReplaceAll(strings.ReplaceAll(string(output), "(", "["), ")", "]")

		var languages []string

		// Decode JSON
		if err := json.Unmarshal([]byte(jsonString), &languages); err != nil {
			return nil
		}

		if len(languages) == 0 {
			return nil
		}

		lang := strings.TrimSpace(languages[0])

		return &lang

	case "linux":
		file, err := os.Open("/etc/locale.conf")

		if err != nil {
			return nil
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "LANG=") {
				lang := strings.TrimPrefix(line, "LANG=")

				return &lang
			}
		}

		if err := scanner.Err(); err != nil {
			return nil
		}
	}

	return nil
}

// GetLanguage retrieves the current language setting of the operating system.
// It checks the environment variable "LANG" first and returns its value if available.
//
// On Windows, it executes a PowerShell command to get the installed UI culture name.
// On macOS, it reads the default Apple languages using the `defaults` command and
// returns the first language in the list.
// On Linux, it reads the "/etc/locale.conf" file to find the language setting.
//
// Returns a pointer to a string containing the language code (e.g., "en-us") or nil
// if the language cannot be determined or an error occurs during execution.
func getLanguage() *string {
	if lang := getLanguageViaApi(); lang != nil {
		Debug("getLanguageViaApi: %s\n", *lang)
		return lang
	}

	if lang := getLanguageViaEnv(); lang != nil {
		Debug("getLanguageViaEnv: %s\n", *lang)
		return lang
	}

	return nil
}

// IsSimplifiedChinese checks if the current language is Simplified Chinese.
// It returns true if the language is either "zh_CN" or "zh-Hans-CN",
// and false if the language is nil or does not match the specified values.
func IsSimplifiedChinese() bool {
	lang := getLanguage()

	if lang == nil {
		Debug("lang: nil\n")
		return false
	}

	Debug("lang: %s\n", *lang)

	simplifiedChineseSet := []string{"zh_CN", "zh-CN", "zh-Hans-CN"}

	for _, v := range simplifiedChineseSet {
		if strings.Contains(strings.ToLower(*lang), strings.ToLower(v)) {
			return true
		}
	}

	return false
}
