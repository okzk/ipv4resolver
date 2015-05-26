package ipv4resolver

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
*/
import "C"

import (
	"net"
	"syscall"
	"unsafe"
)

func ResolveHost(name string) (addrs []net.IP, cname string, err error) {

	var res *C.struct_addrinfo
	var hints C.struct_addrinfo

	hints.ai_flags = C.AI_CANONNAME
	hints.ai_family = C.AF_INET
	hints.ai_socktype = C.SOCK_STREAM

	h := C.CString(name)
	defer C.free(unsafe.Pointer(h))
	gerrno, err := C.getaddrinfo(h, nil, &hints, &res)
	if gerrno != 0 {
		var str string
		if gerrno == C.EAI_NONAME {
			str = "no such host"
		} else if gerrno == C.EAI_SYSTEM {
			if err == nil {
				// err should not be nil, but sometimes getaddrinfo returns
				// gerrno == C.EAI_SYSTEM with err == nil on Linux.
				// The report claims that it happens when we have too many
				// open files, so use syscall.EMFILE (too many open files in system).
				// Most system calls would return ENFILE (too many open files),
				// so at the least EMFILE should be easy to recognize if this
				// comes up again. golang.org/issue/6232.
				err = syscall.EMFILE
			}
			str = err.Error()
		} else {
			str = C.GoString(C.gai_strerror(gerrno))
		}
		return nil, "", &net.DNSError{Err: str, Name: name}
	}
	defer C.freeaddrinfo(res)
	if res != nil {
		cname = C.GoString(res.ai_canonname)
		if cname == "" {
			cname = name
		}
		if len(cname) > 0 && cname[len(cname)-1] != '.' {
			cname += "."
		}
	}
	for r := res; r != nil; r = r.ai_next {
		// We only asked for SOCK_STREAM, but check anyhow.
		if r.ai_socktype != C.SOCK_STREAM {
			continue
		}
		switch r.ai_family {
		default:
			continue
		case C.AF_INET:
			sa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(r.ai_addr))
			addrs = append(addrs, copyIP(sa.Addr[:]))
		}
	}
	return addrs, cname, nil
}

func copyIP(x net.IP) net.IP {
	if len(x) < 16 {
		return x.To16()
	}
	y := make(net.IP, len(x))
	copy(y, x)
	return y
}
