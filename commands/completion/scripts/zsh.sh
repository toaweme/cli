#compdef {{.AppName}}
compdef _{{.AppName}} {{.AppName}}

_{{.AppName}}() {
    local requestComp out directive lastParam lastChar
    local -a completions

    words=("${=words[1,CURRENT]}")
    lastParam="${words[-1]}"
    lastChar="${lastParam[-1]}"

    requestComp="${words[1]} __complete ${words[2,-1]}"
    if [[ "${lastChar}" = "" ]]; then
        requestComp="${requestComp} \"\""
    fi

    out=$(eval "${requestComp}" 2>/dev/null)

    local lastLine
    while IFS='\n' read -r line; do
        lastLine="${line}"
    done < <(printf "%s\n" "${out[@]}")

    if [[ "${lastLine[1]}" = : ]]; then
        directive="${lastLine[2,-1]}"
        local suffix
        (( suffix=${#lastLine}+2 ))
        out="${out[1,-$suffix]}"
    else
        directive=0
    fi

    local -a descs
    while IFS='\n' read -r comp; do
        [[ -n "${comp}" ]] || continue
        comp="${comp//:/\\:}"
        comp="${comp//$'\t'/:}"
        descs+=("${comp}")
    done < <(printf "%s\n" "${out[@]}")

    if _describe '' descs; then
        return 0
    fi

    if (( (directive & 4) == 0 )); then
        _arguments '*:filename:_files'
    fi
}

if [[ "${funcstack[1]}" = "_{{.AppName}}" ]]; then
    _{{.AppName}}
fi
