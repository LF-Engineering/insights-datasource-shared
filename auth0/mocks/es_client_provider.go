// Code generated by mockery v2.3.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// ESClientProvider is an autogenerated mock type for the ESClientProvider type
type ESClientProvider struct {
	mock.Mock
}

// CreateDocument provides a mock function with given fields: index, documentID, body
func (_m *ESClientProvider) CreateDocument(index string, documentID string, body []byte) ([]byte, error) {
	ret := _m.Called(index, documentID, body)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, string, []byte) []byte); ok {
		r0 = rf(index, documentID, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, []byte) error); ok {
		r1 = rf(index, documentID, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateIndex provides a mock function with given fields: index, body
func (_m *ESClientProvider) CreateIndex(index string, body []byte) ([]byte, error) {
	ret := _m.Called(index, body)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, []byte) []byte); ok {
		r0 = rf(index, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []byte) error); ok {
		r1 = rf(index, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: index, query, result
func (_m *ESClientProvider) Get(index string, query map[string]interface{}, result interface{}) error {
	ret := _m.Called(index, query, result)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, map[string]interface{}, interface{}) error); ok {
		r0 = rf(index, query, result)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Search provides a mock function with given fields: index, query
func (_m *ESClientProvider) Search(index string, query map[string]interface{}) ([]byte, error) {
	ret := _m.Called(index, query)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, map[string]interface{}) []byte); ok {
		r0 = rf(index, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, map[string]interface{}) error); ok {
		r1 = rf(index, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}