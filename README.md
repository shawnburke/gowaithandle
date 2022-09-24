# gowaithandle

Go has great concurrency primatives but some are a bit too...primative.

I spent many years writing C# against .NET but then joined Uber and did almost exclusively Go.

One thing that I would constantly miss were the `EventHandle` classes from `System.Threading` because they allowed repeated signaling of the same handle. So I wrote this, but now that I did, I'm not sure it's _that_ useful over the primatives.

But it was fun, so here you go.

They are written using (mostly) lockless synchronization and so should be fast and lightweight.
## Usage

The main classes are:

* `AutoResetEvent` - lets a single thread through for each call to `Set` then toggles back.
* `ManualResetEvent` - allows toggling to signaled, which will let all threads through until `Reset` is called.

These classes implement the `WaitHandle` interface which supports waiting with a timeout.

```
WaitOne(timeout time.Duration) <-chan bool // false means timed out
```

Plus helpers `WaitAll` or `WaitAny` which are useful for dynamic situations where `case` is not possible.

### AutoResetEvent

`AutoResetEvent` allows signaling that lets a single thread through and toggles back immediately.  This is good if you want to "pulse" work so that a single thread.


```
are := AutoResetEvent{}

go func() {

    counter := 0
    for {

        result := <=are.WaitOne(time.Minute)
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

        result := <=mre.WaitOne(time.Minute)
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


