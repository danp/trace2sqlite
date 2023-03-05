# trace2sqlite

Convert Go execution traces to SQLite databases.

## Usage

Create `trace.db` from `my.trace`:

``` sh
trace2sqlite trace.db my.trace
```

## Quick tour

Using a database generated from [this gotraceui test trace](https://github.com/dominikh/gotraceui/blob/dfa752a689cdbaa1f7d648546e5f1715320439c9/trace/testdata/stress_1_20_good).

```
$ sqlite3 trace.db
sqlite> .schema
CREATE TABLE pcs (pc integer primary key, file text, line integer, fn text);
CREATE TABLE stacks (id integer, depth integer, pc integer, primary key(id, depth));
CREATE TABLE events (id integer primary key, ts integer, type text, p integer, g integer, args text, link integer, stack_id integer);
sqlite> select * from events limit 10;
┌────┬──────┬─────────────┬───┬────┬────────────────────┬──────┬──────────┐
│ id │  ts  │    type     │ p │ g  │        args        │ link │ stack_id │
├────┼──────┼─────────────┼───┼────┼────────────────────┼──────┼──────────┤
│ 0  │ 0    │ GoCreate    │ 1 │ 0  │ {"g":1,"stack":2}  │      │ 1        │
│ 1  │ 164  │ GoWaiting   │ 1 │ 1  │ {"g":1}            │      │          │
│ 2  │ 2235 │ GoCreate    │ 1 │ 0  │ {"g":17,"stack":3} │      │ 1        │
│ 3  │ 2290 │ GoInSyscall │ 1 │ 17 │ {"g":17}           │      │          │
│ 4  │ 4341 │ GoCreate    │ 1 │ 0  │ {"g":2,"stack":4}  │      │ 1        │
│ 5  │ 4396 │ GoWaiting   │ 1 │ 2  │ {"g":2}            │      │          │
│ 6  │ 6357 │ GoCreate    │ 1 │ 0  │ {"g":3,"stack":5}  │      │ 1        │
│ 7  │ 6393 │ GoWaiting   │ 1 │ 3  │ {"g":3}            │ 580  │          │
│ 8  │ 8280 │ GoCreate    │ 1 │ 0  │ {"g":4,"stack":6}  │      │ 1        │
│ 9  │ 8299 │ GoWaiting   │ 1 │ 4  │ {"g":4}            │ 348  │          │
└────┴──────┴─────────────┴───┴────┴────────────────────┴──────┴──────────┘
sqlite> select type, count(*) from events group by 1 order by 2 desc limit 10;
┌──────────────┬──────────┐
│     type     │ count(*) │
├──────────────┼──────────┤
│ ProcStart    │ 52271    │
│ ProcStop     │ 52270    │
│ GoStart      │ 23728    │
│ GoSched      │ 22131    │
│ GoUnblock    │ 7813     │
│ GoBlock      │ 7009     │
│ GoStartLabel │ 6539     │
│ HeapAlloc    │ 1438     │
│ Gomaxprocs   │ 673      │
│ GCSTWStart   │ 670      │
└──────────────┴──────────┘
sqlite> select * from pcs limit 10;
┌─────────┬─────────────────────────────────────────────┬──────┬────────────────────────────┐
│   pc    │                    file                     │ line │             fn             │
├─────────┼─────────────────────────────────────────────┼──────┼────────────────────────────┤
│ 0       │                                             │ 0    │                            │
│ 4223036 │ /home/dominikh/prj/go/src/runtime/chan.go   │ 145  │ runtime.chansend1          │
│ 4226455 │ /home/dominikh/prj/go/src/runtime/chan.go   │ 442  │ runtime.chanrecv1          │
│ 4230932 │ /home/dominikh/prj/go/src/runtime/debug.go  │ 33   │ runtime.GOMAXPROCS         │
│ 4252879 │ /home/dominikh/prj/go/src/runtime/malloc.go │ 932  │ runtime.mallocgc           │
│ 4253942 │ /home/dominikh/prj/go/src/runtime/malloc.go │ 1053 │ runtime.mallocgc           │
│ 4254441 │ /home/dominikh/prj/go/src/runtime/malloc.go │ 1150 │ runtime.mallocgc           │
│ 4254567 │ /home/dominikh/prj/go/src/runtime/malloc.go │ 1171 │ runtime.mallocgc           │
│ 4254582 │ /home/dominikh/prj/go/src/runtime/malloc.go │ 1172 │ runtime.mallocgc           │
│ 4254905 │ /home/dominikh/prj/go/src/runtime/malloc.go │ 1217 │ runtime.deductAssistCredit │
└─────────┴─────────────────────────────────────────────┴──────┴────────────────────────────┘
sqlite> select * from stacks limit 10;
┌────┬───────┬─────────┐
│ id │ depth │   pc    │
├────┼───────┼─────────┤
│ 80 │ 0     │ 4696772 │
│ 80 │ 1     │ 4785133 │
│ 80 │ 2     │ 4785109 │
│ 80 │ 3     │ 4784469 │
│ 80 │ 4     │ 4800349 │
│ 80 │ 5     │ 4800341 │
│ 80 │ 6     │ 4772089 │
│ 80 │ 7     │ 5559190 │
│ 80 │ 8     │ 5559150 │
│ 80 │ 9     │ 5565332 │
└────┴───────┴─────────┘
sqlite> with top_go_create_stacks as (select stack_id, count(*) from events where type='GoCreate' group by 1 order by 2 desc) select * from top_go_create_stacks join stacks on (stack_id=id), pcs on (stacks.pc=pcs.pc and depth=0);
┌──────────┬──────────┬─────┬───────┬─────────┬─────────┬───────────────────────────────────────────────────────┬──────┬────────────────────────────────────┐
│ stack_id │ count(*) │ id  │ depth │   pc    │   pc    │                         file                          │ line │                 fn                 │
├──────────┼──────────┼─────┼───────┼─────────┼─────────┼───────────────────────────────────────────────────────┼──────┼────────────────────────────────────┤
│ 63       │ 10       │ 63  │ 0     │ 5679942 │ 5679942 │ /home/dominikh/prj/go/src/runtime/trace/trace_test.go │ 260  │ runtime/trace_test.TestTraceStress │
│ 23       │ 10       │ 23  │ 0     │ 4308676 │ 4308676 │ /home/dominikh/prj/go/src/runtime/mgc.go              │ 1199 │ runtime.gcBgMarkStartWorkers       │
│ 1        │ 9        │ 1   │ 0     │ 5679707 │ 5679707 │ /home/dominikh/prj/go/src/runtime/trace/trace_test.go │ 226  │ runtime/trace_test.TestTraceStress │
│ 107      │ 1        │ 107 │ 0     │ 5680836 │ 5680836 │ /home/dominikh/prj/go/src/runtime/trace/trace_test.go │ 313  │ runtime/trace_test.TestTraceStress │
│ 86       │ 1        │ 86  │ 0     │ 5680604 │ 5680604 │ /home/dominikh/prj/go/src/runtime/trace/trace_test.go │ 295  │ runtime/trace_test.TestTraceStress │
│ 67       │ 1        │ 67  │ 0     │ 5680298 │ 5680298 │ /home/dominikh/prj/go/src/runtime/trace/trace_test.go │ 283  │ runtime/trace_test.TestTraceStress │
│ 65       │ 1        │ 65  │ 0     │ 5680218 │ 5680218 │ /home/dominikh/prj/go/src/runtime/trace/trace_test.go │ 274  │ runtime/trace_test.TestTraceStress │
│ 18       │ 1        │ 18  │ 0     │ 5679882 │ 5679882 │ /home/dominikh/prj/go/src/runtime/trace/trace_test.go │ 233  │ runtime/trace_test.TestTraceStress │
│ 13       │ 1        │ 13  │ 0     │ 5025742 │ 5025742 │ /home/dominikh/prj/go/src/runtime/trace/trace.go      │ 128  │ runtime/trace.Start                │
└──────────┴──────────┴─────┴───────┴─────────┴─────────┴───────────────────────────────────────────────────────┴──────┴────────────────────────────────────┘
```

## Fun queries

### FlameScope / perf

Generate data suitable for [FlameScope](https://github.com/Netflix/flamescope) (inspired by [traceutils](https://github.com/felixge/traceutils/#flamescope)) with:

``` shell
sqlite3 -tabs trace.db \
  "select 'go 0 [0] ' || printf('%f', ts/1e9) || ': cpu-clock:' || x'0a' || (select group_concat(x'09' || printf('%x %s (go)', stacks.pc, pcs.fn), x'0a') from stacks join pcs on (pcs.pc=stacks.pc) where stacks.id=stack_id) from events where type='CPUSample' order by id" \
  > trace.perf
```

Then get trace.perf where FlameScope can read it.

Example from encoding/json benchmarks:

![](https://user-images.githubusercontent.com/2182/222985527-2bf892f0-dc98-498d-a07f-cea359ea6142.png)

## Thanks

The bulk of the work in parsing and making traces more logically useful comes from the [gotraceui](https://github.com/dominikh/gotraceui) project.
This project really just puts the result of that into SQLite.
