#! /bin/zsh
#
# Instance statistics for a given period (default: 4 last weeks)
#
# Usage: $0 [[--server INSTANCE] Number_of_weeks]
#
# Mikael

if [[ $1 == "--server" || $1 == "-i" ]]; then
    opt=("--server" "$2")
    shift 2
fi

w=${1:-4}

TMPL='({{(.date | fromunix).Format "2006-01-02"}}) {{.instance_name}}: {{printf "%.0f users, %.0f statuses\n" .users .statuses}}'

typeset -i wa="$w"
while (( wa >= 0 )); do
    when="$wa weeks ago"
    s=$(date +%s -d "$when")
    stats="$(madonctl instance ${opt[*]} --stats --template "$TMPL" \
            --start "$(( s-3600 ))" --end   "$s" | tail -1)"
    if [[ -n $stats ]]; then
        echo "$when $stats"
    fi
    (( wa-=1 ))
done
