" ~/.vim/syntax/ain.vim
if exists("b:current_syntax")
  finish
endif

" Headings
syntax match ainHeading /^\s*\[\(config\|host\|query\|headers\|method\|body\|backend\|backendoptions\)\]\s*\ze\(\s*#\|\s*$\)\c/
highlight link ainHeading Keyword

" Escapes
syntax match ainEscape /\\`/
syntax match ainEscape /`\${/
syntax match ainEscape /`\$(/
syntax match ainEscape /`#/
highlight link ainEscape Normal

" Envvars: ${VAR}
syntax region ainEnvvar start=+\${+ end=+}+ contains=ainEnvvarEscape
syntax match ainEnvvarEscape /`}/ contained
highlight link ainEnvvar Keyword
highlight link ainEnvvarEscape Keyword

" Executables: $(command)
syntax region ainExec start=+\$(+ end=+)+ contains=ainExecEscape,ainSQ,ainDQ
syntax match ainExecEscape /`)/ contained
highlight link ainExec Identifier
highlight link ainExecEscape Identifier

" Single-quoted strings inside executables
syntax region ainSQ start=+'+ end=+'+ contains=ainSQEscape contained
syntax match ainSQEscape /\\'/ contained
highlight link ainSQ String
highlight link ainSQEscape String

" Double-quoted strings inside executables
syntax region ainDQ start=+"+ end=+"+ contains=ainDQEscape contained
syntax match ainDQEscape /\\"/ contained
highlight link ainDQ String
highlight link ainDQEscape String

" Comments # comment
syntax match ainComment /#.*/
highlight link ainComment Comment

let b:current_syntax = "ain"
