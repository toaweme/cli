_{{.AppName}}_completions() {
    local cur prev words cword
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -n =: || return
    else
        COMPREPLY=()
        _get_comp_words_by_ref -n =: cur prev words cword
    fi

    words=("${words[@]:0:$cword+1}")
    local args=("${words[@]:1}")

    local lastParam="${words[$((${#words[@]}-1))]}"
    local lastChar="${lastParam:$((${#lastParam}-1)):1}"

    local requestComp="${words[0]} __complete ${args[*]}"
    if [[ -z ${cur} && ${lastChar} != = ]]; then
        requestComp="${requestComp} ''"
    fi

    local out directive
    out=$(eval "${requestComp}" 2>/dev/null)

    directive="${out##*:}"
    out="${out%%:*}"
    if [[ "${directive}" == "${out}" ]]; then
        directive=0
    fi

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4

    if (( (directive & shellCompDirectiveError) != 0 )); then
        return
    fi

    if (( (directive & shellCompDirectiveNoSpace) != 0 )); then
        compopt -o nospace 2>/dev/null
    fi

    if (( (directive & shellCompDirectiveNoFileComp) != 0 )); then
        compopt +o default 2>/dev/null
    fi

    local completions=()
    while IFS='' read -r comp; do
        [[ -z "${comp}" ]] && continue
        completions+=("${comp%%$'\t'*}")
    done <<< "${out}"

    IFS=$'\n' read -ra COMPREPLY -d '' \
        < <(IFS=$'\n'; compgen -W "${completions[*]}" -- "${cur}")
}

complete -o default -F _{{.AppName}}_completions {{.AppName}}
