package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

const XCLIP_CMD string = "xclip"

type XClipCmd = uint8

const (
	XClipCmdSelection XClipCmd = iota
	XClipCmdClipboardPrimary
)

func XClipCmdFromArg() XClipCmd {
	if len(os.Args) < 2 {
		return XClipCmdClipboardPrimary
	}

	arg := os.Args[1]

	if arg == "s" {
		return XClipCmdSelection
	}

	if arg == "c" {
		return XClipCmdClipboardPrimary
	}

	// Default is clipboard
	return XClipCmdClipboardPrimary
}

const DEBUG = true

func log(format string, a ...any) {
	if !DEBUG {
		return
	}

	s := ""

	if format[len(format)-1] != '\n' {
		s = "\n"
	}

	fmt.Printf(fmt.Sprintf("[DEBUG] %s%s", s, format), a...)
}

func xclipCmd(cmd XClipCmd) (*bytes.Buffer, error) {
	args := []string{}

	switch cmd {
	case XClipCmdSelection:
		{
			args = append(args, "-o", "-selection", "primary")
		}

	case XClipCmdClipboardPrimary:
		{
			args = append(args, "-o", "-selection", "clipboard")
		}
	}

	stdout := bytes.NewBuffer([]byte{})

	command := exec.Command(XCLIP_CMD, args...)
	command.Stdout = stdout

	err := command.Run()

	log("Running command: %s\n", command.String())

	if err != nil {
		return stdout, err
	}

	return stdout, nil
}

func main() {
	_, err := exec.LookPath(XCLIP_CMD)
	if err != nil {
		fmt.Fprintln(os.Stderr, "xclip not found. Please install 'xclip'")
		os.Exit(1)
	}

	clipOrSelection := XClipCmdFromArg()

	selection, err := xclipCmd(clipOrSelection)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err while executing xclip: %+v\n", err)
		os.Exit(1)
	}

	data, err := io.ReadAll(selection)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read from stdout buffer: %+v\n", err)
		os.Exit(1)
	}

	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok {
		fmt.Fprintf(os.Stderr, "EnvVar API_KEY not found")
		os.Exit(1)
	}

	log("Output: %s", string(data))

	openAiApiReq := OpenAIAPIRequest{
		Model: "chatgpt-4o-latest",
		Store: true,
		Messages: GetConverstaionMessages(RequestMessage{
			Role:    "user",
			Content: string(data),
		}),
	}

	resp, err := SendOpenAIRequest(openAiApiReq, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "OpenAPI req failed: %+v", err)
		os.Exit(1)
	}

	content := resp.Choices[0].Message.Content

	fmt.Println(content)
}
