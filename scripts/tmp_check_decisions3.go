package main
import (
  "database/sql"
  "fmt"
  "log"
  _ "modernc.org/sqlite"
)
func main(){
 db,err:=sql.Open("sqlite","data/data.db"); if err!=nil{log.Fatal(err)}; defer db.Close()
 rows,err:=db.Query(`SELECT id, cycle_number, LENGTH(raw_response), LENGTH(cot_trace), substr(raw_response,1,300), substr(cot_trace,1,180) FROM decision_records WHERE IFNULL(TRIM(decision_json),'')='' AND success=1 ORDER BY id DESC LIMIT 3`)
 if err!=nil{log.Fatal(err)}
 defer rows.Close()
 for rows.Next(){
   var id,cyc,lr,lc int
   var rr,ct string
   if err:=rows.Scan(&id,&cyc,&lr,&lc,&rr,&ct); err!=nil{log.Fatal(err)}
   fmt.Printf("id=%d cycle=%d raw_len=%d cot_len=%d\nraw_head=%q\ncot_head=%q\n---\n",id,cyc,lr,lc,rr,ct)
 }
}
