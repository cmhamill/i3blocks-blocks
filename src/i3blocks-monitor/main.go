package main

import (
    "fmt"
    "os/exec"
    "strings"
)

import "github.com/godbus/dbus"


func signal_i3blocks(i int) error {
    return exec.Command("pkill", fmt.Sprintf("-RTMIN+%d", i), "i3blocks").Run()
}

func main() {
    conn, _ := dbus.SystemBus()

    conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
        strings.Join([]string{
            "type='signal'",
            "path='/'",
            "sender='net.connman'",
            "interface='net.connman.Manager'",
            "member='PropertyChanged'",
            "arg0='State'",
        }, ","),
    )
    conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
    strings.Join([]string{
            "type='signal'",
            "path='/org/freedesktop/UPower/devices/DisplayDevice'",
            "interface='org.freedesktop.DBus.Properties'",
            "member='PropertiesChanged'",
        }, ","),
    )

    c := make(chan *dbus.Signal, 10)
    conn.Signal(c)
    for v := range c {
        switch v.Name {
            case "net.connman.Manager.PropertyChanged": signal_i3blocks(1)
            case "org.freedesktop.DBus.Properties.PropertiesChanged":
                switch v.Path {
                    case "/org/freedesktop/UPower/devices/DisplayDevice":
                        signal_i3blocks(3)
                }
        }
    }
}
