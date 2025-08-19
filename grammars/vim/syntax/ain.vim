" ~/.vim/syntax/ain.vim
if exists("b:current_syntax")
  finish
endif

" Headings
syntax match ainHeading /^\s*\[\(config\|host\|query\|headers\|method\|body\|backend\|backendoptions\)\]\s*\ze\(\s*#\|\s*$\)\c/
highlight link ainHeading Keyword

" Escapes
syntax match ainEscape /\\`/
syntax match ainEnvvarEscape /`\${/
syntax match ainEscape /`\$(/
syntax match ainEscape /`#/
highlight link ainEscape Normal
highlight link ainEnvvarEscape Normal

" Envvars: ${VAR}
syntax region ainEnvvar start=+\${+ end=+}+ contains=ainEnvvarEndEscape
syntax match ainEnvvarEndEscape /`}/ contained
highlight link ainEnvvar Identifier
highlight link ainEnvvarEndEscape Identifier

syntax match ainEscapeContained /\\`/ contained
syntax match ainEnvvarEscapeContained /`\${/ contained

" Executables: $(command)
syntax region ainExec start=+\$(+ end=+)+ contains=ainEscapeContained,ainEnvvarEscapeContained,ainEnvvar,ainExecEscape,ainSQ,ainDQ
syntax match ainExecEscape /`)/ contained
highlight link ainExec Type
highlight link ainExecEscape Type
highlight link ainEnvvarEscapeContained Type
highlight link ainEscapeContained Type

" Single-quoted strings inside executables
syntax region ainSQ start=+'+ end=+'+ contains=ainEscapeContained,ainEnvvarEscapeContained,ainEnvvar,ainSQEscape contained
syntax match ainSQEscape /\\'/ contained
highlight link ainSQ String
highlight link ainSQEscape String
highlight link ainEnvvarEscapeContained String
highlight link ainEscapeContained String

" Double-quoted strings inside executables
syntax region ainDQ start=+"+ end=+"+ contains=ainEscapeContained,ainEnvvarEscapeContained,ainEnvvar,ainDQEscape contained
syntax match ainDQEscape /\\"/ contained
highlight link ainDQ String
highlight link ainDQEscape String
highlight link ainEnvvarEscapeContained String
highlight link ainEscapeContained String

" Comments # comment
syntax match ainComment /#.*/
highlight link ainComment Comment

let b:current_syntax = "ain"
