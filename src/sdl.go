package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const WindowScaleFactor = 30

type UserPrompt struct {
	Cancel          bool
	Prompt          string
	CopyToClipboard bool
	// Where to get context from, selection or clipboard. Default clipboard
	XClipCmd XClipCmd
}

func parseCommands(currentString string, userPrompt *UserPrompt) {
	i := 0

	currentString = strings.TrimSpace(currentString)

	for i < len(currentString) {
		if currentString[i] == '/' && i < len(currentString)-1 {
			switch currentString[i+1] {
			case 'c':
				userPrompt.CopyToClipboard = true

			case 's':
				userPrompt.XClipCmd = XClipCmdSelection
			}

			i += 2
			continue
		}

		break
	}

	userPrompt.Prompt = currentString[i:]
}

func renderText(text string, renderer *sdl.Renderer, font *ttf.Font) error {
	if len(text) == 0 {
		return nil
	}

	// Render text
	color := sdl.Color{R: 0, G: 0, B: 0, A: 255}

	surface, err := font.RenderUTF8Blended(text, color)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create surface for text. Err: +%v", err)
		return err
	}

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture from surface. Err: +%v", err)
		return err
	}

	err = renderer.Copy(texture, nil, &sdl.Rect{X: 10, Y: 10, W: surface.W, H: surface.H})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to copy texture. Err: +%v", err)
		return err
	}

	return nil
}

// Returns the prompt
func CreateWindow() UserPrompt {
	sdl.Init(sdl.INIT_EVERYTHING)

	window, err := sdl.CreateWindow(
		"",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		16*WindowScaleFactor,
		9*WindowScaleFactor,
		sdl.WINDOW_SHOWN|sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_BORDERLESS,
	)
	if err != nil {
		log.Fatal().Msg("Failed to create window")
	}

	renderer, err := sdl.CreateRenderer(window, 0, sdl.RENDERER_ACCELERATED)
	if err != nil {
		sdl.Quit()
		log.Fatal().Msg("Failed to create renderer")
	}

	if renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND) != nil {
		log.Fatal().Msg("Failed to set blend mode")
	}

	if ttf.Init() != nil {
		log.Fatal().Msg("Failed to init TTF")
	}

	rwops, err := sdl.RWFromMem(fontData)
	if err != nil {
		log.Fatal().Msg("Failed to init font")
	}

	font, err := ttf.OpenFontRW(rwops, 1, 14)
	if err != nil {
		log.Fatal().Msgf("Failed to load font. Err: +%v\n", err)
	}

	prompt := []byte{}
	currentString := string(prompt)

	quit := false

	userPrompt := UserPrompt{
		// cancel by defeault
		Cancel:          true,
		Prompt:          "",
		CopyToClipboard: false,
		XClipCmd:        XClipCmdClipboardPrimary,
	}

	for !quit {
		event := sdl.PollEvent()

		if event != nil {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				quit = true

			case *sdl.KeyboardEvent:
				{
					if e.Type == sdl.KEYUP {
						break
					}

					// Only handle ASCII
					if e.Keysym.Sym > 127 {
						break
					}

					if e.Repeat != 0 {
						break
					}

					switch sdl.GetKeyName(e.Keysym.Sym) {
					case "Escape":
						quit = true

					case "Return":
						{
							userPrompt.Cancel = false
							parseCommands(currentString, &userPrompt)

							quit = true
						}

					case "Backspace":
						{
							prompt = prompt[:len(prompt)-1]
							currentString = string(prompt)
						}

					default:
						{
							prompt = append(prompt, byte(e.Keysym.Sym))
							currentString = string(prompt)
						}
					}
				}
			}
		}

		renderer.SetDrawColor(255, 255, 255, 255)
		renderer.Clear()

		if renderText(currentString, renderer, font) != nil {
			quit = true
		}

		renderer.Present()
	}

	sdl.Quit()

	return userPrompt
}
