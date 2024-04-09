package uuid

import (
	"crypto/md5"

	"github.com/google/uuid"
)

// GetMD5UUID hashes the given string into md5 and returns it as uuid
func GetMD5UUID(str string) (uuid.UUID, error) {
	// To maintain backward compatibility, we are using md5 hash of the string
	// We are mimicking github.com/gofrs/uuid behavior:
	//
	// md5Sum := md5.Sum([]byte(str))
	// u, err := uuid.FromBytes(md5Sum[:])

	// u.SetVersion(uuid.V4)
	// u.SetVariant(uuid.VariantRFC4122)

	// google/uuid doesn't allow us to modify the version and variant,
	// so we are doing it manually, using gofrs/uuid library implementation.
	md5Sum := md5.Sum([]byte(str)) // skipcq: GO-S1023
	// SetVariant: VariantRFC4122
	md5Sum[8] = md5Sum[8]&(0xff>>2) | (0x02 << 6)
	// SetVersion: Version 4
	version := byte(4)
	md5Sum[6] = (md5Sum[6] & 0x0f) | (version << 4)

	return uuid.FromBytes(md5Sum[:])
}
