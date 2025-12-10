package jsonrs

import "io"

// switcher is a JSON implementation that switches between different JSON implementations, based on the configuration.
type switcher struct {
	marshallerFn      func() string
	unmarshallerFn    func() string
	validatorFn       func() string
	marshallerImpls   map[string]Marshaller
	unmarshallerImpls map[string]Unmarshaller
	validatorImpls    map[string]Validator
}

func (s *switcher) Marshal(v any) ([]byte, error) {
	return s.marshaller().Marshal(v)
}

func (s *switcher) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return s.marshaller().MarshalIndent(v, prefix, indent)
}

func (s *switcher) Unmarshal(data []byte, v any) error {
	return s.unmarshaller().Unmarshal(data, v)
}

func (s *switcher) MarshalToString(v any) (string, error) {
	return s.marshaller().MarshalToString(v)
}

func (s *switcher) NewDecoder(r io.Reader) Decoder {
	return s.unmarshaller().NewDecoder(r)
}

func (s *switcher) NewEncoder(w io.Writer) Encoder {
	return s.marshaller().NewEncoder(w)
}

func (s *switcher) Valid(data []byte) bool {
	return s.validator().Valid(data)
}

func (s *switcher) marshaller() Marshaller {
	if impl, ok := s.marshallerImpls[s.marshallerFn()]; ok {
		return impl
	}
	return s.marshallerImpls[DefaultLib]
}

func (s *switcher) unmarshaller() Unmarshaller {
	if impl, ok := s.unmarshallerImpls[s.unmarshallerFn()]; ok {
		return impl
	}
	return s.unmarshallerImpls[DefaultLib]
}

func (s *switcher) validator() Validator {
	if impl, ok := s.validatorImpls[s.validatorFn()]; ok {
		return impl
	}
	return s.validatorImpls[DefaultLib]
}
