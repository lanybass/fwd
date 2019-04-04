package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)
var sigs chan os.Signal

func dialTcpTimeout(proto, laddr, raddr string, timeout time.Duration) (net.Conn, error) {
	localaddr, err := net.ResolveTCPAddr(proto, laddr);
	errHandler(err)
	d := net.Dialer{Timeout: timeout, LocalAddr: localaddr}
	return d.Dial(proto, raddr)
}

func getLocalAddrs() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var list []net.IP
	for _, addr := range addrs {
		v := addr.(*net.IPNet)
		if v.IP.To4() != nil {
			list = append(list, v.IP)
		}
	}
	return list, nil
}

func fwd(src net.Conn, remote string, proto string, cliAddr string, connTimeout int) {
	//dst, err := net.Dial(proto, remote)
	dst, err := dialTcpTimeout(proto, cliAddr, remote, time.Second * time.Duration(connTimeout))
	if err != nil {
		errPrinter(err)
		src.Close()
		return
	}
	go func() {
		_, err = io.Copy(src, dst)
		errPrinter(err)
	}()
	go func() {
		_, err = io.Copy(dst, src)
		errPrinter(err)
	}()
}

func errHandler(err error) {
	if err != nil {
		color.Set(color.FgRed)
		fmt.Fprintf(os.Stderr, "[Error] %s\n", err.Error())
		color.Unset()
		os.Exit(1)
	}
}

// TODO: merge error handling functions
func errPrinter(err error) {
	if err != nil {
		color.Set(color.FgRed)
		fmt.Fprintf(os.Stderr, "[Error] %s\n", err.Error())
		color.Unset()
	}
}

func tcpStart(from string, to string,cliAddr string, connTimeout int) {
	proto := "tcp"

	localAddress, err := net.ResolveTCPAddr(proto, from)
	errHandler(err)

	remoteAddress, err := net.ResolveTCPAddr(proto, to)
	errHandler(err)

	listener, err := net.ListenTCP(proto, localAddress)
	errHandler(err)

	defer listener.Close()

	fmt.Printf("Forwarding %s traffic from '%v' to '%v'\n", proto, localAddress, remoteAddress)
	color.Set(color.FgYellow)
	fmt.Println("<CTRL+C> to exit")
	fmt.Println()
	color.Unset()

	for {
		src, err := listener.Accept()
		errHandler(err)
		fmt.Printf("New connection established from '%v'\n", src.RemoteAddr())
		go fwd(src, to, proto, cliAddr, connTimeout)
	}
}

func udpStart(from string, to string) {
	proto := "udp"

	localAddress, err := net.ResolveUDPAddr(proto, from)
	errHandler(err)

	remoteAddress, err := net.ResolveUDPAddr(proto, to)
	errHandler(err)

	listener, err := net.ListenUDP(proto, localAddress)
	errHandler(err)
	defer listener.Close()

	dst, err := net.DialUDP(proto, nil, remoteAddress)
	errHandler(err)
	defer dst.Close()

	fmt.Printf("Forwarding %s traffic from '%v' to '%v'\n", proto, localAddress, remoteAddress)
	color.Set(color.FgYellow)
	fmt.Println("<CTRL+C> to exit")
	fmt.Println()
	color.Unset()

	buf := make([]byte, 512)
	for {
		rnum, err := listener.Read(buf[0:])
		errHandler(err)

		_, err = dst.Write(buf[:rnum])
		errHandler(err)

		fmt.Printf("%d bytes forwared\n", rnum)
	}
}

func ctrlc() {
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		color.Set(color.FgGreen)
		fmt.Println("\nExecution stopped by", sig)
		color.Unset()
		os.Exit(0)
	}()
}

func RunProxy() (err error){
	app := cli.NewApp()
	app.Name = "fwd"
	app.Version = "1.0.0"
	app.Usage = "The little forwarder that could"
	app.UsageText = "fwd --from localhost:2222 --to 192.168.1.254:22"
	app.Copyright = "MIT License"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Joel Bastos",
			Email: "kintoandar@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "from, f",
			Value:  "127.0.0.1:8000",
			EnvVar: "FWD_FROM",
			Usage:  "source HOST:PORT",
		},
		cli.StringFlag{
			Name:   "to, t",
			EnvVar: "FWD_TO",
			Usage:  "destination HOST:PORT",
		},
		cli.StringFlag{
			Name:   "cli_addr, ca",
			EnvVar: "CLI_ADDR",
			Usage:  "client addr HOST:PORT",
		},
		cli.IntFlag{
			Name:   "conn_timeout, ct",
			EnvVar: "CONN_TIMEOUT",
			Usage:  "conn timeout seconds",
		},
		cli.BoolFlag{
			Name:  "list, l",
			Usage: "list local addresses",
		},
		cli.BoolFlag{
			Name:  "udp, u",
			Usage: "enable udp forwarding (tcp by default)",
		},
		cli.BoolFlag{
			Name:  "build, b",
			Usage: "build information",
		},
	}
	sigs = make(chan os.Signal, 1)
	app.Action = func(c *cli.Context) error {
		defer color.Unset()
		color.Set(color.FgGreen)
		if c.Bool("list") {
			list, err := getLocalAddrs()
			errHandler(err)
			fmt.Println("Available local addresses:")
			color.Unset()
			for _, ip := range list {
				fmt.Println(ip)
			}
			return nil
		} else if c.Bool("build") {
			fmt.Println("Built with " + runtime.Version() + " for " + runtime.GOOS + "/" + runtime.GOARCH)
			color.Unset()
			return nil

		} else if c.String("to") == "" {
			color.Unset()
			cli.ShowAppHelp(c)
			return nil
		} else {
			ctrlc()
			if c.Bool("udp") {
				udpStart(c.String("from"), c.String("to"))

			} else {
				tcpStart(c.String("from"), c.String("to"), c.String("cli_addr"), c.Int("conn_timeout"))
			}
			return nil
		}
	}
	return app.Run(os.Args)
}
