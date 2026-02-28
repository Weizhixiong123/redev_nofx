package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "modernc.org/sqlite"
)

func qInt(db *sql.DB, sqls string, args ...any) int64 {
    var v sql.NullInt64
    if err := db.QueryRow(sqls, args...).Scan(&v); err != nil {
        log.Printf("query failed: %v | %s", err, sqls)
        return -1
    }
    if !v.Valid { return 0 }
    return v.Int64
}

func main() {
    db, err := sql.Open("sqlite", "data/data.db")
    if err != nil { log.Fatal(err) }
    defer db.Close()

    fmt.Println("== Tables check ==")
    var exists int
    _ = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='decision_records'").Scan(&exists)
    fmt.Printf("decision_records exists: %v\n", exists == 1)
    if exists != 1 { return }

    fmt.Println("\n== decision_records basic stats ==")
    total := qInt(db, "SELECT COUNT(*) FROM decision_records")
    ok := qInt(db, "SELECT COUNT(*) FROM decision_records WHERE success=1")
    fail := qInt(db, "SELECT COUNT(*) FROM decision_records WHERE success=0")
    fmt.Printf("total=%d success=%d fail=%d\n", total, ok, fail)

    fmt.Println("\n== Null/empty field checks ==")
    fmt.Printf("empty decision_json=%d\n", qInt(db, "SELECT COUNT(*) FROM decision_records WHERE IFNULL(TRIM(decision_json),'')=''"))
    fmt.Printf("empty raw_response=%d\n", qInt(db, "SELECT COUNT(*) FROM decision_records WHERE IFNULL(TRIM(raw_response),'')=''"))
    fmt.Printf("empty cot_trace=%d\n", qInt(db, "SELECT COUNT(*) FROM decision_records WHERE IFNULL(TRIM(cot_trace),'')=''"))
    fmt.Printf("empty execution_log=%d\n", qInt(db, "SELECT COUNT(*) FROM decision_records WHERE IFNULL(TRIM(execution_log),'')=''"))
    fmt.Printf("empty candidate_coins=%d\n", qInt(db, "SELECT COUNT(*) FROM decision_records WHERE IFNULL(TRIM(candidate_coins),'')=''"))

    fmt.Println("\n== JSON validity checks ==")
    fmt.Printf("invalid candidate_coins json=%d\n", qInt(db, "SELECT COUNT(*) FROM decision_records WHERE json_valid(candidate_coins)=0"))
    fmt.Printf("invalid execution_log json=%d\n", qInt(db, "SELECT COUNT(*) FROM decision_records WHERE json_valid(execution_log)=0"))
    fmt.Printf("invalid decisions json=%d\n", qInt(db, "SELECT COUNT(*) FROM decision_records WHERE json_valid(decisions)=0"))

    fmt.Println("\n== Cycle continuity check (per trader duplicates/gaps) ==")
    rows, err := db.Query("SELECT trader_id, COUNT(*) c, MIN(cycle_number), MAX(cycle_number), COUNT(DISTINCT cycle_number) dc FROM decision_records GROUP BY trader_id ORDER BY c DESC")
    if err != nil { log.Fatal(err) }
    defer rows.Close()
    for rows.Next() {
        var trader string
        var c, minc, maxc, dc int64
        if err := rows.Scan(&trader, &c, &minc, &maxc, &dc); err != nil { log.Fatal(err) }
        expected := maxc - minc + 1
        dup := c - dc
        gaps := expected - dc
        if gaps < 0 { gaps = 0 }
        fmt.Printf("trader=%s rows=%d cycle_range=[%d,%d] distinct=%d dup=%d gaps=%d\n", trader, c, minc, maxc, dc, dup, gaps)
    }

    fmt.Println("\n== Latest 5 records quick view ==")
    rows2, err := db.Query("SELECT id, trader_id, cycle_number, timestamp, success, LENGTH(decision_json), LENGTH(raw_response), LENGTH(cot_trace) FROM decision_records ORDER BY timestamp DESC LIMIT 5")
    if err != nil { log.Fatal(err) }
    defer rows2.Close()
    for rows2.Next() {
        var id, cycle int64
        var trader, ts string
        var success int
        var l1, l2, l3 sql.NullInt64
        if err := rows2.Scan(&id, &trader, &cycle, &ts, &success, &l1, &l2, &l3); err != nil { log.Fatal(err) }
        fmt.Printf("id=%d trader=%s cycle=%d ts=%s ok=%d decision_json_len=%d raw_len=%d cot_len=%d\n", id, trader, cycle, ts, success, l1.Int64, l2.Int64, l3.Int64)
    }
}
