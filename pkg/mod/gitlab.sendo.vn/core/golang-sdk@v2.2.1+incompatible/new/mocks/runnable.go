// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Runnable is an autogenerated mock type for the Runnable type
type Runnable struct {
	mock.Mock
}

// Configure provides a mock function with given fields:
func (_m *Runnable) Configure() error {
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
func (_m *Runnable) InitFlags() {
	_m.Called()
}

// Name provides a mock function with given fields:
func (_m *Runnable) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Run provides a mock function with given fields:
func (_m *Runnable) Run() error {
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
func (_m *Runnable) Stop() <-chan bool {
	ret := _m.Called()

	var r0 <-chan bool
	if rf, ok := ret.Get(0).(func() <-chan bool); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan bool)
		}
	}

	return r0
}
