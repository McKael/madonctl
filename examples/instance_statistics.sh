#! /bin/sh
#
# Instance statistics for a given period (default: the last hour)
#
# Usage: $0 [startdate [enddate]]
# (The timestamps must be accepted by date -d)
#
# Examples:
# ./instance_statistics.sh
# ./instance_statistics.sh "30 minutes ago"
# ./instance_statistics.sh "1 hour ago" "30 minutes ago"
# ./instance_statistics.sh "2017-04-30 18:55:00" "2017-04-30 19:05:00"
#
# Mikael

start=${1:-1 hour ago}
end=${2:-now}

TMPL='{{.date | fromunix}}: {{.users}} users, {{.statuses}} statuses{{"\n"}}'
madonctl instance --stats --template "$TMPL" \
    --start "$(date +%s -d "$start")" \
    --end   "$(date +%s -d "$end")"
