// Notes: https://promisesaplus.com/
package future

import (
	"reflect"
	"sync"
)

type Promise struct {
	wg sync.WaitGroup

	tResolvedWrapper reflect.Type
	vResolvedWrapper reflect.Value

	tRejectedWrapper reflect.Type
	vRejectedWrapper reflect.Value

	resolvedFuncs []fnInfoType
	rejectedFuncs []fnInfoType
}

type fnInfoType struct {
	v reflect.Value
	t reflect.Type
}

func Future(fn interface{}) (q *Promise) {
	exec, q := FutureDeferred(fn)
	exec(true)
	return q
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

	v := reflect.ValueOf(fn)

	return func(bAsync bool) {
		if bAsync {
			q.wg.Add(1)
			go func() {
				v.Call([]reflect.Value{q.vResolvedWrapper, q.vRejectedWrapper})
				q.wg.Done()
			}()
		} else {
			v.Call([]reflect.Value{q.vResolvedWrapper, q.vRejectedWrapper})
		}

		return

	}, q
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
	return wrappedFunc(this.tResolvedWrapper, args, this.resolvedFuncs)
}

func (this *Promise) rejectedWrapped(args []reflect.Value) (results []reflect.Value) {
	return wrappedFunc(this.tRejectedWrapper, args, this.rejectedFuncs)
}

func wrappedFunc(tFn reflect.Type, args []reflect.Value, funcInfos []fnInfoType) (results []reflect.Value) {
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

func (this *Promise) Then(fulfilledFn, rejectedFn interface{}) (q *Promise) {
	t1 := reflect.TypeOf(fulfilledFn)
	t2 := reflect.TypeOf(rejectedFn)

	v1 := reflect.ValueOf(fulfilledFn)
	v2 := reflect.ValueOf(rejectedFn)

	bValid := ((t1.Kind() == reflect.Func) || (t1.Kind() == reflect.Ptr) && v1.IsNil()) &&
		((t2.Kind() == reflect.Func) || ((t2.Kind() == reflect.Ptr) && v2.IsNil()))
	if !bValid {
		panic("fulfilledFn and rejectedFn should be functions")
	}

	if !v1.IsNil() {
		this.initResolvedIfNeeded(t1)
		this.resolvedFuncs = append(this.resolvedFuncs, fnInfoType{
			v: v1, t: t1,
		})
	}

	if !v2.IsNil() {
		this.initRejectedIfNeeded(t2)
		this.rejectedFuncs = append(this.rejectedFuncs, fnInfoType{
			v: v2, t: t2,
		})
	}

	return this
}

func (this Promise) Wait() {
	this.wg.Wait()
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
