package bananaphone

import (
	"encoding/binary"
	"errors"
	"unicode/utf16"

	"github.com/awgh/rawreader"
)

// Image - contains information about a loaded image
type Image struct {
	BaseAddr uint64
	Size     uint64
}

// InMemLoads - returns a map of image full name to the related Image struct
func InMemLoads() (map[string]Image, error) {
	retMap := make(map[string]Image)
	peb := GetPEB()
	rawr := rawreader.New(peb, 0x30)
	ptr := make([]byte, 8)
	xerr := errors.New("raw read underflow")

	//PEB->LDR
	n, err := rawr.ReadAt(ptr, 0x18)
	if n != 8 {
		return nil, xerr
	}
	if err != nil {
		return nil, err
	}
	ptrLdr := uintptr(binary.LittleEndian.Uint64(ptr))

	//LDR->InMemoryOrderModuleList
	rawr = rawreader.New(ptrLdr, 0x30)
	n, err = rawr.ReadAt(ptr, 0x20)
	if n != 8 {
		return nil, xerr
	}
	if err != nil {
		return nil, err
	}
	ptrList := uintptr(binary.LittleEndian.Uint64(ptr))
	ptrNext := ptrList
	for {
		//Read an element
		rawr = rawreader.New(ptrNext, 0x120)
		n, err = rawr.ReadAt(ptr, 0x0)
		if n != 8 {
			return nil, xerr
		}
		if err != nil {
			return nil, err
		}
		ptrNext = uintptr(binary.LittleEndian.Uint64(ptr))
		if ptrNext == ptrList {
			break
		}
		n, err = rawr.ReadAt(ptr, 0x30)
		if n != 8 {
			return nil, xerr
		}
		if err != nil {
			return nil, err
		}
		dllBase := uintptr(binary.LittleEndian.Uint64(ptr))
		n, err = rawr.ReadAt(ptr, 0x40)
		if n != 8 {
			return nil, xerr
		}
		if err != nil {
			return nil, err
		}
		sizeOfImage := binary.LittleEndian.Uint64(ptr)

		namelen := make([]byte, 2)
		n, err = rawr.ReadAt(namelen, 0x48)
		if n != 2 {
			return nil, xerr
		}
		if err != nil {
			return nil, err
		}
		fullNameLen := int(binary.LittleEndian.Uint16(namelen))

		n, err = rawr.ReadAt(ptr, 0x48+4+4) // skip 2 ushorts (len and maxlen) and align to 8
		if n != 8 {
			return nil, xerr
		}
		if err != nil {
			return nil, err
		}
		ptrFullName := uintptr(binary.LittleEndian.Uint64(ptr))

		rawr = rawreader.New(ptrFullName, fullNameLen)
		name := make([]byte, fullNameLen)
		n, err = rawr.ReadAt(name, 0x0)
		if n != fullNameLen {
			return nil, xerr
		}
		if err != nil {
			return nil, err
		}
		nameW := make([]uint16, fullNameLen/2)
		for i := 0; i < fullNameLen; i = i + 2 {
			nameW[i/2] = binary.LittleEndian.Uint16(name[i:])
		}
		fullName := string(utf16.Decode(nameW))
		retMap[fullName] = Image{BaseAddr: uint64(dllBase), Size: sizeOfImage}
	}
	return retMap, nil
}
