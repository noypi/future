// Notes: https://promisesaplus.com/
package future

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/pkg/errors"
)

type Promise struct {
	wg sync.WaitGroup

	tResolvedWrapper reflect.Type
	vResolvedWrapper reflect.Value

	tRejectedWrapper reflect.Type
	vRejectedWrapper reflect.Value

	tFn reflect.Type
	vFn reflect.Value

	resolvedFuncs []*fnInfoType
	rejectedFuncs []*fnInfoType

	state   FinalState
	results []reflect.Value
	catch   func(err error, o interface{})
	finally func(FinalState, ...interface{})
}

type FinalState int

const (
	FinalUnknown FinalState = iota
	FinalResolved
	FinalRejected
	FinalRecovered
)

func Future(fn interface{}) (q *Promise) {
	exec, q := FutureDeferred(fn)
	exec(true)
	return q
}

var defaultFinally = func(FinalState, ...interface{}) {}
var defaultCatch = func(err error, o interface{}) {
	log.Println(fmt.Sprintf("%+v", err), " => ", o)
}

func FutureDeferred(fn interface{}) (exec func(bAsync bool), q *Promise) {
	t := reflect.TypeOf(fn)

	bValid := (t.Kind() == reflect.Func) &&
		(t.NumIn() == 2) &&
		(t.In(0).Kind() == reflect.Func) &&
		(t.In(1).Kind() == reflect.Func)

	if !bValid {
		panic("Future should pass a function in form of func(resolveFunc, rejectFunc)")
	}

	q = &Promise{}
	q.state = FinalUnknown
	q.finally = defaultFinally
	q.catch = defaultCatch

	v := reflect.ValueOf(fn)
	q.tFn = t
	q.vFn = v

	return q.exec, q
}

func (this *Promise) Then(fulfilledFn, rejectedFn interface{}) (q *Promise) {

	this.OnSuccess(fulfilledFn)
	this.OnFail(rejectedFn)

	return this
}

func (q *Promise) exec(bAsync bool) {

	defer func() {
		o := recover()
		if nil != o {
			q.state = FinalRecovered
			err := errors.WithStack(errors.New("Promise Recovered"))
			q.catch(err, o)
		}

		q.finally(q.state, vsToInterface(q.results)...)
	}()
	if bAsync {
		q.wg.Add(1)
		go func() {
			q.vFn.Call([]reflect.Value{q.vResolvedWrapper, q.vRejectedWrapper})
			q.wg.Done()
		}()
	} else {
		q.vFn.Call([]reflect.Value{q.vResolvedWrapper, q.vRejectedWrapper})
	}

}

func (this *Promise) OnSuccess(fulfilledFn interface{}) (q *Promise) {
	arr, tFn, _, bAdded := fnInfoTypeArr(this.resolvedFuncs).Append(fulfilledFn)
	if bAdded {
		this.initResolvedIfNeeded(tFn)
		this.resolvedFuncs = arr
	}
	return this
}

func (this *Promise) OnFail(fulfilledFn interface{}) (q *Promise) {
	arr, tFn, _, bAdded := fnInfoTypeArr(this.rejectedFuncs).Append(fulfilledFn)
	if bAdded {
		this.initRejectedIfNeeded(tFn)
		this.rejectedFuncs = arr
	}

	return this
}

func (this *Promise) SetCatch(recoverFn func(error, interface{})) {
	this.catch = recoverFn
}

func (this *Promise) SetFinally(finallyFn func(FinalState, ...interface{})) {
	this.finally = finallyFn
}

func (this Promise) Wait() {
	this.wg.Wait()
}

// returns first promise to finish or rejected
func Race(qs ...*Promise) (q *Promise) {
	var l sync.Mutex
	var bDone bool
	ch := make(chan *Promise)

	fn := func(q1 *Promise) {
		fncb := func() {
			l.Lock()
			if !bDone {
				bDone = true
				ch <- q1
			}
			l.Unlock()
		}
		q1.Then(fncb, fncb)
	}

	for _, q0 := range qs {
		fn(q0)
	}

	q = <-ch
	close(ch)
	q.Wait()

	return
}

type fnInfoType struct {
	v reflect.Value
	t reflect.Type
}

type fnInfoTypeArr []*fnInfoType

func (this fnInfoTypeArr) Append(fulfilledFn interface{}) (arr fnInfoTypeArr, tFn reflect.Type, vFn reflect.Value, bAdded bool) {
	tFn = reflect.TypeOf(fulfilledFn)
	vFn = reflect.ValueOf(fulfilledFn)

	bValid := ((tFn.Kind() == reflect.Func) || (tFn.Kind() == reflect.Ptr) && vFn.IsNil())

	if !bValid {
		panic("param should be a function")
	}

	if !vFn.IsNil() {
		bAdded = true
		arr = append(this, &fnInfoType{v: vFn, t: tFn})
	}

	return
}

func (this *Promise) initResolvedIfNeeded(tResolved reflect.Type) {
	if nil == this.tResolvedWrapper {
		this.tResolvedWrapper = tResolved
		this.vResolvedWrapper = createVFuncWrapper(tResolved, this.resolvedWrapped)
	}
}

func (this *Promise) initRejectedIfNeeded(tRejected reflect.Type) {
	if nil == this.tRejectedWrapper {
		this.tRejectedWrapper = tRejected
		this.vRejectedWrapper = createVFuncWrapper(tRejected, this.rejectedWrapped)
	}
}

func (this *Promise) resolvedWrapped(args []reflect.Value) (results []reflect.Value) {
	this.state = FinalResolved
	results = wrappedFunc(this.tResolvedWrapper, args, this.resolvedFuncs)
	this.results = results
	return
}

func (this *Promise) rejectedWrapped(args []reflect.Value) (results []reflect.Value) {
	this.state = FinalRejected
	results = wrappedFunc(this.tRejectedWrapper, args, this.rejectedFuncs)
	this.results = results
	return
}

func wrappedFunc(tFn reflect.Type, args []reflect.Value, funcInfos []*fnInfoType) (results []reflect.Value) {
	var res0 []reflect.Value
	for i, fnInfo := range funcInfos {
		args0 := res0
		if 0 == i {
			args0 = args
		}
		t := fnInfo.t
		argsFit := make([]reflect.Value, t.NumIn())
		fitFnArgs(t, t.In, argsFit, args0)
		res0 = fnInfo.v.Call(argsFit)
	}

	results = make([]reflect.Value, tFn.NumOut())
	fitFnArgs(tFn, tFn.Out, results, res0)

	return results
}

func createVFuncWrapper(tFn reflect.Type, vWrappedFn func(args []reflect.Value) (results []reflect.Value)) (vFn reflect.Value) {
	newF1t := reflect.FuncOf(
		getFuncTypeIns(tFn),
		getFuncTypeOuts(tFn),
		tFn.IsVariadic(),
	)

	return reflect.MakeFunc(newF1t, vWrappedFn)
}

func fitFnArgs(tFn reflect.Type, dirFn func(int) reflect.Type, toLenArr, fromArr []reflect.Value) {
	for i := 0; i < len(toLenArr); i++ {
		tparam := dirFn(i)
		if i < len(fromArr) {
			if tparam == fromArr[i].Type() {
				toLenArr[i] = fromArr[i]
			} else {
				toLenArr[i] = reflect.New(tparam).Elem()
			}
		} else {
			toLenArr[i] = reflect.New(tparam).Elem()
		}
	}
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

func vsToInterface(vs []reflect.Value) (arr []interface{}) {
	arr = make([]interface{}, len(vs))
	for i, v := range vs {
		arr[i] = v
	}

	return
}
