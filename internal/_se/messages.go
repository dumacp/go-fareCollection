package se

type MsgClose struct{}
type MsgOpen struct{}
type MsgAck struct {
	Error string
}
type MsgEncryptRequest struct {
	MsgID    string
	KeySlot  int
	DevInput []byte
	Data     []byte
	IV       []byte
}
type MsgEncryptResponse struct {
	MsgID  string
	Cipher []byte
}
type MsgDecryptRequest struct {
	MsgID    string
	KeySlot  int
	DevInput []byte
	Data     []byte
	IV       []byte
}
type MsgDecryptResponse struct {
	MsgID string
	Plain []byte
}
type MsgDumpSecretKeyRequest struct {
	KeySlot int
}
type MsgDumpSecretKeyResponse struct {
	Data []byte
}
type MsgCreateEntryKeyRequest struct {
	ContainKeysID []string
	EntryKeyID    string
	ChangeKeySlot int
	Keys          []byte
	AuthHost      bool
	OfflineKey    bool
	PICCKey       bool
	Alg           string
}
type MsgEntryKeyRequest struct {
	EntryKeyID string
	PersoKeyID string
	TargetSlot int
}
type MsgEntryKeyResponse struct {
	Error []byte
	Data  []byte
}
type MsgCreateKeyRequest struct {
	KeySlot int
	Alg     string
}
type MsgImportKeyRequest struct {
	KeySlot int
	Data    []byte
	Alg     string
}
type MsgEnableKeysRequest struct{}
type MsgEnableKeysResponse struct {
	Data []int
}
