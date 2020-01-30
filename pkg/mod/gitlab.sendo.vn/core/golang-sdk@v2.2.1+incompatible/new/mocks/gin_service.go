// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// GinService is an autogenerated mock type for the GinService type
type GinService struct {
	mock.Mock
}

// isGinService provides a mock function with given fields:
func (_m *GinService) isGinService() {
	_m.Called()
}

// Cleanup provides a mock function with given fields:
func (_m *GinService) Cleanup() {
	_m.Called()
}

// Configure provides a mock function with given fields:
func (_m *GinService) Configure() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InitFlags provides a mock function with given fields:
func (_m *GinService) InitFlags() {
	_m.Called()
}

// Port provides a mock function with given fields:
func (_m *GinService) Port() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Run provides a mock function with given fields:
func (_m *GinService) Run() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stop provides a mock function with given fields:
func (_m *GinService) Stop() {
	_m.Called()
}
