package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const XCLIP_CMD string = "xclip"

type XClipCmd = uint8

const (
	XClipCmdSelection XClipCmd = iota
	XClipCmdClipboardPrimary
)

var APIKey string

func XClipCmdFromArg() XClipCmd {
	if len(os.Args) < 2 {
		return XClipCmdClipboardPrimary
	}

	arg := os.Args[1]

	os.Args = os.Args[1:]

	if arg == "s" {
		return XClipCmdSelection
	}

	if arg == "c" {
		return XClipCmdClipboardPrimary
	}

	// Default is clipboard
	return XClipCmdClipboardPrimary
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

	log.Debug().Msgf("Running command: %s\n", command.String())

	if err != nil {
		return stdout, err
	}

	return stdout, nil
}

func openConvo(content string) {
	fileName := "/tmp/thingy.md"

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to open file: '%s'", fileName)
	}
	log.Info().Msgf("Opened file '%s' for writing", fileName)

	n, err := file.WriteString(content)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write to file")
	}
	log.Info().Msgf("Wrote %d bytes to file %s", n, fileName)

	nvim, err := exec.LookPath("nvim")

	cmd := exec.Command("gnome-terminal", "--window", "--maximize", "--", nvim, fileName)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Info().Str("cmd", cmd.String()).Msg("Running command")
	err = cmd.Run()

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to launch nvim")
	}
}

func copyOutputToClipboard(content string) {
	log.Info().Str("output", content).Msg("Copying to clipboard")

	cmd := exec.Command("xclip", "-selection", "clipboard")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get stdin pipe")
	}

	if err := cmd.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run command")
	}

	_, err = stdin.Write([]byte(content))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write to cmd stdin")
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		log.Fatal().Err(err).Msg("Command failed")
	}
}

func setupLogger() {
	home, ok := os.LookupEnv("HOME")
	if !ok {
		log.Fatal().Msg("EnvVar HOME not found")
	}

	logFile, err := os.OpenFile(
		fmt.Sprintf("%s/agent.log", home),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open log file")
	}

	zerolog.TimeFieldFormat = time.RFC3339
	multi := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
		logFile,
	)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("Successfully set up logging")
}

func main() {
	setupLogger()

	_, err := exec.LookPath(XCLIP_CMD)
	if err != nil {
		log.Fatal().Err(err).Msg("xclip not found. Please install 'xclip'")
	}

	clipOrSelection := XClipCmdFromArg()

	selection, err := xclipCmd(clipOrSelection)
	if err != nil {
		log.Fatal().Err(err).Msg("Err while executing xclip")
	}

	data, err := io.ReadAll(selection)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read from stdout buffer")
	}

	log.Debug().Msgf("Output: %s", string(data))

	userPrompt := UserPrompt{}

	if len(os.Args) > 0 && os.Args[0] == "floating" {
		userPrompt = CreateWindow()
		userPrompt.Prompt += "\n"

		if userPrompt.Cancel {
			return
		}
	}

	openAiApiReq := OpenAIAPIRequest{
		Model: "chatgpt-4o-latest",
		Store: true,
		Messages: GetConverstaionMessages(RequestMessage{
			Role:    "user",
			Content: userPrompt.Prompt + string(data),
		}),
	}

	log.Info().Interface("message", openAiApiReq.Messages).Msg("Prompt")

	resp, err := SendOpenAIRequest(openAiApiReq, APIKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "OpenAPI req failed: %+v", err)
		os.Exit(1)
	}

	content := resp.Choices[0].Message.Content

	if userPrompt.CopyToClipboard {
		copyOutputToClipboard(content)
	} else {
		openConvo(content)
	}
}
