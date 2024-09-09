package encrypt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

/*
BenchmarkEncryptDecrypt
BenchmarkEncryptDecrypt/SMALL_AESCFB_AES128
BenchmarkEncryptDecrypt/SMALL_AESCFB_AES128-12         	  941356	      1252 ns/op	    1152 B/op	      15 allocs/op
BenchmarkEncryptDecrypt/SMALL_AESCFB_AES192
BenchmarkEncryptDecrypt/SMALL_AESCFB_AES192-12         	  907440	      1249 ns/op	    1280 B/op	      15 allocs/op
BenchmarkEncryptDecrypt/SMALL_AESCFB_AES256
BenchmarkEncryptDecrypt/SMALL_AESCFB_AES256-12         	  881331	      1429 ns/op	    1408 B/op	      15 allocs/op
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES128
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES128-12         	  803842	      1444 ns/op	    1616 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES192
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES192-12         	  805350	      1443 ns/op	    1744 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES256
BenchmarkEncryptDecrypt/SMALL_AESGCM_AES256-12         	  744871	      1516 ns/op	    1872 B/op	      13 allocs/op
BenchmarkEncryptDecrypt/LARGE_AESCFB_AES128
BenchmarkEncryptDecrypt/LARGE_AESCFB_AES128-12         	     315	   3674803 ns/op	 2106469 B/op	      15 allocs/op
BenchmarkEncryptDecrypt/LARGE_AESCFB_AES192
BenchmarkEncryptDecrypt/LARGE_AESCFB_AES192-12         	     285	   4111928 ns/op	 2106595 B/op	      15 allocs/op
BenchmarkEncryptDecrypt/LARGE_AESCFB_AES256
BenchmarkEncryptDecrypt/LARGE_AESCFB_AES256-12         	     284	   4189400 ns/op	 2106725 B/op	      15 allocs/op
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
		{[]byte("small payload"), "SMALL_AESCFB_AES128", EncryptionAlgoAESCFB, EncryptionLevelAES128},
		{[]byte("small payload"), "SMALL_AESCFB_AES192", EncryptionAlgoAESCFB, EncryptionLevelAES192},
		{[]byte("small payload"), "SMALL_AESCFB_AES256", EncryptionAlgoAESCFB, EncryptionLevelAES256},
		{[]byte("small payload"), "SMALL_AESGCM_AES128", EncryptionAlgoAESGCM, EncryptionLevelAES128},
		{[]byte("small payload"), "SMALL_AESGCM_AES192", EncryptionAlgoAESGCM, EncryptionLevelAES192},
		{[]byte("small payload"), "SMALL_AESGCM_AES256", EncryptionAlgoAESGCM, EncryptionLevelAES256},
		{make([]byte, 2*1024*1024), "LARGE_AESCFB_AES128", EncryptionAlgoAESCFB, EncryptionLevelAES128},
		{make([]byte, 2*1024*1024), "LARGE_AESCFB_AES192", EncryptionAlgoAESCFB, EncryptionLevelAES192},
		{make([]byte, 2*1024*1024), "LARGE_AESCFB_AES256", EncryptionAlgoAESCFB, EncryptionLevelAES256},
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
