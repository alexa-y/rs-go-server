package crypto

type Cipher interface {
	Next() uint32
}
