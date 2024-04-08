package uuid_test

import (
	"crypto/md5"
	"testing"

	gofrsuuid "github.com/gofrs/uuid"
	gluuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/uuid"
)

var (
	uuidGOOGLE gluuid.UUID
	uuidGOFRS  gofrsuuid.UUID
)

func init() {
	gluuid.EnableRandPool()
}

func FuzzGetMD5UUID(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add(gluuid.New().String())

	f.Fuzz(func(t *testing.T, data string) {
		googleMd5, err := uuid.GetMD5UUID(data)
		require.NoError(t, err)

		md5Sum := md5.Sum([]byte(data))
		gofrsMd5, err := gofrsuuid.FromBytes(md5Sum[:])
		require.NoError(t, err)
		gofrsMd5.SetVersion(gofrsuuid.V4)
		gofrsMd5.SetVariant(gofrsuuid.VariantRFC4122)

		require.Equal(t, gofrsMd5.String(), googleMd5.String())
	})
}

func Test_fastUUID(t *testing.T) {
	t.Run("test google conversion gofrs", func(t *testing.T) {
		uuidGOOGLE = gluuid.New()
		b, err := uuidGOOGLE.MarshalBinary()
		require.NoError(t, err)
		uuidGOFRS = gofrsuuid.FromBytesOrNil(b)
		require.Equal(t, uuidGOOGLE.String(), uuidGOFRS.String())
	})
}

func BenchmarkUUID(t *testing.B) {
	t.Run("google uuid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uuidGOOGLE = gluuid.New()
		}
	})

	t.Run("google uuid str gofrs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uuidGOFRS = gofrsuuid.FromStringOrNil(gluuid.New().String())
		}
	})

	t.Run("google uuid bin gofrs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b, _ := gluuid.New().MarshalBinary()
			uuidGOFRS = gofrsuuid.FromBytesOrNil(b)
		}
	})

	t.Run("gofrs uuid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uuidGOFRS = gofrsuuid.Must(gofrsuuid.NewV4())
		}
	})
}
