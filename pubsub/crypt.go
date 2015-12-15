package pubsub

import (
	"github.com/dgryski/dkeyczar"
	"github.com/golang/protobuf/proto"
)

// Encrypter is a helper struct to
// the repetitive encrypt work used in publisher
// implementations.
type Encrypter struct {
	e dkeyczar.Encrypter
}

// NewCrypter will read in the given key file and
// use dkeyczar to set up an encyrpter and a decrypter.
func NewEncrypter(keyFile string) (*Encrypter, error) {
	var err error
	c := &Encrypter{}
	reader := dkeyczar.NewFileReader(keyFile)
	if c.e, err = dkeyczar.NewEncrypter(reader); err != nil {
		return c, err
	}

	c.e.SetEncoding(dkeyczar.NO_ENCODING)
	return c, nil
}

// Encrypt will marshal the given protobuf message, encrypt it
// and put it in a SessionMessage.
func (c *Encrypter) Encrypt(m proto.Message) (*SessionMessage, error) {
	mb, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	enc, material, err := dkeyczar.NewSessionEncrypter(c.e)
	if err != nil {
		return nil, err
	}

	var eb string
	if eb, err = enc.Encrypt(mb); err != nil {
		return nil, err
	}

	sm := &SessionMessage{
		EncryptedMaterial: []byte(material),
		EncryptedMessages: [][]byte{[]byte(eb)},
	}
	return sm, nil
}

// Crypter is a helper struct to encapsulate all
// the repetitive crypto work used in pubsub
// implementations.
type Decrypter struct {
	c dkeyczar.Crypter
}

// NewCrypter will read in the given key file and
// use dkeyczar to set up an encyrpter and a decrypter.
func NewDecrypter(keyFile string) (*Decrypter, error) {
	var err error
	c := &Decrypter{}
	reader := dkeyczar.NewFileReader(keyFile)
	if c.c, err = dkeyczar.NewCrypter(reader); err != nil {
		return c, err
	}

	c.c.SetEncoding(dkeyczar.NO_ENCODING)
	return c, nil
}

// Decrypt will attempt to proto Unmarshal the given
// byte array into a SessionMessage, decrypt the payload
// and return it's []byte.
func (c *Decrypter) Decrypt(msg []byte) ([]byte, error) {
	var (
		sm  SessionMessage
		out []byte
	)
	err := proto.Unmarshal(msg, &sm)
	if err != nil {
		return out, err
	}

	sd, err := dkeyczar.NewSessionDecrypter(c.c, string(sm.EncryptedMaterial))
	if err != nil {
		return out, err
	}

	sd.SetEncoding(dkeyczar.NO_ENCODING)
	buffer, err := sd.Decrypt(string(sm.EncryptedMessages[0]))
	if err != nil {
		return out, err
	}

	return buffer, nil
}
