# Go WaitHandle Library (gowaithandle)
_An ode to .NET Framework synchronization classes_

![Go Test](https://github.com/shawnburke/gowaithandle/actions/workflows/go.yml/badge.svg)  

![Go Report Card](https://goreportcard.com/badge/github.com/shawnburke/gowaithandle)

[![codecov](https://codecov.io/gh/shawnburke/gowaithandle/branch/main/graph/badge.svg)](https://codecov.io/gh/shawnburke/gowaithandle)

Go has great concurrency primatives but some are a bit too...primative.

At Microsoft, I was on the original .NET team and wrote exclusively C# for many years after that. In 2015, I joined Uber and for my 5+ years there I wrote almost exclusively Go.

There are lots of good things about Go, but something I continually miss from .NET are the `System.Threading.EventHandle` classes, because they made it simple to configure a a handle and hand it out to callers via simple interface that abstracted the actual behavior of the handle. Whether it was a manual reset, or a semaphore, or something else, the caller would wait on it in the same way (`WaitOne`). This also allowed repeated signalling of that handle, which is something that gets complicated in Go; you can only close out a channel once, and the semantics of multiple receivers are fixed.

While it is true that the Go primatives are very powerful and it is not terribly difficult to mimic each of these things with channels, `sync.Mutex`, and `sync.WaitGroup`, it often feels like more complexity in areas that are prone to subtle mistakes and nasty race conditions

So it seemed like a fun exercise to implement this functionality in Go.

They are written using Go idioms and (mostly) lockless synchronization and so should be fast and lightweight.

## Usage

The main classes are:

* [`AutoResetEvent`](#autoresetevent) - lets a single thread through for each call to `Set` then toggles back.
* [`ManualResetEvent`](#manualresetevent) - allows toggling to signaled, which will let all threads through until `Reset` is called.
* [`WaitGroup`](#waitgroup) - derived implementation of `sync.WaitGroup` that returns a channel and supports timeout and cancel behavior
* [`Sempahore`](#sempahore) - semaphore that can be used to limit access to a resource or implement throttling or concurrency limitation.

These classes implement the `WaitHandle` interface which supports waiting with `WaitOne(context.Context)` allowing timeouts, deadlines, and cancellation.

Basic usage against this interafce is always the same, which is to create the object then waiters call `WaitOne`.

```
eh := &gowaithandle.AutoResetEvent{}

go func() {

    // wait a second then fail
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
    ok := <-eh.WaitOne()

    if ok {
        // we got it
        doSomething()
        return
    }
    log.Println("ERROR: failed to get handle")

}()

// allow the thread to continue
eh.Set()

// create another one
mh := &gowaithandle.ManualResetEvent{}

go func() {
    ok := <-gowaithandle.WaitAll(context.Background(), eh, mh)
    log.Println("All signaled!")
}()

// signal them both will allow the above to continue
eh.Set()
mh.Set()

```

Plus helpers `WaitAll` or `WaitAny` which are useful for dynamic situations where `case` is not possible.

### AutoResetEvent

`AutoResetEvent` allows signaling that lets a single thread through and toggles back immediately.  This is good if you want to "pulse" work so that a single thread.


```
are := AutoResetEvent{}

go func() {

    counter := 0
    for {

        ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
        defer cancel()
        result := <=are.WaitOne(ctx)
        if !result {
            log.Println("Timeout")
            continue
        }

        Println("Loop", counter)
        counter++
    }
}()

for i := 0; i < 3 i++ {
    are.Set()
}

```

Will print

```
Loop 0
Loop 1
Loop 2
Timeout
Timeout
...
```

### ManualResetEvent

`ManualResetEvent` is an event handle that is toggled to signaled or not. When signaled, all threads are let through until `Reset` is called.

```
mre := ManualResetEvent{}

go func() {

    counter := 0
    for {

        ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
        defer cancel()
        result := <=mre.WaitOne(ctx)
        if !result {
            log.Println("Timeout")
            continue
        }

        Println("Loop", counter)
        counter++
        time.Sleep(time.Second / 10)
    }
}()

mre.Set()
time.Sleep(time.Second / 2)
mre.Reset()

```

Prints (somethimg similar to)

```
Loop 0
Loop 1
Loop 2
Loop 3
Loop 4
Timeout
Timeout
...
```

### WaitGroup

This package also includes a derived implementation of `sync.WaitGroup` that also supports channel-based waiting and timeout.  It's interface and semantics are otherwise the same as the standard class.


```
wg := WaitGroup{}
other := make(chan struct{})

for i := 0; i < n; i++ {
    wg.Add()
    go func() {
        defer wg.Done()
        // do some stuff
    }()
}

ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

// give those threads 1 second to complete
select {
    case result :=<-wg.Done(ctx):
        fmt.Println("All threads completed=", result)
    case <-other:
        fmt.Println("Something else happened before threads were done")
}

```

This class also implements `WaitHandle` so can be used in the helpers:

```
    wg := WaitGroup{}
    wg.Add(1)
    are := AutoResetEvent{}

    go func() {
        defer wg.Done()
        // do some stuff
    }()

    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
    defer cancel()

    // wait until the waitgroup is done and the AutoResetEvent signals
    result := <- wg.WaitAll(ctx, wg, are)

    fmt.Println("Did it work?", result)
```

### Sempahore

Implementation of a sempahore usable for limiting resources such as throttling or concurrency limiting.

The `Sempaphore` also implements `WaitHandle` and can be used in `WaitAny` or `WaitAll`.

To release a resource, call `Release`.

```
s := NewSemaphore(2)

// take two resources
<-s.WaitOne(context.Background())
<-s.WaitOne(context.Background())

// try for one second to get a third, which will time out
ctx, cancel := context.WithTimeout(time.Second)
defer cancel()
if result := <-WaitOne(ctx); ! result {
    fmt.Println("Timeout!")
}

s.Release()

// now this will succeed
result := <-s.WaitOne(context.Background())
fmt.Println("Done", result)  // prints 'Done true'

```

This makes it very easy to implement something like throttling for an HTTP connection.

Below we demonstrate how to build a simple middleware function that would limit connections and
have proper behavior for timeouts and connection resets. If a connection was waiting on the sempahore and was dropped by the client, this code would give up rather than continue trying to aquire the resource.

```
    // limit an http route to 4 concurrent requests with a middleware
    // component
    var throttle = NewSempaphore(4)

    func httpThrottleMiddleware(w http.ResponseWriter, r *http.Request, next http.Handler) {
        
        // we pass the request context here so if the request is disconnected
        // or times out we don't take a resource
        if <- throttle.WaitOne(r.Context()) {
            defer throttle.Release()
            
            next.ServeHttp(w, r)
        } else {
            w.WriteHeader(429)
        }
    }
```