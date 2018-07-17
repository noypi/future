// Notes: https://promisesaplus.com/
package future

import (
	"reflect"
)

type Promise struct {
	tFulfilledFn reflect.Type
	tRejectedFn  reflect.Type
	vFulfilledFn reflect.Value
	vRejectedFn  reflect.Value
}

func defaultRejected() {}

var vDefaultRejectedFn = reflect.ValueOf(defaultRejected)

func Future(fn interface{}) (exec func(), q *Promise) {
	t := reflect.TypeOf(fn)

	bValid := (t.Kind() == reflect.Func) &&
		(t.NumIn() == 2) &&
		(t.In(0).Kind() == reflect.Func) &&
		(t.In(1).Kind() == reflect.Func)

	if !bValid {
		panic("Future should pass a function in form of func(resolveFunc, rejectFunc)")
	}

	q = &Promise{}

	v := reflect.ValueOf(fn)

	return func() {
		//TODO make async
		v.Call([]reflect.Value{q.vFulfilledFn, q.vRejectedFn})
	}, q
}

func (this *Promise) Then(fulfilledFn, rejectedFn interface{}) (q *Promise) {
	t1 := reflect.TypeOf(fulfilledFn)
	if nil == rejectedFn {
		rejectedFn = defaultRejected
	}

	t2 := reflect.TypeOf(rejectedFn)
	bValid := (t1.Kind() == reflect.Func) &&
		(t2.Kind() == reflect.Func)
	if !bValid {
		panic("fulfilledFn and rejectedFn should be functions")
	}

	v1 := reflect.ValueOf(fulfilledFn)
	v2 := reflect.ValueOf(rejectedFn)

	if this.vFulfilledFn.IsValid() {
		newF1t := reflect.FuncOf(
			getFuncTypeIns(this.tFulfilledFn),
			getFuncTypeOuts(this.tFulfilledFn),
			false,
		)

		oldF1v := this.vFulfilledFn
		this.vFulfilledFn = reflect.MakeFunc(newF1t, func(args []reflect.Value) (results []reflect.Value) {
			return v1.Call(oldF1v.Call(args))
		})
	} else {
		this.vFulfilledFn = v1
	}
	this.tFulfilledFn = t1

	if this.vRejectedFn.IsValid() {
		newF2t := reflect.FuncOf(
			getFuncTypeIns(this.tRejectedFn),
			getFuncTypeOuts(this.tRejectedFn),
			false,
		)

		oldF2v := this.vRejectedFn
		this.vRejectedFn = reflect.MakeFunc(newF2t, func(args []reflect.Value) (results []reflect.Value) {
			return v2.Call(oldF2v.Call(args))
		})

	} else {
		this.vRejectedFn = v2
	}
	this.tRejectedFn = t2

	return this
}

func getFuncTypeIns(t reflect.Type) (ts []reflect.Type) {
	for i := 0; i < t.NumIn(); i++ {
		ts = append(ts, t.In(i))
	}
	return
}

func getFuncTypeOuts(t reflect.Type) (ts []reflect.Type) {
	for i := 0; i < t.NumOut(); i++ {
		ts = append(ts, t.Out(i))
	}
	return
}
