package encrypt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

/*
BenchmarkEncryptDecrypt
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES128
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES128-12         	  803842	      1444 ns/op	    1616 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES192
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES192-12         	  805350	      1443 ns/op	    1744 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES256
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES256-12         	  744871	      1516 ns/op	    1872 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/LARGE_AESGCM_AES128
BenchmarkEncryptDecrypt/LARGE_AESGCM_AES128-12         	    1900	    614516 ns/op	 4204053 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/LARGE_AESGCM_AES192
BenchmarkEncryptDecrypt/LARGE_AESGCM_AES192-12         	    1755	    672776 ns/op	 4204180 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/LARGE_AESGCM_AES256
BenchmarkEncryptDecrypt/LARGE_AESGCM_AES256-12         	    1624	    723403 ns/op	 4204308 B/op	      13 allocs/op
*/
func BenchmarkEncryptDecrypt(b *testing.B) {
	tests := []struct {
		payload []byte
		name    string
		algo    EncryptionAlgorithm
		level   EncryptionLevel
	}{
		{[]byte("small payload"), "SMALL_AESGCM_AES128", EncryptionAlgoAESGCM, EncryptionLevelAES128},
		{[]byte("small payload"), "SMALL_AESGCM_AES192", EncryptionAlgoAESGCM, EncryptionLevelAES192},
		{[]byte("small payload"), "SMALL_AESGCM_AES256", EncryptionAlgoAESGCM, EncryptionLevelAES256},
		{make([]byte, 2*1024*1024), "LARGE_AESGCM_AES128", EncryptionAlgoAESGCM, EncryptionLevelAES128},
		{make([]byte, 2*1024*1024), "LARGE_AESGCM_AES192", EncryptionAlgoAESGCM, EncryptionLevelAES192},
		{make([]byte, 2*1024*1024), "LARGE_AESGCM_AES256", EncryptionAlgoAESGCM, EncryptionLevelAES256},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			encrypter, err := New(tt.algo, tt.level)
			require.NoError(b, err)

			key, err := generateRandomString(int(tt.level / 8))
			require.NoError(b, err)

			plaintext := tt.payload

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
