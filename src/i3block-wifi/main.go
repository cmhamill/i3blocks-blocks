package main

import (
    "fmt"
    "io/ioutil"
    "os/exec"
    "strings"
)

import "bar"

const (
    WLAN_ICON = ""
    WLAN_IFACE = "wlan0"
    IWGETID_BIN = "/sbin/iwgetid"
)

type NoWirelessNetwork error;

func GetWirelessNetwork() (string, error) {
    cmd := exec.Command(IWGETID_BIN, "-s")
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return "", err
    }

    if err := cmd.Start(); err != nil {
        return "", err
    }

    rawessid, err := ioutil.ReadAll(stdout)
    if err != nil {
        return "", nil
    }

    essid := strings.TrimSpace(string(rawessid))

    if err := cmd.Wait(); err != nil {
        return essid, err
    }

    return essid, nil
}

func main() {
    out := bar.Output{}

    network, err := GetWirelessNetwork()
    if err != nil {
        out.Separator = bar.NewBool(false)
        out.SeparatorBlockWidth = bar.NewWidth(0)
    } else {
        out.FullText = fmt.Sprintf("%s %s", WLAN_ICON, network)
        out.ShortText = out.FullText
    }

    txt, _ := out.ToJSON()
    fmt.Println(txt)
}
