package main

/*
Device			Bytes		Bits
-----------------------------------
iCE40-LP 384	7417		59336
iCE40-LP 640	32306		258448
iCE40-LP/HX 1K	32303		258424
iCE40-LP/HX 4K	135183		1081464
iCE40-LP/HX 8K	135183		1081464
iCE40LM 1K		68177		545416
iCE40LM 2K		68177		545416
iCE40LM 4K		68176		545408
iCE5LP 1K		71342		570736
iCE5LP 2K		71342		570736
iCE5LP 4K		71342		570736
iCE40UL 640		30942		247536
iCE40UL 1K		30942		247536
iCE40UP 3K		104161		833288
iCE40UP 5K		104161		833288
*/

import (
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("iceflash <port> [command]")
		fmt.Println("	-f              Firmware version (default command)")
		fmt.Println("	-s              Serial number of Flash chip")
		fmt.Println("	-h              Halt/Reset the iCE40")
		fmt.Println("	-g              Go/Start the iCE40")
		fmt.Println("	-c              Check status of CDONE pin of the iCE40")
		fmt.Println("	-e              Erase flash")
		fmt.Println("	-1              Delay 1 second")
		fmt.Println("	-w <filename>   Write flash")
		fmt.Println("	-r <filename>   Read flash")
		fmt.Println("	-t [<filename>] Verify crc32 checksum")
		os.Exit(1)
	}

	// Set up options for, and open the serial port
	options := serial.OpenOptions{
		PortName:              os.Args[1],
		BaudRate:              115200,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       0,
		InterCharacterTimeout: 100 * 100, // 100ms timeout
	}
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	defer port.Close()

	// Execute the default command if no commands are given
	if len(os.Args) == 2 {
		getFWstring(port)
		os.Exit(0)
	}

	filename := ""
	for i := 2; i < len(os.Args); i++ {

		switch os.Args[i] {
		case "-1": // Delay 1 second
			time.Sleep(1000 * time.Millisecond)

		case "-h": // Halt ICE40
			haltICE(port)

		case "-g": // Start ICE40
			startICE(port)

		case "-c": // Check status of CDONE
			getCDONE(port)

		case "-f": // Get firmware version
			getFWstring(port)

		case "-s": // Get flash chip serial number
			showSerial(port)

		case "-e": // Erase flash
			eraseFlash(port)

		case "-r": // Read from flash
			if i == len(os.Args)-1 || os.Args[i+1][0] == '-' {
				log.Fatalln("Filename missing for -r")
			}
			i++
			readFlash(port, os.Args[i], 32303)

		case "-w": // Write to flash
			if i == len(os.Args)-1 || os.Args[i+1][0] == '-' {
				log.Fatalln("Filename missing for -w")
			}
			i++
			filename = os.Args[i]
			writeFlash(port, filename)

		case "-t": // Verify CRC32 checksum
			if filename == "" {
				if i == len(os.Args)-1 || os.Args[i+1][0] == '-' {
					log.Fatalln("Filename missing for -t")
				}
				i++
				filename = os.Args[i]
			}
			testFlash(port, filename, 32303)
		}
	}
}

//
//
//
func getFWstring(port io.ReadWriteCloser) {
	sendByte(port, '!')
	reply := read(port, 1)
	replyLen := int(reply[0])
	serial := read(port, replyLen)
	fmt.Println(string(serial))
}

//
//
//
func haltICE(port io.ReadWriteCloser) {
	sendByte(port, 'h')
	reply := read(port, 1)
	replyLen := int(reply[0])
	status := read(port, replyLen)
	if status[0] != 'R' {
		log.Fatalln("Didn't receive expected result from haltICE")
	}
}

//
//
//
func startICE(port io.ReadWriteCloser) {
	sendByte(port, 'g')
	reply := read(port, 1)
	replyLen := int(reply[0])
	status := read(port, replyLen)
	if status[0] != 'r' {
		log.Fatalln("Didn't receive expected result from startICE")
	}
}

//
//
//
func getCDONE(port io.ReadWriteCloser) {
	sendByte(port, 'c')
	reply := read(port, 1)
	replyLen := int(reply[0])
	status := read(port, replyLen)
	if status[0] == 'c' {
		fmt.Println("STOPPED")
	} else if status[0] == 'C' {
		fmt.Println("ACTIVE")
	} else {
		log.Fatalln("Didn't receive expected result from getCDONE")
	}
}

