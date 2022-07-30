#!/bin/bash

dump_all(){
    sqlfile=$1/sqlite_latest.sql
    sqlite3 crud.sqlite .schema | grep -v sqlite_sequence > $sqlfile

    cat > $1/sqlite_latest.go  << EOF
package $2

const SQLITE_LATEST = \`
EOF
cat $sqlfile | grep -v DROP >> $1/sqlite_latest.go
cat >> $1/sqlite_latest.go  << EOF
\`

EOF


    drop_sql=`cat $sqlfile | grep TABLE | awk '{print "DROP TABLE IF EXISTS "$6";"}' `
    cat >> $1/sqlite_latest.go  << EOF
const SQLITE_DROP = \`
$drop_sql
\`

EOF

    delete_sql=`cat $sqlfile | grep TABLE | awk '{print "DELETE FROM "$6";"}' `
    cat >> $1/sqlite_latest.go  << EOF
const SQLITE_CLEAR = \`
$delete_sql
\`
EOF

}

dump_all ./testsql testsql
