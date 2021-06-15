package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func main() {

	// potrebujeme index rozhrania
	ethIndex, err := strconv.Atoi(os.Args[1])
	if err != nil {
		doError("Neviem precitat index rozhrania")
	}

	// najprv si vytvorime socket typu Netlink
	sock, err := unix.Socket(
		unix.AF_NETLINK,
		unix.SOCK_RAW,
		unix.NETLINK_ROUTE,
	)

	if err != nil {
		doError("Nedokazal som vyrobit socket")
	}

	defer unix.Close(sock)

	// Potrebujeme este zavolat Bind() pre socket
	err = unix.Bind(sock, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
	})

	if err != nil {
		doError("Nedokazal som pripravit socket")
	}

	// vytvorime hlavicku spravy
	length := unix.SizeofNlMsghdr + unix.SizeofIfInfomsg

	header := &unix.NlMsghdr{
		Len:   uint32(length),
		Type:  uint16(unix.RTM_NEWLINK),
		Flags: uint16(unix.NLM_F_REQUEST) | uint16(unix.NLM_F_ACK),
		Seq:   1,
	}

	// ... a tzv. payload spravy
	payload := &unix.IfInfomsg{
		Family: unix.AF_UNSPEC,
		Change: unix.IFF_UP,
		Flags:  unix.IFF_UP,
		Index:  int32(ethIndex), // index rozrania ktore chceme zapnut
	}

	// spravu musime serializovat na pole bajtov.
	msg := make([]byte, length)
	copy(msg[0:unix.SizeofNlMsghdr], (*(*[unix.SizeofNlMsghdr]byte)(unsafe.Pointer(header)))[:])
	copy(msg[unix.SizeofNlMsghdr:length], (*(*[unix.SizeofIfInfomsg]byte)(unsafe.Pointer(payload)))[:])

	//spravu odosleme
	err = unix.Sendto(sock, msg, 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK})
	if err != nil {
		doError("Neviem odoslat spravu")
	}

	// teraz precitame odpoved
	var rb [1024]byte
	nr, _, err := unix.Recvfrom(sock, rb[:], 0)
	if err != nil {
		fmt.Println("Neviem nacitat odpoved")
	}

	rb2 := make([]byte, nr)
	copy(rb2, rb[:nr])

	// nacitanu odpoved potrebujeme spracovat a sparsovat
	resp, err := syscall.ParseNetlinkMessage(rb2)
	if err != nil {
		doError("Neviem sparsovat odpoved")
	}

	if resp[0].Header.Type != unix.NLMSG_ERROR {
		doError("Netlink vratil nespravnu odpoved")
	}

	// deserializacia nacitanych bajtov na msgerr
	errMsg := (*unix.NlMsgerr)(unsafe.Pointer(&resp[0].Data[0]))

	// spracovanie odpovede
	if errMsg.Error != 0 {
		doError("Netlink vratil chybu")
	}

	fmt.Println("Rozhranie je zapnute. Skontroluj cez 'ip link'")
}

func doError(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
