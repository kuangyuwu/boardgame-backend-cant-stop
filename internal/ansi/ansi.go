package ansi

import (
	"fmt"
	"strings"
)

type Color interface {
	fgColorCode() string
	bgColorCode() string
}

type effect string

const (
	Bold      effect = "1"
	Underline effect = "4"
)

func SetFont(fg Color, bg Color, effects ...effect) func(string) string {
	codes := make([]string, 0)
	if _, ok := fg.(defaultColor); !ok {
		codes = append(codes, fg.fgColorCode())
	}
	if _, ok := bg.(defaultColor); !ok {
		codes = append(codes, bg.bgColorCode())
	}
	for _, e := range effects {
		codes = append(codes, string(e))
	}
	if len(codes) == 0 {
		return func(s string) string { return s }
	}
	ansiCode := fmt.Sprintf("\033[%sm", strings.Join(codes, ";"))
	fontSetter := func(s string) string {
		return fmt.Sprintf("%s%s\033[0m", ansiCode, s)
	}
	return fontSetter
}

type defaultColor struct{}

var DefaultColor = defaultColor{}

func (c defaultColor) fgColorCode() string {
	return ""
}

func (c defaultColor) bgColorCode() string {
	return ""
}

type color4Bit int8

const (
	Black color4Bit = 30 + iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BrightBlack color4Bit = 82 + iota
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

func (c color4Bit) fgColorCode() string {
	return fmt.Sprintf("%d", c)
}

func (c color4Bit) bgColorCode() string {
	return fmt.Sprintf("%d", c+10)
}

type Color8Bit uint8

func (c Color8Bit) fgColorCode() string {
	return fmt.Sprintf("38;5;%d", c)
}

func (c Color8Bit) bgColorCode() string {
	return fmt.Sprintf("48;5;%d", c)
}

type ColorRGB [3]uint8

func HexToRGB(hex int) ColorRGB {
	if hex < 0 || hex > 0xffffff {
		panic("ansifont: invalid hex code (must be between 0x000000 and 0xffffff)")
	}
	r := uint8((hex & 0xff0000) >> 16)
	g := uint8((hex & 0x00ff00) >> 8)
	b := uint8(hex & 0x0000ff)
	return ColorRGB{r, g, b}
}

func (c ColorRGB) fgColorCode() string {
	return fmt.Sprintf("38;2;%d;%d;%d", c[0], c[1], c[2])
}

func (c ColorRGB) bgColorCode() string {
	return fmt.Sprintf("38;2;%d;%d;%d", c[0], c[1], c[2])
}
