package demo_test

import "errors"

func (deployment *deployment) releaseAddress(address string) error {
	listener, exists := deployment.reserved[address]
	if !exists {
		return nil
	}

	delete(deployment.reserved, address)

	return listener.Close()
}

func (deployment *deployment) releaseAddresses() error {
	var failures []error
	for address := range deployment.reserved {
		failures = append(failures, deployment.releaseAddress(address))
	}

	return errors.Join(failures...)
}
