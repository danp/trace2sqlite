package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"honnef.co/go/gotraceui/trace"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("need db and trace")
	}

	f, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	t, err := trace.Parse(f, func(f float64) {})
	if err != nil {
		log.Fatal(err)
	}

	conn, err := sqlite.OpenConn(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	end := sqlitex.Transaction(conn)

	fmt.Println("pcs")
	if err := sqlitex.Execute(conn, "create table pcs (pc integer primary key, file text, line integer, fn text)", nil); err != nil {
		log.Fatal(err)
	}
	for pc, fr := range t.PCs {
		o := &sqlitex.ExecOptions{Named: map[string]any{
			"$pc":   pc,
			"$file": fr.File,
			"$line": fr.Line,
			"$fn":   fr.Fn,
		}}
		if err := sqlitex.Execute(conn, "insert into pcs values ($pc, $file, $line, $fn)", o); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("stacks")
	if err := sqlitex.Execute(conn, "create table stacks (id integer, depth integer, pc integer, primary key(id, depth))", nil); err != nil {
		log.Fatal(err)
	}
	for si, s := range t.Stacks {
		for d, pc := range s {
			o := &sqlitex.ExecOptions{Named: map[string]any{"$id": si, "$depth": d, "$pc": pc}}
			if err := sqlitex.Execute(conn, "insert into stacks values ($id, $depth, $pc)", o); err != nil {
				log.Fatal(err)
			}
		}
	}

	fmt.Println("events")
	if err := sqlitex.Execute(conn, "create table events (id integer primary key, ts integer, type text, p integer, g integer, args text, link integer, stack_id integer)", nil); err != nil {
		log.Fatal(err)
	}
	for ei, e := range t.Events {
		d := trace.EventDescriptions[e.Type]
		args := make(map[string]uint64)
		for i, a := range d.Args {
			args[a] = e.Args[i]
		}
		ab, err := json.Marshal(args)
		if err != nil {
			log.Fatal(err)
		}
		o := &sqlitex.ExecOptions{Named: map[string]any{
			"$id":       ei,
			"$ts":       e.Ts,
			"$type":     d.Name,
			"$p":        e.P,
			"$g":        e.G,
			"$args":     string(ab),
			"$link":     nil,
			"$stack_id": nil,
		}}
		if e.Link >= 0 {
			o.Named["$link"] = e.Link
		}
		if e.StkID > 0 {
			o.Named["$stack_id"] = e.StkID
		}

		if err := sqlitex.Execute(conn, "insert into events values ($id, $ts, $type, $p, $g, $args, $link, $stack_id)", o); err != nil {
			log.Fatal(err)
		}
	}

	err = nil
	end(&err)
}
