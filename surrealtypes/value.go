package surrealtypes

type SurrealMarshalable interface {
	MarshalSurreal() ([]byte, error)
	//UnmarshalSurreal([]byte) error
}
