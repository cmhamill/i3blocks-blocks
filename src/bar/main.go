package bar

import (
    "bytes"
    "encoding/json"
    "os"
    "os/exec"
    "strconv"
)

type Alignment  string
type Bool       *bool
type Markup     string
type Width      *int
type MouseButton int

type Point struct {
    X int
    Y int
}

type Input struct {
    BlockName       *string
    BlockInstance   *string
    MouseButton     *MouseButton
    MouseLocation   *Point
}

type Output struct {
    FullText            string      `json:"full_text"`
    ShortText           string      `json:"short_text,omitempty"`
    Color               string      `json:"color,omitempty"`
    MinWidth            Width       `json:"min_width,omitempty"`
    Align               Alignment   `json:"align,omitempty"`
    Name                string      `json:"name,omitempty"`
    Instance            string      `json:"instance,omitempty"`
    Urgent              Bool        `json:"urgent,omitempty"`
    Separator           Bool        `json:"separator,omitempty"`
    SeparatorBlockWidth Width       `json:"separator_block_width,omitempty"`
    Markup              Markup      `json:"markup,omitempty"`
}

func (o *Output) ToJSON() (string, error) {
    b, err := json.Marshal(o)
    b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
    b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
    b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
    return string(b), err
}

const (
    MouseButtonLeft     MouseButton = 1
    MouseButtonMiddle   MouseButton = 2
    MouseButtonRight    MouseButton = 3
    MouseScrollUp       MouseButton = 4
    MouseScrollDown     MouseButton = 5
)

const (
    AlignLeft   Alignment = "left"
    AlignCenter Alignment = "center"
    AlignRight  Alignment = "right"
)

const (
    MarkupPango Markup = "pango"
    MarkupNone  Markup = "none"
)

func NewBool(b bool) Bool {
    return &b
}

func NewWidth(i int) Width {
    return &i
}

func GetInput() *Input {
    i := Input{}

    if v := os.Getenv("BLOCK_NAME"); v != "" {
        i.BlockName = &v
    }

    if v := os.Getenv("BLOCK_INSTANCE"); v != "" {
        i.BlockInstance = &v
    }

    if v := os.Getenv("BLOCK_BUTTON"); v != "" {
        b, _ := strconv.Atoi(v);
        mb := MouseButton(b)
        i.MouseButton = &mb
    }

    if x := os.Getenv("BLOCK_X"); x != "" {
        if y := os.Getenv("BLOCK_Y"); y != "" {
            X, _ := strconv.Atoi(x)
            Y, _ := strconv.Atoi(y)

            i.MouseLocation = &Point{X: X, Y: Y}
        }
    }

    return &i
}

func MoreInfo(title string, info string) error {
    return exec.Command(
        "notify-send", "-u", "normal", "-t", "5000", title, info,
    ).Run()
}
