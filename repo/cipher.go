package repo

type Cipher interface {
	Next() uint32
}
