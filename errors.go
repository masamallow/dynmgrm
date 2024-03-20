package dynmgrm

import "errors"

var (
	ErrCollectionAlreadyContainsItem = errors.New("collection already contains item")
	ErrFailedToCast                  = errors.New("failed to cast")
	ErrDynmgrmAreNotSupported        = errors.New("dynmgrm are not supported this operation")
)
