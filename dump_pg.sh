#!/bin/bash

dump_clear(){
    ofile=$1
    used=sed
    if [ `uname` == "Darwin" ];then
        used=gsed
    fi
    $used -i 's/public\.//g' $ofile
    $used -i 's/ Owner: .*//g' $ofile
    $used -i 's/ALTER TABLE .* OWNER TO .*;//g' $ofile
    $used -i 's/-- Dumped from database.*//g' $ofile
    $used -i 's/-- Dumped by pg_dump.*//g' $ofile   
    $used -i ':a;N;$!ba;s/\n\n\n\n/\n\n/g' $ofile
    $used -i ':a;N;$!ba;s/START WITH 1\n/START WITH 1000\n/g' $ofile
    $used -i ':a;N;$!ba;s/SET[^=]*=[^\n]*\n//g' $ofile
    $used -i ':a;N;$!ba;s/ALTER TABLE ONLY [^\n]* DROP [^\n]*\n//g' $ofile
    $used -i ':a;N;$!ba;s/SELECT pg_catalog[^\n]*\n//g' $ofile
    $used -i 's/ALTER TABLE/ALTER TABLE IF EXISTS/g' $ofile
    $used -i 's/DROP INDEX/DROP INDEX IF EXISTS/g' $ofile
    $used -i 's/DROP SEQUENCE/DROP SEQUENCE IF EXISTS/g' $ofile
    $used -i 's/DROP TABLE/DROP TABLE IF EXISTS/g' $ofile
}

dump_all(){
    tmpfile=pg_latest.sql
    ssh psql.loc "docker exec postgres pg_dump -s -c -U dev -d crud -f /tmp/$tmpfile"
    ssh psql.loc "docker cp postgres:/tmp/$tmpfile /tmp/"
    scp psql.loc:/tmp/$tmpfile $1/
    dump_clear $1/$tmpfile


    cat > $1/pg_latest.go  << EOF
package $2

const PG_LATEST = \`
EOF
    cat $1/$tmpfile | grep -v DROP >> $1/pg_latest.go
    cat >> $1/pg_latest.go  << EOF
\`

EOF

    cat $1/$tmpfile | grep DROP > $1/pg_latest_clear.sql

    cat >> $1/pg_latest.go  << EOF
const PG_DROP = \`
EOF

    cat $1/pg_latest_clear.sql >> $1/pg_latest.go 
    cat >> $1/pg_latest.go  << EOF
\`

EOF

    cat >> $1/pg_latest.go  << EOF
const PG_CLEAR = \`
EOF
    cat $1/$tmpfile | grep 'DROP TABLE' | sed 's/DROP TABLE IF EXISTS/DELETE FROM/' >> $1/pg_latest.go

    cat >> $1/pg_latest.go  << EOF
\`
EOF

}

case $1 in
upgrade)
     dump_upgrade $2 "$3"
;;
clear)
    dump_clear $2
;;
*)
    dump_all ./testsql testsql
;;
esac

