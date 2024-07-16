package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var blacklist = map[string]struct{}{
	"cmp_name": {}, "cmp_id": {},
	"system": {}, "grpc_log": {},
	"http-scheme": {}, "http-proto": {},
	"remote-addr": {}, "user-agent": {},
	"unit": {}, "status": {},
	"span.kind": {}, "num_migrations": {},
	"service": {}, "user_id": {},
	"query": {}, "is_read": {},
	"offset": {}, "limit": {},
	"sort": {}, "result": {},
	"name": {}, "version": {},
	"git_commit": {}, "go_version": {},
	"count": {}, "num_workspaces": {},
	"level": {}, "category": {}, "correlation_id": {}, "uri": {}, "ts": {}, "msg": {}, "layer": {},
}

var whitelist = map[string]struct{}{
	"error": {}, "errorVerbose": {},
	"panic_error": {}, "panic_stack_trace": {},
	"caller": {},
}

var colorMap = map[string]string{
	"default":      "\033[39m",
	"black":        "\033[30m",
	"red":          "\033[31m",
	"green":        "\033[32m",
	"yellow":       "\033[33m",
	"blue":         "\033[34m",
	"magenta":      "\033[35m",
	"cyan":         "\033[36m",
	"lightGray":    "\033[37m",
	"darkGray":     "\033[90m",
	"lightRed":     "\033[91m",
	"lightGreen":   "\033[92m",
	"lightYellow":  "\033[93m",
	"lightBlue":    "\033[94m",
	"lightMagenta": "\033[95m",
	"lightCyan":    "\033[96m",
	"white":        "\033[97m",
	"end":          "\033[0m",
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var logObject map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logObject); err != nil {
			fmt.Println(line)
			continue
		}

		// Determine color based on log level
		var color string
		switch logObject["level"] {
		case "fatal", "error":
			color = colorMap["red"]
		case "warn":
			color = colorMap["yellow"]
		case "info":
			color = colorMap["green"]
		case "debug":
			color = colorMap["cyan"]
		default:
			color = ""
		}
		colorEnd := ""
		if color != "" {
			colorEnd = colorMap["end"]
		}

		// Print the formatted log message
		fmt.Printf("%s%s%s - %s\n", color, logObject["ts"], colorEnd, logObject["msg"])

		for key, value := range logObject {
			if _, ok := whitelist[key]; ok {
				continue
			}
			if _, ok := blacklist[key]; ok {
				continue
			}
			if strValue, ok := value.(string); ok && strValue != "" {
				msg := "    " + strings.ReplaceAll(strValue, "\n", "\n    ")
				fmt.Printf("%s%s%s\n", colorMap["darkGray"], key, colorMap["end"])
				fmt.Printf("%s%s%s\n", colorMap["darkGray"], msg, colorMap["end"])
			}
		}

		for key := range whitelist {
			if value, ok := logObject[key]; ok {
				if strValue, ok := value.(string); ok && strValue != "" {
					msg := "    " + strings.ReplaceAll(strValue, "\n", "\n    ")
					fmt.Printf("%s%s%s\n", colorMap["darkGray"], key, colorMap["end"])
					fmt.Printf("%s%s%s\n", colorMap["darkGray"], msg, colorMap["end"])
				}
			}
		}

		// Insert a blank line after each log entry for separation
		fmt.Println("")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
