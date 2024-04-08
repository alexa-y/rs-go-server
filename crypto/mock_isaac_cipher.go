package crypto

type mockISAACCipher struct{}

func NewMockISAACCipher(seed []uint32) Cipher {
	return &mockISAACCipher{}
}

func (c *mockISAACCipher) Next() uint32 {
	return 0
}
