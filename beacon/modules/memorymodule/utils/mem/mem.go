package mem

import (
	"unsafe"
)

type usp = unsafe.Pointer
type size_t = int //always, we pass into functions with the result of len(slice),  which is int

//just as C's memcpy, make sure the dest and src capcity >= len
//Memcpy can not handle dest and src overlap condition
func Memcpy(dest, src unsafe.Pointer, len size_t) unsafe.Pointer {

	cnt := len >> 3
	var i size_t = 0
	for i = 0; i < cnt; i++ {
		var pdest *uint64 = (*uint64)(usp(uintptr(dest) + uintptr(8*i)))
		var psrc *uint64 = (*uint64)(usp(uintptr(src) + uintptr(8*i)))
		*pdest = *psrc
	}
	left := len & 7
	for i = 0; i < left; i++ {
		var pdest *uint8 = (*uint8)(usp(uintptr(dest) + uintptr(8*cnt+i)))
		var psrc *uint8 = (*uint8)(usp(uintptr(src) + uintptr(8*cnt+i)))

		*pdest = *psrc
	}
	return dest
}

func memcpyH(dest, src unsafe.Pointer, len size_t) unsafe.Pointer {

	cnt := len >> 3
	var i size_t = 0
	for i = 1; i <= cnt; i++ {
		var pdest *uint64 = (*uint64)(usp(uintptr(dest) + uintptr(len-8*i)))
		var psrc *uint64 = (*uint64)(usp(uintptr(src) + uintptr(len-8*i)))
		*pdest = *psrc
	}
	left := len & 7
	for i = 1; i <= left; i++ {
		var pdest *uint8 = (*uint8)(usp(uintptr(dest) + uintptr(len-8*cnt-i)))
		var psrc *uint8 = (*uint8)(usp(uintptr(src) + uintptr(len-8*cnt-i)))

		*pdest = *psrc
	}
	return dest
}

//just AS C's memmove, make sure the dest and src capcity >= len
//Memmove can  handle dest and src overlap condition
func Memmove(dest, src unsafe.Pointer, len size_t) unsafe.Pointer {
	if dest == src {
		return dest
	}
	if uintptr(src) < uintptr(dest) && uintptr(dest) <= uintptr(src)+uintptr(len) {
		return memcpyH(dest, src, len)
	}

	return Memcpy(dest, src, len)
}

func Memcmp(dest, src unsafe.Pointer, len size_t) int {

	cnt := len >> 3
	var i size_t = 0
	for i = 0; i < cnt; i++ {
		var pdest *uint64 = (*uint64)(usp(uintptr(dest) + uintptr(8*i)))
		var psrc *uint64 = (*uint64)(usp(uintptr(src) + uintptr(8*i)))
		switch {
		case *pdest < *psrc:
			return -1
		case *pdest > *psrc:
			return 1
		default:
		}
	}

	left := len & 7
	for i = 0; i < left; i++ {
		var pdest *uint8 = (*uint8)(usp(uintptr(dest) + uintptr(8*cnt+i)))
		var psrc *uint8 = (*uint8)(usp(uintptr(src) + uintptr(8*cnt+i)))
		switch {
		case *pdest < *psrc:
			return -1
		case *pdest > *psrc:
			return 1
		default:
		}
	}
	return 0
}

func Memset(dest unsafe.Pointer, ch int8, len size_t) unsafe.Pointer {

	left := len & 7
	cnt := len >> 3
	if cnt > 0 {
		left += 8
	}
	var i size_t = 0
	for i = 0; i < left; i++ {
		var pdest *int8 = (*int8)(usp(uintptr(dest) + uintptr(i)))
		*pdest = ch
	}
	if cnt < 2 {
		return dest
	}
	var pfirst *int64 = (*int64)(dest)

	for i = 0; i < cnt-1; i++ {
		var pdest *int64 = (*int64)(usp(uintptr(dest) + uintptr(left+8*i)))
		*pdest = *pfirst
	}

	return dest
}
