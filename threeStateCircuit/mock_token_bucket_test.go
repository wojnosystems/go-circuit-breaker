package threeStateCircuit

type tokenBucketAlwaysFails struct{}

func (b *tokenBucketAlwaysFails) Allowed(_ uint64) bool {
	return false
}

type tokenBucketAlwaysSucceeds struct{}

func (s *tokenBucketAlwaysSucceeds) Allowed(_ uint64) bool {
	return true
}
