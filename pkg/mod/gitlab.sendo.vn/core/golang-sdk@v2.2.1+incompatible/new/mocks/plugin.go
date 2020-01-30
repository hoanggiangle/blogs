// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import generator "gitlab.sendo.vn/core/golang-sdk/new/protoc-gen-sendo/generator"
import mock "github.com/stretchr/testify/mock"

// Plugin is an autogenerated mock type for the Plugin type
type Plugin struct {
	mock.Mock
}

// Generate provides a mock function with given fields: file
func (_m *Plugin) Generate(file *generator.FileDescriptor) {
	_m.Called(file)
}

// GenerateImports provides a mock function with given fields: file
func (_m *Plugin) GenerateImports(file *generator.FileDescriptor) {
	_m.Called(file)
}

// Init provides a mock function with given fields: g
func (_m *Plugin) Init(g *generator.Generator) {
	_m.Called(g)
}

// Name provides a mock function with given fields:
func (_m *Plugin) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