//
//
//
func eraseFlash(port io.ReadWriteCloser) {
	fmt.Println("Erasing flash")
	sendByte(port, 'e')
	reply := read(port, 1)
	replyLen := int(reply[0])
	status := read(port, replyLen)
	if status[0] != 'O' {
		log.Fatalln("Didn't receive expected result from eraseFlash")
	}
}

//
//
//
func showSerial(port io.ReadWriteCloser) {
	sendByte(port, 's')
	reply := read(port, 1)
	replyLen := int(reply[0])
	serial := read(port, replyLen)
	dst := make([]byte, 2*replyLen)
	hex.Encode(dst, serial)
	fmt.Println(string(dst))
}

//
//
//
func readFlash(port io.ReadWriteCloser, filename string, bytes uint16) {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	var address uint32 = 0
	for bytes > 0 {
		var l uint16 = bytes
		if l > 256 {
			l = 256
		}
		sendByte(port, 'r')
		sendUint32(port, address)
		sendUint16(port, uint16(l))
		reply := read(port, int(l))
		//		fmt.Printf("%s\n", hex.Dump(reply))
		_, err = file.Write(reply)
		if err != nil {
			log.Fatal(err)
		}
		address += uint32(l)
		bytes -= l
	}
}

//
//
//
func writeFlash(port io.ReadWriteCloser, filename string) {
	var address uint32 = 0
	buf := make([]byte, 256)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	st, _ := file.Stat()
	fmt.Printf("Uploading %d bytes to flash\n", st.Size())
	for {
		l, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
			break
		}

		sendByte(port, 'w')
		sendUint32(port, address)
		sendUint16(port, uint16(l))
		for i := 0; i < l; i++ {
			sendByte(port, buf[i])
		}
		time.Sleep(3 * time.Millisecond)
		reply := read(port, 1)
		replyLen := int(reply[0])
		status := read(port, replyLen)
		if status[0] != 'O' {
			log.Fatalln("Didn't receive expected result from writeFlash")
		}
		address += uint32(l)
	}
}

//
//
//
func testFlash(port io.ReadWriteCloser, filename string, bytes uint16) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	st, _ := file.Stat()
	l := st.Size()
	var address uint32 = 0

	h := crc32.NewIEEE()
	io.Copy(h, file)
	filehash := h.Sum32()

	sendByte(port, 't')
	sendUint32(port, address)
	sendUint16(port, uint16(l))
	reply := read(port, 1)
	replyLen := int(reply[0])
	hash := read(port, replyLen)
	flashhash := uint32(hash[0])<<24 + uint32(hash[1])<<16 + uint32(hash[2])<<8 + uint32(hash[3])
	if flashhash == filehash {
		fmt.Println("VERIFIED")
	} else {
		log.Fatalln("VERIFICATION FAILED")
	}
}

func sendByte(port io.ReadWriteCloser, value byte) {
	outb := []byte{value}
	cnt, err := port.Write(outb)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}
	if cnt != 1 {
		log.Fatalf("port.Write didn't return 1 bytes written, got %d trying to send %d", cnt, value)
	}
}

func read(port io.ReadWriteCloser, length int) []byte {
	inb := []byte{0}
	outb := make([]byte, 0)
	for i := 0; i < length; i++ {
		ni, err := port.Read(inb)
		if err != nil {
			log.Fatalf("port.Read: %v", err)
		}
		if ni != 1 {
			log.Fatalf("port.Read didn't read 1 byte")
		}
		outb = append(outb, inb[0])
	}
	return outb
}

func sendUint32(port io.ReadWriteCloser, v uint32) {
	sendByte(port, byte((v>>24)&0xFF))
	sendByte(port, byte((v>>16)&0xFF))
	sendByte(port, byte((v>>8)&0xFF))
	sendByte(port, byte((v>>0)&0xFF))
}

func sendUint16(port io.ReadWriteCloser, v uint16) {
	sendByte(port, byte((v>>8)&0xFF))
	sendByte(port, byte((v>>0)&0xFF))
}
