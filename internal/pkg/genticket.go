package pkg

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateTicketID(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range length {
		// Mengambil angka acak yang aman secara kriptografi
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		// Mengambil karakter dari charset berdasarkan angka acak
		result[i] = charset[num.Int64()]
	}

	// Menambahkan prefix "T-" di depan
	return fmt.Sprintf("T-%s", string(result)), nil
}
