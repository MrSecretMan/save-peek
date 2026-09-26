package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MrSecretMan/save-peek/internal/stardew"
	webui "github.com/MrSecretMan/save-peek/internal/web"
)

func main() {
	var (
		lan     = flag.Bool("lan", false, "listen on the local network instead of localhost only")
		port    = flag.Int("port", 8273, "web server port")
		saveDir = flag.String("save-dir", "", "extra directory to search for Stardew saves")
	)
	flag.Parse()

	save, err := stardew.FindLatest(*saveDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "savepeek:", err)
		fmt.Fprintln(os.Stderr, "use -save-dir /path/to/StardewValley/Saves if your saves are somewhere unusual")
		os.Exit(1)
	}

	progress, err := stardew.Parse(save)
	if err != nil {
		log.Fatal(err)
	}

	host := "127.0.0.1"
	if *lan {
		host = "0.0.0.0"
	}
	addr := fmt.Sprintf("%s:%d", host, *port)

	fmt.Printf("farm:  %s\n", pick(progress.FarmName, save.Folder))
	fmt.Printf("save:  %s %d, Year %d\n", pick(progress.Season, "?"), progress.Day, progress.Year)
	if *lan {
		fmt.Println("lan:   anyone on this network can open the page")
		for _, ip := range localIPv4() {
			fmt.Printf("web:   http://%s:%d\n", ip, *port)
		}
	} else {
		fmt.Printf("web:   http://127.0.0.1:%d\n", *port)
		fmt.Println("hint:  use -lan if you want to open it from your phone")
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           webui.Server{Save: save}.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func localIPv4() []string {
	ifaces, _ := net.Interfaces()
	var out []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ip := net.ParseIP(strings.Split(addr.String(), "/")[0])
			if ip != nil && ip.To4() != nil {
				out = append(out, ip.String())
			}
		}
	}
	return out
}

func pick(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
