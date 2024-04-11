package crypto

import "rs-go-server/repo"

type mockISAACCipher struct{}

func NewMockISAACCipher(seed []uint32) repo.Cipher {
	return &mockISAACCipher{}
}

func (c *mockISAACCipher) Next() uint32 {
	return 0
}
