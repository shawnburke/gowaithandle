# gowaithandle

Go has great concurrency primatives but some are a bit too...primative.

I spent many years writing C# against .NET but then joined Uber and did almost exclusively Go.

One thing that I would constantly miss were the `EventHandle` classes from `System.Threading` because they allowed repeated signaling of the same handle. After writing this, I've gained a better appreciation for those Go primatives: it's true that with a good understanding of them you can mimic most of these behaviors. However, it requires a good understanding and sometimes it's nice to have things wrapped up.

And it wa a fun exercise, so here you go.

They are written using Go idioms and (mostly) lockless synchronization and so should be fast and lightweight.

## Usage

The main classes are:

* `AutoResetEvent` - lets a single thread through for each call to `Set` then toggles back.
* `ManualResetEvent` - allows toggling to signaled, which will let all threads through until `Reset` is called.

These classes implement the `WaitHandle` interface which supports waiting with `context.Context` which supports timeouts, deadlines, and cancellation.

```
WaitOne(ctx context.Context) <-chan bool // false means timed out
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
