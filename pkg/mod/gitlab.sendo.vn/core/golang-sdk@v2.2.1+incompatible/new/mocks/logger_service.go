// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import logger "gitlab.sendo.vn/core/golang-sdk/new/logger"
import mock "github.com/stretchr/testify/mock"

// LoggerService is an autogenerated mock type for the LoggerService type
type LoggerService struct {
	mock.Mock
}

// GetLogger provides a mock function with given fields: prefix
func (_m *LoggerService) GetLogger(prefix string) logger.Logger {
	ret := _m.Called(prefix)

	var r0 logger.Logger
	if rf, ok := ret.Get(0).(func(string) logger.Logger); ok {
		r0 = rf(prefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(logger.Logger)
		}
	}

	return r0
}
