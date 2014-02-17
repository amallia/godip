package main

import (
        "bufio"
        "fmt"
        "io"
        "log"
        "net"
        "os"
        "strings"
	"crypto/md5"
	"time"
	"os/exec"

        "github.com/vaughan0/go-ini"
)

type domain struct {
        user     string
        password string
	script string
}

var domains = make(map[string]*domain)

func (d *domain) auth(salt, passwd string) bool {
        h := md5.New()
        io.WriteString(h, d.password+"."+salt)
        if fmt.Sprintf("%x", h.Sum(nil)) == passwd {
                return true
        } else {
                return false
        }
}

func script(cmd string) (err error) {
 	out, err := exec.Command(cmd).CombinedOutput()
	log.Printf("%s\n", out)
	return
}

func main() {
        ln, err := net.Listen("tcp", ":3496")
        if err != nil {
                // handle error
                panic(err)
        }
	log.Println("listening on :3496")
	// check ini
        file, err := ini.LoadFile("domains.ini")
        info, _ := os.Stat("domains.ini")
        if info.Mode() != 256 {
                println("Set chmod 400 to domains.ini before use it")
                return
        }
	// load domain from ini
        for name, section := range file {
                h := md5.New()
                io.WriteString(h, section["password"])
		d := domains[name]
		if d != nil {
			println("domain duplicated, check the sections")
			return
		}
                domains[name] = &domain{section["user"], fmt.Sprintf("%x", h.Sum(nil)), section["script"]}
        }

        for {
                conn, err := ln.Accept()
                if err != nil {
                        log.Printf("%v", err)
                }
                go func(c net.Conn) {
                        defer c.Close()

                        log.Printf("Incoming connection from %s", c.RemoteAddr())
			h := md5.New()
			io.WriteString(h, time.Now().String())
			salt :=  fmt.Sprintf("%x", h.Sum(nil))[:10]
                        c.Write([]byte(salt))
                        
                        rw := bufio.NewReader(c)
                        buf := make([]byte, 1024)
                        n, err := rw.Read(buf)
                        if err != nil {
                                log.Printf("%v", err)
                        }
                        msg := strings.Split(string(buf[0:n]), ":")

			// check update message form
                        if len(msg) == 5 {
                                if d := domains[msg[2]]; d == nil {
                                        c.Write([]byte(string("1")))
                                } else {
                                        if d.auth(salt, msg[1]) {
                                                c.Write([]byte(string("0")))
                                        } else {
                                                c.Write([]byte(string("1")))
                                        }
                                }
                        }
                        if len(msg) == 4 {                         
				switch msg[3][:1] {
                                case "0":
                                if d := domains[msg[2]]; d == nil {
                                        c.Write([]byte(string("1")))
                                } else {
                                        if d.auth(salt, msg[1]) {
                                                c.Write([]byte(string("0")))
                                        } else {
                                                c.Write([]byte(string("1")))
                                        }
                                }
                                case "1":
                                        if d := domains[msg[2]]; d == nil {
                                                c.Write([]byte(string("1")))
                                        } else {
                                                if d.auth(salt, msg[1]) {
                                                        c.Write([]byte(string("2")))
                                                } else {
                                                        c.Write([]byte(string("1")))
                                                }
                                        }
                                case "2":
                                        if d := domains[msg[2]]; d == nil {
                                                c.Write([]byte(string("1")))
                                        } else {
                                                if d.auth(salt, msg[1]) {
							host := strings.Split(c.RemoteAddr().String(), ":")
							err := script(domains[msg[2]].script)
							if err == nil {
                                                        c.Write([]byte(string("0:") + host[0]))
							} else {
							c.Write([]byte(string("1")))
							}
                                                } else {
                                                        c.Write([]byte(string("1")))
                                                }
                                        }
                                }
                        }
                }(conn)
        }
}
