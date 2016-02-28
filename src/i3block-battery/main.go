package main

import (
    "fmt"
    "path"
    "strconv"
    "strings"
    "time"
)

import "github.com/godbus/dbus"

import "bar"

const (
    UPOWER_BUS = "org.freedesktop.UPower"
    UPOWER_PATH = "/org/freedesktop/UPower"
)

const (
    ICON_AC = ""
    ICON_BAT_ABSENT = ""
    ICON_BAT_EMPTY = ""
    ICON_BAT_ONE_QUARTER = ""
    ICON_BAT_HALF = ""
    ICON_BAT_THREE_QUARTERS = ""
    ICON_BAT_FULL = ""
)

const (
    COLOR_INFO_NORMAL = "#839496"
    COLOR_NORMAL = "#586e75"
    COLOR_BAD = "#dc322f"
    COLOR_DEGRADED = "#b58900"
    COLOR_GOOD = "#859900"
)

var BAT_NAME = map[string]string{
    "BAT0": "Internal Battery",
    "BAT1": "Removable Battery",
    "":     "Total",
}

type DeviceType uint32
const (
    DeviceTypeUnknown DeviceType = iota
    DeviceTypeLinePower
    DeviceTypeBattery
    DeviceTypeUPS
    DeviceTypeMonitor
    DeviceTypeMouse
    DeviceTypeKeyboard
    DeviceTypePDA
    DeviceTypePhone
    DeviceTypeMediaPlayer
    DeviceTypeTablet
    DeviceTypeComputer
)

type DeviceState uint32
const (
    DeviceStateUnknown DeviceState = iota
    DeviceStateCharging
    DeviceStateDischarging
    DeviceStateEmpty
    DeviceStateFullyCharged
    DeviceStatePendingCharge
    DeviceStatePendingDischarge
)

func (s DeviceState) String() string {
    switch s {
        case DeviceStateUnknown:            return "Unknown"
        case DeviceStateCharging:           return "Charging"
        case DeviceStateDischarging:        return "Discharging"
        case DeviceStateEmpty:              return "Empty"
        case DeviceStateFullyCharged:       return "Full"
        case DeviceStatePendingCharge:      return "Pending Charge"
        case DeviceStatePendingDischarge:   return "Pending Discharge"
    }
    return ""
}

type PowerWarning uint32
const (
    PowerWarningUnknown PowerWarning = iota
    PowerWarningNone
    PowerWarningDischarging
    PowerWarningLow
    PowerWarningCritical
    PowerWarningAction
)

func (w PowerWarning) String() string {
    switch w {
        case PowerWarningUnknown:       return "Unknown"
        case PowerWarningNone:          return "None"
        case PowerWarningDischarging:   return "Discharging"
        case PowerWarningLow:           return "Low"
        case PowerWarningCritical:      return "Critical"
        case PowerWarningAction:        return "Action"
    }
    return ""
}

type Power struct {
    conn            *dbus.Conn
    OnBattery       bool
    DisplayDevice   *Battery
    Batteries       []*Battery
}

func NewPower() *Power {
    p := &Power{}
    p.conn, _ = dbus.SystemBus()

    onbatt, _ := p.conn.Object(UPOWER_BUS, UPOWER_PATH).GetProperty(
        UPOWER_BUS + ".OnBattery",
    )

    p.OnBattery     = onbatt.Value().(bool)
    p.DisplayDevice = NewBattery(p.conn, "DisplayDevice")

    return p
}

func (p *Power) PopulateBatteries() {
    var devices []dbus.ObjectPath
    p.conn.Object(UPOWER_BUS, UPOWER_PATH).Call(
        UPOWER_BUS + ".EnumerateDevices", 0,
    ).Store(&devices)

    p.Batteries = []*Battery{}
    for _, pth := range devices {
        typ, _ := p.conn.Object(UPOWER_BUS, pth).GetProperty(
            UPOWER_BUS + ".Device.Type",
        )

        if DeviceType(typ.Value().(uint32)) == DeviceTypeBattery {
            p.Batteries = append(p.Batteries,
                NewBattery(p.conn, path.Base(string(pth))),
            )
        }
    }
}

type Battery struct {
    ID              string
    State           DeviceState
    IsPresent       bool
    Percentage      int
    Warning         PowerWarning
    TimeToEmpty     *time.Duration
    TimeToFull      *time.Duration
}

