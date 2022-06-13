#!/bin/bash

dump_clear(){
    ofile=$1
    used=sed
    if [ `uname` == "Darwin" ];then
        used=gsed
    fi
    $used -i 's/public\.//g' $ofile
    $used -i 's/ Owner: dev//g' $ofile
    $used -i 's/ALTER TABLE .* OWNER TO .*;//g' $ofile
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

tmpfile=crud.sql
ssh loc "docker exec postgres pg_dump -s -c -U dev -d crud -f /tmp/$tmpfile"
ssh loc "docker cp postgres:/tmp/$tmpfile /tmp/"
scp loc:/tmp/$tmpfile ./
dump_clear $tmpfile

cat > db_test.go  << EOF
package gen

var LATEST = \`
EOF
cat $tmpfile | grep -v DROP >> db_test.go
cat >> db_test.go  << EOF
\`

EOF

cat $tmpfile | grep DROP > clear.sql

cat >> db_test.go  << EOF
var DROP = \`
EOF

cat clear.sql >> db_test.go 
cat >> db_test.go  << EOF
\`

EOF

cat >> db_test.go  << EOF
const CLEAR = \`
EOF
cat $tmpfile | grep 'DROP TABLE' | sed 's/DROP TABLE IF EXISTS/DELETE FROM/' >> db_test.go

cat >> db_test.go  << EOF
\`
EOF