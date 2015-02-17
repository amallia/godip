package main

import (
        "bufio"
        "crypto/md5"
        "fmt"
        "io"
        "log"
        "net"
        "os"
        "os/exec"
        "strings"
        "time"

	"github.com/dgv/godip/deps/ini"
)

type domain struct {
        user     string
        password string
        script   string
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

func script(hostname, ip, cmd string) (err error) {
        cmd = strings.Replace(cmd, "<domain>", hostname, -1)
        cmd = strings.Replace(cmd, "<ip>", ip, -1)
        c := strings.Split(cmd, " ")
        out, err := exec.Command("/bin/sh", c...).CombinedOutput()
        if err != nil {
                log.Printf("%v\n", err)
        } else {
                log.Printf("%s\n", out)
        }
        return
}

func badlogin(c net.Conn, host string) {
	      log.Printf("[%s] bad login", host)
              c.Write([]byte(string("1")))
}

func main() {
        ln, err := net.Listen("tcp", ":3495")
        if err != nil {
                // handle error
                panic(err)
        }
        log.Println("listening on :3495")
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
                        remoteHost := strings.Split(c.RemoteAddr().String(), ":")
                        host := remoteHost[0]
                        defer func (c net.Conn, host string) {
				log.Printf("[%s] close", host)
				c.Close()
			}(c, host)

                        h := md5.New()
                        io.WriteString(h, time.Now().String())
                        salt := fmt.Sprintf("%x", h.Sum(nil))[:10]
                        c.Write([]byte(salt))

                        rw := bufio.NewReader(c)
                        buf := make([]byte, 1024)
                        n, err := rw.Read(buf)
                        if err != nil {
                                log.Printf("%v", err)
                        }
                        msg := strings.Split(string(buf[0:n]), ":")
                        log.Printf("[%s] new conn", host)
                        // check update message form
                        if len(msg) == 5 {
                                if d := domains[msg[2]]; d == nil {
                                       badlogin(c, host)
                                } else {
                                        if d.auth(salt, msg[1]) {
                                                err := script(msg[2], host, domains[msg[2]].script)
                                                if err == nil {
                                                        c.Write([]byte(string("0")))
                                                } else {
                                                        badlogin(c, host)
                                                }
                                        } else {
                                                badlogin(c, host)
                                        }
                                }
                        }
                        if len(msg) == 4 {
                                switch msg[3][:1] {
                                case "0":
                                        if d := domains[msg[2]]; d == nil {
                                               badlogin(c, host)
                                        } else {
                                                if d.auth(salt, msg[1]) {
                                                        err := script(msg[2], host, domains[msg[2]].script)
                                                        if err == nil {
                                                                c.Write([]byte(string("0")))
                                                        } else {
                                                                badlogin(c, host)
                                                        }
						}
                                        }
                                case "1":
                                        if d := domains[msg[2]]; d == nil {
                                                badlogin(c, host)
                                        } else {
                                                if d.auth(salt, msg[1]) {
                                                        err := script(msg[2], host, domains[msg[2]].script)
                                                        if err == nil {
                                                                c.Write([]byte(string("2")))
                                                        } else {
                                                                badlogin(c, host)
                                                        }
                                                } else {
                                                       badlogin(c, host)
                                                }
                                        }
                                case "2":
                                        if d := domains[msg[2]]; d == nil {
                                                badlogin(c, host)
                                        } else {
                                                if d.auth(salt, msg[1]) {
                                                        err := script(msg[2], host, domains[msg[2]].script)
                                                        if err == nil {
                                                                c.Write([]byte(string("0:") + host))
                                                        } else {
                                                                badlogin(c, host)
                                                        }
                                                } else {
                                                       badlogin(c, host)
                                                }
                                        }
                                }
                        }
                }(conn)
        }
}
