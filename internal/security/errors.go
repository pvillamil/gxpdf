package security

import "errors"

var (
	// ErrMissingFileID is returned when FileID is not provided for encryption.
	ErrMissingFileID = errors.New("file ID is required for encryption")

	// ErrInvalidKeyLength is returned when an unsupported key length is used.
	ErrInvalidKeyLength = errors.New("key length must be 40, 128, or 256 bits")

	// ErrInvalidPassword is returned when password verification fails.
	ErrInvalidPassword = errors.New("invalid password")

	// ErrPasswordRequired is returned when a password is needed to open an encrypted PDF.
	ErrPasswordRequired = errors.New("password required for encrypted PDF")

	// ErrUnsupportedVersion is returned when encryption version is not supported.
	ErrUnsupportedVersion = errors.New("unsupported encryption version")

	// ErrInvalidPadding is returned when PKCS#7 padding is invalid.
	ErrInvalidPadding = errors.New("invalid PKCS#7 padding")

	// ErrDataTooShort is returned when encrypted data is shorter than expected.
	ErrDataTooShort = errors.New("encrypted data too short")
)