func NewBattery(conn *dbus.Conn, devname string) *Battery {
    dev := conn.Object(UPOWER_BUS,
        dbus.ObjectPath(path.Join(UPOWER_PATH, "devices", devname)),
    )
    id, _           := dev.GetProperty(UPOWER_BUS + ".Device.NativePath")
    state, _        := dev.GetProperty(UPOWER_BUS + ".Device.State")
    present, _      := dev.GetProperty(UPOWER_BUS + ".Device.IsPresent")
    percent, _      := dev.GetProperty(UPOWER_BUS + ".Device.Percentage")
    warning, _      := dev.GetProperty(UPOWER_BUS + ".Device.WarningLevel")
    timeempty, _    := dev.GetProperty(UPOWER_BUS + ".Device.TimeToEmpty")
    timefull, _     := dev.GetProperty(UPOWER_BUS + ".Device.TimeToFull")

    b := Battery{
        ID:         id.Value().(string),
        State:      DeviceState(state.Value().(uint32)),
        IsPresent:  present.Value().(bool),
        Percentage: int(percent.Value().(float64)),
        Warning:    PowerWarning(warning.Value().(uint32)),
    }

    te := timeempty.Value().(int64)
    if te != 0 {
        t, _ := time.ParseDuration(strconv.FormatInt(te, 10) + "s")
        b.TimeToEmpty = &t
    }

    tf := timefull.Value().(int64)
    if tf != 0 {
        t, _ := time.ParseDuration(strconv.FormatInt(tf, 10) + "s")
        b.TimeToFull = &t
    }

    return &b
}

func (b *Battery) Info() string {
    battery_present := true
    info := []string{}

    title_icon := ""
    title_icon_color := COLOR_INFO_NORMAL
    if !b.IsPresent {
        battery_present = false
        title_icon = ICON_BAT_ABSENT + " "
        title_icon_color = COLOR_DEGRADED
    }
    info = append(info, fmt.Sprintf(
        `<span foreground='%s'>%s</span><i>%s</i>`,
        title_icon_color, title_icon, BAT_NAME[b.ID],
    ))
    if !battery_present {
        info = append(info, "")
        return strings.Join(info, "\n")
    }

    state_color := COLOR_INFO_NORMAL
    switch b.Warning {
        case PowerWarningCritical:  state_color = COLOR_BAD
        case PowerWarningLow:       state_color = COLOR_DEGRADED
    }
    info = append(info,
        fmt.Sprintf(`State: <span foreground='%s'>%s</span>`, state_color, b.State),
    )

    info = append(info, fmt.Sprintf("Capacity: %d%%", b.Percentage))

    if b.TimeToFull != nil {
        info = append(info, fmt.Sprintf(
            "Time until full: %s", FmtDuration(b.TimeToFull),
        ))
    } else if b.TimeToEmpty != nil {
        info = append(info, fmt.Sprintf(
            "Time until empty: %s", FmtDuration(b.TimeToEmpty),
        ))
    }

    info = append(info, "")
    return strings.Join(info, "\n")
}

func FmtDuration(t *time.Duration) string {
    out := ""

    hours := int(t.Hours())
    minutes := int(t.Minutes()) % 60

    if hours > 0 {
        out += fmt.Sprintf("%d hours", hours)
        if minutes > 0 {
            out += ", "
        }
    }

    switch minutes {
        case 0:     break
        case 1:     out += "1 minute"
        default:    out += fmt.Sprintf("%d minutes", minutes)
    }

    return out
}

func main() {
    args := bar.GetInput()
    out := bar.Output{}

    p := NewPower()

    text := []string{}

    ac_color := COLOR_NORMAL
    if !p.OnBattery {
        if p.DisplayDevice.State == DeviceStateCharging {
            ac_color = COLOR_GOOD
        }
        text = append(text,
            fmt.Sprintf(`<span foreground='%s'>%s</span>`, ac_color, ICON_AC),
        )
    }

    var batt_icon string
    switch n := p.DisplayDevice.Percentage; {
        case n <= 13:
            batt_icon = ICON_BAT_EMPTY
        case n <= 38:
            batt_icon = ICON_BAT_ONE_QUARTER
        case n <= 63:
            batt_icon = ICON_BAT_HALF
        case n <= 88:
            batt_icon = ICON_BAT_THREE_QUARTERS
        default:
            batt_icon = ICON_BAT_FULL
    }

    batt_color := COLOR_NORMAL
    switch p.DisplayDevice.Warning {
        case PowerWarningLow:       batt_color = COLOR_DEGRADED
        case PowerWarningCritical:  batt_color = COLOR_BAD
    }
    if p.DisplayDevice.State == DeviceStateFullyCharged {
        batt_color = COLOR_GOOD
    }

    text = append(text,
        fmt.Sprintf(`<span foreground='%s'>%s</span>`, batt_color, batt_icon),
    )

    out.FullText = strings.Join(text, " ")
    out.ShortText = out.FullText
    out.Markup = bar.MarkupPango

    txt, _ := out.ToJSON()
    fmt.Println(txt)

    if args.MouseButton != nil {
        info := p.DisplayDevice.Info() + "\n"

        if *args.MouseButton != bar.MouseButtonLeft {
            p.PopulateBatteries()
            removable_battery_present := false

            for _, b := range p.Batteries {
                info += b.Info() + "\n"
                if BAT_NAME[b.ID] == "Removable Battery" {
                    removable_battery_present = true
                }
            }
            if !removable_battery_present {
                b := Battery{ID: "BAT1", IsPresent: false}
                info += b.Info() + "\n"
            }
        }

        bar.MoreInfo("Battery Info", info)
    }
}
