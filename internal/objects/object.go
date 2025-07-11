package objects

type Object interface {
	Type() string
	Serialize() ([]byte, error)
	GetHash() string
}
