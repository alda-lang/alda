package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

func NormalizeHostString(host string) string {
	if !strings.HasPrefix(host, "http://") &&
		!strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}

	host = strings.TrimRight(host, "/")

	return host
}

func SanitizeError(err error) error {
	if strings.Contains(err.Error(), "connection refused") {
		return errors.New("No Alda server running.")
	} else {
		return err
	}
}

func Msg(host string, port int, msg string, args ...interface{}) {
	host = strings.Replace(host, "http://", "", 1)
	host = strings.Replace(host, "https://", "", 1)
	host = strings.TrimRight(host, "/")
	if host == "localhost" {
		host = ""
	} else {
		host = host + ":"
	}

	hostAndPort := color.BlueString(host + strconv.Itoa(port))
	prefix := fmt.Sprintf("[%s] ", hostAndPort)

	fmt.Fprintf(color.Output, prefix+msg+"\n", args...)
}

func Error(host string, port int, msg string, args ...interface{}) {
	fmtString := fmt.Sprintf("%s %s", color.RedString("ERROR"), msg)
	Msg(host, port, fmtString, args...)
}
