// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/harrybrwn/db (interfaces: Rows,Pingable)
//
// Generated by this command:
//
//	mockgen -package=mockrows -destination ./mockrows/rows.go . Rows,Pingable
//

// Package mockrows is a generated GoMock package.
package mockrows

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockRows is a mock of Rows interface.
type MockRows struct {
	ctrl     *gomock.Controller
	recorder *MockRowsMockRecorder
	isgomock struct{}
}

// MockRowsMockRecorder is the mock recorder for MockRows.
type MockRowsMockRecorder struct {
	mock *MockRows
}

// NewMockRows creates a new mock instance.
func NewMockRows(ctrl *gomock.Controller) *MockRows {
	mock := &MockRows{ctrl: ctrl}
	mock.recorder = &MockRowsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRows) EXPECT() *MockRowsMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockRows) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockRowsMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockRows)(nil).Close))
}

// Err mocks base method.
func (m *MockRows) Err() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Err")
	ret0, _ := ret[0].(error)
	return ret0
}

// Err indicates an expected call of Err.
func (mr *MockRowsMockRecorder) Err() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Err", reflect.TypeOf((*MockRows)(nil).Err))
}

// Next mocks base method.
func (m *MockRows) Next() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Next")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Next indicates an expected call of Next.
func (mr *MockRowsMockRecorder) Next() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*MockRows)(nil).Next))
}

// Scan mocks base method.
func (m *MockRows) Scan(arg0 ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan.
func (mr *MockRowsMockRecorder) Scan(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*MockRows)(nil).Scan), arg0...)
}

// MockPingable is a mock of Pingable interface.
type MockPingable struct {
	ctrl     *gomock.Controller
	recorder *MockPingableMockRecorder
	isgomock struct{}
}

// MockPingableMockRecorder is the mock recorder for MockPingable.
type MockPingableMockRecorder struct {
	mock *MockPingable
}

// NewMockPingable creates a new mock instance.
func NewMockPingable(ctrl *gomock.Controller) *MockPingable {
	mock := &MockPingable{ctrl: ctrl}
	mock.recorder = &MockPingableMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPingable) EXPECT() *MockPingableMockRecorder {
	return m.recorder
}

// Ping mocks base method.
func (m *MockPingable) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockPingableMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockPingable)(nil).Ping))
}

// PingContext mocks base method.
func (m *MockPingable) PingContext(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingContext", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PingContext indicates an expected call of PingContext.
func (mr *MockPingableMockRecorder) PingContext(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingContext", reflect.TypeOf((*MockPingable)(nil).PingContext), arg0)
}
