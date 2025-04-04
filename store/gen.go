package store

import (
	"github.com/twmb/murmur3"
	"slices"
)

const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const RADIX = 62

func GenerateDigest(url string) string {
	hash := murmur3.StringSum64(url)
	return encodeToString(hash)
}

func encodeToString(num uint64) string {
	var bytes []byte
	for num > 0 {
		bytes = append(bytes, chars[num%RADIX])
		num /= RADIX
	}
	slices.Reverse(bytes)
	return string(bytes)
}
