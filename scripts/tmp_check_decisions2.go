package main
import (
  "database/sql"
  "fmt"
  "log"
  _ "modernc.org/sqlite"
)
func main(){
 db,err:=sql.Open("sqlite","data/data.db"); if err!=nil{log.Fatal(err)}; defer db.Close()
 q:=`SELECT success, COUNT(*) FROM decision_records WHERE IFNULL(TRIM(decision_json),'')='' GROUP BY success ORDER BY success` 
 rows,err:=db.Query(q); if err!=nil{log.Fatal(err)}; defer rows.Close()
 fmt.Println("empty decision_json by success:")
 for rows.Next(){var s int; var c int; rows.Scan(&s,&c); fmt.Printf("success=%d count=%d\n",s,c)}
 fmt.Println("\nlatest 8 empty decision_json rows:")
 rows2,err:=db.Query(`SELECT id, cycle_number, success, error_message, execution_log FROM decision_records WHERE IFNULL(TRIM(decision_json),'')='' ORDER BY id DESC LIMIT 8`); if err!=nil{log.Fatal(err)}; defer rows2.Close()
 for rows2.Next(){var id,cyc int; var s int; var em,el string; rows2.Scan(&id,&cyc,&s,&em,&el); if len(el)>140 {el=el[:140]+"..."}; if len(em)>80 {em=em[:80]+"..."}; fmt.Printf("id=%d cycle=%d success=%d err=%q log=%q\n",id,cyc,s,em,el)}
}
