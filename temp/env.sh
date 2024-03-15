#!/bin/bash
export COMMAPI_HTTPD_HARDCODE_TOKEN=token-value
export COMMAPI_INTRASERVICE_DSN="server=sql-03.watcom.local;user id=commonapi;password=commonapi;port=1433;database=Intraservice;"
export COMMAPI_INTRASERVICE_URL="https://helpdesk.watcom.ru"
export COMMAPI_INTRASERVICE_USER=ias@watcom.local
export COMMAPI_INTRASERVICE_PASS=sfg\$%3@45%
export COMMAPI_CMINFO_DSN="sqlserver://commonapi:commonapi@sql2-caravan.watcom.local:1433?database=CM_Info&connection_timeout=0&encrypt=disable"
export COMMAPI_REF_DSN="postgres://evolution:evolution@dworker-01.watcom.local:5432/evolution?sslmode=disable&pool_max_conns=2"