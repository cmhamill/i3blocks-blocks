package main

import (
    "bufio"
    "fmt"
    "os/exec"
    "strconv"
    "strings"
)

import "bar"

type Arity int

type Mixer struct {
    Name string
    Arity Arity
    Volume int
    Muted bool
}

const (
    ICON_VOL_OFF = ""
    ICON_VOL_DOWN = ""
    ICON_VOL_UP = ""
    ICON_MIC_OFF = ""
    ICON_MIC_ON = ""
)

const (
    COLOR_NORMAL = "#586e75"
    COLOR_MUTE = "#b58900"
)

const (
    ArityMono Arity = iota
    ArityStereo Arity = iota
)

func parse_amixer_output(name string) (*Mixer, error) {
    m := &Mixer{Name: name}

    cmd := exec.Command("/usr/bin/amixer", "-M", "sget", m.Name)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }

    if err := cmd.Start(); err != nil {
        return nil, err
   }

    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        t := scanner.Text()

        if strings.HasPrefix(strings.TrimSpace(t), "Playback channels:") ||
                strings.HasPrefix(strings.TrimSpace(t), "Capture channels:") {
            if strings.TrimSpace(strings.Split(t, ":")[1]) == "Mono" {
                m.Arity = ArityMono
            } else {
                m.Arity = ArityStereo
            }
        } else if (m.Arity == ArityMono && strings.HasPrefix(strings.TrimSpace(t), "Mono:")) ||
                (m.Arity == ArityStereo && strings.HasPrefix(strings.TrimSpace(t), "Front Left:")) {
            parts := strings.Split(t, "[")
            m.Volume, _ = strconv.Atoi(strings.TrimSuffix(strings.Split(parts[1], "]")[0], "%"))

            muted := strings.TrimSuffix(parts[len(parts)-1], "]")
            if muted == "off" {
                m.Muted = true
            } else {
                m.Muted = false
            }
        }

        if err := scanner.Err(); err != nil {
            return nil, err
        }
    }

    if err := cmd.Wait(); err != nil {
        return m, err
    }

    return m, nil
}

func main() {
    args := bar.GetInput()
    out := bar.Output{}

    master, _ := parse_amixer_output("Master")
    mic, _ := parse_amixer_output("Capture")

    master_icon := ICON_VOL_OFF
    master_color := COLOR_NORMAL

    if master.Muted {
        master_color = COLOR_MUTE
    }
    switch v := master.Volume; {
        case v <= 33:           master_icon = ICON_VOL_OFF
        case v > 33 && v <= 66: master_icon = ICON_VOL_DOWN
        case v > 66:            master_icon = ICON_VOL_UP
    }

    mic_icon := ICON_MIC_ON
    mic_color := COLOR_NORMAL
    if mic.Muted {
        mic_icon = ICON_MIC_OFF
        mic_color = COLOR_MUTE
    }

    if args.MouseButton != nil {
        info := fmt.Sprintf("Master: %d%%", master.Volume)
        if master.Muted {
            info += ", muted"
        }
        info += "\n"

        info += fmt.Sprintf("Microphone: %d%%", mic.Volume)
        if mic.Muted {
            info += ", muted"
        }
        info += "\n"

        bar.MoreInfo("Volume Info", info)
    }

    out.FullText = fmt.Sprintf(
        `<span foreground='%s'>%s</span> <span foreground='%s'>%s</span>`,
        master_color, master_icon, mic_color, mic_icon,
    )
    out.ShortText = out.FullText
    out.Markup = bar.MarkupPango

    txt, _ := out.ToJSON()
    fmt.Println(txt)

}
