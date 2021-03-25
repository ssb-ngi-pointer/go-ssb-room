// Code generated by counterfeiter. DO NOT EDIT.
package mockdb

import (
	"context"
	"sync"

	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb"
)

type FakeAuthWithSSBService struct {
	CheckTokenStub        func(context.Context, string) (int64, error)
	checkTokenMutex       sync.RWMutex
	checkTokenArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	checkTokenReturns struct {
		result1 int64
		result2 error
	}
	checkTokenReturnsOnCall map[int]struct {
		result1 int64
		result2 error
	}
	CreateTokenStub        func(context.Context, int64) (string, error)
	createTokenMutex       sync.RWMutex
	createTokenArgsForCall []struct {
		arg1 context.Context
		arg2 int64
	}
	createTokenReturns struct {
		result1 string
		result2 error
	}
	createTokenReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	RemoveTokenStub        func(context.Context, string) error
	removeTokenMutex       sync.RWMutex
	removeTokenArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	removeTokenReturns struct {
		result1 error
	}
	removeTokenReturnsOnCall map[int]struct {
		result1 error
	}
	WipeTokensForMemberStub        func(context.Context, int64) error
	wipeTokensForMemberMutex       sync.RWMutex
	wipeTokensForMemberArgsForCall []struct {
		arg1 context.Context
		arg2 int64
	}
	wipeTokensForMemberReturns struct {
		result1 error
	}
	wipeTokensForMemberReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeAuthWithSSBService) CheckToken(arg1 context.Context, arg2 string) (int64, error) {
	fake.checkTokenMutex.Lock()
	ret, specificReturn := fake.checkTokenReturnsOnCall[len(fake.checkTokenArgsForCall)]
	fake.checkTokenArgsForCall = append(fake.checkTokenArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.CheckTokenStub
	fakeReturns := fake.checkTokenReturns
	fake.recordInvocation("CheckToken", []interface{}{arg1, arg2})
	fake.checkTokenMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeAuthWithSSBService) CheckTokenCallCount() int {
	fake.checkTokenMutex.RLock()
	defer fake.checkTokenMutex.RUnlock()
	return len(fake.checkTokenArgsForCall)
}

func (fake *FakeAuthWithSSBService) CheckTokenCalls(stub func(context.Context, string) (int64, error)) {
	fake.checkTokenMutex.Lock()
	defer fake.checkTokenMutex.Unlock()
	fake.CheckTokenStub = stub
}

func (fake *FakeAuthWithSSBService) CheckTokenArgsForCall(i int) (context.Context, string) {
	fake.checkTokenMutex.RLock()
	defer fake.checkTokenMutex.RUnlock()
	argsForCall := fake.checkTokenArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeAuthWithSSBService) CheckTokenReturns(result1 int64, result2 error) {
	fake.checkTokenMutex.Lock()
	defer fake.checkTokenMutex.Unlock()
	fake.CheckTokenStub = nil
	fake.checkTokenReturns = struct {
		result1 int64
		result2 error
	}{result1, result2}
}

func (fake *FakeAuthWithSSBService) CheckTokenReturnsOnCall(i int, result1 int64, result2 error) {
	fake.checkTokenMutex.Lock()
	defer fake.checkTokenMutex.Unlock()
	fake.CheckTokenStub = nil
	if fake.checkTokenReturnsOnCall == nil {
		fake.checkTokenReturnsOnCall = make(map[int]struct {
			result1 int64
			result2 error
		})
	}
	fake.checkTokenReturnsOnCall[i] = struct {
		result1 int64
		result2 error
	}{result1, result2}
}

func (fake *FakeAuthWithSSBService) CreateToken(arg1 context.Context, arg2 int64) (string, error) {
	fake.createTokenMutex.Lock()
	ret, specificReturn := fake.createTokenReturnsOnCall[len(fake.createTokenArgsForCall)]
	fake.createTokenArgsForCall = append(fake.createTokenArgsForCall, struct {
		arg1 context.Context
		arg2 int64
	}{arg1, arg2})
	stub := fake.CreateTokenStub
	fakeReturns := fake.createTokenReturns
	fake.recordInvocation("CreateToken", []interface{}{arg1, arg2})
	fake.createTokenMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeAuthWithSSBService) CreateTokenCallCount() int {
	fake.createTokenMutex.RLock()
	defer fake.createTokenMutex.RUnlock()
	return len(fake.createTokenArgsForCall)
}

func (fake *FakeAuthWithSSBService) CreateTokenCalls(stub func(context.Context, int64) (string, error)) {
	fake.createTokenMutex.Lock()
	defer fake.createTokenMutex.Unlock()
	fake.CreateTokenStub = stub
}

func (fake *FakeAuthWithSSBService) CreateTokenArgsForCall(i int) (context.Context, int64) {
	fake.createTokenMutex.RLock()
	defer fake.createTokenMutex.RUnlock()
	argsForCall := fake.createTokenArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeAuthWithSSBService) CreateTokenReturns(result1 string, result2 error) {
	fake.createTokenMutex.Lock()
	defer fake.createTokenMutex.Unlock()
	fake.CreateTokenStub = nil
	fake.createTokenReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeAuthWithSSBService) CreateTokenReturnsOnCall(i int, result1 string, result2 error) {
	fake.createTokenMutex.Lock()
	defer fake.createTokenMutex.Unlock()
	fake.CreateTokenStub = nil
	if fake.createTokenReturnsOnCall == nil {
		fake.createTokenReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.createTokenReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeAuthWithSSBService) RemoveToken(arg1 context.Context, arg2 string) error {
	fake.removeTokenMutex.Lock()
	ret, specificReturn := fake.removeTokenReturnsOnCall[len(fake.removeTokenArgsForCall)]
	fake.removeTokenArgsForCall = append(fake.removeTokenArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.RemoveTokenStub
	fakeReturns := fake.removeTokenReturns
	fake.recordInvocation("RemoveToken", []interface{}{arg1, arg2})
	fake.removeTokenMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeAuthWithSSBService) RemoveTokenCallCount() int {
	fake.removeTokenMutex.RLock()
	defer fake.removeTokenMutex.RUnlock()
	return len(fake.removeTokenArgsForCall)
}

func (fake *FakeAuthWithSSBService) RemoveTokenCalls(stub func(context.Context, string) error) {
	fake.removeTokenMutex.Lock()
	defer fake.removeTokenMutex.Unlock()
	fake.RemoveTokenStub = stub
}

func (fake *FakeAuthWithSSBService) RemoveTokenArgsForCall(i int) (context.Context, string) {
	fake.removeTokenMutex.RLock()
	defer fake.removeTokenMutex.RUnlock()
	argsForCall := fake.removeTokenArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeAuthWithSSBService) RemoveTokenReturns(result1 error) {
	fake.removeTokenMutex.Lock()
	defer fake.removeTokenMutex.Unlock()
	fake.RemoveTokenStub = nil
	fake.removeTokenReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAuthWithSSBService) RemoveTokenReturnsOnCall(i int, result1 error) {
	fake.removeTokenMutex.Lock()
	defer fake.removeTokenMutex.Unlock()
	fake.RemoveTokenStub = nil
	if fake.removeTokenReturnsOnCall == nil {
		fake.removeTokenReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.removeTokenReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeAuthWithSSBService) WipeTokensForMember(arg1 context.Context, arg2 int64) error {
	fake.wipeTokensForMemberMutex.Lock()
	ret, specificReturn := fake.wipeTokensForMemberReturnsOnCall[len(fake.wipeTokensForMemberArgsForCall)]
	fake.wipeTokensForMemberArgsForCall = append(fake.wipeTokensForMemberArgsForCall, struct {
		arg1 context.Context
		arg2 int64
	}{arg1, arg2})
	stub := fake.WipeTokensForMemberStub
	fakeReturns := fake.wipeTokensForMemberReturns
	fake.recordInvocation("WipeTokensForMember", []interface{}{arg1, arg2})
	fake.wipeTokensForMemberMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeAuthWithSSBService) WipeTokensForMemberCallCount() int {
	fake.wipeTokensForMemberMutex.RLock()
	defer fake.wipeTokensForMemberMutex.RUnlock()
	return len(fake.wipeTokensForMemberArgsForCall)
}

func (fake *FakeAuthWithSSBService) WipeTokensForMemberCalls(stub func(context.Context, int64) error) {
	fake.wipeTokensForMemberMutex.Lock()
	defer fake.wipeTokensForMemberMutex.Unlock()
	fake.WipeTokensForMemberStub = stub
}

func (fake *FakeAuthWithSSBService) WipeTokensForMemberArgsForCall(i int) (context.Context, int64) {
	fake.wipeTokensForMemberMutex.RLock()
	defer fake.wipeTokensForMemberMutex.RUnlock()
	argsForCall := fake.wipeTokensForMemberArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeAuthWithSSBService) WipeTokensForMemberReturns(result1 error) {
	fake.wipeTokensForMemberMutex.Lock()
	defer fake.wipeTokensForMemberMutex.Unlock()
	fake.WipeTokensForMemberStub = nil
	fake.wipeTokensForMemberReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAuthWithSSBService) WipeTokensForMemberReturnsOnCall(i int, result1 error) {
	fake.wipeTokensForMemberMutex.Lock()
	defer fake.wipeTokensForMemberMutex.Unlock()
	fake.WipeTokensForMemberStub = nil
	if fake.wipeTokensForMemberReturnsOnCall == nil {
		fake.wipeTokensForMemberReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.wipeTokensForMemberReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeAuthWithSSBService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.checkTokenMutex.RLock()
	defer fake.checkTokenMutex.RUnlock()
	fake.createTokenMutex.RLock()
	defer fake.createTokenMutex.RUnlock()
	fake.removeTokenMutex.RLock()
	defer fake.removeTokenMutex.RUnlock()
	fake.wipeTokensForMemberMutex.RLock()
	defer fake.wipeTokensForMemberMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeAuthWithSSBService) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ roomdb.AuthWithSSBService = new(FakeAuthWithSSBService)
