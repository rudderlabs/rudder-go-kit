package encrypt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

/*
BenchmarkEncryptDecrypt/AESCFB_AES128
BenchmarkEncryptDecrypt/AESCFB_AES128-12         	  573664	      2027 ns/op	    1600 B/op	      15 allocs/op
BenchmarkEncryptDecrypt/AESCFB_AES192
BenchmarkEncryptDecrypt/AESCFB_AES192-12         	  557448	      2079 ns/op	    1728 B/op	      15 allocs/op
BenchmarkEncryptDecrypt/AESCFB_AES256
BenchmarkEncryptDecrypt/AESCFB_AES256-12         	  531597	      2210 ns/op	    1856 B/op	      15 allocs/op
BenchmarkEncryptDecrypt/AESGCM_AES128
BenchmarkEncryptDecrypt/AESGCM_AES128-12         	  723628	      1626 ns/op	    2480 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/AESGCM_AES192
BenchmarkEncryptDecrypt/AESGCM_AES192-12         	  679282	      1641 ns/op	    2608 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/AESGCM_AES256
BenchmarkEncryptDecrypt/AESGCM_AES256-12         	  687800	      1720 ns/op	    2736 B/op	      13 allocs/op
*/
func BenchmarkEncryptDecrypt(b *testing.B) {
	tests := []struct {
		name  string
		algo  EncryptionAlgorithm
		level EncryptionLevel
	}{
		{"AESCFB_AES128", EncryptionAlgoAESCFB, EncryptionLevelAES128},
		{"AESCFB_AES192", EncryptionAlgoAESCFB, EncryptionLevelAES192},
		{"AESCFB_AES256", EncryptionAlgoAESCFB, EncryptionLevelAES256},
		{"AESGCM_AES128", EncryptionAlgoAESGCM, EncryptionLevelAES128},
		{"AESGCM_AES192", EncryptionAlgoAESGCM, EncryptionLevelAES192},
		{"AESGCM_AES256", EncryptionAlgoAESGCM, EncryptionLevelAES256},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			encrypter, err := New(tt.algo, tt.level)
			require.NoError(b, err)

			key, err := generateRandomString(int(tt.level / 8))
			require.NoError(b, err)

			plaintext := loremIpsumDolor

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ciphertext, err := encrypter.Encrypt(plaintext, key)
				require.NoError(b, err)

				_, err = encrypter.Decrypt(ciphertext, key)
				require.NoError(b, err)
			}
		})
	}
}
